BEGIN;

CREATE TABLE users (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  email character varying(254) UNIQUE NOT NULL CHECK (email != ''),
  password bytea NOT NULL CHECK (password != ''),
  created timestamp with time zone DEFAULT current_timestamp NOT NULL,
  updated timestamp with time zone,
  activation_token uuid DEFAULT gen_random_uuid(),
  password_reset_token uuid,
  active boolean DEFAULT FALSE NOT NULL
);

COMMENT ON COLUMN users.activation_token IS 'used just for initial account activation';

CREATE INDEX ON users (active, activation_token);

COMMIT;
