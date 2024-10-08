

select * from job_types;

select * from events;

select * from event_details;

select * from event_details WHERE region_name = 'mdw' order by event_time DESC fetch first 10 rows only;


SELECT DISTINCT pipeline, model FROM pipelines WHERE region = 'mdw';

SELECT orchestrator,
	avg(success_rate) as success_rate,
	avg(seg_duration) as seg_duration,
	avg(inference_time) as inference_time,
	avg(round_trip_time) as round_trip_time
	FROM event_details
GROUP BY orchestrator;

SELECT orchestrator,
       avg(success_rate) as success_rate,avg(seg_duration) as seg_duration,avg(round_trip_time) as round_trip_time
FROM event_details WHERE region_name = 'mdw' AND EXTRACT(EPOCH FROM event_time) >= 1724867119 AND EXTRACT(EPOCH FROM event_time) <= 1724953519 GROUP BY orchestrator


SELECT payload
FROM event_details
WHERE
    orchestrator = '0x02b6aac33a397aaadee5227c70c69bb97f2cc529' AND region_name = 'mdw'
ORDER BY
    event_time DESC;

DELETE FROM event_details WHERE id > 126;

select * from event_details
WHERE
    region_name = 'mdw'
ORDER BY
    event_time' DESC;



