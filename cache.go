package sample1

import (
	"fmt"
	"sync"
	"time"
)

// PriceService is a service that we can use to get prices for the items
// Calls to this service are expensive (they take time)
type PriceService interface {
	GetPriceFor(itemCode string) (float64, error)
}

// TransparentCache is a cache that wraps the actual service
// The cache will remember prices we ask for, so that we don't have to wait on every call
// Cache should only return a price if it is not older than "maxAge", so that we don't get stale prices
type TransparentCache struct {
	actualPriceService PriceService
	maxAge             time.Duration
	prices             map[string]CachedPrice
	mutex              sync.RWMutex
}

type CachedPrice struct {
	price   float64
	savedAt time.Time
}

func NewTransparentCache(actualPriceService PriceService, maxAge time.Duration) *TransparentCache {
	return &TransparentCache{
		actualPriceService: actualPriceService,
		maxAge:             maxAge,
		prices:             make(map[string]CachedPrice),
		mutex:              sync.RWMutex{},
	}
}

func (c *TransparentCache) readFromCache(itemCode string) (CachedPrice, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	cachedPrice, ok := c.prices[itemCode]
	return cachedPrice, ok
}

func (c *TransparentCache) writeToCache(itemCode string, cachedPrice CachedPrice) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.prices[itemCode] = cachedPrice
}

func (c *TransparentCache) isCachedPriceExpired(cachedPrice CachedPrice) bool {
	return cachedPrice.savedAt.Add(c.maxAge).Before(time.Now())
}

// GetPriceFor gets the price for the item, either from the cache or the actual service if it was not cached or too old
func (c *TransparentCache) GetPriceFor(itemCode string) (float64, error) {
	cachedPrice, ok := c.readFromCache(itemCode)
	if ok && !c.isCachedPriceExpired(cachedPrice) {
		return cachedPrice.price, nil
	}
	price, err := c.actualPriceService.GetPriceFor(itemCode)
	if err != nil {
		return 0, fmt.Errorf("getting price from service : %v", err.Error())
	}
	newCachedPrice := CachedPrice{
		price:   price,
		savedAt: time.Now(),
	}
	c.writeToCache(itemCode, newCachedPrice)
	return price, nil
}

// GetPricesFor gets the prices for several items at once, some might be found in the cache, others might not
// If any of the operations returns an error, it should return an error as well
func (c *TransparentCache) GetPricesFor(itemCodes ...string) ([]float64, error) {
	results := []float64{}

	for _, itemCode := range itemCodes {
		// TODO: parallelize this, it can be optimized to not make the calls to the external service sequentially
		price, err := c.GetPriceFor(itemCode)
		if err != nil {
			return []float64{}, err
		}
		results = append(results, price)
	}
	return results, nil
}
