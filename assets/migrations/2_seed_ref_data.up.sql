DO $$
BEGIN

  -- Insert initial data into job_types
  INSERT INTO job_types (name) VALUES ('transcoding');
  INSERT INTO job_types (name) VALUES ('ai');

  -- Insert initial data into regions
  INSERT INTO regions (name, display_name, job_type_id) VALUES ('GLOBAL', 'Global', 1);
  INSERT INTO regions (name, display_name, job_type_id) VALUES ('GLOBAL', 'Global', 2);
  INSERT INTO regions (name, display_name, job_type_id) VALUES ('ATL', 'Atlanta', 1);
  INSERT INTO regions (name, display_name, job_type_id) VALUES ('ATL2', 'Atlanta 2', 1);
  INSERT INTO regions (name, display_name, job_type_id) VALUES ('MDW', 'Chicago', 1);
  INSERT INTO regions (name, display_name, job_type_id) VALUES ('MDW', 'Chicago', 2);
  INSERT INTO regions (name, display_name, job_type_id) VALUES ('FRA', 'Frankfurt', 1);
  INSERT INTO regions (name, display_name, job_type_id) VALUES ('FRA', 'Frankfurt', 2);
  INSERT INTO regions (name, display_name, job_type_id) VALUES ('HOU', 'Houston', 1);
  INSERT INTO regions (name, display_name, job_type_id) VALUES ('LON', 'London', 1);
  INSERT INTO regions (name, display_name, job_type_id) VALUES ('LAX', 'Los Angeles', 1);
  INSERT INTO regions (name, display_name, job_type_id) VALUES ('LAX', 'Los Angeles', 2);
  INSERT INTO regions (name, display_name, job_type_id) VALUES ('MAD', 'Madrid', 1);
  INSERT INTO regions (name, display_name, job_type_id) VALUES ('MIA', 'Miami', 1);
  INSERT INTO regions (name, display_name, job_type_id) VALUES ('MOS2', 'Moscow', 1);
  INSERT INTO regions (name, display_name, job_type_id) VALUES ('ASH', 'Nashua', 1);
  INSERT INTO regions (name, display_name, job_type_id) VALUES ('NYC', 'New York City', 1);
  INSERT INTO regions (name, display_name, job_type_id) VALUES ('PRG', 'Prague', 1);
  INSERT INTO regions (name, display_name, job_type_id) VALUES ('SJO', 'San Jose', 1);
  INSERT INTO regions (name, display_name, job_type_id) VALUES ('SAO', 'São Paulo', 1);
  INSERT INTO regions (name, display_name, job_type_id) VALUES ('SEA', 'Seattle', 1);
  INSERT INTO regions (name, display_name, job_type_id) VALUES ('SIN', 'Singapore', 1);
  INSERT INTO regions (name, display_name, job_type_id) VALUES ('STO', 'Stockholm', 1);
  INSERT INTO regions (name, display_name, job_type_id) VALUES ('SYD', 'Sydney', 1);
  INSERT INTO regions (name, display_name, job_type_id) VALUES ('HND', 'Tokyo', 1);
  INSERT INTO regions (name, display_name, job_type_id) VALUES ('TOR', 'Toronto', 1);



EXCEPTION
    WHEN OTHERS THEN
        ROLLBACK;
        RAISE NOTICE 'Creating the job_types and regions failed. Error: %, SQLSTATE: %', SQLERRM, SQLSTATE;
END $$;