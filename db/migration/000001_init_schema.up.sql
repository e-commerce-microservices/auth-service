CREATE TABLE
    session (
        "id" uuid PRIMARY KEY,
        "user_id" serial8 NOT NULL,
        "refresh_token" varchar NOT NULL,
        "user_agent" varchar NOT NULL,
        "client_ip" varchar NOT NULL,
        "expires_at" timestamptz NOT NULL,
        "created_at" timestamptz NOT NULL DEFAULT (now())
    );
