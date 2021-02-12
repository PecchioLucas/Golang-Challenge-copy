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
	actualPriceService  PriceService
	maxAge              time.Duration
	prices              map[string]CachedPrice
	mutex               sync.RWMutex
	maxParallelRoutines int
}

// CachedPrice is the struct thats saved in a cache and contains the price
type CachedPrice struct {
	price   float64
	savedAt time.Time
}

// NewTransparentCache is the constructor for TransparentCache
func NewTransparentCache(actualPriceService PriceService, maxAge time.Duration) *TransparentCache {
	return &TransparentCache{
		actualPriceService:  actualPriceService,
		maxAge:              maxAge,
		prices:              make(map[string]CachedPrice),
		mutex:               sync.RWMutex{},
		maxParallelRoutines: 4,
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
	type priceResult struct {
		itemCode string
		price    float64
		err      error
	}

	results := []float64{}

	resultChannel := make(chan priceResult, len(itemCodes))
	semaphoreChannel := make(chan struct{}, c.maxParallelRoutines)

	waitGroup := sync.WaitGroup{}
	for _, itemCode := range itemCodes {
		waitGroup.Add(1)
		semaphoreChannel <- struct{}{}
		go func(itemCode string, resultsChannel chan priceResult, waitGroup *sync.WaitGroup, semaphoreChannel chan struct{}) {
			price, err := c.GetPriceFor(itemCode)
			resultsChannel <- priceResult{
				itemCode: itemCode,
				price:    price,
				err:      err,
			}
			waitGroup.Done()
			<-semaphoreChannel
		}(itemCode, resultChannel, &waitGroup, semaphoreChannel)
	}
	waitGroup.Wait()

	resultErrors := []error{}
	for i := 0; i < len(itemCodes); i++ {
		result := <-resultChannel
		if result.err != nil {
			resultErrors = append(resultErrors, result.err)
		}
		if len(resultErrors) == 0 {
			results = append(results, result.price)
		}
	}

	if len(resultErrors) > 0 {
		return []float64{}, fmt.Errorf("%d errors ocurred", len(resultErrors))
	}

	return results, nil
}
