package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	config "example.com/demo/config"
	m "example.com/demo/model"
	"github.com/redis/go-redis/v9"
)

type CurrencyPair struct {
	Base  string
	Quote string
}

type exchangeRateCache interface {
	Get(ctx context.Context, key string) *redis.StringCmd
	Set(ctx context.Context, key string, value any, expiration time.Duration) *redis.StatusCmd
}

func (p CurrencyPair) cacheKey() string {
	return fmt.Sprintf("fx:%s:%s", p.Base, p.Quote)
}

type ExchangeRateRepository struct {
	mu    sync.RWMutex
	rates map[string]m.ExchangeRateV2
	redis exchangeRateCache
}

var ExchangeRateRepo *ExchangeRateRepository

func NewExchangeRateRepository(redisClient exchangeRateCache) *ExchangeRateRepository {
	return &ExchangeRateRepository{
		rates: make(map[string]m.ExchangeRateV2),
		redis: redisClient,
	}
}

func (e *ExchangeRateRepository) getExchangeRate(pair CurrencyPair) *m.ExchangeRateV2 {
	exchangeKey := pair.cacheKey()

	e.mu.RLock()
	rate, exists := e.rates[exchangeKey]
	e.mu.RUnlock()

	if !exists || time.Since(rate.LastUpdated) > config.LOCAL_EXCHANGE_RATE_EXPIRATION {
		return e.updateExchangeRate(pair)
	}

	return &rate
}

func (e *ExchangeRateRepository) updateExchangeRate(pair CurrencyPair) *m.ExchangeRateV2 {
	exchangeKey := pair.cacheKey()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if value, err := e.redis.Get(ctx, exchangeKey).Result(); err == nil {
		rateValue, err := strconv.ParseFloat(value, 64)
		if err != nil {
			log.Printf("failed to parse exchange rate value for %s: %v", exchangeKey, err)
			return nil
		}

		newRate := m.ExchangeRateV2{
			Base:        pair.Base,
			Quote:       pair.Quote,
			Rate:        rateValue,
			LastUpdated: time.Now(),
		}

		e.mu.Lock()
		e.rates[exchangeKey] = newRate
		e.mu.Unlock()

		return &newRate
	}

	return nil
}

func (e *ExchangeRateRepository) ConvertPrice(price int, pair CurrencyPair) (int, error) {
	rate := e.getExchangeRate(pair)
	if rate == nil {
		return 0, fmt.Errorf("exchange rate %s/%s not found", pair.Base, pair.Quote)
	}

	convertedPrice := float64(price) * rate.Rate
	return int(convertedPrice), nil
}

func (e *ExchangeRateRepository) refreshExchangeRates() error {
	client := &http.Client{Timeout: 10 * time.Second}

	resp, err := client.Get(config.EXCHANGE_RATE_API_URL)
	if err != nil {
		return fmt.Errorf("fetch exchange rates: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("exchange rate API returned status: %s", resp.Status)
	}

	var rates []m.ExchangeRateV2
	if err := json.NewDecoder(resp.Body).Decode(&rates); err != nil {
		return fmt.Errorf("decode exchange rates: %w", err)
	}
	if e.redis == nil {
		return fmt.Errorf("redis client is not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for _, rate := range rates {
		pair := CurrencyPair{Base: rate.Base, Quote: rate.Quote}
		key := pair.cacheKey()
		value := strconv.FormatFloat(rate.Rate, 'f', -1, 64)
		expiration := 2 * config.REFRESH_EXCHANGE_RATE_INTERVAL

		if err := e.redis.Set(ctx, key, value, expiration).Err(); err != nil {
			return fmt.Errorf("cache exchange rate %s: %w", key, err)
		}
	}
	return nil
}

func (e *ExchangeRateRepository) StartExchangeRateUpdater() func() {
	ticker := time.NewTicker(config.REFRESH_EXCHANGE_RATE_INTERVAL)
	done := make(chan struct{})
	stopped := make(chan struct{})

	if err := e.refreshExchangeRates(); err != nil {
		log.Printf("exchange rate update failed: %v", err)
	}

	go func() {
		defer close(stopped)
		for {
			select {
			case <-done:
				return
			case tick := <-ticker.C:
				if err := e.refreshExchangeRates(); err != nil {
					log.Printf("exchange rate update failed: %v", err)
				} else {
					log.Printf("exchange rates updated at %s", tick.Format(time.RFC3339))
				}
			}
		}
	}()

	return func() {
		ticker.Stop()
		close(done)
		<-stopped
	}
}
