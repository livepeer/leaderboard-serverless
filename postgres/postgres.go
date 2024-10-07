package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/livepeer/leaderboard-serverless/assets"
	"github.com/livepeer/leaderboard-serverless/common"
	"github.com/livepeer/leaderboard-serverless/db/cache"
	"github.com/livepeer/leaderboard-serverless/db/interfaces"
	"github.com/livepeer/leaderboard-serverless/models"
)

type Item struct {
	ID    int
	Stats models.Stats
}

type DB struct {
	pool             *pgxpool.Pool
	connectionString string
	internalCache    cache.Cache
	dbJobManager     interfaces.DBManager
}

var configuredDbTimeout = common.EnvOrDefault("DB_TIMEOUT", 20).(int)
var defaultTimeout = time.Duration(configuredDbTimeout) * time.Second

func Start(connectionUrl string, internalCache cache.Cache, dbJobManager interfaces.DBManager) (*DB, error) {
	common.Logger.Info("Creating connection to database")
	var err error
	ctx, cancel := WithTimeout()
	defer cancel()
	pool, err := pgxpool.Connect(ctx, connectionUrl)
	if err != nil {
		return nil, err
	}
	db := &DB{pool, connectionUrl, internalCache, dbJobManager}

	if err := db.ensureDatabase(); err != nil {
		return nil, err
	}

	if err := db.runMigrations(); err != nil {
		return nil, err
	}
	common.Logger.Info("Database connection successfully created.")
	return db, nil
}

func (db *DB) Close() {
	common.Logger.Debug("Closing database connection pool")
	db.pool.Close()
}

func (db *DB) withConnection(fn func(ctx context.Context, conn *pgxpool.Conn) error) error {
	ctx, cancel := WithTimeout()
	defer cancel()
	conn, err := db.pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer func() {
		common.Logger.Debug("Releasing database connection")
		conn.Release()
	}()

	return fn(ctx, conn)
}

func (db *DB) InsertStats(stats *models.Stats) error {
	err := db.withConnection(func(ctx context.Context, conn *pgxpool.Conn) error {
		qry := `INSERT INTO events(event_time, orchestrator, region_id, payload) 
						SELECT 
							CURRENT_TIMESTAMP, $1, regions.id, $2
						FROM 
								job_types
						JOIN
								regions ON regions.name = $3  AND regions.job_type_id = job_types.id
						WHERE 
								job_types.name = $4`
		common.Logger.Debug("Inserting stats: %v", stats)
		_, err := conn.Exec(ctx, qry, stats.Orchestrator, stats, stats.Region, stats.JobType())
		if err != nil {
			common.Logger.Error("Failed to insert stats: %v", err)
		}
		return err
	})
	return err
}

// BestOrchRegion returns the best region for a given orchestrator and job type in the past 24 hours
func (db *DB) BestAIRegion(orchestratorId string) (*models.Stats, error) {

	since := common.GetDefaultSince()

	// we will return the best region for AI jobs by adding
	// a sort order on success_rate and round_trip_time
	query := &models.StatsQuery{
		Orchestrator: orchestratorId,
		Since:        since,
		Until:        time.Now().UTC(),
		JobType:      models.AI,
		SortFields: []models.StatsQuerySortField{
			models.NewSortField("success_rate", models.SortOrderDesc),
			models.NewSortField("round_trip_time", models.SortOrderAsc),
		},
		Limit: 1,
	}
	aggrStatsResults, err := db.AggregatedStats(query)
	if err != nil {
		return nil, err
	}
	if len(aggrStatsResults.Stats) > 1 {
		return nil, fmt.Errorf("too many stats objects returned.  Found %d stats when searching for the best AI region for orchestrator. Expected 1", len(aggrStatsResults.Stats))
	}
	if len(aggrStatsResults.Stats) == 0 {
		common.Logger.Debug("No best AI region stats found for orchestrator %v", orchestratorId)
		return nil, nil
	}
	return aggrStatsResults.Stats[0], nil
}

func (db *DB) MedianRTT(statsQuery *models.StatsQuery) (float64, error) {

	//make a copy of the query and then clear out any query fields
	//that might interfere with the median RTT query
	statsQueryCopy := *statsQuery
	statsQueryCopy.Limit = 0
	statsQueryCopy.SortFields = nil

	medianRTT := -1.0
	err := setJobTypeIfEmpty(&statsQueryCopy)
	if err != nil {
		return medianRTT, err
	}

	err = db.withConnection(func(ctx context.Context, conn *pgxpool.Conn) error {
		baseSQLQuery := `SELECT PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY COALESCE(round_trip_time, 0)) AS median_round_trip_time FROM event_details WHERE round_trip_time != 0 AND success_rate = 1 AND event_time >= $1 AND event_time <= $2`
		finalQuery, args := db.buildAggregateQueryArgs(&statsQueryCopy, baseSQLQuery, nil)

		common.Logger.Debug("Running query: %v with args: %v", finalQuery, args)
		rows, err := conn.Query(ctx, finalQuery, args...)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var (
				medianRTTCol sql.NullFloat64
			)
			if err := rows.Scan(&medianRTTCol); err != nil {
				return err
			}
			medianRTT = db.extractFloat64(medianRTTCol)
			common.Logger.Debug("Determined media rtt of: %d ", medianRTT)
		}
		return nil
	})
	return medianRTT, err
}

func (db *DB) AggregatedStats(statsQuery *models.StatsQuery) (*models.AggregatedStatsResults, error) {
	aggregatedStatsResults := models.AggregatedStatsResults{
		Stats: []*models.Stats{},
	}

	err := setJobTypeIfEmpty(statsQuery)
	if err != nil {
		return &aggregatedStatsResults, err
	}

	err = db.withConnection(func(ctx context.Context, conn *pgxpool.Conn) error {

		baseSQLQuery := `SELECT orchestrator, payload->>'model' as model, payload->>'pipeline' as pipeline, region_name as region, job_type_name as job_type, AVG(COALESCE(success_rate, 0))  as success_rate, AVG(COALESCE(seg_duration, 0)) as seg_duration, AVG(COALESCE(round_trip_time, 0)) as round_trip_time FROM event_details WHERE event_time >= $1 AND event_time <= $2`
		groupFields := []string{"orchestrator", "region", "job_type", "payload->>'model'", "payload->>'pipeline'"}
		finalQuery, args := db.buildAggregateQueryArgs(statsQuery, baseSQLQuery, groupFields)

		common.Logger.Debug("Running query: %v with args: %v", finalQuery, args)
		rows, err := conn.Query(ctx, finalQuery, args...)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var (
				orchestrator  sql.NullString
				model         sql.NullString
				pipeline      sql.NullString
				region        sql.NullString
				job_type      sql.NullString
				successRate   sql.NullFloat64
				segDuration   sql.NullFloat64
				roundTripTime sql.NullFloat64
			)
			if err := rows.Scan(&orchestrator, &model, &pipeline, &region, &job_type, &successRate, &segDuration, &roundTripTime); err != nil {
				return err
			}
			common.Logger.Trace("Found stats for orchestrator %v, region %v, job_type %v, ", orchestrator, region, job_type)
			aggregatedStatsResults.Stats = append(aggregatedStatsResults.Stats, &models.Stats{
				Orchestrator:  db.extractString(orchestrator),
				Region:        db.extractString(region),
				SuccessRate:   db.extractFloat64(successRate),
				SegDuration:   db.extractFloat64(segDuration),
				RoundTripTime: db.extractFloat64(roundTripTime),
				//model and pipeline are added here to ensure
				//the JobType() function can understand the type of job
				Model:    db.extractString(model),
				Pipeline: db.extractString(pipeline),
			})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	//calculate median RTT for the aggregated stats
	aggregatedStatsResults.MedianRTT, err = db.MedianRTT(statsQuery)
	common.Logger.Debug("Returning %d aggregated stats", len(aggregatedStatsResults.Stats))
	return &aggregatedStatsResults, err
}

// helper function to build the query arguments for the aggregated stats and related mediaRTT queries
func (db *DB) buildAggregateQueryArgs(query *models.StatsQuery, baseQuery string, groupFields []string) (string, []interface{}) {
	args := []interface{}{query.Since, query.Until}

	paramNumber := 3
	if query.Orchestrator != "" {
		baseQuery += fmt.Sprintf(` AND orchestrator = $%d`, paramNumber)
		args = append(args, query.Orchestrator)
		paramNumber++
	}
	if query.Region != "" {
		baseQuery += fmt.Sprintf(` AND region_name = $%d`, paramNumber)
		args = append(args, query.Region)
		paramNumber++
	}
	if query.Pipeline != "" {
		baseQuery += fmt.Sprintf(` AND payload->>'pipeline' = $%d`, paramNumber)
		args = append(args, query.Pipeline)
		paramNumber++
	}
	if query.Model != "" {
		baseQuery += fmt.Sprintf(` AND payload->>'model' = $%d`, paramNumber)
		args = append(args, query.Model)
		paramNumber++
	}
	if query.JobType != models.Unknown {
		baseQuery += fmt.Sprintf(" AND job_type_name = '%s'", query.JobType.String())
	}
	if groupFields != nil && len(groupFields) > 0 {
		baseQuery += ` GROUP BY `
		for i, field := range groupFields {
			if i > 0 {
				baseQuery += `, `
			}
			baseQuery += field
		}
	}

	if query.SortFields != nil && len(query.SortFields) > 0 {
		baseQuery += " ORDER BY "
		for i, field := range query.SortFields {
			if i > 0 {
				baseQuery += ", "
			}
			baseQuery += field.String()
		}
	}
	if query.Limit > 0 {
		baseQuery += fmt.Sprintf(" LIMIT %d", query.Limit)
	}
	return baseQuery, args
}

func (db *DB) RawStats(query *models.StatsQuery) ([]*models.Stats, error) {

	stats := []*models.Stats{}
	err := setJobTypeIfEmpty(query)
	if err != nil {
		return nil, err
	}

	err = db.withConnection(func(ctx context.Context, conn *pgxpool.Conn) error {
		baseQuery := `SELECT payload FROM event_details WHERE orchestrator = $1 AND event_time >= $2 AND event_time <= $3`
		args := []interface{}{query.Orchestrator, query.Since, query.Until}

		paramNumber := 4
		if query.Region != "" {
			baseQuery += fmt.Sprintf(` AND region_name = $%d`, paramNumber)
			args = append(args, query.Region)
			paramNumber++
		}
		if query.Pipeline != "" {
			baseQuery += fmt.Sprintf(` AND payload->>'pipeline' = $%d`, paramNumber)
			args = append(args, query.Pipeline)
			paramNumber++
		}
		if query.Model != "" {
			baseQuery += fmt.Sprintf(` AND payload->>'model' = $%d`, paramNumber)
			args = append(args, query.Model)
			paramNumber++
		}
		if query.JobType != models.Unknown {
			baseQuery += fmt.Sprintf(" AND job_type_name = '%s'", query.JobType.String())
		}
		baseQuery += " ORDER BY event_time DESC"
		common.Logger.Debug("Running query: %v with args: %v", baseQuery, args)
		rows, err := conn.Query(ctx, baseQuery, args...)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var stat models.Stats
			if err := rows.Scan(&stat); err != nil {
				return err
			}
			stats = append(stats, &stat)
		}
		return nil
	})
	return stats, err
}

// Regions returns the regions from the database or the cache if available
func (db *DB) Regions() ([]*models.Region, error) {

	//check the cache for non-expired regions
	cacheResults := db.internalCache.GetRegions()
	if cacheResults.CacheHit && !cacheResults.CacheExpired {
		return cacheResults.Results.([]*models.Region), nil
	}

	// the cache has expired or is empty, so before we retrieve the regions from the database
	// we will run any job manager activities to ensure the regions are up to date
	db.dbJobManager.UpdateRegions()

	regions, err := db.retrieveRegionsFromStore()
	//update the cache with the new regions if there was no error
	if err == nil {
		db.internalCache.UpdateRegions(regions)
	} else {
		//since we got an error, we will invalidate the cache to ensure we don't keep returning stale data
		common.Logger.Error("Failed to retrieve regions from the database.  Cache will be invalidated.  Error: %v", err)
		db.internalCache.InvalidateRegionsCache()
	}

	return regions, err
}

// retrieveRegionsFromStore retrieves the regions from the database without using the cache
func (db *DB) retrieveRegionsFromStore() ([]*models.Region, error) {
	var regions []*models.Region
	err := db.withConnection(func(ctx context.Context, conn *pgxpool.Conn) error {
		qry := "SELECT r.name, r.display_name, jt.name AS type FROM regions r INNER JOIN job_types jt ON jt.id = r.job_type_id"
		rows, err := conn.Query(ctx, qry)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var region models.Region
			if err := rows.Scan(&region.Name, &region.DisplayName, &region.Type); err != nil {
				return err
			}
			regions = append(regions, &region)
		}
		if err := rows.Err(); err != nil {
			return err
		}
		return nil
	})
	return regions, err
}

// InsertRegions inserts regions into the database and returns the number of regions inserted and processed
func (db *DB) InsertRegions(regions []*models.Region) (int, int) {
	regionsInserted := 0
	regionsProcessed := 0
	db.withConnection(func(ctx context.Context, conn *pgxpool.Conn) error {
		for _, region := range regions {
			qry := `INSERT INTO regions(name, display_name, job_type_id)
							SELECT 
								$1, $2, jt.id 
							FROM 
								job_types jt 
							WHERE 
								jt.name = $3`
			_, err := db.pool.Exec(ctx, qry, region.Name, region.DisplayName, region.Type)
			regionsProcessed++
			if err != nil {
				common.Logger.Error("failed to insert region (%s): %v  Skipping...", region.Name, err)
				continue
			}
			regionsInserted++
		}
		common.Logger.Debug("Inserted %d out of %d regions", regionsInserted, len(regions))
		return nil
	})

	//update the cache if new regions were inserted
	if regionsInserted > 0 {
		newRegions, err := db.retrieveRegionsFromStore()
		if err != nil {
			common.Logger.Error("Failed to retrieve regions while updating the cache after inserting a new region.  Cache will be invalidated.  Error: %v", err)
			db.internalCache.InvalidateRegionsCache()
		} else {
			db.internalCache.UpdateRegions(newRegions)
		}
	}
	return regionsInserted, regionsProcessed
}

func (db *DB) Pipelines(query *models.StatsQuery) ([]*models.Pipeline, error) {

	//check the cache for non-expired regions
	cacheResults := db.internalCache.GetPipelines()
	if cacheResults.CacheHit && !cacheResults.CacheExpired {
		return cacheResults.Results.([]*models.Pipeline), nil
	}

	pipelines := []*models.Pipeline{}

	err := db.withConnection(func(ctx context.Context, conn *pgxpool.Conn) error {
		qry :=
			`SELECT
        e.payload ->> 'pipeline' AS pipeline,
        ARRAY_AGG(DISTINCT e.payload ->> 'model') AS models,
        ARRAY_AGG(DISTINCT r.name) AS regions
			FROM events e
			JOIN regions r ON e.region_id = r.id
			WHERE
					e.payload ? 'pipeline' AND
					e.event_time >= $1 AND
					e.event_time <= $2`

		params := []interface{}{query.Since, query.Until}

		// Add region filter if provided
		if query.Region != "" {
			qry += ` AND r.name = $3`
			params = append(params, query.Region)
		}

		qry += ` GROUP BY pipeline ORDER BY pipeline`

		common.Logger.Debug("Running query: %v with args: %v, %v, %v", qry, query.Since, query.Until, query.Region)
		rows, err := conn.Query(ctx, qry, params...)

		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var pipeline models.Pipeline
			var modelsArr []string
			var regionsArr []string
			if err := rows.Scan(&pipeline.Name, &modelsArr, &regionsArr); err != nil {
				return err
			}
			pipeline.Models = modelsArr
			pipeline.Regions = regionsArr
			pipelines = append(pipelines, &pipeline)
		}

		if err := rows.Err(); err != nil {
			return err
		}
		return nil
	})
	return pipelines, err
}

func (db *DB) ensureDatabase() error {
	common.Logger.Info("Ensuring the database exists")
	return db.withConnection(func(ctx context.Context, conn *pgxpool.Conn) error {
		_, err := conn.Exec(ctx, `CREATE DATABASE leaderboard`)
		if err != nil && !strings.Contains(err.Error(), "already exists") {
			return err
		}
		return nil
	})
}

func (db *DB) runMigrations() error {
	common.Logger.Info("Running database migrations to ensure the schema is up to date")

	migrationFs := assets.GetMigrations()
	d, err := iofs.New(migrationFs, assets.Path)
	if err != nil {
		common.Logger.Fatal("Failed to create database migration source instance: %v", err)
	}
	m, err := migrate.NewWithSourceInstance("iofs", d, db.connectionString)
	if err != nil {
		common.Logger.Fatal("Failed to create database migrator instance: %v", err)
		return err
	}
	defer m.Close()

	//configure the logger for the migrate tool
	m.Log = &Log{}

	version, dirty, _ := m.Version()
	common.Logger.Info("Current migration version in the database is: %v", version)

	if dirty {
		msg := "database is in a dirty state and requires intervention to run additional migrations"
		common.Logger.Fatal(msg)
		return errors.New(msg)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		common.Logger.Fatal("Failed to run database migrations for db url: %s -> %v", db.connectionString, err)
		return err
	}

	versionNew, _, _ := m.Version()
	if version == versionNew {
		common.Logger.Info("After applying migrations, database version remained the same at version %v", versionNew)
	} else {
		common.Logger.Info("Database migrated to version %v", version)
	}
	return nil
}

// define a Logger for migrate tool
type Log struct{}

func (l *Log) Verbose() bool {
	return true
}

func (l *Log) Printf(msg string, v ...interface{}) {
	common.Logger.Debug(fmt.Sprintf("[MIGRATOR] %s", msg), v...)
}

// WithTimeout returns a context with the standard timeout for the db.
func WithTimeout() (context.Context, context.CancelFunc) {
	context, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	return context, func() {
		common.Logger.Debug("Calling cancel on context")
		cancel()
	}
}

// setJobTypeIfEmptyadjusts the query to ensure that the job type is set
// Transocding -> When the model and pipeline parameters are empty
// AI -> When pipeline or model is set
func setJobTypeIfEmpty(query *models.StatsQuery) error {
	if query == nil {
		return errors.New("query cannot be nil")
	}

	//job type is already set, no need to adjust
	if query.JobType != models.Unknown {
		return nil
	}

	// validate and adjust the job type
	if query.Model != "" || query.Pipeline != "" {
		if query.JobType == models.Transcoding {
			return errors.New("job type is set to a Transcoding, but model or pipeline is set")
		}
		query.JobType = models.AI
	} else if query.Model == "" && query.Pipeline == "" {
		if query.JobType == models.AI {
			return errors.New("job type is set to AI, but model or pipeline is not set")
		}
		//default to transcoding tp maintain backwards compatibility
		//when we are sure it is not a request related to AI jobs
		query.JobType = models.Transcoding
	}
	common.Logger.Debug("Job type corrected and set to %v", query.JobType)
	return nil
}

// extractFloat64 extracts the float64 value from a sql.NullFloat64
func (db *DB) extractFloat64(column sql.NullFloat64) float64 {
	if column.Valid {
		return column.Float64
	}
	return 0
}

// extractString extracts the string value from a sql.NullString
func (db *DB) extractString(column sql.NullString) string {
	if column.Valid {
		return column.String
	}
	return ""
}
