package domain

import (
	"time"

	"github.com/google/uuid"
)

// ─── Users ───────────────────────────────────────────────────────────────────

type Role string

const (
	RoleAdmin Role = "admin"
	RoleAgent Role = "agent"
)

type Availability string

const (
	AvailabilityOnline  Availability = "online"
	AvailabilityBusy    Availability = "busy"
	AvailabilityOffline Availability = "offline"
)

type User struct {
	ID           uuid.UUID    `json:"id"`
	Name         string       `json:"name"`
	Email        string       `json:"email"`
	PasswordHash string       `json:"-"`
	Role         Role         `json:"role"`
	Availability Availability `json:"availability"`
	AvatarURL    *string      `json:"avatar_url"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
}

// ─── Companies ───────────────────────────────────────────────────────────────

type Company struct {
	ID               uuid.UUID      `json:"id"`
	Name             string         `json:"name"`
	Domain           *string        `json:"domain"`
	Phone            *string        `json:"phone"`
	Website          *string        `json:"website"`
	Industry         *string        `json:"industry"`
	Description      *string        `json:"description"`
	CustomAttributes map[string]any `json:"custom_attributes"`
	ContactsCount    int            `json:"contacts_count"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
}

// ─── Contacts ────────────────────────────────────────────────────────────────

type ContactStatus string

const (
	ContactStatusNovo        ContactStatus = "novo"
	ContactStatusQualificado ContactStatus = "qualificado"
	ContactStatusProposta    ContactStatus = "proposta"
	ContactStatusNegociacao  ContactStatus = "negociacao"
	ContactStatusFechado     ContactStatus = "fechado"
	ContactStatusPerdido     ContactStatus = "perdido"
)

type Contact struct {
	ID               uuid.UUID      `json:"id"`
	Name             string         `json:"name"`
	Email            *string        `json:"email"`
	Phone            *string        `json:"phone"`
	Company          *string        `json:"company"`
	CompanyID        *uuid.UUID     `json:"company_id"`
	AvatarURL        *string        `json:"avatar_url"`
	Status           ContactStatus  `json:"status"`
	Score            int16          `json:"score"`
	AssignedTo       *uuid.UUID     `json:"assigned_to"`
	CustomAttributes map[string]any `json:"custom_attributes"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	Labels           []Label        `json:"labels,omitempty"`
	AssignedUser     *User          `json:"assigned_user,omitempty"`
	CompanyObj       *Company       `json:"company_obj,omitempty"`
}

// ─── Inboxes ─────────────────────────────────────────────────────────────────

type ChannelType string

const (
	ChannelManual   ChannelType = "manual"
	ChannelPhone    ChannelType = "phone"
	ChannelEmail    ChannelType = "email"
	ChannelWhatsApp ChannelType = "whatsapp"
	ChannelWeb      ChannelType = "web"
	ChannelQuePasa  ChannelType = "quepasa"
)

type WhatsAppSettings struct {
	PhoneNumber         string    `json:"phone_number"`
	PhoneNumberID       string    `json:"phone_number_id"`
	BusinessAccountID   string    `json:"business_account_id"`
	APIKey              string    `json:"api_key"`
	WebhookVerifyToken  string    `json:"webhook_verify_token"`
	MessageTemplates    []any     `json:"message_templates,omitempty"`
	TemplatesLastSynced *time.Time `json:"templates_last_synced,omitempty"`
}

type QuePasaSettings struct {
	BotToken    string `json:"bot_token"`
	PhoneNumber string `json:"phone_number"`
	BaseURL     string `json:"base_url"`
}

type Inbox struct {
	ID          uuid.UUID   `json:"id"`
	Name        string      `json:"name"`
	ChannelType ChannelType `json:"channel_type"`
	Settings    interface{} `json:"settings"`
	CreatedAt   time.Time   `json:"created_at"`
}

// ─── Conversations ────────────────────────────────────────────────────────────

type Priority string

const (
	PriorityLow    Priority = "low"
	PriorityMedium Priority = "medium"
	PriorityHigh   Priority = "high"
	PriorityUrgent Priority = "urgent"
)

type ConversationStatus string

const (
	ConvStatusOpen     ConversationStatus = "open"
	ConvStatusResolved ConversationStatus = "resolved"
	ConvStatusPending  ConversationStatus = "pending"
	ConvStatusSnoozed  ConversationStatus = "snoozed"
)

type Conversation struct {
	ID               uuid.UUID          `json:"id"`
	ContactID        uuid.UUID          `json:"contact_id"`
	InboxID          uuid.UUID          `json:"inbox_id"`
	AssignedTo       *uuid.UUID         `json:"assigned_to"`
	Status           ConversationStatus `json:"status"`
	Priority         Priority           `json:"priority"`
	Subject          *string            `json:"subject"`
	Meta             interface{}        `json:"meta"`
	CustomAttributes map[string]any     `json:"custom_attributes"`
	CreatedAt        time.Time          `json:"created_at"`
	UpdatedAt        time.Time          `json:"updated_at"`
	LastActivityAt   time.Time          `json:"last_activity_at"`
	Contact          *Contact           `json:"contact,omitempty"`
	Inbox            *Inbox             `json:"inbox,omitempty"`
	AssignedUser     *User              `json:"assigned_user,omitempty"`
	Labels           []Label            `json:"labels,omitempty"`
	UnreadCount      int                `json:"unread_count,omitempty"`
	LastMessage      *Message           `json:"last_message,omitempty"`
}

// ─── Messages ─────────────────────────────────────────────────────────────────

type SenderType string

const (
	SenderAgent   SenderType = "agent"
	SenderContact SenderType = "contact"
	SenderSystem  SenderType = "system"
)

type ContentType string

const (
	ContentText     ContentType = "text"
	ContentNote     ContentType = "note"
	ContentActivity ContentType = "activity"
)

type Message struct {
	ID             uuid.UUID   `json:"id"`
	ConversationID uuid.UUID   `json:"conversation_id"`
	SenderType     SenderType  `json:"sender_type"`
	SenderID       *uuid.UUID  `json:"sender_id"`
	Content        string      `json:"content"`
	ContentType    ContentType `json:"content_type"`
	Attachments    interface{} `json:"attachments"`
	SourceID       *string     `json:"source_id,omitempty"`
	ExternalStatus string      `json:"external_status,omitempty"`
	CreatedAt      time.Time   `json:"created_at"`
	Sender         *User       `json:"sender,omitempty"`
}

// ─── Labels ───────────────────────────────────────────────────────────────────

type Label struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Color       string    `json:"color"`
	Description *string   `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

// ─── Webhooks ─────────────────────────────────────────────────────────────────

type Webhook struct {
	ID            uuid.UUID `json:"id"`
	URL           string    `json:"url"`
	Subscriptions []string  `json:"subscriptions"`
	Secret        *string   `json:"secret,omitempty"`
	Enabled       bool      `json:"enabled"`
	CreatedAt     time.Time `json:"created_at"`
}

// ─── Notes ────────────────────────────────────────────────────────────────────

type Note struct {
	ID        uuid.UUID  `json:"id"`
	ContactID uuid.UUID  `json:"contact_id"`
	UserID    *uuid.UUID `json:"user_id"`
	Content   string     `json:"content"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	Author    *User      `json:"author,omitempty"`
}

// ─── Custom Attributes ────────────────────────────────────────────────────────

type CustomAttributeDefinition struct {
	ID            uuid.UUID `json:"id"`
	EntityType    string    `json:"entity_type"`
	AttributeKey  string    `json:"attribute_key"`
	DisplayName   string    `json:"display_name"`
	AttributeType string    `json:"attribute_type"`
	Options       []string  `json:"options"`
	CreatedAt     time.Time `json:"created_at"`
}

// ─── SSE Events ───────────────────────────────────────────────────────────────

type SSEEvent struct {
	Type    string `json:"type"`
	Payload any    `json:"payload"`
}

// ─── Pagination ───────────────────────────────────────────────────────────────

type Pagination struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
	Total int `json:"total"`
}

type PagedResult[T any] struct {
	Data       []T        `json:"data"`
	Pagination Pagination `json:"pagination"`
}
