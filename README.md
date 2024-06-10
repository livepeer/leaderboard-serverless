# Leaderboard-Serverless

## Deployments

- Production API: https://leaderboard-serverless.vercel.app/api/
- staging API: https://staging-leaderboard-serverless.vercel.app/api/

## Development

All endpoints live in the `/api` folder, every file represents an endpoint and should export a function following the `net/http` handler interface 

```
Handler(w http.ResponseWriter, r *http.Request)
```

#### Install Vercel CLI 

```
$ npm i -g vercel
```

#### Run Dev Server

```
$ vercel dev
```

This will start a local server on `localhost:3000`.

##### ENV Variables

**Storage Option** 
For the endpoints to run you must define a storage options through environment variables. 

Either one of  
- MONGO=<mongodb+srv://...>
- POSTGRES=<postgres://...>

**Secret** 
The POST endpoint requires HMAC authentication. The server can be started with a `SECRET`, any clienting posting data will have to use the same `SECRET` to create a HMAC message based on the `SECRET` that can be provided in the `Authorization` header. 

## Storage Options

Both MongoDB and Postgres are supported, your favorite storage layer is enabled simply through choosing either to connect to through its respective environment variable.

## API

#### `GET /api/aggregated_stats?orchestrator=<orchAddr>&region=<region_code>&since=<timestamp>&until=<timestamp>`

- The orchestrator to get aggregated stats for. If `orchestrator` is not provided the response will include aggregated scores for all orchestrators

- The region to get aggregated stats for. If `region` is not provided all regions will be returned in the response. Region must be one of `"FRA", "MDW", "SIN"`.

- The timestamp to evaluate the query from. If neither `since` nor `until` are provided it will return the results fore the last 24 hours. If `until` is provided but `since` is not, it will return all results before the `until` timestamp.


**example response** 

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

#### `GET /api/raw_stats?orchestrator=<orchAddr>&region=<region_code>&since=<timestamp>`

- The orchestrator's address to check raw stats for. If no parameter for `orchestrator` is provided the request will return `400 Bad Request`

- The region to check stats for. If `region` is not provided all regions will be returned in the response.

- The timestamp to evaluate the query from. If `since` is not provided it will return the results fore the last 24 hours. 
 

**example response**

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

#### POST `/api/post_stats`

Request object 

```
type Stats struct {
	ID                primitive.ObjectID `json:"-" bson:"_id,omitempty"`
	Region            string             `json:"region" bson:"-"`
	Orchestrator      string             `json:"orchestrator" bson:"orchestrator"`
	SegmentsSent      int                `json:"segments_sent" bson:"segments_sent"`
	SegmentsReceived  int                `json:"segments_received" bson:"segments_received"`
	SuccessRate       float64            `json:"success_rate" bson:"success_rate"`
	SegDuration    int64              `json:"seg_duration" bson:"seg_duration"`
	UploadTime     int64              `json:"upload_time" bson:"upload_time"`
	DownloadTime   int64              `json:"download_time" bson:"download_time"`
	TranscodeTime  int64              `json:"transcode_time" bson:"avg_transcode_time"`
	Errors            []string           `json:"errors" bson:"errors"`
	Timestamp         int64              `json:"timestamp" bson:"timestamp"`
}
```

## Deployment Using Vercel

These serverless functions are deployed using Vercel.

The function signatures for the serverless handlers follow the standard `net/http` interface rather than the `aws-lambda-go` interface. 

For more information, check out https://vercel.com/docs/runtimes#official-runtimes/go .
