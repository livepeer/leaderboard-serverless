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

	orchestrators := []string{"0xda43d85b8d419a9c51bbf0089c9bd5169c23f2f9", "0xe0a4a877cd0a07da7c08dffebc2546a4713147f2", "0x9c10672cee058fd658103d90872fe431bb6c0afa", "0xa5e37e0ba14655e92deff29f32adbc7d09b8a2cf", "0x4ff088ac5422f994486663ff903b040692797168", "0xe9e284277648fcdb09b8efc1832c73c09b5ecf59", "0x9d2b4e5c4b1fd81d06b883b0aca661b771c39ea3", "0x525419ff5707190389bfb5c87c375d710f5fcb0e", "0x4f4758f7167b18e1f5b3c1a7575e3eb584894dbc", "0xa20416801ac2eacf2372e825b4a90ef52490c2bb"}

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
