-- USERS TABLE --
CREATE TABLE IF NOT EXISTS users (
    id SERIAL,
    username TEXT NOT NULL,
    password TEXT NOT NULL,
    role TEXT NOT NULL
);

DO $$ 
BEGIN
    BEGIN
        ALTER TABLE users ADD CONSTRAINT pk_users PRIMARY KEY (id);
    EXCEPTION WHEN duplicate_object THEN NULL;
    END;

    BEGIN
        ALTER TABLE users ADD CONSTRAINT unique_users_username UNIQUE (username);
    EXCEPTION WHEN duplicate_object THEN NULL;
    END;

    BEGIN
        ALTER TABLE users ADD CONSTRAINT chk_users_role CHECK (role IN ('admin', 'user'));
    EXCEPTION WHEN duplicate_object THEN NULL;
    END;
END $$;

-- ROOMS TABLE --
CREATE TABLE IF NOT EXISTS rooms (
    id SERIAL,
    name TEXT NOT NULL,
    capacity INT NOT NULL
);

DO $$ 
BEGIN
    BEGIN
        ALTER TABLE rooms ADD CONSTRAINT pk_rooms PRIMARY KEY (id);
    EXCEPTION WHEN duplicate_object THEN NULL;
    END;

    BEGIN
        ALTER TABLE rooms ADD CONSTRAINT unique_rooms_name UNIQUE (name);
    EXCEPTION WHEN duplicate_object THEN NULL;
    END;

    BEGIN
        ALTER TABLE rooms ADD CONSTRAINT chk_rooms_capacity CHECK (capacity > 0);
    EXCEPTION WHEN duplicate_object THEN NULL;
    END;
END $$;

-- BOOKINGS TABLE --
CREATE TABLE IF NOT EXISTS bookings (
    id SERIAL,
    room_id INT NOT NULL,
    user_id INT NOT NULL,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NOT NULL
);

DO $$ 
BEGIN
    BEGIN
        ALTER TABLE bookings ADD CONSTRAINT pk_bookings PRIMARY KEY (id);
    EXCEPTION WHEN duplicate_object THEN NULL;
    END;

    BEGIN
        ALTER TABLE bookings ADD CONSTRAINT fk_bookings_room 
        FOREIGN KEY (room_id) REFERENCES rooms(id) ON DELETE CASCADE;
    EXCEPTION WHEN duplicate_object THEN NULL;
    END;

    BEGIN
        ALTER TABLE bookings ADD CONSTRAINT fk_bookings_user 
        FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
    EXCEPTION WHEN duplicate_object THEN NULL;
    END;

    BEGIN
        ALTER TABLE bookings ADD CONSTRAINT chk_bookings_time CHECK (start_time < end_time);
    EXCEPTION WHEN duplicate_object THEN NULL;
    END;

    BEGIN
        ALTER TABLE bookings ADD CONSTRAINT unique_bookings 
        UNIQUE (room_id, start_time, end_time);
    EXCEPTION WHEN duplicate_object THEN NULL;
    END;
END $$;

-- AUDIT LOGS TABLE --
CREATE TABLE IF NOT EXISTS audit_logs (
    id SERIAL,
    user_id INT,
    action TEXT NOT NULL,
    details JSONB,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

DO $$ 
BEGIN
    BEGIN
        ALTER TABLE audit_logs ADD CONSTRAINT pk_audit_logs PRIMARY KEY (id);
    EXCEPTION WHEN duplicate_object THEN NULL;
    END;

    BEGIN
        ALTER TABLE audit_logs ADD CONSTRAINT fk_audit_logs_user 
        FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL;
    EXCEPTION WHEN duplicate_object THEN NULL;
    END;

    BEGIN
        ALTER TABLE audit_logs ADD CONSTRAINT chk_audit_logs_action 
        CHECK (char_length(action) > 0);
    EXCEPTION WHEN duplicate_object THEN NULL;
    END;
END $$;
