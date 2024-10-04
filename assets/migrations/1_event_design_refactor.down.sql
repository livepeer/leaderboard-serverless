-- Drop views
DROP VIEW IF EXISTS pipelines;
DROP VIEW IF EXISTS event_details;

-- Drop indexes
DROP INDEX IF EXISTS idx_events_region_id;
DROP INDEX IF EXISTS idx_events_orchestrator;
DROP INDEX IF EXISTS idx_events_timestamp;
DROP INDEX IF EXISTS idx_events_payload_pipeline;
DROP INDEX IF EXISTS idx_events_payload_model;

-- Drop tables
DROP TABLE IF EXISTS events;
DROP TABLE IF EXISTS regions;
DROP TABLE IF EXISTS job_types;