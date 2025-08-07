-- Comprehensive Database Schema for Bookmark Sync Service
-- Migration: 01_init_users.sql
-- Description: Complete database schema with all tables, indexes, and constraints

-- Enable necessary extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Drop existing tables if they exist (for clean migration)
DROP TABLE IF EXISTS follows CASCADE;
DROP TABLE IF EXISTS sync_events CASCADE;
DROP TABLE IF EXISTS comments CASCADE;
DROP TABLE IF EXISTS bookmark_collections CASCADE;
DROP TABLE IF EXISTS collections CASCADE;
DROP TABLE IF EXISTS bookmarks CASCADE;
DROP TABLE IF EXISTS users CASCADE;

-- Users table
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    email VARCHAR(255) UNIQUE NOT NULL,
    username VARCHAR(255) UNIQUE NOT NULL,
    display_name VARCHAR(255),
    avatar TEXT,
    supabase_id VARCHAR(255) UNIQUE NOT NULL,
    preferences JSONB DEFAULT '{}',
    last_active_at TIMESTAMP WITH TIME ZONE
);

-- Bookmarks table
CREATE TABLE bookmarks (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    url TEXT NOT NULL,
    title TEXT NOT NULL,
    description TEXT,
    favicon TEXT,
    screenshot TEXT,
    metadata JSONB DEFAULT '{}',
    tags JSONB DEFAULT '[]',
    save_count INTEGER DEFAULT 0,
    like_count INTEGER DEFAULT 0,
    comment_count INTEGER DEFAULT 0,
    status VARCHAR(50) DEFAULT 'active',
    last_accessed_at TIMESTAMP WITH TIME ZONE,
    last_checked_at TIMESTAMP WITH TIME ZONE
);

-- Collections table
CREATE TABLE collections (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    color VARCHAR(7),
    icon VARCHAR(50),
    parent_id INTEGER REFERENCES collections(id) ON DELETE CASCADE,
    visibility VARCHAR(50) DEFAULT 'private',
    share_link VARCHAR(255) UNIQUE,
    metadata JSONB DEFAULT '{}'
);

-- Comments table
CREATE TABLE comments (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    bookmark_id INTEGER NOT NULL REFERENCES bookmarks(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    parent_id INTEGER REFERENCES comments(id) ON DELETE CASCADE,
    is_moderated BOOLEAN DEFAULT FALSE
);

-- Sync events table
CREATE TABLE sync_events (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    device_id VARCHAR(255) NOT NULL,
    event_type VARCHAR(50) NOT NULL,
    resource_type VARCHAR(50) NOT NULL,
    resource_id INTEGER NOT NULL,
    changes JSONB DEFAULT '{}',
    processed BOOLEAN DEFAULT FALSE,
    conflict_resolved BOOLEAN DEFAULT FALSE
);

-- Follows table
CREATE TABLE follows (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    follower_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    following_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    notification_settings JSONB DEFAULT '{}',
    UNIQUE(follower_id, following_id)
);

-- Many-to-many relationship table for bookmarks and collections
CREATE TABLE bookmark_collections (
    bookmark_id INTEGER NOT NULL REFERENCES bookmarks(id) ON DELETE CASCADE,
    collection_id INTEGER NOT NULL REFERENCES collections(id) ON DELETE CASCADE,
    PRIMARY KEY (bookmark_id, collection_id)
);

-- Create indexes for performance
-- User indexes
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_supabase_id ON users(supabase_id);
CREATE INDEX idx_users_last_active ON users(last_active_at);
CREATE INDEX idx_users_deleted_at ON users(deleted_at);

-- Bookmark indexes
CREATE INDEX idx_bookmarks_user_id ON bookmarks(user_id);
CREATE INDEX idx_bookmarks_url ON bookmarks(url);
CREATE INDEX idx_bookmarks_status ON bookmarks(status);
CREATE INDEX idx_bookmarks_created_at ON bookmarks(created_at);
CREATE INDEX idx_bookmarks_user_created ON bookmarks(user_id, created_at);
CREATE INDEX idx_bookmarks_deleted_at ON bookmarks(deleted_at);

-- Full-text search indexes for bookmarks
CREATE INDEX idx_bookmarks_title_gin ON bookmarks USING gin(to_tsvector('english', title));
CREATE INDEX idx_bookmarks_description_gin ON bookmarks USING gin(to_tsvector('english', description));

-- Collection indexes
CREATE INDEX idx_collections_user_id ON collections(user_id);
CREATE INDEX idx_collections_parent_id ON collections(parent_id);
CREATE INDEX idx_collections_visibility ON collections(visibility);
CREATE INDEX idx_collections_share_link ON collections(share_link);
CREATE INDEX idx_collections_deleted_at ON collections(deleted_at);

-- Comment indexes
CREATE INDEX idx_comments_bookmark_id ON comments(bookmark_id);
CREATE INDEX idx_comments_user_id ON comments(user_id);
CREATE INDEX idx_comments_parent_id ON comments(parent_id);
CREATE INDEX idx_comments_deleted_at ON comments(deleted_at);

-- Sync event indexes
CREATE INDEX idx_sync_events_user_device ON sync_events(user_id, device_id);
CREATE INDEX idx_sync_events_processed ON sync_events(processed);
CREATE INDEX idx_sync_events_created_at ON sync_events(created_at);
CREATE INDEX idx_sync_events_deleted_at ON sync_events(deleted_at);

-- Follow indexes
CREATE INDEX idx_follows_follower_id ON follows(follower_id);
CREATE INDEX idx_follows_following_id ON follows(following_id);
CREATE INDEX idx_follows_deleted_at ON follows(deleted_at);

-- Bookmark-Collection relationship indexes
CREATE INDEX idx_bookmark_collections_bookmark_id ON bookmark_collections(bookmark_id);
CREATE INDEX idx_bookmark_collections_collection_id ON bookmark_collections(collection_id);

-- Create functions for updated_at triggers
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers for updated_at
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_bookmarks_updated_at BEFORE UPDATE ON bookmarks FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_collections_updated_at BEFORE UPDATE ON collections FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_comments_updated_at BEFORE UPDATE ON comments FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_sync_events_updated_at BEFORE UPDATE ON sync_events FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_follows_updated_at BEFORE UPDATE ON follows FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Create health check function
CREATE OR REPLACE FUNCTION health_check()
RETURNS TABLE(status TEXT, timestamp TIMESTAMP WITH TIME ZONE, tables_count INTEGER) AS $$
DECLARE
    table_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO table_count
    FROM information_schema.tables
    WHERE table_schema = 'public' AND table_type = 'BASE TABLE';

    RETURN QUERY SELECT 'healthy'::TEXT, NOW(), table_count;
END;
$$ LANGUAGE plpgsql;

-- Insert seed data for development
INSERT INTO users (email, username, display_name, supabase_id, preferences) VALUES
('admin@example.com', 'admin', 'Administrator', 'admin-supabase-id', '{"theme": "dark", "gridSize": "medium", "defaultView": "grid"}'),
('user1@example.com', 'user1', 'Test User 1', 'user1-supabase-id', '{"theme": "light", "gridSize": "large", "defaultView": "list"}'),
('user2@example.com', 'user2', 'Test User 2', 'user2-supabase-id', '{"theme": "auto", "gridSize": "small", "defaultView": "grid"}');

INSERT INTO collections (user_id, name, description, color, icon, visibility) VALUES
(1, 'Development Resources', 'Useful development tools and resources', '#3B82F6', 'code', 'private'),
(1, 'Design Inspiration', 'Beautiful designs and UI/UX resources', '#8B5CF6', 'palette', 'public'),
(2, 'Learning Materials', 'Educational content and tutorials', '#10B981', 'book', 'private');

INSERT INTO bookmarks (user_id, url, title, description, tags, status) VALUES
(1, 'https://github.com', 'GitHub', 'The world''s leading software development platform', '["development", "git", "collaboration"]', 'active'),
(1, 'https://stackoverflow.com', 'Stack Overflow', 'The largest online community for developers', '["development", "programming", "help"]', 'active'),
(2, 'https://developer.mozilla.org', 'MDN Web Docs', 'Resources for developers, by developers', '["documentation", "web", "javascript"]', 'active'),
(2, 'https://go.dev', 'The Go Programming Language', 'Build fast, reliable, and efficient software at scale', '["golang", "programming", "backend"]', 'active');

-- Associate bookmarks with collections
INSERT INTO bookmark_collections (bookmark_id, collection_id) VALUES
(1, 1), -- GitHub -> Development Resources
(2, 1), -- Stack Overflow -> Development Resources
(3, 3), -- MDN -> Learning Materials
(4, 3); -- Go -> Learning Materials

-- Add comments to tables for documentation
COMMENT ON TABLE users IS 'User accounts with Supabase integration';
COMMENT ON TABLE bookmarks IS 'User bookmarks with metadata and social features';
COMMENT ON TABLE collections IS 'Bookmark collections with hierarchy support';
COMMENT ON TABLE comments IS 'Comments on bookmarks with threading support';
COMMENT ON TABLE sync_events IS 'Synchronization events for cross-device sync';
COMMENT ON TABLE follows IS 'User following relationships for social features';
COMMENT ON TABLE bookmark_collections IS 'Many-to-many relationship between bookmarks and collections';

-- Migration completed
SELECT 'Database schema migration completed successfully!' as status;