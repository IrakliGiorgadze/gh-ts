-- Cleanup script for test users
-- This script removes test users created during testing
-- Use with caution - only run in development/test environments

-- Delete users with test email patterns
DELETE FROM users 
WHERE email LIKE '%test%' 
   OR email LIKE '%example.com%'
   OR email LIKE '%@test%.com%'
   OR email LIKE 'user%@test%.com%'
   OR email LIKE 'duptest%@example.com%'
   OR email LIKE 'test%@example.com%'
   OR email LIKE 'validtest%@example.com%'
   OR email LIKE 'newuser%@example.com%'
   OR email LIKE 'casetest%@example.com%';

-- Show remaining count
SELECT COUNT(*) as remaining_test_users 
FROM users 
WHERE email LIKE '%test%' OR email LIKE '%example.com%';

