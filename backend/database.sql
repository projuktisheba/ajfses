-- ==========================================
-- PostgreSQL Database & User Creation Script
-- ==========================================
-- How to run this script:
--   psql -U postgres -f database.sql
-- If shows FATAL:  Peer authentication failed for user "postgres"
--   sudo -i -u postgres
--   psql -f database.sql

-- disconnect
    SELECT pg_terminate_backend(pid)
    FROM pg_stat_activity
    WHERE datname = 'ajfses_main_db'
    AND pid <> pg_backend_pid();


-- (Run as a PostgreSQL superuser, e.g. 'postgres')
CREATE USER ajfses_sudo_user WITH PASSWORD '58yubDDQ6EeVvnXxeAU8GXwiP3j8xtjx';

-- Create a new database owned by that user
CREATE DATABASE ajfses_main_db OWNER ajfses_sudo_user;

-- Grant all privileges on the database to the user
GRANT ALL PRIVILEGES ON DATABASE ajfses_main_db TO ajfses_sudo_user;

-- (Optional) Verify ownership later:
--   \l   → list databases
--   \du  → list roles/users
