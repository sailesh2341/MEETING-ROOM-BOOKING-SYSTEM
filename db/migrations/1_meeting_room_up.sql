CREATE EXTENSION IF NOT EXISTS btree_gist;

CREATE TABLE IF NOT EXISTS users (
    id SERIAL,
    username TEXT NOT NULL,
    password TEXT NOT NULL,
    role TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
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
        ALTER TABLE users ADD CONSTRAINT chk_users_username CHECK (char_length(trim(username)) > 0);
    EXCEPTION WHEN duplicate_object THEN NULL;
    END;

    BEGIN
        ALTER TABLE users ADD CONSTRAINT chk_users_password CHECK (char_length(password) > 0);
    EXCEPTION WHEN duplicate_object THEN NULL;
    END;

    BEGIN
        ALTER TABLE users ADD CONSTRAINT chk_users_role CHECK (role IN ('admin', 'user'));
    EXCEPTION WHEN duplicate_object THEN NULL;
    END;
END $$;

CREATE TABLE IF NOT EXISTS rooms (
    id SERIAL,
    name TEXT NOT NULL,
    capacity INT NOT NULL,
    location TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
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
        ALTER TABLE rooms ADD CONSTRAINT chk_rooms_name CHECK (char_length(trim(name)) > 0);
    EXCEPTION WHEN duplicate_object THEN NULL;
    END;

    BEGIN
        ALTER TABLE rooms ADD CONSTRAINT chk_rooms_capacity CHECK (capacity > 0);
    EXCEPTION WHEN duplicate_object THEN NULL;
    END;
END $$;

CREATE TABLE IF NOT EXISTS bookings (
    id SERIAL,
    room_id INT NOT NULL,
    user_id INT NOT NULL,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NOT NULL,
    status TEXT NOT NULL DEFAULT 'booked',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
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
        ALTER TABLE bookings ADD CONSTRAINT chk_bookings_status CHECK (status IN ('booked', 'cancelled'));
    EXCEPTION WHEN duplicate_object THEN NULL;
    END;

    BEGIN
        ALTER TABLE bookings ADD CONSTRAINT unique_bookings UNIQUE (room_id, user_id, start_time, end_time, status);
    EXCEPTION WHEN duplicate_object THEN NULL;
    END;

    BEGIN
        ALTER TABLE bookings ADD CONSTRAINT no_booking_overlap
        EXCLUDE USING gist (
            room_id WITH =,
            tsrange(start_time, end_time, '[)') WITH &&
        )
        WHERE (status = 'booked');
    EXCEPTION WHEN duplicate_object THEN NULL;
    END;
END $$;

CREATE TABLE IF NOT EXISTS audit_logs (
    id SERIAL,
    user_id INT,
    action TEXT NOT NULL,
    details JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
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
        ALTER TABLE audit_logs ADD CONSTRAINT chk_audit_logs_action CHECK (char_length(trim(action)) > 0);
    EXCEPTION WHEN duplicate_object THEN NULL;
    END;
END $$;

CREATE INDEX IF NOT EXISTS idx_rooms_capacity ON rooms (capacity);
CREATE INDEX IF NOT EXISTS idx_bookings_room_time ON bookings (room_id, start_time, end_time);
CREATE INDEX IF NOT EXISTS idx_bookings_user ON bookings (user_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs (created_at DESC);
