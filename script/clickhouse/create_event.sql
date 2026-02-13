CREATE TABLE behavior_event
(
    event_id String,
    event_type LowCardinality(String),

    user_id String,

    resource_type LowCardinality(String),
    resource_id String,

    timestamp Int64,
    event_time DateTime,

    source LowCardinality(String),
    client LowCardinality(String)
)
    ENGINE = MergeTree
PARTITION BY toYYYYMMDD(event_time)
ORDER BY (event_type, resource_type, resource_id, event_time);
