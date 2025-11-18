CREATE TABLE proxies (
    url TEXT PRIMARY KEY,
    score INTEGER NOT NULL,
    alive BOOLEAN NOT NULL,
    last_test TIMESTAMP NOT NULL,
    usage_count INTEGER NOT NULL,
    fail_count INTEGER NOT NULL,
    success_count INTEGER NOT NULL,
    latency_ms INTEGER
);
