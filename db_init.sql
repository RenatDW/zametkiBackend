CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE notes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title TEXT NOT NULL,
    owner_id UUID REFERENCES users(id),

    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE note_members (
    note_id UUID REFERENCES notes(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,

    role TEXT NOT NULL DEFAULT 'editor',

    PRIMARY KEY (note_id, user_id)
);

CREATE TABLE note_updates (
    id BIGSERIAL PRIMARY KEY,

    note_id UUID REFERENCES notes(id) ON DELETE CASCADE,

    -- CRDT delta (Yjs update)
    update BYTEA NOT NULL,

    -- кто отправил (опционально)
    client_id UUID,

    created_at TIMESTAMP DEFAULT NOW()
);


CREATE TABLE note_snapshots (
    note_id UUID PRIMARY KEY REFERENCES notes(id) ON DELETE CASCADE,

    snapshot BYTEA NOT NULL,

    -- до какого update включительно этот snapshot валиден
    last_update_id BIGINT NOT NULL,

    updated_at TIMESTAMP DEFAULT NOW()
);