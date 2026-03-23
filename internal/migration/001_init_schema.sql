-- RSS Reader Database Initialization Script
-- Generated from GORM models
-- Database: PostgreSQL

-- Enable UUID extension (optional, for future use)
-- CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ============================================
-- Users Table
-- ============================================
CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    email VARCHAR(255) NOT NULL,
    password VARCHAR(255) NOT NULL,
    CONSTRAINT uk_users_email UNIQUE (email)
);

-- Index for soft delete queries
CREATE INDEX idx_users_deleted_at ON users(deleted_at);

-- ============================================
-- Feeds Table
-- ============================================
CREATE TABLE IF NOT EXISTS feeds (
    id BIGSERIAL PRIMARY KEY,
    url VARCHAR(2048) NOT NULL,
    title VARCHAR(500),
    category VARCHAR(255),
    icon_url VARCHAR(2048),
    user_id BIGINT NOT NULL,
    last_fetch TIMESTAMP WITH TIME ZONE,
    CONSTRAINT fk_feeds_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT uk_user_feed UNIQUE (user_id, url)
);

-- Indexes for common queries
CREATE INDEX idx_feeds_user_id ON feeds(user_id);
CREATE INDEX idx_feeds_category ON feeds(category);

-- ============================================
-- Articles Table
-- ============================================
CREATE TABLE IF NOT EXISTS articles (
    id BIGSERIAL PRIMARY KEY,
    feed_id BIGINT NOT NULL,
    title VARCHAR(1000) NOT NULL,
    link VARCHAR(2048),
    description TEXT,
    content TEXT,
    pub_date TIMESTAMP WITH TIME ZONE,
    is_read BOOLEAN DEFAULT FALSE,
    user_id BIGINT NOT NULL,
    summary TEXT,
    key_points TEXT,
    CONSTRAINT fk_articles_feed FOREIGN KEY (feed_id) REFERENCES feeds(id) ON DELETE CASCADE,
    CONSTRAINT fk_articles_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Indexes for common queries
CREATE INDEX idx_articles_feed_id ON articles(feed_id);
CREATE INDEX idx_articles_user_id ON articles(user_id);
CREATE INDEX idx_articles_title ON articles(title);
CREATE INDEX idx_articles_pub_date ON articles(pub_date DESC);
CREATE INDEX idx_articles_is_read ON articles(is_read);

-- ============================================
-- Tags Table
-- ============================================
CREATE TABLE IF NOT EXISTS tags (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255),
    user_id BIGINT NOT NULL,
    CONSTRAINT fk_tags_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Indexes for common queries
CREATE INDEX idx_tags_user_id ON tags(user_id);
CREATE INDEX idx_tags_name ON tags(name);

-- ============================================
-- Article_Tags Junction Table (Many-to-Many)
-- ============================================
CREATE TABLE IF NOT EXISTS article_tags (
    article_id BIGINT NOT NULL,
    tag_id BIGINT NOT NULL,
    PRIMARY KEY (article_id, tag_id),
    CONSTRAINT fk_article_tags_article FOREIGN KEY (article_id) REFERENCES articles(id) ON DELETE CASCADE,
    CONSTRAINT fk_article_tags_tag FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
);

-- Indexes for junction table
CREATE INDEX idx_article_tags_tag_id ON article_tags(tag_id);

-- ============================================
-- Comments and Notes
-- ============================================
-- 
-- Schema Design Notes:
-- 
-- 1. Users: Uses GORM's soft delete (deleted_at column)
-- 2. Feeds: Composite unique index on (user_id, url) to allow same URL for different users
-- 3. Articles: Contains summary and key_points fields for AI-generated content
-- 4. Tags: Many-to-many relationship with articles through article_tags junction table
-- 5. Foreign Keys: All use CASCADE delete to maintain referential integrity
-- 
-- Performance Considerations:
-- 
-- - Indexes added on foreign keys and frequently queried columns
-- - pub_date index is DESC for efficient chronological ordering
-- - Composite unique constraints prevent duplicate data
-- 
-- Migration Version: 001
-- Created: 2026-03-23
