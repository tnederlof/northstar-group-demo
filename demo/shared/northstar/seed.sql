-- =============================================================================
-- Northstar Group Demo Seed Data
-- =============================================================================
-- This seed file populates the Fider database with realistic enterprise data
-- for the "Northstar ToolsHub" internal platform feedback portal.
--
-- Run with: psql -U fider -d fider < seed.sql

BEGIN;

-- -----------------------------------------------------------------------------
-- Clear existing data (for idempotent seeding)
-- -----------------------------------------------------------------------------
TRUNCATE TABLE comments, post_votes, post_tags, posts, tags, users, tenants CASCADE;

-- Reset sequences
ALTER SEQUENCE tenants_id_seq RESTART WITH 1;
ALTER SEQUENCE users_id_seq RESTART WITH 1;
ALTER SEQUENCE tags_id_seq RESTART WITH 1;
ALTER SEQUENCE posts_id_seq RESTART WITH 1;
ALTER SEQUENCE comments_id_seq RESTART WITH 1;

-- -----------------------------------------------------------------------------
-- Tenant: Northstar ToolsHub
-- -----------------------------------------------------------------------------
INSERT INTO tenants (
    id, name, subdomain, status, is_private, custom_css, welcome_message, 
    cname, logo_bkey, invitation, created_at, locale, is_email_auth_allowed,
    is_feed_enabled, prevent_indexing, allowed_schemes, is_moderation_enabled, is_pro
)
VALUES (
    1,
    'Northstar ToolsHub',
    'toolshub',
    1, -- Active
    false,
    '',
    'Welcome to Northstar ToolsHub! This is our internal platform for collecting feedback and feature requests from teams across Northstar Group. Share your ideas to help us improve our internal tools.',
    '',
    '',
    'We review submissions weekly in Platform Engineering stand-up.',
    NOW(),
    'en',           -- locale
    true,           -- is_email_auth_allowed
    true,           -- is_feed_enabled
    false,          -- prevent_indexing
    '',             -- allowed_schemes
    false,          -- is_moderation_enabled
    false           -- is_pro
);

-- -----------------------------------------------------------------------------
-- Users
-- -----------------------------------------------------------------------------
-- Roles: 1=Visitor, 2=Collaborator, 3=Administrator

INSERT INTO users (id, tenant_id, name, email, role, status, avatar_type, avatar_bkey, created_at) VALUES
(1, 1, 'Alex Rivera', 'alex.rivera@northstar.io', 3, 1, 0, '', NOW()),      -- Administrator, Platform Eng Lead
(2, 1, 'Sarah Chen', 'sarah.chen@northstar.io', 2, 1, 0, '', NOW()),        -- Collaborator, PM Digital
(3, 1, 'Marcus Wright', 'marcus.wright@northstar.io', 1, 1, 0, '', NOW()),  -- Visitor, Ops Analyst Logistics
(4, 1, 'Jennifer Patel', 'jennifer.patel@northstar.io', 1, 1, 0, '', NOW()); -- Visitor, Finance Systems Lead

-- -----------------------------------------------------------------------------
-- Tags
-- -----------------------------------------------------------------------------
INSERT INTO tags (id, tenant_id, name, slug, color, is_public, created_at) VALUES
(1, 1, 'Logistics Ops', 'logistics-ops', '2196F3', true, NOW()),
(2, 1, 'Digital Platform', 'digital-platform', '9C27B0', true, NOW()),
(3, 1, 'Finance Systems', 'finance-systems', '4CAF50', true, NOW()),
(4, 1, 'Security', 'security', 'F44336', true, NOW()),
(5, 1, 'UX', 'ux', 'FF9800', true, NOW()),
(6, 1, 'Integration', 'integration', '00BCD4', true, NOW());

-- -----------------------------------------------------------------------------
-- Posts
-- -----------------------------------------------------------------------------
-- Status: 0=Open, 1=Planned, 2=Started, 3=Completed, 4=Declined

-- Open (5)
INSERT INTO posts (id, tenant_id, number, title, slug, description, status, user_id, created_at, is_approved) VALUES
(1, 1, 1, 'Okta SSO Session Timeout Too Short', 'okta-sso-session-timeout', 'The 15-minute session timeout for Okta SSO is causing frequent re-authentication during long workflows. This is especially problematic for warehouse floor staff who step away briefly.', 0, 3, NOW() - INTERVAL '14 days', true),
(2, 1, 2, 'PagerDuty Integration for Critical Alerts', 'pagerduty-integration', 'Need ability to route critical platform alerts directly to PagerDuty for on-call rotation. Currently manually monitoring Slack which leads to delayed responses.', 0, 1, NOW() - INTERVAL '12 days', true),
(3, 1, 3, 'Dark Mode Support', 'dark-mode-support', 'Many engineers work late hours and would appreciate a dark mode option to reduce eye strain. Other internal tools already support this.', 0, 2, NOW() - INTERVAL '10 days', true),
(4, 1, 4, 'Mobile App Crashes on Android 14', 'mobile-app-crashes-android', 'The ToolsHub mobile app consistently crashes when viewing attachments on Android 14 devices. Affects our field logistics team significantly.', 0, 3, NOW() - INTERVAL '8 days', true),
(5, 1, 5, 'ServiceNow Ticket Auto-Creation', 'servicenow-auto-creation', 'When a feature request reaches "Planned" status, automatically create a linked ServiceNow ticket for tracking in our ITSM system.', 0, 4, NOW() - INTERVAL '5 days', true);

-- Planned (4)
INSERT INTO posts (id, tenant_id, number, title, slug, description, status, user_id, created_at, is_approved, response, response_user_id, response_date) VALUES
(6, 1, 6, 'SAP S/4HANA Real-Time Sync', 'sap-s4hana-sync', 'Integration with SAP S/4HANA for real-time inventory and order data. Current nightly batch sync causes data staleness issues.', 1, 3, NOW() - INTERVAL '30 days', true, 'Approved for Q2 roadmap. Working with SAP integration team on API access.', 1, NOW() - INTERVAL '5 days'),
(7, 1, 7, 'Bulk CSV Import for Feature Requests', 'bulk-csv-import', 'Allow importing multiple feature requests via CSV upload. Useful during planning sessions when capturing many ideas at once.', 1, 2, NOW() - INTERVAL '25 days', true, 'Scheduled for next sprint. Will support standard CSV format.', 1, NOW() - INTERVAL '3 days'),
(8, 1, 8, 'Splunk Integration for Audit Logs', 'splunk-audit-logs', 'Export all user actions to Splunk for compliance and security monitoring. Required for SOX audit trail.', 1, 4, NOW() - INTERVAL '20 days', true, 'Working with InfoSec on requirements. Target completion: end of quarter.', 1, NOW() - INTERVAL '7 days'),
(9, 1, 9, 'Azure AD Group-Based Access Control', 'azure-ad-group-sync', 'Sync permissions from Azure AD security groups instead of managing access manually. Would reduce admin overhead significantly.', 1, 1, NOW() - INTERVAL '18 days', true, 'In planning phase. Coordinating with Identity team on group mappings.', 1, NOW() - INTERVAL '2 days');

-- Started (1)
INSERT INTO posts (id, tenant_id, number, title, slug, description, status, user_id, created_at, is_approved, response, response_user_id, response_date) VALUES
(10, 1, 10, 'API Rate Limiting Dashboard', 'api-rate-limiting', 'Need visibility into API rate limits and current usage. Several teams have hit limits unexpectedly during peak operations.', 2, 1, NOW() - INTERVAL '45 days', true, 'Development started. Dashboard will show real-time usage and configurable alerts.', 1, NOW() - INTERVAL '10 days');

-- Completed (4)
INSERT INTO posts (id, tenant_id, number, title, slug, description, status, user_id, created_at, is_approved, response, response_user_id, response_date) VALUES
(11, 1, 11, 'Global Search Across All Modules', 'global-search', 'Ability to search across feature requests, comments, and user profiles from a single search bar.', 3, 2, NOW() - INTERVAL '60 days', true, 'Deployed in v2.3.0. Search now covers all content types with filters.', 1, NOW() - INTERVAL '15 days'),
(12, 1, 12, 'SendGrid Email Notifications', 'sendgrid-notifications', 'Replace legacy SMTP with SendGrid for better deliverability and tracking of notification emails.', 3, 1, NOW() - INTERVAL '55 days', true, 'Migration complete. Email delivery rates improved from 94% to 99.2%.', 1, NOW() - INTERVAL '20 days'),
(13, 1, 13, 'Two-Factor Authentication', 'two-factor-auth', 'Add 2FA support for enhanced account security, especially for administrator accounts.', 3, 4, NOW() - INTERVAL '50 days', true, 'Launched! All admin accounts now require 2FA. Optional for other users.', 1, NOW() - INTERVAL '25 days'),
(14, 1, 14, 'Keyboard Navigation Shortcuts', 'keyboard-shortcuts', 'Add keyboard shortcuts for power users: j/k for navigation, ? for help, / for search focus.', 3, 2, NOW() - INTERVAL '40 days', true, 'Shipped in v2.2.0. Press ? anywhere to see available shortcuts.', 1, NOW() - INTERVAL '30 days');

-- Declined (2)
INSERT INTO posts (id, tenant_id, number, title, slug, description, status, user_id, created_at, is_approved, response, response_user_id, response_date) VALUES
(15, 1, 15, 'Blockchain-Based Supply Chain Tracking', 'blockchain-tracking', 'Implement blockchain for immutable supply chain audit trail across all logistics operations.', 4, 3, NOW() - INTERVAL '35 days', true, 'After evaluation, the complexity and cost outweigh benefits for our use case. Our current audit system meets compliance requirements.', 1, NOW() - INTERVAL '28 days'),
(16, 1, 16, 'AI Chatbot for Support', 'ai-chatbot-support', 'Add an AI-powered chatbot to handle common support questions automatically.', 4, 2, NOW() - INTERVAL '22 days', true, 'Declined for now. Our support volume doesn''t justify the investment. Will revisit if ticket volume increases significantly.', 1, NOW() - INTERVAL '18 days');

-- Additional open post
INSERT INTO posts (id, tenant_id, number, title, slug, description, status, user_id, created_at, is_approved) VALUES
(17, 1, 17, 'Custom Dashboard Widgets', 'custom-dashboard-widgets', 'Allow users to customize their dashboard with configurable widgets for metrics relevant to their role.', 0, 4, NOW() - INTERVAL '3 days', true);

-- -----------------------------------------------------------------------------
-- Post Tags
-- -----------------------------------------------------------------------------
-- post_tags requires: post_id, tag_id, created_at, created_by_id, tenant_id
INSERT INTO post_tags (post_id, tag_id, created_at, created_by_id, tenant_id) VALUES
(1, 4, NOW(), 1, 1),   -- Okta SSO -> Security
(2, 6, NOW(), 1, 1),   -- PagerDuty -> Integration
(3, 5, NOW(), 1, 1),   -- Dark Mode -> UX
(4, 5, NOW(), 1, 1),   -- Mobile Crashes -> UX
(4, 1, NOW(), 1, 1),   -- Mobile Crashes -> Logistics Ops
(5, 6, NOW(), 1, 1),   -- ServiceNow -> Integration
(6, 1, NOW(), 1, 1),   -- SAP S/4HANA -> Logistics Ops
(6, 6, NOW(), 1, 1),   -- SAP S/4HANA -> Integration
(7, 5, NOW(), 1, 1),   -- Bulk CSV -> UX
(8, 4, NOW(), 1, 1),   -- Splunk -> Security
(8, 3, NOW(), 1, 1),   -- Splunk -> Finance Systems
(9, 4, NOW(), 1, 1),   -- Azure AD -> Security
(10, 2, NOW(), 1, 1),  -- API Rate Limiting -> Digital Platform
(11, 5, NOW(), 1, 1),  -- Global Search -> UX
(12, 6, NOW(), 1, 1),  -- SendGrid -> Integration
(13, 4, NOW(), 1, 1),  -- 2FA -> Security
(14, 5, NOW(), 1, 1),  -- Keyboard Shortcuts -> UX
(15, 1, NOW(), 1, 1),  -- Blockchain -> Logistics Ops
(16, 2, NOW(), 1, 1),  -- AI Chatbot -> Digital Platform
(17, 5, NOW(), 1, 1);  -- Custom Dashboard -> UX

-- -----------------------------------------------------------------------------
-- Comments
-- -----------------------------------------------------------------------------
-- comments requires: is_approved (NOT NULL)
INSERT INTO comments (id, tenant_id, post_id, content, user_id, created_at, is_approved) VALUES
-- Okta SSO comments
(1, 1, 1, 'This is affecting warehouse efficiency. Staff are re-authenticating 5-6 times per shift.', 3, NOW() - INTERVAL '13 days', true),
(2, 1, 1, 'We can look at extending the session timeout to 2 hours with activity-based refresh. Will need InfoSec approval.', 1, NOW() - INTERVAL '12 days', true),

-- SAP S/4HANA comments
(3, 1, 6, 'The nightly sync causes issues every Monday when weekend orders pile up. Real-time would be huge.', 3, NOW() - INTERVAL '28 days', true),
(4, 1, 6, 'API documentation from SAP team received. Initial POC looks promising.', 1, NOW() - INTERVAL '6 days', true),

-- API Rate Limiting comments
(5, 1, 10, 'We hit the rate limit during Black Friday and had no visibility. This is critical.', 2, NOW() - INTERVAL '40 days', true),
(6, 1, 10, 'Adding per-team quotas as well so you can see your allocation vs usage.', 1, NOW() - INTERVAL '9 days', true),

-- Global Search comments
(7, 1, 11, 'Love this feature! Found an old discussion I was looking for in seconds.', 2, NOW() - INTERVAL '14 days', true),

-- Two-Factor Auth comments
(8, 1, 13, 'Can we get hardware key support (YubiKey) in addition to TOTP?', 4, NOW() - INTERVAL '45 days', true),
(9, 1, 13, 'YubiKey support added in v2.3.1. Thanks for the suggestion!', 1, NOW() - INTERVAL '10 days', true);

-- -----------------------------------------------------------------------------
-- Post Votes (sample votes to match vote counts)
-- -----------------------------------------------------------------------------
-- Note: In a real scenario, we'd have more granular vote tracking.
-- These are simplified to show the system working.

-- post_votes requires: tenant_id (NOT NULL)
INSERT INTO post_votes (post_id, user_id, created_at, tenant_id) VALUES
(1, 2, NOW() - INTERVAL '13 days', 1),
(1, 4, NOW() - INTERVAL '12 days', 1),
(3, 1, NOW() - INTERVAL '9 days', 1),
(3, 3, NOW() - INTERVAL '8 days', 1),
(3, 4, NOW() - INTERVAL '7 days', 1),
(6, 2, NOW() - INTERVAL '25 days', 1),
(6, 4, NOW() - INTERVAL '24 days', 1),
(10, 2, NOW() - INTERVAL '40 days', 1),
(10, 3, NOW() - INTERVAL '38 days', 1),
(10, 4, NOW() - INTERVAL '35 days', 1),
(11, 1, NOW() - INTERVAL '55 days', 1),
(11, 3, NOW() - INTERVAL '50 days', 1),
(11, 4, NOW() - INTERVAL '48 days', 1),
(13, 2, NOW() - INTERVAL '45 days', 1),
(13, 3, NOW() - INTERVAL '42 days', 1);

-- Update sequences to next available ID
SELECT setval('tenants_id_seq', (SELECT MAX(id) FROM tenants));
SELECT setval('users_id_seq', (SELECT MAX(id) FROM users));
SELECT setval('tags_id_seq', (SELECT MAX(id) FROM tags));
SELECT setval('posts_id_seq', (SELECT MAX(id) FROM posts));
SELECT setval('comments_id_seq', (SELECT MAX(id) FROM comments));

COMMIT;

-- Verification queries (optional, comment out for production)
-- SELECT 'Tenants:', COUNT(*) FROM tenants;
-- SELECT 'Users:', COUNT(*) FROM users;
-- SELECT 'Tags:', COUNT(*) FROM tags;
-- SELECT 'Posts:', COUNT(*) FROM posts;
-- SELECT 'Comments:', COUNT(*) FROM comments;
-- SELECT 'Post Tags:', COUNT(*) FROM post_tags;
-- SELECT 'Post Votes:', COUNT(*) FROM post_votes;
