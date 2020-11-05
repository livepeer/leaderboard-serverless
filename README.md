# Leaderboard-Serverless

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

This will start a local server on `localhost:3000`

##### ENV Variables

For the endpoints to run you must define a storage options through environment variables. 

Either one of  
- MONGO=<mongodb+srv://...>
- POSTGRES=<postgres://...>

## Storage Options

Both MongoDB and Postgres are supported, your favorite storage layer is enabled simply through choosing either to connect to through its respective environment variable.

## API

#### `GET /api/aggregated_stats?orchestrator=<orchAddr>&region=<Airport_code>&since=<timestamp>`

- If `orchestrator` is not provided the response will include aggregated scores for all orchestrators

- If `region` is not provided all regions will be returned in the response. "GLOBAL" would be included as well which would be the average of the regions. 

- If `since` is not provided we will default to something sensible, for example 24h. 

**example response** 

```
{
   "<orchAddr>": {
    "MDW": {
    	"score": 5.5,
      "success_rate": 91.5
    },
    "FRA": {
    	"score": 2.5,
      "success_rate": 100
    },
    "SIN": {
    	"score": 6.6,
			"success_rate": 93
    }
  },
   "<orchAddr2>": {
  		...
  },
	...
}
```

#### `GET /api/raw_stats?orchestrator=<orchAddr>&region=<Airport_code>&since=<timestamp>`

If no parameter for `orchestrator` is provided the request will return `400 Bad Request` 

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
        "avg_seg_duration": number,
        "avg_upload_time": number,
        "avg_upload_score": number,
        "avg_download_time": number,
        "avg_download_score": number,
        "avg_transcode_time": number,
        "avg_transcode_score": number,
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
	AvgSegDuration    int64              `json:"avg_seg_duration" bson:"avg_seg_duration"`
	AvgUploadTime     int64              `json:"avg_upload_time" bson:"avg_upload_time"`
	AvgUploadScore    float64            `json:"avg_upload_score" bson:"avg_upload_score"`
	AvgDownloadTime   int64              `json:"avg_download_time" bson:"avg_download_time"`
	AvgDownloadScore  float64            `json:"avg_download_score" bson:"avg_download_score"`
	AvgTranscodeTime  int64              `json:"avg_transcode_time" bson:"avg_transcode_time"`
	AvgTranscodeScore float64            `json:"avg_transcode_score" bson:"avg_transcode_score"`
	RoundTripScore    float64            `json:"round_trip_score" bson:"round_trip_score"`
	Errors            []string           `json:"errors" bson:"errors"`
	Timestamp         int64              `json:"timestamp" bson:"timestamp"`
}
```

Region must be one of `"FRA", "MDW", "SIN"`

## Deployment Using Vercel

These serverless functions are deployed using Vercel.

The function signatures for the serverless handlers follow the standard `net/http` interface rather than the `aws-lambda-go` interface. 

For more information, check out https://vercel.com/docs/runtimes#official-runtimes/go .