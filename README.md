# Leaderboard-Serverless

## API

### `GET /aggregated_stats?orchestrator=<orchAddr>&region=<Airport_code>&since=<timestamp>`

- If `orchestrator` is not provided the response will include aggregated scores for all orchestrators

- If `region is not provided all regions will be returned in the response. "GLOBAL" would be included as well which would be the average of the regions. 

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

### `GET /raw_stats?orchestrator=<orchAddr>&region=<Airport_code>&since=<timestamp>`

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

### POST `/post_stats

## Deployment Using Vercel

These serverless functions are deployed using Vercel.

The function signatures for the serverless handlers follow the standard `net/http` interface rather than the `aws-lambda-go` interface. 

For more information, check out https://vercel.com/docs/runtimes#official-runtimes/go .