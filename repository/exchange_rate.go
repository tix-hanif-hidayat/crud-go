package repository

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	config "example.com/demo/config"
	"example.com/demo/db"
	m "example.com/demo/model"
)

type ExchangeRateRepository struct {
	mu    sync.RWMutex
	rates []m.ExchangeRate
}

var ExchangeRateRepo = NewExchangeRateRepository()

func NewExchangeRateRepository() *ExchangeRateRepository {
	return &ExchangeRateRepository{
		rates: []m.ExchangeRate{
			{
				BaseCurrency:  "IDR",
				QuoteCurrency: "USD",
				Rate:          1.32,
				LastUpdated:   time.Now(),
			},
		},
	}
}

func (e *ExchangeRateRepository) getExchangeRate(baseCurrency, quoteCurrency string) *m.ExchangeRate {
	e.mu.RLock()
	defer e.mu.RUnlock()

	for _, rate := range e.rates {
		if rate.BaseCurrency == baseCurrency && rate.QuoteCurrency == quoteCurrency {
			return &rate
		}
	}
	return nil
}

func (e *ExchangeRateRepository) updateExchangeRate(baseCurrency, quoteCurrency string) *m.ExchangeRate {
	ctx := context.Background()
	e.mu.Lock()
	defer e.mu.Unlock()

	for i, rate := range e.rates {
		if rate.BaseCurrency == baseCurrency && rate.QuoteCurrency == quoteCurrency {
			value, err := db.Redis.Get(ctx, "fx:"+baseCurrency+":"+quoteCurrency).Result()
			if err != nil {
				return nil
			}

			e.rates[i].Rate, err = strconv.ParseFloat(value, 64)
			if err != nil {
				return nil
			}

			e.rates[i].LastUpdated = time.Now()
			updatedRate := e.rates[i]
			return &updatedRate
		}
	}
	return nil
}

func (e *ExchangeRateRepository) ConvertPrice(price int, baseCurrency, quoteCurrency string) (int, error) {
	rate := e.getExchangeRate(baseCurrency, quoteCurrency)
	if rate == nil {
		return 0, fmt.Errorf("exchange rate %s/%s not found", baseCurrency, quoteCurrency)
	}

	if time.Since(rate.LastUpdated) > config.CURRENY_EXPIRATION {
		updatedRate := e.updateExchangeRate(baseCurrency, quoteCurrency)
		if updatedRate == nil {
			return 0, fmt.Errorf("failed to update exchange rate %s/%s", baseCurrency, quoteCurrency)
		}
		rate = updatedRate
	}

	convertedPrice := float64(price) * rate.Rate
	return int(convertedPrice), nil
}
