-- =========================
-- UP
-- =========================
CREATE TABLE claims (
    id BIGSERIAL PRIMARY KEY,
    vin TEXT NOT NULL,
    retail_date DATE NOT NULL,
    ro_open_date DATE NOT NULL,
    ro_close_date DATE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- индексы
CREATE INDEX idx_claims_vin ON claims(vin);
CREATE INDEX idx_claims_ro_open_date ON claims(ro_open_date);

-- =========================
-- DOWN (на будущее)
-- =========================
-- DROP TABLE claims;