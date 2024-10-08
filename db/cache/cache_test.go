package cache

import (
	"os"
	"sync"
	"testing"
	"time"

	"github.com/livepeer/leaderboard-serverless/common"
	"github.com/livepeer/leaderboard-serverless/models"
)

func TestMain(m *testing.M) {

	//check if LOG_LEVEL is set.  If not, set it to DEBUG.
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "debug"
	}
	common.Logger.SetLevel(logLevel)

	setCacheTimeouts("1")
	defer setCacheTimeouts("60")

	// Run the tests
	code := m.Run()

	// Exit with the appropriate code
	os.Exit(code)
}

func setCacheTimeouts(length string) {
	os.Setenv("REGIONS_CACHE_TIMEOUT", length)
	os.Setenv("PIPELINES_CACHE_TIMEOUT", length)
}

func TestInvalidateRegionsCache(t *testing.T) {

	cache := NewCache()
	cache.UpdateRegions([]*models.Region{{Name: "us-east-1"}})

	cache.InvalidateRegionsCache()

	cacheResult := cache.GetRegions()

	if cacheResult.Results != nil && len(cacheResult.Results.([]*models.Region)) > 0 {
		t.Errorf("expected nil regions, got %v", cacheResult.Results)
	}
	if !cacheResult.LastUpdate.IsZero() {
		t.Errorf("expected zero LastUpdate, got %v", cacheResult.LastUpdate)
	}
}

func TestGetRegionsCacheHit(t *testing.T) {
	cache := NewCache()
	testNewRegion := &models.Region{
		Name:        "NPL",
		DisplayName: "Northpole",
		Type:        models.Transcoding.String(),
	}
	cache.UpdateRegions([]*models.Region{testNewRegion})

	cacheResult := cache.GetRegions()
	if cacheResult.Results == nil {
		t.Errorf("expected non-nil regions, got nil")
	}
	if !cacheResult.CacheHit {
		t.Errorf("expected cache hit to be true, got false")
	}
	if cacheResult.CacheExpired {
		t.Errorf("expected cache not expired, but got expired")
	}

	//make sure the cache result has all the regions fields
	//and is an exact copy of the original
	if cacheResult.Results.([]*models.Region)[0].Name != testNewRegion.Name {
		t.Errorf("expected region name to be 'us-east-1', got '%s'", cacheResult.Results.([]*models.Region)[0].Name)
	}
	if cacheResult.Results.([]*models.Region)[0].DisplayName != testNewRegion.DisplayName {
		t.Errorf("expected region display name to be 'US East (N. Virginia)', got '%s'", cacheResult.Results.([]*models.Region)[0].DisplayName)
	}
	if cacheResult.Results.([]*models.Region)[0].Type != testNewRegion.Type {
		t.Errorf("expected region type to be 'ai', got '%s'", cacheResult.Results.([]*models.Region)[0].Type)
	}

}

func TestGetRegionsCacheExpired(t *testing.T) {
	cache := NewCache()
	cache.UpdateRegions([]*models.Region{{Name: "us-east-1"}})

	time.Sleep(cache.regionsCacheTimeout + time.Second)

	cacheResult := cache.GetRegions()
	if cacheResult.Results == nil {
		t.Errorf("expected non-nil regions, got nil")
	}
	if !cacheResult.CacheHit {
		t.Errorf("expected cache hit to be false, got true")
	}
	if !cacheResult.CacheExpired {
		t.Errorf("expected cache to be expired, got not expired")
	}
}

func TestUpdateRegions(t *testing.T) {
	cache := NewCache()
	newRegions := []*models.Region{{Name: "us-east-1"}}
	cache.UpdateRegions(newRegions)

	cacheResult := cache.GetRegions()
	if cacheResult.Results == nil {
		t.Errorf("expected non-nil regions, got nil")
	}
	if len(cacheResult.Results.([]*models.Region)) != len(newRegions) {
		t.Errorf("expected %d regions, got %d", len(newRegions), len(cacheResult.Results.([]*models.Region)))
	}
}

func TestInvalidatePipelinesCache(t *testing.T) {
	cache := NewCache()
	cache.UpdatePipelines([]*models.Pipeline{{Name: "test-pipeline"}})

	cache.InvalidatePipelinesCache()

	cacheResult := cache.GetPipelines()

	if cacheResult.Results != nil && len(cacheResult.Results.([]*models.Pipeline)) > 0 {
		t.Errorf("expected nil pipelines, got %v", cacheResult.Results)
	}
	if !cacheResult.LastUpdate.IsZero() {
		t.Errorf("expected zero LastUpdate, got %v", cacheResult.LastUpdate)
	}
}

func TestGetPipelinesCacheHit(t *testing.T) {
	cache := NewCache()
	testPipeline := &models.Pipeline{
		Name:   "test-pipeline",
		Models: []string{"test-model"},
		Regions: []string{
			"test-region",
		},
	}
	cache.UpdatePipelines([]*models.Pipeline{testPipeline})

	cacheResult := cache.GetPipelines()
	if cacheResult.Results == nil {
		t.Errorf("expected non-nil pipelines, got nil")
	}
	if !cacheResult.CacheHit {
		t.Errorf("expected cache hit to be true, got false")
	}
	if cacheResult.CacheExpired {
		t.Errorf("expected cache not expired, but got expired")
	}

	//make sure the cache result has all the pipelines fields
	//and is an exact copy of the original
	if cacheResult.Results.([]*models.Pipeline)[0].Name != testPipeline.Name {
		t.Errorf("expected pipeline name to be 'test-pipeline', got '%s'", cacheResult.Results.([]*models.Pipeline)[0].Name)
	}
	if cacheResult.Results.([]*models.Pipeline)[0].Models[0] != testPipeline.Models[0] {
		t.Errorf("expected pipeline model to be 'test-model', got '%s'", cacheResult.Results.([]*models.Pipeline)[0].Models[0])
	}
	if cacheResult.Results.([]*models.Pipeline)[0].Regions[0] != testPipeline.Regions[0] {
		t.Errorf("expected pipeline region to be 'test-region', got '%s'", cacheResult.Results.([]*models.Pipeline)[0].Regions[0])
	}
}

func TestGetPipelinesCacheExpired(t *testing.T) {
	cache := NewCache()
	cache.UpdatePipelines([]*models.Pipeline{{Name: "test-pipeline"}})

	time.Sleep(cache.pipelinesCacheTimeout + time.Second)

	cacheResult := cache.GetPipelines()
	if cacheResult.Results == nil {
		t.Errorf("expected non-nil pipelines, got nil")
	}
	if !cacheResult.CacheHit {
		t.Errorf("expected cache hit to be false, got true")
	}
	if !cacheResult.CacheExpired {
		t.Errorf("expected cache to be expired, got not expired")
	}
}

func TestUpdatePipelines(t *testing.T) {
	cache := NewCache()
	newPipelines := []*models.Pipeline{{Name: "test-pipeline"}}
	cache.UpdatePipelines(newPipelines)

	cacheResult := cache.GetPipelines()
	if cacheResult.Results == nil {
		t.Errorf("expected non-nil pipelines, got nil")
	}
	if len(cacheResult.Results.([]*models.Pipeline)) != len(newPipelines) {
		t.Errorf("expected %d pipelines, got %d", len(newPipelines), len(cacheResult.Results.([]*models.Pipeline)))
	}
}

func TestCacheEvictionAndUpdateAfterExpiration(t *testing.T) {
	cache := NewCache()
	cache.UpdateRegions([]*models.Region{{Name: "us-east-1"}})

	time.Sleep(cache.regionsCacheTimeout + time.Second)

	cacheResult := cache.GetRegions()
	if cacheResult.Results == nil {
		t.Errorf("expected non-nil regions, got nil")
	}
	if !cacheResult.CacheExpired {
		t.Errorf("expected cache to be expired, got not expired")
	}

	newRegions := []*models.Region{{Name: "us-west-2"}}
	cache.UpdateRegions(newRegions)

	cacheResult = cache.GetRegions()
	if cacheResult.CacheExpired {
		t.Errorf("expected cache not to be expired after update, got expired")
	}
	if len(cacheResult.Results.([]*models.Region)) != len(newRegions) {
		t.Errorf("expected %d regions, got %d", len(newRegions), len(cacheResult.Results.([]*models.Region)))
	}
	if cacheResult.Results.([]*models.Region)[0].Name != "us-west-2" {
		t.Errorf("expected region name to be 'us-west-2', got '%s'", cacheResult.Results.([]*models.Region)[0].Name)
	}
}

func TestCacheRetentionOfExpiredData(t *testing.T) {
	cache := NewCache()
	cache.UpdatePipelines([]*models.Pipeline{{Name: "initial-pipeline"}, {Name: "second-pipeline"}})

	time.Sleep(cache.pipelinesCacheTimeout + time.Second)

	cacheResult := cache.GetPipelines()
	if cacheResult.Results == nil {
		t.Errorf("expected non-nil pipelines, got nil")
	}
	if !cacheResult.CacheExpired {
		t.Errorf("expected cache to be expired, got not expired")
	}
	if len(cacheResult.Results.([]*models.Pipeline)) != 2 {
		t.Errorf("expected 2 pipelines, got %d", len(cacheResult.Results.([]*models.Pipeline)))
	}
	if cacheResult.Results.([]*models.Pipeline)[0].Name != "initial-pipeline" {
		t.Errorf("expected pipeline name to be 'initial-pipeline', got '%s'", cacheResult.Results.([]*models.Pipeline)[0].Name)
	}
}

func TestConcurrentAccess(t *testing.T) {

	setCacheTimeouts("60")

	cache := NewCache()

	//prime the cache with data
	regions := []*models.Region{{Name: "region-prime"}}
	cache.UpdateRegions(regions)

	var wg sync.WaitGroup
	concurrentGoroutines := 10

	// Updating regions concurrently
	for i := 0; i < concurrentGoroutines; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			common.Logger.Debug("Updating regions by goroutine %d", i)
			regions := []*models.Region{{Name: "region-%d"}}
			cache.UpdateRegions(regions)
			common.Logger.Debug("Done updating regions by goroutine %d", i)
		}(i)
	}

	// Reading regions concurrently
	for i := 0; i < concurrentGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cacheResult := cache.GetRegions()
			common.Logger.Debug("Reading regions by goroutine %d", i)
			if cacheResult.Results == nil {
				t.Errorf("expected non-nil regions, got nil")
			}
			if cacheResult.CacheExpired {
				t.Errorf("expected cache not expired during concurrent access, but got expired")
			}
			common.Logger.Debug("Done reading regions by goroutine %d", i)
		}()
	}

	wg.Wait()
}

func TestConcurrentAccessWithInvalidation(t *testing.T) {
	cache := NewCache()

	//prime the cache with data
	regions := []*models.Region{{Name: "region-prime"}}
	cache.UpdateRegions(regions)

	var wg sync.WaitGroup
	concurrentGoroutines := 10
	expiredCacheHits := 0
	totalCacheHits := 0

	// Updating regions concurrently
	for i := 0; i < concurrentGoroutines; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			regions := []*models.Region{{Name: "region-%d"}}
			cache.UpdateRegions(regions)
		}(i)
	}

	// Concurrent invalidation and reading
	for i := 0; i < concurrentGoroutines; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			if i%2 == 0 {
				cache.InvalidateRegionsCache()
			} else {
				cacheResult := cache.GetRegions()
				if !cacheResult.CacheExpired && cacheResult.Results == nil {
					t.Errorf("expected non-nil regions if cache is not expired, got nil")
				}
				if cacheResult.CacheExpired && cacheResult.Results != nil {
					t.Errorf("expected nil regions if cache is expired, got %v", cacheResult.Results)
				}
			}

			cacheResult := cache.GetRegions()
			if cacheResult.CacheExpired {
				expiredCacheHits++
			}
			if cacheResult.CacheHit {
				totalCacheHits++
			}
		}(i)
	}

	wg.Wait()

	//make sure we got some cache invalidations
	common.Logger.Debug("Expired cache hits: %d", expiredCacheHits)
	if expiredCacheHits == 0 {
		t.Errorf("expected some cache invalidations, got none")
	}

	//make sure we got hits on the cache for every lookup
	if totalCacheHits != concurrentGoroutines {
		t.Errorf("expected cache hits for every lookup, got %d", totalCacheHits)
	}

}
