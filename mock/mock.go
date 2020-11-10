package main

import (
	"log"
	"math/rand"
	"time"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/livepeer/leaderboard-serverless/db"
	"github.com/livepeer/leaderboard-serverless/models"
)

const addressSize = 20

func main() {
	store, err := db.Start()
	if err != nil {
		log.Fatal(err)
	}

	orchestrators := []string{}

	for i := 0; i < 10; i++ {
		orchestrators = append(orchestrators, randAddress())
	}

	for _, o := range orchestrators {
		for i := 0; i < 15; i++ {

			stats := &models.Stats{
				Region:            models.Regions[i%2],
				Orchestrator:      o,
				SegmentsSent:      rand.Int(),
				SegmentsReceived:  rand.Int(),
				SuccessRate:       rand.Float64() * 100,
				AvgSegDuration:    rand.Float64(),
				AvgUploadTime:     rand.Float64(),
				AvgUploadScore:    rand.Float64() * 10,
				AvgDownloadTime:   rand.Float64(),
				AvgDownloadScore:  rand.Float64() * 10,
				AvgTranscodeTime:  rand.Float64(),
				AvgTranscodeScore: rand.Float64() * 10,
				RoundTripScore:    rand.Float64() * 10,
				Timestamp:         time.Now().Unix(),
			}

			if err := store.InsertStats(stats); err != nil {
				log.Fatal(err)
			}
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
