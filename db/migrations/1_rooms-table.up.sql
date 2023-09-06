CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS rooms (
    id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    name TEXT NOT NULL,
    code TEXT NOT NULL UNIQUE,
    created TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE OR REPLACE FUNCTION generate_unique_code()
RETURNS VARCHAR(4) AS
$$
DECLARE
    new_code VARCHAR(4);
BEGIN
    LOOP
        -- Generate a random 6-letter string
        new_code := upper(substr(md5(random()::text), 1, 4));

        -- Check if the generated string already exists in column1 of table1
        IF NOT EXISTS (SELECT 1 FROM rooms WHERE code = new_code) THEN
            RETURN new_code; -- Return the unique string
        END IF;
    END LOOP;
END;
$$
LANGUAGE plpgsql;

ALTER TABLE rooms ALTER COLUMN code SET DEFAULT generate_unique_code();