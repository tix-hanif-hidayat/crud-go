package router

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"testing"
	"time"

	"example.com/demo/config"
	"example.com/demo/model"
	"example.com/demo/repository"
)

type testRoomStore struct{}

func (testRoomStore) GetAllRooms(*int) ([]model.Room, error) {
	return []model.Room{{Id: 1, HotelId: 1, Name: "Room 1", Price: 250}}, nil
}

func (testRoomStore) GetRoomById(id int) (*model.Room, error) {
	if id != 1 {
		return nil, nil
	}
	return &model.Room{Id: 1, HotelId: 1, Name: "Room 1", Price: 250}, nil
}

func (testRoomStore) CreateRoom(model.Room) (*model.Room, error) {
	panic("not used in this test")
}

func (testRoomStore) UpdateRoom(int, model.Room) (*model.Room, error) {
	panic("not used in this test")
}

func (testRoomStore) DeleteRoom(int) (bool, error) {
	panic("not used in this test")
}

func TestConcurrentRoomPriceRequests(t *testing.T) {
	originalExchangeRateRepo := repository.ExchangeRateRepo
	originalRoomRepo := repository.RoomRepo
	originalCurrencyExpiration := config.CURRENY_EXPIRATION
	repository.ExchangeRateRepo = repository.NewExchangeRateRepository()
	repository.RoomRepo = testRoomStore{}
	config.CURRENY_EXPIRATION = time.Hour
	defer func() {
		repository.ExchangeRateRepo = originalExchangeRateRepo
		repository.RoomRepo = originalRoomRepo
		config.CURRENY_EXPIRATION = originalCurrencyExpiration
	}()

	handler := MainRouter()
	const requestCount = 10_000
	start := make(chan struct{})
	errors := make(chan string, requestCount)

	var wg sync.WaitGroup
	wg.Add(requestCount)
	for i := 0; i < requestCount; i++ {
		go func() {
			defer wg.Done()
			<-start

			request := httptest.NewRequest(http.MethodGet, "/api/rooms/1/price?currency=USD", nil)
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)

			if response.Code != http.StatusOK {
				errors <- "unexpected status: " + strconv.Itoa(response.Code)
				return
			}
			if got, want := response.Body.String(), "330\n"; got != want {
				errors <- "unexpected price response: " + got
			}
		}()
	}

	close(start)
	wg.Wait()
	close(errors)

	for err := range errors {
		t.Error(err)
	}
}
