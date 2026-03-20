-- Project Memory
CREATE TABLE project_memories (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    project_id BIGINT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    content TEXT NOT NULL DEFAULT '',
    updated_by BIGINT REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(project_id)
);

-- Notifications
CREATE TABLE notifications (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type TEXT NOT NULL CHECK (type IN ('review_request', 'comment', 'mention', 'stage_change', 'invite')),
    read BOOLEAN NOT NULL DEFAULT FALSE,
    actor_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    change_id BIGINT REFERENCES changes(id) ON DELETE CASCADE,
    project_id BIGINT REFERENCES projects(id) ON DELETE CASCADE,
    stage TEXT,
    comment_preview TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_notifications_user_id ON notifications(user_id);
CREATE INDEX idx_notifications_user_read ON notifications(user_id, read);
CREATE INDEX idx_notifications_created_at ON notifications(created_at);

-- Add change_type to changes
ALTER TABLE changes ADD COLUMN change_type TEXT NOT NULL DEFAULT 'feature';
