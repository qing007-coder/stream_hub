CREATE TABLE IF NOT EXISTS stream_hub.admin_logs (
    event_time DateTime64(3),
    level LowCardinality(String),
    admin_id String,
    admin_email String,
    ip String,
    action String,
    target_type String,
    target_id String,
    detail String,
    result String,
    module LowCardinality(String)
) ENGINE = MergeTree()
PARTITION BY toYYYYMMDD(event_time)
ORDER BY (event_time, admin_id);