package postgres

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	"github.com/livepeer/leaderboard-serverless/models"

	_ "github.com/lib/pq"
)

type Item struct {
	ID    int
	Stats models.Stats
}

type DB struct {
	*sql.DB
}

func Start() (*DB, error) {
	sql, err := sql.Open("postgres", os.Getenv("POSTGRES"))
	if err != nil {
		return nil, err
	}
	db := &DB{sql}

	if err := db.ensureDatabase(); err != nil {
		return nil, err
	}

	if err := db.ensureTables(); err != nil {
		return nil, err
	}
	return db, nil
}

func (db *DB) InsertStats(stats *models.Stats) error {
	qry := fmt.Sprintf("INSERT INTO %s (stats) VALUES($1)", stats.Region)
	_, err := db.Exec(qry, stats)
	return err
}

func (db *DB) AggregatedStats(orch, region string, since, until int64) ([]*models.Stats, error) {
	qry := fmt.Sprintf(`SELECT stats->>'orchestrator', 
	avg(CAST(stats->>'success_rate' as FLOAT)) as success_rate, 
	avg(CAST(stats->>'seg_duration' as FLOAT)) as seg_duration, 
	avg(CAST(stats->>'round_trip_time' as FLOAT)) as round_trip_time  
	FROM %v WHERE (stats->>'timestamp')::int >= %v AND (stats->>'timestamp')::int <= %v`,
		region, since, until)
	if orch != "" {
		qry += fmt.Sprintf(`AND stats->>'orchestrator' = '%v' `, orch)
	}

	qry += `GROUP BY stats->>'orchestrator'`

	rows, err := db.Query(qry)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	stats := []*models.Stats{}
	for rows.Next() {
		var (
			id            string
			successRate   sql.NullFloat64
			segDuration   sql.NullFloat64
			roundTripTime sql.NullFloat64
		)
		if err := rows.Scan(&id, &successRate, &segDuration, &roundTripTime); err != nil {
			return nil, err
		}
		stats = append(stats, &models.Stats{
			ID:            id,
			SuccessRate:   successRate.Float64,
			SegDuration:   segDuration.Float64,
			RoundTripTime: roundTripTime.Float64,
		})
	}
	return stats, nil
}

func (db *DB) RawStats(orch, region string, since, until int64) ([]*models.Stats, error) {
	qry := fmt.Sprintf(`SELECT stats FROM %v WHERE stats->>'orchestrator' = '%v' AND stats->>'timestamp' >= '%v' AND stats->>'timestamp' <= '%v' ORDER BY stats->>'timestamp' DESC`, region, orch, since, until)
	rows, err := db.Query(qry)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	stats := []*models.Stats{}
	for rows.Next() {
		var stat models.Stats
		if err := rows.Scan(&stat); err != nil {
			return nil, err
		}
		stats = append(stats, &stat)
	}
	return stats, nil
}

func (db *DB) ensureDatabase() error {
	_, err := db.Exec(`CREATE DATABASE leaderboard`)
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		db.Close()
		return err
	}
	return nil
}

func (db *DB) ensureTables() error {
	for _, region := range models.Regions {
		_, err := db.Exec(fmt.Sprintf(`
		CREATE TABLE %s (
			id SERIAL PRIMARY KEY,
			stats JSONB
		);
		`, region))
		if err != nil && !strings.Contains(err.Error(), "already exists") {
			db.Close()
			return err
		}
	}
	return nil
}
