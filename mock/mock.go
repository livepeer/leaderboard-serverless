package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"livepeer.org/leaderboard/models"
	"livepeer.org/leaderboard/mongo"
)

const addressSize = 20

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	if err := mongo.Start(ctx); err != nil {
		log.Fatal(err)
	}

	orchestrators := []string{}

	for i := 0; i < 10; i++ {
		orchestrators = append(orchestrators, randAddress())
	}

	for _, o := range orchestrators {
		for i := 0; i < 15; i++ {
			queryCtx, queryCancel := context.WithTimeout(context.Background(), 5*time.Second)

			stats := models.Stats{
				Region:            mongo.Regions[i%2],
				Orchestrator:      o,
				SegmentsSent:      rand.Int(),
				SegmentsReceived:  rand.Int(),
				SuccessRate:       rand.Float64() * 100,
				AvgSegDuration:    rand.Int63(),
				AvgUploadTime:     rand.Int63(),
				AvgUploadScore:    rand.Float64() * 10,
				AvgDownloadTime:   rand.Int63(),
				AvgDownloadScore:  rand.Float64() * 10,
				AvgTranscodeTime:  rand.Int63(),
				AvgTranscodeScore: rand.Float64() * 10,
				RoundTripScore:    rand.Float64() * 10,
				Timestamp:         time.Now().Unix(),
			}

			result, err := mongo.DB.Collection(stats.Region).InsertOne(queryCtx, stats)
			fmt.Println(result, err)
			queryCancel()
		}
	}

}

func randAddress() string {
	return ethcommon.BytesToAddress(randBytes(addressSize)).Hex()
}

func randBytes(size uint) []byte {
	x := make([]byte, size, size)
	for i := 0; i < len(x); i++ {
		x[i] = byte(rand.Uint32())
	}
	return x
}
