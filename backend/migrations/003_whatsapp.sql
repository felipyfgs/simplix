-- +goose Up
-- +goose StatementBegin

ALTER TABLE messages ADD COLUMN source_id VARCHAR(255);
ALTER TABLE messages ADD COLUMN external_status VARCHAR(20) NOT NULL DEFAULT 'sent';
CREATE UNIQUE INDEX idx_msg_source_id ON messages(source_id) WHERE source_id IS NOT NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_msg_source_id;
ALTER TABLE messages DROP COLUMN IF EXISTS external_status;
ALTER TABLE messages DROP COLUMN IF EXISTS source_id;

-- +goose StatementEnd
