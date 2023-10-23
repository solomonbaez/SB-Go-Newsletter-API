CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

DELETE FROM newsletter_issues
WHERE newsletter_issue_id = '00000000-0000-0000-0000-000000000000'::uuid;