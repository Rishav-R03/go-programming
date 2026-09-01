CREATE TABLE IF NOT EXISTS request_metrics (
    id          BIGSERIAL PRIMARY KEY,
    client_id   TEXT NOT NULL,
    ts          TIMESTAMPTZ NOT NULL,
    status      TEXT NOT NULL,
    latency_ms  DOUBLE PRECISION NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_request_metrics_client_ts
    ON request_metrics (client_id, ts DESC);

CREATE TABLE IF NOT EXISTS client_policies (
    client_id       TEXT PRIMARY KEY,
    req_limit       INTEGER NOT NULL,
    window_seconds  INTEGER NOT NULL
);
