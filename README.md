# Leaderboard-Serverless

## Overview

This repository contains the Livepeer Leaderboard API server - A set of HTTP Rest Endpoints that provide statistics from job testing the Livepeer Orchestrator Network. 
see [figure 1](#figure-1---logical-overview-of-livepeers-testing).

### Key Features Supported
* APIs support POST (submit new test job statistics) and GET (query test job statistics).
  * Includes Raw Stats - query every job test results.
  * Includes Aggregated Stats - query the aggregated raw stats  
* Supports AI and Transcoding Test Job Types.
* Uses Postgres for data storage.
* Has a database migration tool to support different applications upgrades.
* Deploy using Docker or Binaries.

The serverless functions can be deployed using `vercel-cli` or self-hosted. 

### Production API Consumers 
* Livepeer Inc - https://explorer.livepeer.org (Performance Leaderboard)
* Interptr- https://interptr-latest-test-streams.vercel.app (Transcoding Test Stream UI)
* Livepeer.Cloud SPE - https://inspector.livepeer.cloud/ (AI and Transcoding Test Job UI) 


## Required Software

This software can run on many operating systems.  Make sure you have the below software installed on you system before proceeding.

* **Git** 
  * [Install Guide](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git)
* **Go** 1.23.1 or newer
  * https://go.dev/doc/install
* **Postgres**
  * A postgres database is required. 
  * The database must be running prior to running the API server
  * see section *"Local Postgres local DB (optional)"*
* Vercel CLI
  * https://vercel.com/docs/cli
  * Optional, unless you deploy to Vercel
* Docker
  * Optional, unless you run a local Postgres DB

## Development

### Clone the repo

`git clone https://github.com/livepeer/leaderboard-serverless.git`
`cd <repo-folder>`

### Run Build

Run the go build:

`go build -o leaderboard-serverless main.go`

When the build completes, check for the binary file:
`leaderboard-serverless`

### Local Postgres local DB (optional)
The `<repo-folder>/docker-compose.yml` file will allow you to spin up Postgres with a database 
called `leaderboad` and user named `leaderboard` and password `leaderboard` **DO NOT USE THESE VALUES FOR PRODUCTION!**

To run a local DB run the following docker command

`docker compose up -d`

This will enable the database to run on 

`<your ip>:5432`

When configuring the API Server, you will use the following environment variable:

`POSTGRES=postgres://leaderboard:leaderboard@<your ip>:5432/leaderboard?sslmode=disable;`

### Configure Environment Variables

*Note: Make sure to set the environment variables for your specific operating system*

#### Required 
* `POSTGRES=postgres://<YOUR_DATABASE_URL>` - The database conneciton URL. For example: postgres://leaderboard:leaderboard@localhost:5432/leaderboard
* `SECRET=<YOUR_SECRET KEY>`
  - The `SECRET` environment variable is part of the API security. The `/api/post_stats` HTTP POST Endpoint requires HMAC authentication.Any HTTP client posting data will have to use the same `SECRET` to create a HMAC message based on the `SECRET` that must be provided in the HTTP Authorization Header. The value of `SECRET` needs to be changed to your own private value. **DO NOT USE THIS VALUE FOR PRODUCTION!**

#### Optional
* `START_TIME_WINDOW` - The lookback period in hours for retrieving stats in aggregate or raw stats. Default is 24h.
* `DB_TIMEOUT` - The time in seconds used for database operations before they will timeout. Default is 20s.
* `LOG_LEVEL`  - The logging level of the application. Default is INFO.
* `SECRET` - The secret used in HTTP Authorization headers to authenitcate callers of protected endpoints.  See the section on Endpoint Security.  This is optional is you do not intend to post stats.
* `REGIONS_CACHE_TIMEOUT` - The timeout for the application to cache regions before retrieving them from the database.  The default is 60 seconds.
* `PIPELINES_CACHE_TIMEOUT` - The timeout for the application to cache pipelines before retrieving them from the database.  The default is 60 seconds.
* `CATALYST_REGION_URL` - A custom URL point to the Catlyst JSON representing regions to be inserted into the database.

### Run the App

Next, execute the `leaderboard-serverless` binary

Your console will show the following output:
```
2024/09/06 17:09:50 Setting up log level to default: INFO
time=2024-09-06T17:09:50.771Z level=INFO msg="Server starting on port 8080"
```

In your browser, go to `http://<YOUR_SERVER_IP>:8080/api/regions` to verify the server works

You will see similar JSON output:
```
{
  "regions": [
    {
      "id": "TOR",
      "name": "Toronto",
      "type": "transcoding"
    },
    {
      "id": "HND",
      "name": "Tokyo",
      "type": "transcoding"
    },
    ...
  ]
}
```

### Troubleshooting Tip

If your `POSTGRES` environment variable is misconfigured, you may see an error like this:

```
failed to connect to `host=localhost user=leaderboard database=leaderboard`: dial error (dial tcp [::1]:5432: connect: connection refused)
```

Make sure you can connect to the database url defined with proper user and password credentials.

### Running The Unit Tests

Tests can be run with the following command in the project root:

```
go build && go test -p 1-v ./...
```

Since we use an emebedded database for the entire test run between packages, test packages must be run one at a time (-p 1 flag). 

See here for more: https://github.com/fergusstrange/embedded-postgres/issues/115

## Production

Livepeer Inc hosts a version of this API to support the Livepeer Explorer Performance Leaderboard.

### Livepeer Inc's API
- Production API: https://leaderboard-serverless.vercel.app/api/
- Staging API: https://staging-leaderboard-serverless.vercel.app/api/

## API Reference

All APIs start with `/api/`

#### `GET /api/aggregated_stats?orchestrator=<orchAddr>&region=<region_code>&since=<timestamp>&until=<timestamp>`

| Parameter         | Description                                                                                                                                                           |
|-------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `orchestrator`     | The orchestrator to get aggregated stats for. If `orchestrator` is not provided, the response will include aggregated scores for all orchestrators.                    |
| `region`          | The region to get aggregated stats for. If `region` is not provided, all regions will be returned in the response. Region must be a registered region in the database.  For example `"FRA", "MDW", "SIN"`.         |
| `since`           | The timestamp to evaluate the query from. If neither `since` nor `until` are provided, it will return the results starting from the time period specified by the environment variable `START_TIME_WINDOW` or its default.                                |
| `until`           | If `until` is provided but `since` is not, it will return all results before the `until` timestamp.                                                                     |


#### Transcoding Response 

```
{
   "<orchAddr>": {
    "MDW": {
      "total_score": 5.5,
      "latency_score": 6.01,
      "success_rate": 91.5
    },
    "FRA": {
    	"total_score": 2.5,
      "latency_score": 2.5,
      "success_rate": 100
    },
    "SIN": {
    	"total_score": 6.6,
      "latency_score": 7.10
			"success_rate": 93
    }
  },
   "<orchAddr2>": {
  		...
  },
	...
}
```

#### AI Request 

This is similar to the above `aggregated_stats` requests with the added parameters documented below.

#### `GET /api/aggregated_stats?orchestrator=<orchAddr>&region=<region_code>&since=<timestamp>&until=<timestamp>&model=<model>&pipeline=<pipeline>`

| Parameter         | Description                                                                                                                          |
|-------------------|--------------------------------------------------------------------------------------------------------------------------------------|
| `model`           | The model to check stats for. Model is a required field. If `model` is not provided, you will get an HTTP status code 400 (Bad Request). |
| `pipeline`        | The pipeline to check stats for. Pipeline is a required field. If `pipeline` is not provided, you will get an HTTP status code 400 (Bad Request). |


#### AI Response 

```
{
  "0x10742714f33f3d804e3fa489618b5c3ca12a6df7": {
    "FRA": {
      "success_rate": 1,
      "round_trip_score": 0.742521971754293,
      "score": 0.909882690114002
    },
    "LAX": {
      "success_rate": 1,
      "round_trip_score": 0.844420265972075,
      "score": 0.945547093090226
    },
    "MDW": {
      "success_rate": 1,
      "round_trip_score": 0.797933017387645,
      "score": 0.929276556085676
    }
  }
}
```

#### `GET /api/raw_stats?orchestrator=<orchAddr>&region=<region_code>&since=<timestamp>`

| Parameter         | Description                                                                                                                           |
|-------------------|---------------------------------------------------------------------------------------------------------------------------------------|
| `orchestrator`     | The orchestrator's address to check raw stats for. If no parameter for `orchestrator` is provided, the request will return `400 Bad Request`. |
| `region`          | The region to check stats for. If `region` is not provided, all regions will be returned in the response.                               |
| `since`           | The timestamp to evaluate the query from. If `since` is not provided, it will return the results starting from time period specified by the environment variable `START_TIME_WINDOW` or its default.                 |

#### Transcoding Response

For each region return an array of the metrics from the 'metrics gathering' section as a "raw dump"

```
{
 "FRA": [
    {
	    "timestamp": number,
        "segments_sent": number,
        "segments_received": number,
        "success_rate": number,
        "seg_duration": number,
        "upload_time": number,
        "download_time": number,
        "transcode_time": number,
        "round_trip_time": number,
        "errors": Array
      }
   ],
   "MDW": [...],
   "SIN": [...]
}
```

#### AI Request 

This is similar to the above `raw_stats` requests with the added parameters documented below.

#### `GET /api/raw_stats?orchestrator=<orchAddr>&region=<region_code>&since=<timestamp>&until=<timestamp>&model=<model>&pipeline=<pipeline>`

| Parameter         | Description                                                                                                                               |
|-------------------|-------------------------------------------------------------------------------------------------------------------------------------------|
| `model`           | The model to check stats for. Model is an optional field.                                                                                 |
| `pipeline`        | The pipeline to check stats for. Pipeline is a required field. If pipeline is not provided, you will get Transcoding data (see raw_stats for Transcoding). |

#### AI Response 

```
{
  "FRA": [
    {
      "region": "FRA",
      "orchestrator": "0x10742714f33f3d804e3fa489618b5c3ca12a6df7",
      "success_rate": 1,
      "round_trip_time": 7.236450406,
      "errors": [],
      "timestamp": 1726864722,
      "model": "stabilityai/stable-video-diffusion-img2vid-xt-1-1",
      "model_is_warm": true,
      "pipeline": "Image to video",
      "input_parameters": "{\"fps\":8,\"height\":256,\"model_id\":\"stabilityai/stable-video-diffusion-img2vid-xt-1-1\",\"motion_bucket_id\":127,\"noise_aug_strength\":0.065,\"width\":256}",
      "response_payload": "{\"images\":[{\"nsfw\":false,\"seed\":1384909895,\"url\":\"/stream/112b6ad2/772ed708.mp4\"}]}\n"
    },
    {
      "region": "FRA",
      "orchestrator": "0x10742714f33f3d804e3fa489618b5c3ca12a6df7",
      "success_rate": 1,
      "round_trip_time": 7.333097532,
      "errors": [],
      "timestamp": 1726857456,
      "model": "stabilityai/stable-video-diffusion-img2vid-xt-1-1",
      "model_is_warm": true,
      "pipeline": "Image to video",
      "input_parameters": "{\"fps\":8,\"height\":256,\"model_id\":\"stabilityai/stable-video-diffusion-img2vid-xt-1-1\",\"motion_bucket_id\":127,\"noise_aug_strength\":0.065,\"width\":256}",
      "response_payload": "{\"images\":[{\"nsfw\":false,\"seed\":2533618378,\"url\":\"/stream/774f96b9/105469a0.mp4\"}]}\n"
    }
  ],
  "LAX": [
    {
      "region": "LAX",
      "orchestrator": "0x10742714f33f3d804e3fa489618b5c3ca12a6df7",
      "success_rate": 1,
      "round_trip_time": 4.110541139,
      "errors": [],
      "timestamp": 1726866030,
      "model": "stabilityai/stable-video-diffusion-img2vid-xt-1-1",
      "model_is_warm": true,
      "pipeline": "Image to video",
      "input_parameters": "{\"fps\":8,\"height\":256,\"model_id\":\"stabilityai/stable-video-diffusion-img2vid-xt-1-1\",\"motion_bucket_id\":127,\"noise_aug_strength\":0.065,\"width\":256}",
      "response_payload": "{\"images\":[{\"nsfw\":false,\"seed\":689122349,\"url\":\"/stream/bafdeb1f/5f1f1fee.mp4\"}]}\n"
    },
  ]
}
```


#### POST `/api/post_stats`

This accepts a JSON encododed Stats object that maps to the `Stats` struct below.

```
// Stats are the raw stats per test stream
type Stats struct {
	Region        string  `json:"region" bson:"-"`
	Orchestrator  string  `json:"orchestrator" bson:"orchestrator"`
	SuccessRate   float64 `json:"success_rate" bson:"success_rate"`
	RoundTripTime float64 `json:"round_trip_time" bson:"round_trip_time"`
	Errors        []Error `json:"errors" bson:"errors"`
	Timestamp     int64   `json:"timestamp" bson:"timestamp"`

	// Transcoding stats fields
	SegDuration      float64 `json:"seg_duration,omitempty" bson:"seg_duration,omitempty"`
	SegmentsSent     int     `json:"segments_sent,omitempty" bson:"segments_sent,omitempty"`
	SegmentsReceived int     `json:"segments_received,omitempty" bson:"segments_received,omitempty"`
	UploadTime       float64 `json:"upload_time,omitempty" bson:"upload_time,omitempty"`
	DownloadTime     float64 `json:"download_time,omitempty" bson:"download_time,omitempty"`
	TranscodeTime    float64 `json:"transcode_time,omitempty" bson:"transcode_time,omitempty"`

	// AI stats fields
	Model           string `json:"model,omitempty" bson:"model,omitempty"`
	ModelIsWarm     bool   `json:"model_is_warm,omitempty" bson:"model_is_warm,omitempty"`
	Pipeline        string `json:"pipeline,omitempty" bson:"pipeline,omitempty"`
	InputParameters string `json:"input_parameters,omitempty" bson:"input_parameters,omitempty"`
	ResponsePayload string `json:"response_payload,omitempty" bson:"response_payload,omitempty"`
}
```

#### `GET /api/pipelines?region=<region_code>&since=<timestamp>&until=<timestamp>`

| Parameter         | Description                                                                                                                                                      |
|-------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `region`          | The region to check stats for. If `region` is not provided, all pipelines will be returned in the response.                                                       |
| `since`           | The timestamp to evaluate the query from. If neither `since` nor `until` are provided, it will return the results starting from the time period specified by the environment variable `START_TIME_WINDOW` or its default.                           |
| `until`           | If `until` is provided but `since` is not, it will return all results before the `until` timestamp.                                                                |


This endpoint outputs the pipelines and models in JSON format.

#### Response 
```
{
  "pipelines": [
    {
      "id": "Audio to text",
      "models": [
        "openai/whisper-large-v3"
      ],
      "regions": [
        "FRA",
        "LAX",
        "MDW"
      ]
    },
    {
      "id": "Image to image",
      "models": [
        "ByteDance/SDXL-Lightning",
        "timbrooks/instruct-pix2pix"
      ],
      "regions": [
        "FRA",
        "LAX",
        "MDW"
      ]
    }
    ...
  ]
}
```


#### `GET /api/regions`

This endpoint outputs the regions in JSON format.  It does not take any parameters.

#### Response 
```
{
  "regions": [
    {
      "id": "TOR",
      "name": "Toronto",
      "type": "transcoding"
    },
    {
      "id": "HND",
      "name": "Tokyo",
      "type": "transcoding"
    },
    {
      "id": "SYD",
      "name": "Sydney",
      "type": "transcoding"
    },
    {
      "id": "STO",
      "name": "Stockholm",
      "type": "transcoding"
    },
  ]
}
```

## Database

The database is responsible for storing the results of test data for each job executed as well as some reference data (regions).

The new schema allows for easy addition for Regions and Job Types (AI, Transcoding, etc...)

The new schema has index all queries to allow fast searching on common fields:
orchestrator
timestamp
model/pipeline
region
job type 

![Leaderboard Database Entity Relation Diagram](docs/new-db-entity-relation.png)

## Migrations

As reference data and schema design evolves, it is necessary to deploy these changes to your backend database.  In order to avoid human error and manual tasks, database migrations are automated in this project.  This means one can update DDL and DML in the databsae with the addition of a SQL script.  In other words, you can alter the structure of the database or the data stored in the databse with these migrations.

These scripts must follow the naming convention of <migration_number>_descriptive_text>.<up | down>.sql.  As an example:

```
8_create_users_table.up.sql
8_create_users_table.down.sql
```
This example defines the eighth migration for the datbaase with two migrations, one to upgrade the database and one to revert it.  When the application starts its connection to the database, it will run all upgrade (or up) migrations automatically.

Database migrations are found in the assets/migrations folder.  These are embeded in the golang binary built from this project for ease of access regardless of where the application is deployed.  This also allows Vercel to use these migrations.  The migrations are loaded and processed by the [golang-migrate](https://github.com/golang-migrate/migrate) project.  Please read their documentation for more details.

### Upgrading the database

The database upgrade runs automatically once the first API is called.  This is because the first API call instantiates a connection to the database, which first checks that the database exists and then applies any migrations necessary to the database.

### Downgrading the database

IMPORTANT: This is not a substitute for backing up your database.  Before attempting any of the below steps, ensure you back up your database.

Run the following command using docker to downgrade the database to it's original state (-all) or to a specific version (N).  The example below shows how to downgrade by one version (N=1).

> docker run --rm -it -v ./assets/migrations:/migrations --network host migrate/migrate -path=/migrations/ -verbose -database $POSTGRES down 1

The resulting output:

```2024/09/04 17:00:25 Start buffering 2/d migrate_data_to_events
2024/09/04 17:00:25 Start buffering 1/d event_design_refactor
2024/09/04 17:00:25 Read and execute 2/d migrate_data_to_events
2024/09/04 17:00:25 Finished 2/d migrate_data_to_events (read 8.673955ms, ran 8.698581ms)
2024/09/04 17:00:25 Read and execute 1/d event_design_refactor
2024/09/04 17:00:25 Finished 1/d event_design_refactor (read 26.303185ms, ran 27.874975ms)
2024/09/04 17:00:25 Finished after 55.855055ms
2024/09/04 17:00:25 Closing source and database
```

## Old DB Schema Reference

Each Region was a separate table. The only tests that existed was the Transcoding work. As part of the [Livepeer.Cloud SPE Proposal #2](https://explorer.livepeer.org/treasury/69112973991711207069799657820129915730234258793790128205157315299386501373337), the database was migrated to a table called "EVENTS".

![Leaderboard Database Entity Relation Diagram](docs/old-db-entity-relation.png)

## Appendix

### Figure 1 - Logical Overview of Livepeer's Testing

This repository documents the *"Livepeer API Server"* and *"Database Server"* boxes defined in figure 1.

![Job Tester Logical Architecture](docs/job-tester-logical_architecture.png)