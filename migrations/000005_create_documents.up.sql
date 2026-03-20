CREATE TABLE documents (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    change_id BIGINT NOT NULL REFERENCES changes(id) ON DELETE CASCADE,
    type TEXT NOT NULL CHECK (type IN ('proposal', 'design', 'tasks')),
    title TEXT,
    content TEXT NOT NULL DEFAULT '',
    version INTEGER NOT NULL DEFAULT 1 CHECK (version > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(change_id, type, title)
);

CREATE INDEX idx_documents_change_id ON documents(change_id);

CREATE TABLE document_versions (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    document_id BIGINT NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    version INTEGER NOT NULL CHECK (version > 0),
    user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_document_versions_document_id ON document_versions(document_id);
CREATE INDEX idx_document_versions_user_id ON document_versions(user_id);
