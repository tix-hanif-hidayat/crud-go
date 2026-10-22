package router

import (
	"net/http"

	"example.com/demo/controller"
	"github.com/gorilla/mux"
)

func createRoomRouter(r *mux.Router) {
	r.HandleFunc("/rooms", controller.ListRooms).Methods(http.MethodGet)
	r.HandleFunc("/rooms", controller.AddRoom).Methods(http.MethodPost)
	r.HandleFunc("/rooms/{room_id}", controller.ShowRoom).Methods(http.MethodGet)
	r.HandleFunc("/rooms/{room_id}", controller.ReplaceRoom).Methods(http.MethodPut)
	r.HandleFunc("/rooms/{room_id}", controller.RemoveRoom).Methods(http.MethodDelete)
	r.HandleFunc("/rooms/{room_id}/price", controller.GetRoomPrice).Methods(http.MethodGet)
	r.HandleFunc("/hotels/{hotel_id}/rooms", controller.ListHotelRooms).Methods(http.MethodGet)
}

func MainRouter() http.Handler {
	r := mux.NewRouter()

	api := r.PathPrefix("/api").Subrouter()
	createRoomRouter(api)
	return r
}
