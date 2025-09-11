CREATE TABLE IF NOT EXISTS subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    service_name TEXT NOT NULL,
    price INT NOT NULL,
    user_id UUID NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);

CREATE OR REPLACE FUNCTION trigger_set_timestamp()
    RETURNS TRIGGER AS $$
    BEGIN
        NEW.updated_at = now();
    RETURN NEW;
    END;
    $$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS set_timestamp ON subscriptions;

CREATE TRIGGER set_timestamp
    BEFORE UPDATE ON subscriptions
    FOR EACH ROW
    EXECUTE PROCEDURE trigger_set_timestamp();
