-- Create admin user: Irakli Giorgadze
-- Email: ika.giorgadze@gmail.com
-- Password: @dm!n123
-- Password will be hashed using bcrypt (cost factor 12)

INSERT INTO users (email, name, role, password_h, active, created_at, updated_at)
VALUES (
  'ika.giorgadze@gmail.com',
  'Irakli Giorgadze',
  'admin',
  crypt('@dm!n123', gen_salt('bf', 12)),
  true,
  now(),
  now()
)
ON CONFLICT (email) DO UPDATE
SET 
  role = 'admin',
  password_h = crypt('@dm!n123', gen_salt('bf', 12)),
  active = true,
  updated_at = now();

-- Verify the user was created
SELECT id, email, name, role, active, created_at 
FROM users 
WHERE email = 'ika.giorgadze@gmail.com';

