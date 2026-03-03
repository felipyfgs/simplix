-- +goose Up
ALTER TABLE inboxes DROP CONSTRAINT IF EXISTS inboxes_channel_type_check;
ALTER TABLE inboxes ADD CONSTRAINT inboxes_channel_type_check
  CHECK (channel_type IN ('manual','phone','email','whatsapp','web','quepasa'));

-- +goose Down
ALTER TABLE inboxes DROP CONSTRAINT IF EXISTS inboxes_channel_type_check;
ALTER TABLE inboxes ADD CONSTRAINT inboxes_channel_type_check
  CHECK (channel_type IN ('manual','phone','email','whatsapp','web'));
