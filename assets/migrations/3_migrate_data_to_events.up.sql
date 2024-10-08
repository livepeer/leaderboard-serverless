
-- Purpose: Migrate data from region tables to a central, normalized events table.
DO $$
DECLARE
    region_table_name TEXT;
    table_list TEXT[] := ARRAY['mdw', 'fra', 'sin', 'nyc', 'lax', 'lon', 'prg', 'sao', 'atl', 'hnd', 'mad', 'mia', 'mos2', 'sto', 'syd', 'sea', 'tor', 'sjo', 'ash', 'hou', 'atl2'];
    sql_command TEXT;
BEGIN
    FOREACH region_table_name IN ARRAY table_list
    LOOP
        BEGIN
            -- make sure the region table exists before migrating data
            IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = region_table_name) THEN
                RAISE NOTICE 'Table % does not exist while attempting to migrate region table. Assuming this is a clean install where this table normally does not exist (aka nothing to migrate).', region_table_name;
                CONTINUE;
            END IF;
            sql_command := format('
                INSERT INTO events (event_time, orchestrator, region_id, payload)
                SELECT
                    to_timestamp((stats->>''timestamp'')::bigint) AT TIME ZONE ''UTC'' AS event_time,
                    stats->>''orchestrator'' AS orchestrator,
                    (SELECT id FROM regions WHERE name = upper(''%I'') AND job_type_id = (SELECT id FROM job_types WHERE name = CASE 
                        WHEN stats->>''model'' IS NULL AND stats->>''pipeline'' IS NULL THEN ''transcoding'' 
                        ELSE ''ai'' 
                    END)) AS region_id,
                    stats AS payload
                FROM %I', region_table_name, region_table_name);
            RAISE NOTICE 'Migrating region % to events table', region_table_name;
            EXECUTE sql_command;
        EXCEPTION
            WHEN OTHERS THEN
                RAISE NOTICE 'Migrating data from region tables to events failed for table %. Error: %, SQLSTATE: %', region_table_name, SQLERRM, SQLSTATE;
                RAISE;
        END;
    END LOOP;
    RAISE NOTICE 'Migrating data from region tables to events completed successfully!';
END $$ LANGUAGE plpgsql;