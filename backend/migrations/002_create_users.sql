-- =========================
  -- UP
  -- =========================
  CREATE TABLE users (
      id BIGSERIAL PRIMARY KEY,
      email TEXT NOT NULL,
      password_hash TEXT NOT NULL,
      is_active BOOLEAN NOT NULL DEFAULT TRUE,
      created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
      updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
  );

  -- уникальность email без учета регистра
  CREATE UNIQUE INDEX idx_users_email_lower_unique ON users (lower(email));

  -- =========================
  -- DOWN (на будущее)
  -- =========================
  -- DROP TABLE users;