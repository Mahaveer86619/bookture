-- 1. Setup Extensions
CREATE EXTENSION IF NOT EXISTS vector;
CREATE EXTENSION IF NOT EXISTS age;
CREATE EXTENSION IF NOT EXISTS pgcrypto;
LOAD 'age';
ALTER DATABASE bookture SET search_path = "$user", public, ag_catalog;

-- ==========================================
-- USERS & SOCIAL
-- ==========================================

CREATE TABLE users (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username        TEXT NOT NULL UNIQUE,
    email           TEXT NOT NULL UNIQUE,
    password_hash   TEXT NOT NULL,
    display_name    TEXT,
    bio             TEXT,
    avatar_url      TEXT,
    created_at      TIMESTAMPTZ DEFAULT now(),
    updated_at      TIMESTAMPTZ DEFAULT now(),
    deleted_at      TIMESTAMPTZ DEFAULT NULL
);

CREATE TABLE follows (
    follower_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    following_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at      TIMESTAMPTZ DEFAULT now(),
    deleted_at      TIMESTAMPTZ DEFAULT NULL,
    PRIMARY KEY (follower_id, following_id)
);

-- ==========================================
-- HIERARCHY: Book -> Volume -> Chapter -> Scene
-- ==========================================

DROP TABLE IF EXISTS books CASCADE;
CREATE TABLE books (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title           TEXT NOT NULL,
    author          TEXT,
    description     TEXT,
    source_format   TEXT CHECK (source_format IN ('pdf', 'epub', 'txt')),
    status          TEXT DEFAULT 'draft', 
    is_public       BOOLEAN DEFAULT FALSE, -- Visibility toggle
    created_at      TIMESTAMPTZ DEFAULT now(),
    updated_at      TIMESTAMPTZ DEFAULT now(),
    deleted_at      TIMESTAMPTZ DEFAULT NULL
);

CREATE TABLE volumes (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    book_id         UUID NOT NULL REFERENCES books(id) ON DELETE CASCADE,
    volume_number   INT NOT NULL,
    title           TEXT,
    summary         TEXT,
    created_at      TIMESTAMPTZ DEFAULT now(),
    deleted_at      TIMESTAMPTZ DEFAULT NULL,
    UNIQUE (book_id, volume_number)
);

CREATE TABLE chapters (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    volume_id       UUID NOT NULL REFERENCES volumes(id) ON DELETE CASCADE,
    chapter_number  INT NOT NULL,
    title           TEXT,
    raw_text        TEXT, -- Full text blob for backup
    summary         TEXT,
    created_at      TIMESTAMPTZ DEFAULT now(),
    deleted_at      TIMESTAMPTZ DEFAULT NULL,
    UNIQUE (volume_id, chapter_number)
);

-- ==========================================
-- THE CORE UNIT: SCENES (Narrative Moments)
-- ==========================================
-- A scene is a logical grouping of paragraphs that describes one event.
-- This is what gets an Image and Audio.

CREATE TABLE scenes (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chapter_id          UUID NOT NULL REFERENCES chapters(id) ON DELETE CASCADE,
    scene_index         INT NOT NULL, -- Order within chapter
    
    -- Content
    content_text        TEXT NOT NULL, -- The specific text for this scene
    summary_visual      TEXT,          -- Prompt-ready description
    summary_narrative   TEXT,          -- Story-like summary for "Highlights Mode"
    
    -- Intelligence
    importance_score    FLOAT DEFAULT 0.0, -- 0.0 to 1.0 (For filtering Highlights Mode)
    embedding           VECTOR(1536),      -- For semantic search
    
    -- Assets (Multi-modal)
    image_url           TEXT,
    image_prompt        TEXT,
    audio_url           TEXT, -- TTS output file
    
    created_at          TIMESTAMPTZ DEFAULT now(),
    deleted_at      TIMESTAMPTZ DEFAULT NULL,
    UNIQUE (chapter_id, scene_index)
);

-- Note: Paragraphs table is optional now. 
-- If you need strict parsing lineage, keep it. 
-- If 'scenes' contain the text, you might not need it.

-- ==========================================
-- CHARACTER ENGINE (Evolution & Versioning)
-- ==========================================

-- The "Identity" of the character (Global)
CREATE TABLE characters (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    book_id     UUID NOT NULL REFERENCES books(id) ON DELETE CASCADE,
    name        TEXT NOT NULL,
    base_prompt TEXT, -- Core visual traits (e.g. "black hair, scar")
    created_at  TIMESTAMPTZ DEFAULT now(),
    deleted_at      TIMESTAMPTZ DEFAULT NULL
);

-- The "State" of a character at a specific point in time
-- Example: "Arjun (Vol 1)" vs "Arjun (Vol 3 - Blinded)"
CREATE TABLE character_snapshots (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    character_id    UUID NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    
    -- Scope: This snapshot applies to this Volume (or specific Chapter)
    volume_id       UUID REFERENCES volumes(id) ON DELETE CASCADE,
    chapter_id      UUID REFERENCES chapters(id) ON DELETE CASCADE,
    
    visual_changes  TEXT, -- "Wearing armor now", "Has a beard"
    emotional_state TEXT, -- "Grieving", "Vengeful"
    
    current_prompt  TEXT, -- The merged prompt for image gen: base + changes
    
    created_at      TIMESTAMPTZ DEFAULT now(),
    deleted_at      TIMESTAMPTZ DEFAULT NULL
);

-- ==========================================
-- PIPELINE & LOGS
-- ==========================================

CREATE TABLE job_queue (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    entity_id   UUID NOT NULL, -- ID of book, chapter, or scene
    entity_type TEXT NOT NULL, -- 'book', 'scene', 'character'
    task_type   TEXT NOT NULL, -- 'parse', 'summarize', 'generate_image', 'tts'
    status      TEXT DEFAULT 'pending',
    retry_count INT DEFAULT 0,
    error_log   TEXT,
    created_at  TIMESTAMPTZ DEFAULT now(),
    updated_at  TIMESTAMPTZ DEFAULT now(),
    deleted_at      TIMESTAMPTZ DEFAULT NULL
);

-- Indexes for performance
CREATE INDEX idx_scenes_embedding ON scenes USING hnsw (embedding vector_cosine_ops);
CREATE INDEX idx_scenes_chapter ON scenes(chapter_id);
CREATE INDEX idx_snapshots_char ON character_snapshots(character_id);