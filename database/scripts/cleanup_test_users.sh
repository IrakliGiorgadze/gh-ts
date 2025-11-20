#!/bin/bash
# Cleanup script for test users
# This script removes test users from the database
# Use with caution - only run in development/test environments

set -e

echo "=== Test User Cleanup Script ==="
echo ""
echo "This will delete users with test email patterns:"
echo "  - Emails containing 'test'"
echo "  - Emails containing 'example.com'"
echo "  - Other test patterns"
echo ""
read -p "Are you sure you want to continue? (yes/no): " confirm

if [ "$confirm" != "yes" ]; then
    echo "Aborted."
    exit 1
fi

echo ""
echo "Connecting to database..."

# Get database connection details from docker-compose
DB_USER="${POSTGRES_USER:-ticketuser}"
DB_NAME="${POSTGRES_DB:-ticketing_db}"

# Count before
BEFORE=$(docker-compose exec -T db psql -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT COUNT(*) FROM users WHERE email LIKE '%test%' OR email LIKE '%example.com%';" 2>/dev/null | tr -d ' ')

echo "Found $BEFORE test users to delete."

# Run cleanup
docker-compose exec -T db psql -U "$DB_USER" -d "$DB_NAME" -f - <<EOF
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
EOF

# Count after
AFTER=$(docker-compose exec -T db psql -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT COUNT(*) FROM users WHERE email LIKE '%test%' OR email LIKE '%example.com%';" 2>/dev/null | tr -d ' ')

echo ""
echo "Cleanup complete!"
echo "Deleted: $((BEFORE - AFTER)) users"
echo "Remaining test users: $AFTER"
echo ""

