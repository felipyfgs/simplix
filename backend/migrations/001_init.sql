-- +goose Up
-- +goose StatementBegin

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Users
CREATE TABLE users (
    id            UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    name          VARCHAR(100) NOT NULL,
    email         VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    role          VARCHAR(20)  NOT NULL DEFAULT 'agent' CHECK (role IN ('admin', 'agent')),
    availability  VARCHAR(20)  NOT NULL DEFAULT 'offline' CHECK (availability IN ('online', 'busy', 'offline')),
    avatar_url    VARCHAR(500),
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Contacts (general)
CREATE TABLE contacts (
    id                UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    name              VARCHAR(255) NOT NULL,
    email             VARCHAR(255),
    phone             VARCHAR(50),
    company           VARCHAR(255),
    avatar_url        VARCHAR(500),
    status            VARCHAR(30)  NOT NULL DEFAULT 'novo'
                      CHECK (status IN ('novo','qualificado','proposta','negociacao','fechado','perdido')),
    score             SMALLINT     NOT NULL DEFAULT 0,
    assigned_to       UUID         REFERENCES users(id) ON DELETE SET NULL,
    custom_attributes JSONB        NOT NULL DEFAULT '{}',
    created_at        TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_contacts_status   ON contacts(status);
CREATE INDEX idx_contacts_email    ON contacts(email);
CREATE INDEX idx_contacts_assigned ON contacts(assigned_to);

-- Inboxes (communication channels)
CREATE TABLE inboxes (
    id           UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    name         VARCHAR(100) NOT NULL,
    channel_type VARCHAR(20)  NOT NULL DEFAULT 'manual' CHECK (channel_type IN ('manual','phone','email','whatsapp','web')),
    settings     JSONB        NOT NULL DEFAULT '{}',
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

INSERT INTO inboxes (id, name, channel_type) VALUES (uuid_generate_v4(), 'Geral', 'manual');

-- Conversations
CREATE TABLE conversations (
    id                UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    contact_id        UUID        NOT NULL REFERENCES contacts(id) ON DELETE CASCADE,
    inbox_id          UUID        NOT NULL REFERENCES inboxes(id),
    assigned_to       UUID        REFERENCES users(id) ON DELETE SET NULL,
    status            VARCHAR(20) NOT NULL DEFAULT 'open' CHECK (status IN ('open','resolved','pending','snoozed')),
    priority          VARCHAR(10) NOT NULL DEFAULT 'medium' CHECK (priority IN ('low','medium','high','urgent')),
    subject           VARCHAR(255),
    meta              JSONB       NOT NULL DEFAULT '{}',
    custom_attributes JSONB       NOT NULL DEFAULT '{}',
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_activity_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_conv_contact      ON conversations(contact_id);
CREATE INDEX idx_conv_status       ON conversations(status);
CREATE INDEX idx_conv_assigned     ON conversations(assigned_to);
CREATE INDEX idx_conv_inbox        ON conversations(inbox_id);
CREATE INDEX idx_conv_last_activity ON conversations(last_activity_at DESC);

-- Messages
CREATE TABLE messages (
    id              UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    conversation_id UUID        NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    sender_type     VARCHAR(10) NOT NULL CHECK (sender_type IN ('agent','contact','system')),
    sender_id       UUID,
    content         TEXT        NOT NULL,
    content_type    VARCHAR(20) NOT NULL DEFAULT 'text' CHECK (content_type IN ('text','note','activity')),
    attachments     JSONB       NOT NULL DEFAULT '[]',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_msg_conversation ON messages(conversation_id);
CREATE INDEX idx_msg_created      ON messages(created_at DESC);

-- Labels
CREATE TABLE labels (
    id          UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    name        VARCHAR(50) NOT NULL UNIQUE,
    color       VARCHAR(7)  NOT NULL DEFAULT '#6B7280',
    description TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Contact labels (many-to-many)
CREATE TABLE contact_labels (
    contact_id UUID NOT NULL REFERENCES contacts(id) ON DELETE CASCADE,
    label_id   UUID NOT NULL REFERENCES labels(id)   ON DELETE CASCADE,
    PRIMARY KEY (contact_id, label_id)
);

-- Conversation labels (many-to-many)
CREATE TABLE conversation_labels (
    conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    label_id        UUID NOT NULL REFERENCES labels(id)        ON DELETE CASCADE,
    PRIMARY KEY (conversation_id, label_id)
);

-- Notes (private contact notes)
CREATE TABLE notes (
    id         UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    contact_id UUID        NOT NULL REFERENCES contacts(id) ON DELETE CASCADE,
    user_id    UUID        REFERENCES users(id) ON DELETE SET NULL,
    content    TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_notes_contact ON notes(contact_id);

-- Custom attribute definitions
CREATE TABLE custom_attribute_definitions (
    id             UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    entity_type    VARCHAR(20) NOT NULL CHECK (entity_type IN ('contact','conversation')),
    attribute_key  VARCHAR(50) NOT NULL,
    display_name   VARCHAR(100) NOT NULL,
    attribute_type VARCHAR(20) NOT NULL DEFAULT 'text' CHECK (attribute_type IN ('text','number','boolean','date','list')),
    options        JSONB       NOT NULL DEFAULT '[]',
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (entity_type, attribute_key)
);

-- Webhooks
CREATE TABLE webhooks (
    id            UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    url           VARCHAR(500) NOT NULL,
    subscriptions TEXT[]       NOT NULL DEFAULT '{}',
    secret        VARCHAR(100),
    enabled       BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Settings
CREATE TABLE settings (
    key        VARCHAR(100) PRIMARY KEY,
    value      TEXT         NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

INSERT INTO settings (key, value) VALUES ('app_name', 'Simplix CRM');

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS settings CASCADE;
DROP TABLE IF EXISTS webhooks CASCADE;
DROP TABLE IF EXISTS custom_attribute_definitions CASCADE;
DROP TABLE IF EXISTS notes CASCADE;
DROP TABLE IF EXISTS conversation_labels CASCADE;
DROP TABLE IF EXISTS contact_labels CASCADE;
DROP TABLE IF EXISTS labels CASCADE;
DROP TABLE IF EXISTS messages CASCADE;
DROP TABLE IF EXISTS conversations CASCADE;
DROP TABLE IF EXISTS inboxes CASCADE;
DROP TABLE IF EXISTS contacts CASCADE;
DROP TABLE IF EXISTS users CASCADE;
-- +goose StatementEnd
