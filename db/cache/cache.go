package cache

import (
	"reflect"
	"sync"
	"time"

	"github.com/livepeer/leaderboard-serverless/common"
	"github.com/livepeer/leaderboard-serverless/models"
)

type Cache interface {
	InvalidateRegionsCache()
	GetRegions() CacheResult
	UpdateRegions(newRegions []*models.Region)
	InvalidatePipelinesCache()
	GetPipelines() CacheResult
	UpdatePipelines(newPipelines []*models.Pipeline)
}

type CacheResult struct {
	Results      interface{}
	LastUpdate   time.Time
	CacheHit     bool
	CacheExpired bool
}

type MemCache struct {
	mu                    sync.RWMutex
	regionsCacheTimeout   time.Duration
	regions               []*models.Region
	regionsLastUpdate     time.Time
	pipelinesCacheTimeout time.Duration
	pipelines             []*models.Pipeline
	pipelinesLastUpdate   time.Time
}

func NewCache() *MemCache {
	c := &MemCache{}
	c.regionsCacheTimeout = getCacheTimeout("REGIONS_CACHE_TIMEOUT", 60)
	c.pipelinesCacheTimeout = getCacheTimeout("PIPELINES_CACHE_TIMEOUT", 60)
	return c
}

func getCacheTimeout(envVar string, defaultValue int) time.Duration {
	timeout := common.EnvOrDefault(envVar, defaultValue).(int)
	common.Logger.Info("Cache timeout for %s is set to %d seconds", envVar, timeout)
	return time.Duration(timeout) * time.Second
}

func (c *MemCache) InvalidateRegionsCache() {
	common.Logger.Debug("Invalidating regions cache")
	c.invalidateCacheResult(c.regions)
}

func (c *MemCache) GetRegions() CacheResult {
	return c.getCacheResult(c.regions, c.regionsLastUpdate, c.regionsCacheTimeout)
}

func (c *MemCache) UpdateRegions(newRegions []*models.Region) {
	common.Logger.Debug("Updating regions cache")
	c.updateCache(&c.regions, &c.regionsLastUpdate, newRegions)
}

func (c *MemCache) InvalidatePipelinesCache() {
	common.Logger.Debug("Invalidating pipelines cache")
	c.invalidateCacheResult(c.pipelines)
}

func (c *MemCache) GetPipelines() CacheResult {
	return c.getCacheResult(c.pipelines, c.pipelinesLastUpdate, c.pipelinesCacheTimeout)
}

func (c *MemCache) UpdatePipelines(newPipelines []*models.Pipeline) {
	common.Logger.Debug("Updating pipelines cache")
	c.updateCache(&c.pipelines, &c.pipelinesLastUpdate, newPipelines)
}

/** Utility functions **/

func (c *MemCache) invalidateCacheResult(data interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	zeroTime := time.Time{}
	switch v := data.(type) {
	case []*models.Region:
		c.regions = nil
		c.regionsLastUpdate = zeroTime
		common.Logger.Debug("Invalidated regions cache")
	case []*models.Pipeline:
		c.pipelines = nil
		c.pipelinesLastUpdate = zeroTime
		common.Logger.Debug("Invalidated pipelines cache")
	default:
		common.Logger.Warn("unsupported cache type: %T", v)
	}
}

func (c *MemCache) getCacheResult(data interface{}, lastUpdate time.Time, timeout time.Duration) CacheResult {
	c.mu.RLock()
	defer c.mu.RUnlock()
	common.Logger.Debug("Getting cache result for data: %v, lastUpdate: %v, timeout: %v", data, lastUpdate, timeout)
	return CacheResult{
		// we give the caller a copy of the data in the cache so it can't modify the cache and is unaffected by invalidation
		Results:      c.shallowCopy(data),
		LastUpdate:   lastUpdate,
		CacheHit:     data != nil,
		CacheExpired: c.isCacheExpired(data, lastUpdate, timeout),
	}
}

func (c *MemCache) updateCache(data interface{}, lastUpdate *time.Time, newData interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	switch v := data.(type) {
	case *[]*models.Region:
		c.regions = newData.([]*models.Region)
	case *[]*models.Pipeline:
		c.pipelines = newData.([]*models.Pipeline)
	default:
		common.Logger.Warn("unsupported cache type: %T", v)
	}
	*lastUpdate = time.Now()
}

func (c *MemCache) isCacheExpired(data interface{}, timeLastUpdated time.Time, timeout time.Duration) bool {
	common.Logger.Debug("Checking if cache is expired for data: %v, timeLastUpdated: %v, timeout: %v", data, timeLastUpdated, timeout)
	//if we have no data, it was expired or was never set
	if data == nil {
		return true
	}
	return timeLastUpdated.IsZero() || time.Since(timeLastUpdated) > timeout
}

// shallowCopy returns a shallow copy of the original object
// if structs get more complex, we may need to implement a deep copy
func (c *MemCache) shallowCopy(original interface{}) interface{} {
	return reflect.ValueOf(original).Interface()
}
