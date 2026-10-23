package repository

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"example.com/demo/config"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func newTestExchangeRateRepository(t *testing.T, initialRate string) (*miniredis.Miniredis, *ExchangeRateRepository, CurrencyPair) {
	t.Helper()

	redisServer := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
	t.Cleanup(func() {
		if err := redisClient.Close(); err != nil {
			t.Errorf("close Redis client: %v", err)
		}
	})

	pair := CurrencyPair{Base: "IDR", Quote: "USD"}
	if err := redisServer.Set(pair.cacheKey(), initialRate); err != nil {
		t.Fatalf("set exchange rate in Redis: %v", err)
	}

	return redisServer, NewExchangeRateRepository(redisClient), pair
}

func TestUpdateExchangeRate(t *testing.T) {
	_, repository, pair := newTestExchangeRateRepository(t, "0.000065")

	rate := repository.updateExchangeRate(pair)
	if rate == nil {
		t.Fatal("updateExchangeRate() returned nil")
	}
	if rate.Base != pair.Base {
		t.Errorf("Base = %q, want %q", rate.Base, pair.Base)
	}
	if rate.Quote != pair.Quote {
		t.Errorf("Quote = %q, want %q", rate.Quote, pair.Quote)
	}
	if rate.Rate != 0.000065 {
		t.Errorf("Rate = %v, want %v", rate.Rate, 0.000065)
	}
	if rate.LastUpdated.IsZero() {
		t.Error("LastUpdated is zero")
	}
}

func TestGetExchangeRate(t *testing.T) {
	redisServer, repository, pair := newTestExchangeRateRepository(t, "0.000065")

	rate := repository.getExchangeRate(pair)
	if rate == nil {
		t.Fatal("getExchangeRate() returned nil")
	}
	if rate.Rate != 0.000065 {
		t.Errorf("Rate = %v, want %v", rate.Rate, 0.000065)
	}

	if err := redisServer.Set(pair.cacheKey(), "0.000070"); err != nil {
		t.Fatalf("update exchange rate in Redis: %v", err)
	}

	key := pair.cacheKey()
	repository.mu.Lock()
	cachedRate := repository.rates[key]
	cachedRate.LastUpdated = time.Now().Add(-config.LOCAL_EXCHANGE_RATE_EXPIRATION)
	repository.rates[key] = cachedRate
	repository.mu.Unlock()

	refreshedRate := repository.getExchangeRate(pair)
	if refreshedRate == nil {
		t.Fatal("getExchangeRate() returned nil for expired cached rate")
	}
	if refreshedRate.Rate != 0.000070 {
		t.Errorf("refreshed Rate = %v, want %v", refreshedRate.Rate, 0.000070)
	}
}

func TestGetExchangeRateConcurrent(t *testing.T) {
	const goroutineCount = 10_000

	_, repository, pair := newTestExchangeRateRepository(t, "0.000065")
	start := make(chan struct{})
	errs := make(chan error, goroutineCount)
	var workers sync.WaitGroup
	workers.Add(goroutineCount)

	for i := 0; i < goroutineCount; i++ {
		go func() {
			defer workers.Done()
			<-start

			rate := repository.getExchangeRate(pair)
			if rate == nil {
				errs <- fmt.Errorf("getExchangeRate() returned nil")
				return
			}
			if rate.Rate != 0.000065 {
				errs <- fmt.Errorf("getExchangeRate() returned rate %v, want %v", rate.Rate, 0.000065)
			}
		}()
	}

	close(start)
	workers.Wait()
	close(errs)

	for err := range errs {
		t.Error(err)
	}
}
