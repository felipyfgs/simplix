package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/simplix/api/internal/domain"
	"github.com/simplix/api/internal/repository"
	"github.com/simplix/api/internal/service"
)

type WhatsAppWebhookHandler struct {
	inboxes     *repository.InboxRepo
	contacts    *repository.ContactRepo
	convs       *repository.ConversationRepo
	messages    *repository.MessageRepo
	sseBroker   *service.SSEBroker
	whatsappSvc *service.WhatsAppService
}

func NewWhatsAppWebhookHandler(
	inboxes *repository.InboxRepo,
	contacts *repository.ContactRepo,
	convs *repository.ConversationRepo,
	messages *repository.MessageRepo,
	sse *service.SSEBroker,
	wa *service.WhatsAppService,
) *WhatsAppWebhookHandler {
	return &WhatsAppWebhookHandler{
		inboxes:     inboxes,
		contacts:    contacts,
		convs:       convs,
		messages:    messages,
		sseBroker:   sse,
		whatsappSvc: wa,
	}
}

// Verify handles Meta's webhook verification challenge.
func (h *WhatsAppWebhookHandler) Verify(w http.ResponseWriter, r *http.Request) {
	inboxID, err := uuid.Parse(chi.URLParam(r, "inbox_id"))
	if err != nil {
		http.Error(w, "invalid inbox_id", http.StatusBadRequest)
		return
	}

	ix, err := h.inboxes.FindByID(r.Context(), inboxID)
	if err != nil {
		http.Error(w, "inbox not found", http.StatusNotFound)
		return
	}

	settings := whatsAppSettingsFromInbox(ix)

	mode      := r.URL.Query().Get("hub.mode")
	token     := r.URL.Query().Get("hub.verify_token")
	challenge := r.URL.Query().Get("hub.challenge")

	if mode == "subscribe" && h.whatsappSvc.VerifyToken(settings, token) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(challenge))
		return
	}
	http.Error(w, "verification failed", http.StatusForbidden)
}

// Process handles incoming WhatsApp messages from Meta.
func (h *WhatsAppWebhookHandler) Process(w http.ResponseWriter, r *http.Request) {
	// Always respond 200 immediately so Meta doesn't retry
	w.WriteHeader(http.StatusOK)

	inboxID, err := uuid.Parse(chi.URLParam(r, "inbox_id"))
	if err != nil {
		log.Error().Err(err).Str("inbox_id", chi.URLParam(r, "inbox_id")).Msg("invalid inbox_id in webhook")
		return
	}

	var payload metaWebhookPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		log.Error().Err(err).Msg("failed to decode whatsapp payload")
		return
	}

	for _, entry := range payload.Entry {
		for _, change := range entry.Changes {
			v := change.Value
			if len(v.Messages) == 0 {
				continue
			}
			for _, msg := range v.Messages {
				h.processIncomingMessage(r.Context(), inboxID, v, msg)
			}
		}
	}
}

func (h *WhatsAppWebhookHandler) processIncomingMessage(
	ctx context.Context,
	inboxID uuid.UUID,
	value metaChangeValue,
	msg metaMessage,
) {
	// Deduplicate by source_id
	if msg.ID != "" {
		if _, err := h.messages.FindBySourceID(ctx, msg.ID); err == nil {
			return // already processed
		}
	}

	fromPhone := "+" + strings.TrimPrefix(msg.From, "+")

	// Find or create contact
	contact, err := h.findOrCreateContact(ctx, fromPhone, value.Contacts)
	if err != nil {
		log.Error().Err(err).Str("phone", fromPhone).Msg("find/create contact failed")
		return
	}

	// Find open conversation or create new one
	conv, err := h.findOrCreateConversation(ctx, contact.ID, inboxID)
	if err != nil {
		log.Error().Err(err).Str("inbox_id", inboxID.String()).Msg("find/create conversation failed")
		return
	}

	// Extract text content
	content := extractContent(msg)
	if content == "" {
		return
	}

	// Create message
	m, err := h.messages.CreateWithSourceID(
		ctx, conv.ID, domain.SenderContact, &contact.ID,
		content, domain.ContentText, msg.ID,
	)
	if err != nil {
		log.Error().Err(err).Msg("create whatsapp message failed")
		return
	}

	_ = h.convs.TouchActivity(ctx, conv.ID)

	h.sseBroker.BroadcastAll(domain.SSEEvent{
		Type:    "message.created",
		Payload: m,
	})
}

func (h *WhatsAppWebhookHandler) findOrCreateContact(ctx context.Context, phone string, waContacts []metaContact) (*domain.Contact, error) {
	contacts, _, err := h.contacts.List(ctx, repository.ContactFilter{Query: phone, Page: 1, Limit: 1})
	if err == nil && len(contacts) > 0 {
		return &contacts[0], nil
	}

	name := phone
	for _, wc := range waContacts {
		if wc.Profile.Name != "" {
			name = wc.Profile.Name
			break
		}
	}

	return h.contacts.Create(ctx, &name, nil, &phone, nil, nil)
}

func (h *WhatsAppWebhookHandler) findOrCreateConversation(ctx context.Context, contactID, inboxID uuid.UUID) (*domain.Conversation, error) {
	filter := repository.ConversationFilter{
		ContactID: contactID.String(),
		InboxID:   inboxID.String(),
		Status:    "open",
		Page:      1,
		Limit:     1,
	}
	convs, _, err := h.convs.List(ctx, filter)
	if err == nil && len(convs) > 0 {
		return &convs[0], nil
	}

	return h.convs.Create(ctx, contactID, inboxID, nil)
}

func extractContent(msg metaMessage) string {
	switch msg.Type {
	case "text":
		return msg.Text.Body
	case "image", "video", "audio", "document":
		caption := ""
		switch msg.Type {
		case "image":
			caption = msg.Image.Caption
		case "video":
			caption = msg.Video.Caption
		case "document":
			caption = msg.Document.Caption
		}
		if caption != "" {
			return caption
		}
		return "[" + msg.Type + "]"
	case "location":
		return "[localização]"
	case "contacts":
		return "[contato]"
	default:
		return ""
	}
}

// ─── Meta payload types ───────────────────────────────────────────────────────

type metaWebhookPayload struct {
	Entry []metaEntry `json:"entry"`
}

type metaEntry struct {
	Changes []metaChange `json:"changes"`
}

type metaChange struct {
	Value metaChangeValue `json:"value"`
}

type metaChangeValue struct {
	Messages []metaMessage `json:"messages"`
	Contacts []metaContact `json:"contacts"`
}

type metaMessage struct {
	ID       string       `json:"id"`
	From     string       `json:"from"`
	Type     string       `json:"type"`
	Text     metaText     `json:"text"`
	Image    metaMedia    `json:"image"`
	Video    metaMedia    `json:"video"`
	Audio    metaMedia    `json:"audio"`
	Document metaMedia    `json:"document"`
}

type metaText struct {
	Body string `json:"body"`
}

type metaMedia struct {
	Caption  string `json:"caption"`
	Filename string `json:"filename"`
}

type metaContact struct {
	Profile metaContactProfile `json:"profile"`
	WaID    string             `json:"wa_id"`
}

type metaContactProfile struct {
	Name string `json:"name"`
}
