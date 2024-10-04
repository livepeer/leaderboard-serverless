
-- Create job_types table
CREATE TABLE job_types
(
    id           SERIAL PRIMARY KEY,
    name         VARCHAR(56) NOT NULL UNIQUE
);

-- Create regions table
CREATE TABLE regions
(
    id           SERIAL PRIMARY KEY,
    name         VARCHAR(56) NOT NULL,
    display_name VARCHAR(128) NOT NULL,
    job_type_id  INTEGER REFERENCES job_types (id)
);
ALTER TABLE regions ADD CONSTRAINT unique_name_jobtype UNIQUE (name, job_type_id);


-- create new events table
CREATE TABLE events
(
    id           SERIAL PRIMARY KEY,
    event_time    TIMESTAMPTZ   NOT NULL DEFAULT CURRENT_TIMESTAMP,
    orchestrator VARCHAR(56) NOT NULL,
    region_id    INTEGER REFERENCES regions (id),
    payload      JSONB       NOT NULL
);


-- Create new indexes
CREATE INDEX idx_events_region_id ON events(region_id);
CREATE INDEX idx_events_orchestrator ON events (orchestrator);
CREATE INDEX idx_events_timestamp ON events (event_time);
CREATE INDEX idx_events_payload_pipeline ON events ((payload->>'pipeline'));
CREATE INDEX idx_events_payload_model ON events ((payload->>'model'));


CREATE VIEW event_details AS
SELECT r.name AS region_name,
    j.name AS job_type_name,
    e.id,
    e.event_time,
    CAST(e.payload->>'success_rate' as FLOAT) AS success_rate,
    CAST(e.payload->>'seg_duration' as FLOAT) AS seg_duration,
    CAST(e.payload->>'round_trip_time' as FLOAT) AS round_trip_time,
    e.orchestrator,
    e.payload
FROM events e
        INNER JOIN
    regions r ON e.region_id = r.id
        INNER JOIN
    job_types j ON r.job_type_id = j.id;

-- Create a function to show unique pipelines and models by region and date range
CREATE OR REPLACE FUNCTION get_pipelines(
    start_date TIMESTAMPTZ,
    end_date TIMESTAMPTZ,
    region_name TEXT DEFAULT NULL
)
RETURNS TABLE (
    pipeline TEXT,
    models TEXT[],
    regions TEXT[]
) AS $$
    SELECT
        e.payload ->> 'pipeline' AS pipeline,
        ARRAY_AGG(DISTINCT e.payload ->> 'model') AS models,
        ARRAY_AGG(DISTINCT r.name) AS regions
    FROM events e
    JOIN regions r ON e.region_id = r.id
    WHERE
        e.payload ? 'pipeline' AND
        e.event_time >= start_date AND
        e.event_time <= end_date AND
        (region_name IS NULL OR r.name = region_name)
    GROUP BY pipeline;
$$ LANGUAGE sql;


