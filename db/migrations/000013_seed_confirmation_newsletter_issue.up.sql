CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

INSERT INTO newsletter_issues(
    newsletter_issue_id,
    title,
    text_content,
    html_content,
    published_at
) VALUES (
    '00000000-0000-0000-0000-000000000000'::uuid,
    'Please confirm your subscription',
    'Welcome to our newsletter! Please confirm your subscription at: {{.link}}',
    '<p>Welcome to our newsletter! Please confirm your subscription at: {{.link}}</p>',
    NOW()
);