BEGIN;

CREATE TABLE users (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  email character varying(254) UNIQUE NOT NULL CHECK (email != ''),
  password bytea NOT NULL CHECK (password != ''),
  activation_token uuid DEFAULT gen_random_uuid(),
  recovery_token uuid,
  active boolean DEFAULT FALSE NOT NULL,
  activated timestamp with time zone
  created timestamp with time zone DEFAULT current_timestamp NOT NULL,
  updated timestamp with time zone
  -- confirmation_token varchar(255) NULL,
	-- confirmation_sent_at timestamptz NULL,
	-- recovery_sent_at timestamptz NULL,
	-- email_change_token varchar(255) NULL,
	-- email_change varchar(255) NULL,
	-- email_change_sent_at timestamptz NULL,
	-- last_sign_in_at timestamptz NULL,
);

COMMENT ON COLUMN users.activation_token IS 'used just for initial account activation';

CREATE INDEX ON users (active, activation_token);

COMMIT;
