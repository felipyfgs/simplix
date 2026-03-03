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

type QuePasaWebhookHandler struct {
	inboxes   *repository.InboxRepo
	contacts  *repository.ContactRepo
	convs     *repository.ConversationRepo
	messages  *repository.MessageRepo
	sseBroker *service.SSEBroker
}

func NewQuePasaWebhookHandler(
	inboxes *repository.InboxRepo,
	contacts *repository.ContactRepo,
	convs *repository.ConversationRepo,
	messages *repository.MessageRepo,
	sse *service.SSEBroker,
) *QuePasaWebhookHandler {
	return &QuePasaWebhookHandler{
		inboxes:   inboxes,
		contacts:  contacts,
		convs:     convs,
		messages:  messages,
		sseBroker: sse,
	}
}

// Process handles incoming messages from QuePasa.
func (h *QuePasaWebhookHandler) Process(w http.ResponseWriter, r *http.Request) {
	// Always respond 200 immediately so QuePasa doesn't retry
	w.WriteHeader(http.StatusOK)

	inboxID, err := uuid.Parse(chi.URLParam(r, "inbox_id"))
	if err != nil {
		log.Error().Err(err).Str("inbox_id", chi.URLParam(r, "inbox_id")).Msg("invalid inbox_id in quepasa webhook")
		return
	}

	var payload qpWebhookPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		log.Error().Err(err).Msg("failed to decode quepasa payload")
		return
	}

	if payload.Type != "message" || payload.Message.FromMe {
		return
	}

	h.processIncomingMessage(r.Context(), inboxID, payload.Message)
}

func (h *QuePasaWebhookHandler) processIncomingMessage(ctx context.Context, inboxID uuid.UUID, msg qpMessage) {
	// Deduplicate by source_id
	if msg.ID != "" {
		if _, err := h.messages.FindBySourceID(ctx, msg.ID); err == nil {
			return
		}
	}

	phone := extractQuePasaPhone(msg)
	if phone == "" {
		return
	}

	contact, err := h.findOrCreateContact(ctx, phone, msg.Sender.DisplayName)
	if err != nil {
		log.Error().Err(err).Str("phone", phone).Msg("quepasa: find/create contact failed")
		return
	}

	conv, err := h.findOrCreateConversation(ctx, contact.ID, inboxID)
	if err != nil {
		log.Error().Err(err).Str("inbox_id", inboxID.String()).Msg("quepasa: find/create conversation failed")
		return
	}

	content := msg.Body
	if content == "" {
		return
	}

	m, err := h.messages.CreateWithSourceID(
		ctx, conv.ID, domain.SenderContact, &contact.ID,
		content, domain.ContentText, msg.ID,
	)
	if err != nil {
		log.Error().Err(err).Msg("quepasa: create message failed")
		return
	}

	_ = h.convs.TouchActivity(ctx, conv.ID)

	h.sseBroker.BroadcastAll(domain.SSEEvent{
		Type:    "message.created",
		Payload: m,
	})
}

func (h *QuePasaWebhookHandler) findOrCreateContact(ctx context.Context, phone, displayName string) (*domain.Contact, error) {
	contacts, _, err := h.contacts.List(ctx, repository.ContactFilter{Query: phone, Page: 1, Limit: 1})
	if err == nil && len(contacts) > 0 {
		return &contacts[0], nil
	}

	name := phone
	if displayName != "" {
		name = displayName
	}
	return h.contacts.Create(ctx, &name, nil, &phone, nil, nil)
}

func (h *QuePasaWebhookHandler) findOrCreateConversation(ctx context.Context, contactID, inboxID uuid.UUID) (*domain.Conversation, error) {
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

// extractQuePasaPhone extracts a normalized phone number from the QuePasa message.
// chatId format: "5511999999999@s.whatsapp.net" or sender.id: "5511999999999"
func extractQuePasaPhone(msg qpMessage) string {
	if msg.Sender.ID != "" {
		return "+" + strings.TrimPrefix(msg.Sender.ID, "+")
	}
	// fallback: strip @s.whatsapp.net from chatId
	chatID := strings.Split(msg.ChatID, "@")[0]
	if chatID != "" {
		return "+" + strings.TrimPrefix(chatID, "+")
	}
	return ""
}

// ─── QuePasa payload types ────────────────────────────────────────────────────

type qpWebhookPayload struct {
	WorkspaceID string    `json:"workspaceId"`
	Type        string    `json:"type"`
	Message     qpMessage `json:"message"`
}

type qpMessage struct {
	ID      string   `json:"id"`
	ChatID  string   `json:"chatId"`
	Body    string   `json:"body"`
	FromMe  bool     `json:"fromMe"`
	Sender  qpSender `json:"sender"`
}

type qpSender struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
}
