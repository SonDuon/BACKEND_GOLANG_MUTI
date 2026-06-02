# Database Schema

Tài liệu này tóm tắt schema hiện tại (migrations + models).

**Overview**
- DB uses UUID primary keys (extension uuid-ossp).
- Các bảng chính: `users`, `categories`, `movies`, `movie_categories`, `episodes`, `media_sources`, `watch_history`, `favorites`.

**Users**
- Columns: `id (UUID PK)`, `email (varchar, unique, not null)`, `password_hash (varchar, not null)`, `full_name`, `avatar_url`, `role (user|admin)`, `is_verified (bool)`, `created_at`, `updated_at`, `deleted_at`.
- Indexes: `idx_users_email` on `email`.

**Categories**
- Purpose: unified genre/country/year
- Columns: `id (UUID PK)`, `name (varchar)`, `slug (varchar)`, `type (genre|country|year)`, `created_at`, `updated_at`.
- Constraints: `UNIQUE(slug, type)`.
- Indexes: `idx_categories_type` on `type`.

**Movies**
- Columns: `id (UUID PK)`, `title`, `original_title`, `slug (unique)`, `description (text)`, `thumb_url`, `poster_url`, `backdrop_url`,
  `source (varchar, default 'self')`, `external_id`, `type`, `status (completed|ongoing)`, `release_year (int)`, `rating (decimal(3,2))`,
  `vote_count (int)`, `duration (int)`, `total_episodes (int)`, `last_synced_at (timestamptz)`, `created_at`, `updated_at`, `deleted_at`.
- Constraints: `UNIQUE(source, external_id)` ensures one record per source/external id.
- Indexes: title, slug, source, external_id, release_year, rating (desc), type, last_synced_at (from GORM tag).

**Movie_Categories (junction)**
- Columns: `movie_id (UUID FK -> movies.id)`, `category_id (UUID FK -> categories.id)`
- PK: composite `(movie_id, category_id)`
- FKs: ON DELETE CASCADE
- Indexes: on both movie_id and category_id.

**Episodes**
- Columns: `id (UUID PK)`, `movie_id (UUID FK -> movies.id)`, `title`, `slug`, `episode_number (int)`, `season_number (int, default 1)`, `description`, `thumbnail`, `duration (int)`, `air_date (date)`, `created_at`, `updated_at`.
- Constraints: `UNIQUE(movie_id, slug)` and `UNIQUE(movie_id, season_number, episode_number)`.
- Indexes: `idx_episodes_movie` on `movie_id`.

**Media_Sources**
- Columns: `id (UUID PK)`, `episode_id (UUID FK -> episodes.id)`, `server_name`, `source_type (external_api|self_hosted)`, `source_key`, `quality`, `is_default (bool)`, `created_at`, `updated_at`.
- Indexes: `idx_media_sources_episode` on `episode_id`.

**Watch_History**
- Columns: `id (UUID PK)`, `user_id (UUID FK -> users.id)`, `content_id (UUID)`, `content_type (movie|episode)`, `last_position (int)`, `total_duration (int)`, `is_completed (bool)`, `updated_at`.
- Constraints: `UNIQUE(user_id, content_id, content_type)`.
- Indexes: `idx_watch_history_user` on `(user_id, updated_at DESC)`.

**Favorites**
- Columns: `id (UUID PK)`, `user_id (UUID FK -> users.id)`, `movie_id (UUID FK -> movies.id)`, `created_at`.
- Constraints: `UNIQUE(user_id, movie_id)`.
- Indexes: `idx_favorites_user` on `user_id`.

**Notes / Relationships**
- `movies` 1 — * `episodes` (ON DELETE CASCADE).
- `episodes` 1 — * `media_sources` (ON DELETE CASCADE).
- `movies` * — * `categories` via `movie_categories`.
- `users` * — * `favorites` (user favoriting movies).
- `watch_history` stores progress for both movies and episodes using `content_type` + `content_id`.

**Migration changes**
- `002_add_thumb_url_to_movies` adds `thumb_url` to `movies`.
- `003_add_last_synced_at_to_movies` adds `last_synced_at` (timestamptz) to `movies`.
