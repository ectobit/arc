CREATE TABLE users (
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  email character varying(254) UNIQUE NOT NULL,
  password character(60) NOT NULL,
  created timestamp with time zone DEFAULT current_timestamp NOT NULL,
  updated timestamp with time zone,
  activation_token uuid DEFAULT uuid_generate_v4(),
  password_reset_token uuid,
  active boolean DEFAULT FALSE NOT NULL
);
CREATE INDEX ON users (active, activation_token);
