-- SQL Code Test File - Tests syntax highlighting for SQL language features
-- This file demonstrates various SQL constructs for testing code rendering

-- ============================================================================
-- DDL: Data Definition Language
-- ============================================================================

-- Create database
CREATE DATABASE IF NOT EXISTS agentviewer_test
    CHARACTER SET utf8mb4
    COLLATE utf8mb4_unicode_ci;

USE agentviewer_test;

-- Create enum type (PostgreSQL)
-- CREATE TYPE status_enum AS ENUM ('pending', 'running', 'complete', 'failed');

-- Create tables with various constraints
CREATE TABLE IF NOT EXISTS users (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    uuid CHAR(36) NOT NULL,
    email VARCHAR(255) NOT NULL,
    username VARCHAR(50) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    status ENUM('active', 'inactive', 'suspended') NOT NULL DEFAULT 'active',
    role ENUM('admin', 'moderator', 'user') NOT NULL DEFAULT 'user',
    email_verified_at TIMESTAMP NULL,
    last_login_at TIMESTAMP NULL,
    login_count INT UNSIGNED NOT NULL DEFAULT 0,
    preferences JSON,
    metadata JSON,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,

    PRIMARY KEY (id),
    UNIQUE KEY uk_users_uuid (uuid),
    UNIQUE KEY uk_users_email (email),
    UNIQUE KEY uk_users_username (username),
    INDEX idx_users_status (status),
    INDEX idx_users_created (created_at),
    INDEX idx_users_deleted (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS organizations (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) NOT NULL,
    description TEXT,
    settings JSON NOT NULL DEFAULT '{}',
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    PRIMARY KEY (id),
    UNIQUE KEY uk_orgs_slug (slug)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS user_organizations (
    user_id BIGINT UNSIGNED NOT NULL,
    organization_id BIGINT UNSIGNED NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'member',
    joined_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    PRIMARY KEY (user_id, organization_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS posts (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    author_id BIGINT UNSIGNED NOT NULL,
    organization_id BIGINT UNSIGNED,
    title VARCHAR(500) NOT NULL,
    slug VARCHAR(255) NOT NULL,
    content MEDIUMTEXT NOT NULL,
    excerpt TEXT,
    status ENUM('draft', 'published', 'archived') NOT NULL DEFAULT 'draft',
    view_count INT UNSIGNED NOT NULL DEFAULT 0,
    published_at TIMESTAMP NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    PRIMARY KEY (id),
    UNIQUE KEY uk_posts_slug (slug),
    INDEX idx_posts_author (author_id),
    INDEX idx_posts_status (status),
    INDEX idx_posts_published (published_at),
    FULLTEXT INDEX ft_posts_content (title, content),

    FOREIGN KEY (author_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE SET NULL
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS tags (
    id INT UNSIGNED NOT NULL AUTO_INCREMENT,
    name VARCHAR(100) NOT NULL,
    slug VARCHAR(100) NOT NULL,
    description VARCHAR(255),
    color CHAR(7),

    PRIMARY KEY (id),
    UNIQUE KEY uk_tags_slug (slug)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS post_tags (
    post_id BIGINT UNSIGNED NOT NULL,
    tag_id INT UNSIGNED NOT NULL,

    PRIMARY KEY (post_id, tag_id),
    FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
    FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
) ENGINE=InnoDB;

-- ============================================================================
-- Views
-- ============================================================================

CREATE OR REPLACE VIEW v_published_posts AS
SELECT
    p.id,
    p.title,
    p.slug,
    p.excerpt,
    p.view_count,
    p.published_at,
    u.username AS author_username,
    u.first_name AS author_first_name,
    u.last_name AS author_last_name,
    o.name AS organization_name,
    GROUP_CONCAT(t.name ORDER BY t.name SEPARATOR ', ') AS tags
FROM posts p
INNER JOIN users u ON p.author_id = u.id
LEFT JOIN organizations o ON p.organization_id = o.id
LEFT JOIN post_tags pt ON p.id = pt.post_id
LEFT JOIN tags t ON pt.tag_id = t.id
WHERE p.status = 'published'
    AND p.published_at <= NOW()
    AND u.status = 'active'
    AND u.deleted_at IS NULL
GROUP BY p.id, u.username, u.first_name, u.last_name, o.name;

-- ============================================================================
-- DML: Data Manipulation Language
-- ============================================================================

-- Insert with values
INSERT INTO tags (name, slug, description, color) VALUES
    ('Technology', 'technology', 'Tech-related posts', '#3498db'),
    ('Science', 'science', 'Scientific discoveries', '#2ecc71'),
    ('Programming', 'programming', 'Coding tutorials', '#9b59b6'),
    ('Database', 'database', 'Database topics', '#e74c3c'),
    ('DevOps', 'devops', 'DevOps practices', '#f1c40f');

-- Insert with subquery
INSERT INTO post_tags (post_id, tag_id)
SELECT p.id, t.id
FROM posts p
CROSS JOIN tags t
WHERE p.title LIKE '%SQL%'
    AND t.slug = 'database'
    AND NOT EXISTS (
        SELECT 1 FROM post_tags pt
        WHERE pt.post_id = p.id AND pt.tag_id = t.id
    );

-- Update with join
UPDATE users u
INNER JOIN user_organizations uo ON u.id = uo.user_id
INNER JOIN organizations o ON uo.organization_id = o.id
SET u.preferences = JSON_SET(COALESCE(u.preferences, '{}'), '$.default_org', o.slug)
WHERE o.is_active = TRUE
    AND u.preferences IS NULL OR JSON_EXTRACT(u.preferences, '$.default_org') IS NULL;

-- Update with case expression
UPDATE posts
SET status = CASE
    WHEN published_at IS NOT NULL AND published_at <= NOW() THEN 'published'
    WHEN updated_at < DATE_SUB(NOW(), INTERVAL 30 DAY) THEN 'archived'
    ELSE status
END
WHERE status = 'draft';

-- Delete with subquery
DELETE FROM post_tags
WHERE post_id IN (
    SELECT id FROM posts WHERE status = 'archived'
);

-- ============================================================================
-- Complex Queries
-- ============================================================================

-- CTE (Common Table Expression)
WITH RECURSIVE org_hierarchy AS (
    -- Base case
    SELECT id, name, NULL AS parent_id, 0 AS level, name AS path
    FROM organizations
    WHERE id = 1

    UNION ALL

    -- Recursive case (conceptual - organizations don't have parent_id in our schema)
    SELECT o.id, o.name, oh.id, oh.level + 1, CONCAT(oh.path, ' > ', o.name)
    FROM organizations o
    INNER JOIN org_hierarchy oh ON o.id = oh.id + 1
    WHERE oh.level < 5
),
user_stats AS (
    SELECT
        u.id,
        u.username,
        COUNT(DISTINCT p.id) AS post_count,
        COALESCE(SUM(p.view_count), 0) AS total_views,
        MAX(p.published_at) AS last_post_at
    FROM users u
    LEFT JOIN posts p ON u.id = p.author_id AND p.status = 'published'
    GROUP BY u.id, u.username
)
SELECT
    us.username,
    us.post_count,
    us.total_views,
    us.last_post_at,
    CASE
        WHEN us.post_count >= 100 THEN 'Prolific Writer'
        WHEN us.post_count >= 50 THEN 'Regular Contributor'
        WHEN us.post_count >= 10 THEN 'Active Member'
        WHEN us.post_count >= 1 THEN 'Beginner'
        ELSE 'Lurker'
    END AS user_tier
FROM user_stats us
WHERE us.post_count > 0
ORDER BY us.total_views DESC, us.post_count DESC;

-- Window functions
SELECT
    p.id,
    p.title,
    p.published_at,
    p.view_count,
    u.username AS author,
    ROW_NUMBER() OVER (PARTITION BY p.author_id ORDER BY p.published_at DESC) AS author_post_rank,
    RANK() OVER (ORDER BY p.view_count DESC) AS view_rank,
    DENSE_RANK() OVER (ORDER BY DATE(p.published_at) DESC) AS date_rank,
    SUM(p.view_count) OVER (PARTITION BY p.author_id) AS author_total_views,
    AVG(p.view_count) OVER (PARTITION BY p.author_id) AS author_avg_views,
    SUM(p.view_count) OVER (ORDER BY p.published_at ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW) AS cumulative_views,
    LAG(p.view_count, 1, 0) OVER (PARTITION BY p.author_id ORDER BY p.published_at) AS prev_post_views,
    LEAD(p.view_count, 1, 0) OVER (PARTITION BY p.author_id ORDER BY p.published_at) AS next_post_views,
    FIRST_VALUE(p.title) OVER (PARTITION BY p.author_id ORDER BY p.published_at) AS first_post_title,
    NTILE(4) OVER (ORDER BY p.view_count DESC) AS quartile
FROM posts p
INNER JOIN users u ON p.author_id = u.id
WHERE p.status = 'published'
ORDER BY p.published_at DESC;

-- JSON operations
SELECT
    u.id,
    u.username,
    JSON_EXTRACT(u.preferences, '$.theme') AS theme,
    JSON_EXTRACT(u.preferences, '$.notifications.email') AS email_notifications,
    JSON_KEYS(COALESCE(u.preferences, '{}')) AS preference_keys,
    JSON_LENGTH(COALESCE(u.metadata, '{}')) AS metadata_count,
    JSON_CONTAINS(u.preferences, '"dark"', '$.theme') AS has_dark_theme
FROM users u
WHERE u.preferences IS NOT NULL
    AND JSON_VALID(u.preferences)
    AND JSON_EXTRACT(u.preferences, '$.notifications') IS NOT NULL;

-- Full-text search
SELECT
    p.id,
    p.title,
    p.excerpt,
    MATCH(p.title, p.content) AGAINST('database optimization' IN NATURAL LANGUAGE MODE) AS relevance
FROM posts p
WHERE MATCH(p.title, p.content) AGAINST('database optimization' IN NATURAL LANGUAGE MODE)
ORDER BY relevance DESC
LIMIT 10;

-- ============================================================================
-- Stored Procedures and Functions
-- ============================================================================

DELIMITER //

CREATE FUNCTION get_user_display_name(user_id BIGINT)
RETURNS VARCHAR(255)
DETERMINISTIC
READS SQL DATA
BEGIN
    DECLARE display_name VARCHAR(255);

    SELECT
        CASE
            WHEN first_name IS NOT NULL AND last_name IS NOT NULL
                THEN CONCAT(first_name, ' ', last_name)
            WHEN first_name IS NOT NULL
                THEN first_name
            ELSE username
        END INTO display_name
    FROM users
    WHERE id = user_id;

    RETURN COALESCE(display_name, 'Unknown User');
END//

CREATE PROCEDURE sp_publish_post(
    IN p_post_id BIGINT,
    IN p_user_id BIGINT,
    OUT p_success BOOLEAN,
    OUT p_message VARCHAR(255)
)
BEGIN
    DECLARE v_author_id BIGINT;
    DECLARE v_status VARCHAR(20);

    DECLARE EXIT HANDLER FOR SQLEXCEPTION
    BEGIN
        ROLLBACK;
        SET p_success = FALSE;
        SET p_message = 'An error occurred while publishing the post';
    END;

    START TRANSACTION;

    -- Validate post exists and user is author
    SELECT author_id, status INTO v_author_id, v_status
    FROM posts
    WHERE id = p_post_id
    FOR UPDATE;

    IF v_author_id IS NULL THEN
        SET p_success = FALSE;
        SET p_message = 'Post not found';
        ROLLBACK;
    ELSEIF v_author_id != p_user_id THEN
        SET p_success = FALSE;
        SET p_message = 'Unauthorized: You are not the author of this post';
        ROLLBACK;
    ELSEIF v_status = 'published' THEN
        SET p_success = FALSE;
        SET p_message = 'Post is already published';
        ROLLBACK;
    ELSE
        UPDATE posts
        SET status = 'published',
            published_at = NOW()
        WHERE id = p_post_id;

        SET p_success = TRUE;
        SET p_message = 'Post published successfully';
        COMMIT;
    END IF;
END//

DELIMITER ;

-- ============================================================================
-- Triggers
-- ============================================================================

DELIMITER //

CREATE TRIGGER tr_users_before_update
BEFORE UPDATE ON users
FOR EACH ROW
BEGIN
    SET NEW.updated_at = CURRENT_TIMESTAMP;

    IF NEW.email != OLD.email THEN
        SET NEW.email_verified_at = NULL;
    END IF;
END//

CREATE TRIGGER tr_users_after_login
AFTER UPDATE ON users
FOR EACH ROW
BEGIN
    IF NEW.last_login_at != OLD.last_login_at OR
       (NEW.last_login_at IS NOT NULL AND OLD.last_login_at IS NULL) THEN
        UPDATE users
        SET login_count = login_count + 1
        WHERE id = NEW.id;
    END IF;
END//

DELIMITER ;

-- ============================================================================
-- Transactions
-- ============================================================================

START TRANSACTION;

-- Create a new user
INSERT INTO users (uuid, email, username, password_hash)
VALUES (UUID(), 'test@example.com', 'testuser', 'hashed_password');

SET @new_user_id = LAST_INSERT_ID();

-- Create their organization
INSERT INTO organizations (name, slug)
VALUES ('Test Organization', 'test-org');

SET @new_org_id = LAST_INSERT_ID();

-- Link user to organization
INSERT INTO user_organizations (user_id, organization_id, role)
VALUES (@new_user_id, @new_org_id, 'owner');

COMMIT;

-- ============================================================================
-- Grant permissions
-- ============================================================================

GRANT SELECT, INSERT, UPDATE ON agentviewer_test.users TO 'app_user'@'localhost';
GRANT SELECT ON agentviewer_test.v_published_posts TO 'readonly_user'@'%';
GRANT EXECUTE ON PROCEDURE agentviewer_test.sp_publish_post TO 'app_user'@'localhost';

-- End of SQL test file
