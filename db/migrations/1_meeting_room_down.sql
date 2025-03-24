-- BOOKINGS CONSTRAINTS --
ALTER TABLE bookings DROP CONSTRAINT IF EXISTS fk_bookings_room;
ALTER TABLE bookings DROP CONSTRAINT IF EXISTS fk_bookings_user;
ALTER TABLE bookings DROP CONSTRAINT IF EXISTS pk_bookings;
ALTER TABLE bookings DROP CONSTRAINT IF EXISTS unique_bookings;
ALTER TABLE bookings DROP CONSTRAINT IF EXISTS chk_bookings_time;

-- AUDIT LOGS CONSTRAINTS --
ALTER TABLE audit_logs DROP CONSTRAINT IF EXISTS fk_audit_logs_user;
ALTER TABLE audit_logs DROP CONSTRAINT IF EXISTS pk_audit_logs;
ALTER TABLE audit_logs DROP CONSTRAINT IF EXISTS chk_audit_logs_action;

-- ROOMS CONSTRAINTS --
ALTER TABLE rooms DROP CONSTRAINT IF EXISTS pk_rooms;
ALTER TABLE rooms DROP CONSTRAINT IF EXISTS unique_rooms_name;
ALTER TABLE rooms DROP CONSTRAINT IF EXISTS chk_rooms_capacity;

-- USERS CONSTRAINTS --
ALTER TABLE users DROP CONSTRAINT IF EXISTS pk_users;
ALTER TABLE users DROP CONSTRAINT IF EXISTS unique_users_username;
ALTER TABLE users DROP CONSTRAINT IF EXISTS chk_users_role;

-- Drop tables in reverse order (to avoid dependency issues)
DROP TABLE IF EXISTS audit_logs;
DROP TABLE IF EXISTS bookings;
DROP TABLE IF EXISTS rooms;
DROP TABLE IF EXISTS users;
