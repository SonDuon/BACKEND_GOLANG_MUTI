-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ─────────────────────────────────────
-- 1. USERS
-- ─────────────────────────────────────
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(255),
    avatar_url VARCHAR(500),
    role VARCHAR(20) DEFAULT 'user' CHECK (role IN ('user', 'admin')),
    is_verified BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_users_email ON users(email);

-- ─────────────────────────────────────
-- 2. CATEGORIES (Unified: genre/country/year)
-- ─────────────────────────────────────
CREATE TABLE categories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    slug VARCHAR(100) NOT NULL,
    type VARCHAR(50) NOT NULL CHECK (type IN ('genre', 'country', 'year')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(slug, type)
);

CREATE INDEX idx_categories_type ON categories(type);

-- ─────────────────────────────────────
-- 3. MOVIES
-- ─────────────────────────────────────
CREATE TABLE movies (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title VARCHAR(255) NOT NULL,
    original_title VARCHAR(255),
    slug VARCHAR(255) UNIQUE NOT NULL,
    description TEXT,
    thumb_url VARCHAR(500),
    poster_url VARCHAR(500),
    backdrop_url VARCHAR(500),
    
    -- 🔑 Hybrid Source Tracking
    source VARCHAR(50) NOT NULL DEFAULT 'self',
    external_id VARCHAR(100) NOT NULL,
    
    -- 📊 Metadata
    type VARCHAR(50) NOT NULL CHECK (type IN ('movie', 'series')),
    status VARCHAR(50) CHECK (status IN ('completed', 'ongoing')),
    release_year INTEGER,
    rating DECIMAL(3,2) DEFAULT 0.00,
    vote_count INTEGER DEFAULT 0,
    duration INTEGER,
    total_episodes INTEGER,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    UNIQUE(source, external_id)
);

CREATE INDEX idx_movies_title ON movies(title);
CREATE INDEX idx_movies_slug ON movies(slug);
CREATE INDEX idx_movies_source ON movies(source);
CREATE INDEX idx_movies_external_id ON movies(external_id);
CREATE INDEX idx_movies_release_year ON movies(release_year);
CREATE INDEX idx_movies_rating ON movies(rating DESC);
CREATE INDEX idx_movies_type ON movies(type);

-- ─────────────────────────────────────
-- 4. MOVIE_CATEGORIES (Junction Table)
-- ─────────────────────────────────────
CREATE TABLE movie_categories (
    movie_id UUID NOT NULL REFERENCES movies(id) ON DELETE CASCADE,
    category_id UUID NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
    PRIMARY KEY (movie_id, category_id)
);

CREATE INDEX idx_movie_categories_movie ON movie_categories(movie_id);
CREATE INDEX idx_movie_categories_category ON movie_categories(category_id);

-- ─────────────────────────────────────
-- 5. EPISODES
-- ─────────────────────────────────────
CREATE TABLE episodes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    movie_id UUID NOT NULL REFERENCES movies(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL,
    episode_number INTEGER NOT NULL,
    season_number INTEGER DEFAULT 1,
    description TEXT,
    thumbnail VARCHAR(500),
    duration INTEGER,
    air_date DATE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(movie_id, slug),
    UNIQUE(movie_id, season_number, episode_number)
);

CREATE INDEX idx_episodes_movie ON episodes(movie_id);

-- ─────────────────────────────────────
-- 6. MEDIA_SOURCES (Multi-source video links)
-- ─────────────────────────────────────
CREATE TABLE media_sources (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    episode_id UUID NOT NULL REFERENCES episodes(id) ON DELETE CASCADE,
    server_name VARCHAR(100) NOT NULL,
    source_type VARCHAR(50) NOT NULL CHECK (source_type IN ('external_api', 'self_hosted')),
    source_key VARCHAR(1000) NOT NULL,
    quality VARCHAR(20),
    is_default BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_media_sources_episode ON media_sources(episode_id);

-- ─────────────────────────────────────
-- 7. WATCH_HISTORY
-- ─────────────────────────────────────
CREATE TABLE watch_history (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content_id UUID NOT NULL,
    content_type VARCHAR(20) NOT NULL CHECK (content_type IN ('movie', 'episode')),
    last_position INTEGER DEFAULT 0,
    total_duration INTEGER,
    is_completed BOOLEAN DEFAULT false,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, content_id, content_type)
);

CREATE INDEX idx_watch_history_user ON watch_history(user_id, updated_at DESC);

-- ─────────────────────────────────────
-- 8. FAVORITES
-- ─────────────────────────────────────
CREATE TABLE favorites (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    movie_id UUID NOT NULL REFERENCES movies(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, movie_id)
);

CREATE INDEX idx_favorites_user ON favorites(user_id);