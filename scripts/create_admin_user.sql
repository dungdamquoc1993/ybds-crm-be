-- Connect to the account database
\c ybds_user

-- Create admin role if it doesn't exist
INSERT INTO roles (id, created_at, updated_at, name)
VALUES 
    (gen_random_uuid(), CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, 'admin')
ON CONFLICT (name) DO NOTHING;

-- Create AI agent role if it doesn't exist
INSERT INTO roles (id, created_at, updated_at, name)
VALUES 
    (gen_random_uuid(), CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, 'agent')
ON CONFLICT (name) DO NOTHING;

-- Create admin user with password 'admin123'
-- Note: In a real system, you would hash the password properly
-- Password hash and salt below are for 'admin123' using Argon2id
INSERT INTO users (
    id, 
    created_at, 
    updated_at, 
    username, 
    email, 
    phone, 
    password_hash, 
    salt, 
    is_active
)
VALUES (
    gen_random_uuid(),
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP,
    'admin',
    'admin@example.com',
    '1234567890',
    'bmVCVyqD13HDKvHv7LL+CprWsuQPeJrpGwKe10LlHTE=', -- Hash for 'admin123'
    'c26Qaj0QJhcbf2vZ5BshJA==', -- Salt used for hashing
    true
)
ON CONFLICT (username) DO NOTHING;

-- Get the admin user ID
DO $$
DECLARE
    admin_user_id UUID;
    admin_role_id UUID;
BEGIN
    SELECT id INTO admin_user_id FROM users WHERE username = 'admin';
    SELECT id INTO admin_role_id FROM roles WHERE name = 'admin';
    
    -- Assign admin role to the admin user
    -- First check if the relationship already exists
    IF NOT EXISTS (
        SELECT 1 FROM user_roles 
        WHERE user_id = admin_user_id AND role_id = admin_role_id
    ) THEN
        INSERT INTO user_roles (
            id,
            created_at,
            updated_at,
            user_id,
            role_id
        )
        VALUES (
            gen_random_uuid(),
            CURRENT_TIMESTAMP,
            CURRENT_TIMESTAMP,
            admin_user_id,
            admin_role_id
        );
    END IF;
END $$;

-- Create AI agent user with password 'agent123'
INSERT INTO users (
    id, 
    created_at, 
    updated_at, 
    username, 
    email, 
    phone, 
    password_hash, 
    salt, 
    is_active
)
VALUES (
    gen_random_uuid(),
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP,
    'aiagent',
    'ai@example.com',
    '0987654321',
    '3o8VzcgHgWlIPKs78z/xTcFo40PF+G3oUvQAEaE/5vw=', -- Hash for 'agent123'
    'RDqKJ8eIf8jLNxWGDKZLEQ==', -- Salt used for hashing
    true
)
ON CONFLICT (username) DO NOTHING;

-- Get the agent user ID
DO $$
DECLARE
    agent_user_id UUID;
    agent_role_id UUID;
BEGIN
    SELECT id INTO agent_user_id FROM users WHERE username = 'aiagent';
    SELECT id INTO agent_role_id FROM roles WHERE name = 'agent';
    
    -- Assign agent role to the agent user
    -- First check if the relationship already exists
    IF NOT EXISTS (
        SELECT 1 FROM user_roles 
        WHERE user_id = agent_user_id AND role_id = agent_role_id
    ) THEN
        INSERT INTO user_roles (
            id,
            created_at,
            updated_at,
            user_id,
            role_id
        )
        VALUES (
            gen_random_uuid(),
            CURRENT_TIMESTAMP,
            CURRENT_TIMESTAMP,
            agent_user_id,
            agent_role_id
        );
    END IF;
END $$; 