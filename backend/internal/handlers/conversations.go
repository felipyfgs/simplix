package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/simplix/api/internal/domain"
	"github.com/simplix/api/internal/middleware"
	"github.com/simplix/api/internal/repository"
	"github.com/simplix/api/internal/service"
)

type ConversationHandler struct {
	convs       *repository.ConversationRepo
	messages    *repository.MessageRepo
	inboxes     *repository.InboxRepo
	contacts    *repository.ContactRepo
	whatsappSvc *service.WhatsAppService
	quepasaSvc  *service.QuePasaService
}

func NewConversationHandler(convs *repository.ConversationRepo, msgs *repository.MessageRepo, inboxes *repository.InboxRepo, contacts *repository.ContactRepo, wa *service.WhatsAppService, qp *service.QuePasaService) *ConversationHandler {
	return &ConversationHandler{convs: convs, messages: msgs, inboxes: inboxes, contacts: contacts, whatsappSvc: wa, quepasaSvc: qp}
}

func (h *ConversationHandler) List(w http.ResponseWriter, r *http.Request) {
	page, limit := ParsePage(r)
	q := r.URL.Query()
	filter := repository.ConversationFilter{
		Status:     q.Get("status"),
		InboxID:    q.Get("inbox_id"),
		AssignedTo: q.Get("assigned_to"),
		ContactID:  q.Get("contact_id"),
		Label:      q.Get("label"),
		Page:       page,
		Limit:      limit,
	}
	convs, total, err := h.convs.List(r.Context(), filter)
	if err != nil {
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	if convs == nil {
		convs = []domain.Conversation{}
	}
	JSON(w, http.StatusOK, domain.PagedResult[domain.Conversation]{
		Data:       convs,
		Pagination: domain.Pagination{Page: page, Limit: limit, Total: total},
	})
}

func (h *ConversationHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ContactID string  `json:"contact_id"`
		InboxID   string  `json:"inbox_id"`
		Subject   *string `json:"subject"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "invalid body")
		return
	}
	contactID, err := uuid.Parse(req.ContactID)
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid contact_id")
		return
	}
	var inboxID uuid.UUID
	if req.InboxID != "" {
		inboxID, err = uuid.Parse(req.InboxID)
		if err != nil {
			Error(w, http.StatusBadRequest, "invalid inbox_id")
			return
		}
	} else {
		inboxID, err = h.convs.GetFirstInboxID(r.Context())
		if err != nil {
			Error(w, http.StatusInternalServerError, "no inbox found")
			return
		}
	}
	conv, err := h.convs.Create(r.Context(), contactID, inboxID, req.Subject)
	if err != nil {
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	JSON(w, http.StatusCreated, conv)
}

func (h *ConversationHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid id")
		return
	}
	conv, err := h.convs.FindByID(r.Context(), id)
	if err != nil {
		Error(w, http.StatusNotFound, "conversation not found")
		return
	}
	labels, _ := h.convs.GetLabels(r.Context(), id)
	conv.Labels = labels
	JSON(w, http.StatusOK, conv)
}

func (h *ConversationHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid id")
		return
	}
	var req struct {
		Status     *string `json:"status"`
		AssignedTo *string `json:"assigned_to"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "invalid body")
		return
	}
	if req.Status != nil {
		_ = h.convs.UpdateStatus(r.Context(), id, domain.ConversationStatus(*req.Status))
	}
	if req.AssignedTo != nil {
		aid, err := uuid.Parse(*req.AssignedTo)
		if err == nil {
			_ = h.convs.Assign(r.Context(), id, &aid)
		}
	}
	conv, _ := h.convs.FindByID(r.Context(), id)
	JSON(w, http.StatusOK, conv)
}

func (h *ConversationHandler) ListMessages(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid id")
		return
	}
	msgs, err := h.messages.List(r.Context(), id)
	if err != nil {
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	if msgs == nil {
		msgs = []domain.Message{}
	}
	JSON(w, http.StatusOK, msgs)
}

func (h *ConversationHandler) CreateMessage(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid id")
		return
	}
	userID, _ := middleware.GetUserID(r)
	var req struct {
		Content     string `json:"content"`
		ContentType string `json:"content_type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Content == "" {
		Error(w, http.StatusBadRequest, "content required")
		return
	}
	ct := domain.ContentType(req.ContentType)
	if ct == "" {
		ct = domain.ContentText
	}
	conv, _ := h.convs.FindByID(r.Context(), id)

	msg, err := h.messages.Create(r.Context(), id, domain.SenderAgent, &userID, req.Content, ct)
	if err != nil {
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	_ = h.convs.TouchActivity(r.Context(), id)

	if ct == domain.ContentText && conv != nil && h.inboxes != nil {
		go h.sendWhatsAppIfNeeded(conv, msg)
		go h.sendQuePasaIfNeeded(conv, msg)
	}

	JSON(w, http.StatusCreated, msg)
}

func (h *ConversationHandler) DeleteMessage(w http.ResponseWriter, r *http.Request) {
	convID, _ := uuid.Parse(chi.URLParam(r, "id"))
	msgID, err := uuid.Parse(chi.URLParam(r, "mid"))
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid message id")
		return
	}
	if err := h.messages.Delete(r.Context(), msgID, convID); err != nil {
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *ConversationHandler) Assign(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid id")
		return
	}
	var req struct {
		UserID *string `json:"user_id"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	var uid *uuid.UUID
	if req.UserID != nil {
		aid, err := uuid.Parse(*req.UserID)
		if err == nil {
			uid = &aid
		}
	}
	_ = h.convs.Assign(r.Context(), id, uid)
	conv, _ := h.convs.FindByID(r.Context(), id)
	JSON(w, http.StatusOK, conv)
}

func (h *ConversationHandler) GetLabels(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid id")
		return
	}
	labels, err := h.convs.GetLabels(r.Context(), id)
	if err != nil {
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	if labels == nil {
		labels = []domain.Label{}
	}
	JSON(w, http.StatusOK, labels)
}

func (h *ConversationHandler) SetLabels(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid id")
		return
	}
	var req struct {
		LabelIDs []uuid.UUID `json:"label_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "invalid body")
		return
	}
	if err := h.convs.SetLabels(r.Context(), id, req.LabelIDs); err != nil {
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	labels, _ := h.convs.GetLabels(r.Context(), id)
	if labels == nil {
		labels = []domain.Label{}
	}
	JSON(w, http.StatusOK, labels)
}

func (h *ConversationHandler) sendWhatsAppIfNeeded(conv *domain.Conversation, msg *domain.Message) {
	if h.whatsappSvc == nil {
		return
	}
	ctx := context.Background()

	ix, err := h.inboxes.FindByID(ctx, conv.InboxID)
	if err != nil || ix.ChannelType != domain.ChannelWhatsApp {
		return
	}

	contact, err := h.contacts.FindByID(ctx, conv.ContactID)
	if err != nil || contact.Phone == nil || *contact.Phone == "" {
		return
	}

	settings := whatsAppSettingsFromInbox(ix)
	metaMsgID, err := h.whatsappSvc.SendTextMessage(settings, *contact.Phone, msg.Content)
	if err != nil {
		log.Error().Err(err).Str("conv_id", conv.ID.String()).Msg("whatsapp send failed")
		return
	}
	if metaMsgID != "" {
		_ = h.messages.UpdateSourceID(ctx, msg.ID, metaMsgID)
	}
}

func (h *ConversationHandler) sendQuePasaIfNeeded(conv *domain.Conversation, msg *domain.Message) {
	if h.quepasaSvc == nil {
		return
	}
	ctx := context.Background()

	ix, err := h.inboxes.FindByID(ctx, conv.InboxID)
	if err != nil || ix.ChannelType != domain.ChannelQuePasa {
		return
	}

	contact, err := h.contacts.FindByID(ctx, conv.ContactID)
	if err != nil || contact.Phone == nil || *contact.Phone == "" {
		return
	}

	settings := quePasaSettingsFromInbox(ix)
	qpMsgID, err := h.quepasaSvc.SendTextMessage(settings, *contact.Phone, msg.Content)
	if err != nil {
		log.Error().Err(err).Str("conv_id", conv.ID.String()).Msg("quepasa send failed")
		return
	}
	if qpMsgID != "" {
		_ = h.messages.UpdateSourceID(ctx, msg.ID, qpMsgID)
	}
}
