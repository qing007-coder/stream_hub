CREATE TABLE IF NOT EXISTS stream_hub.user_logs (
    event_time DateTime64(3),
    level LowCardinality(String), -- 使用低基数优化，节省空间并提速
    uid String,
    ip String,
    method String,
    path String,
    status Int16,
    latency Int64,
    message String,
    trace_id String,
    module LowCardinality(String)
) ENGINE = MergeTree()
PARTITION BY toYYYYMMDD(event_time)
ORDER BY (event_time, uid, trace_id);