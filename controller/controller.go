package controller

import (
	"encoding/json"
	"net/http"
	"strconv"

	"example.com/demo/model"
	"example.com/demo/repository"
	"example.com/demo/utils"
)

func ListRooms(w http.ResponseWriter, r *http.Request) {
	rooms, err := repository.RoomRepo.GetAllRooms(nil)
	if err != nil {
		http.Error(w, "failed to list rooms", http.StatusInternalServerError)
		return
	}
	utils.WriteJSON(w, http.StatusOK, rooms)
}

func ListHotelRooms(w http.ResponseWriter, r *http.Request) {
	id, ok := utils.ParsePathID(w, r, "hotel_id")
	if !ok {
		return
	}
	rooms, err := repository.RoomRepo.GetAllRooms(&id)
	if err != nil {
		http.Error(w, "failed to list rooms", http.StatusInternalServerError)
		return
	}
	utils.WriteJSON(w, http.StatusOK, rooms)
}

func ShowRoom(w http.ResponseWriter, r *http.Request) {
	id, ok := utils.ParsePathID(w, r, "room_id")
	if !ok {
		return
	}
	room, err := repository.RoomRepo.GetRoomById(id)
	if err != nil {
		http.Error(w, "failed to get room", http.StatusInternalServerError)
		return
	}
	if room == nil {
		http.NotFound(w, r)
		return
	}
	utils.WriteJSON(w, http.StatusOK, room)
}

func AddRoom(w http.ResponseWriter, r *http.Request) {
	var room model.Room
	if err := json.NewDecoder(r.Body).Decode(&room); err != nil {
		http.Error(w, "invalid room JSON", http.StatusBadRequest)
		return
	}

	created, err := repository.RoomRepo.CreateRoom(room)
	if err != nil {
		http.Error(w, "failed to create room", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Location", "/rooms/"+strconv.Itoa(created.Id))
	utils.WriteJSON(w, http.StatusCreated, created)
}

func ReplaceRoom(w http.ResponseWriter, r *http.Request) {
	id, ok := utils.ParsePathID(w, r, "room_id")
	if !ok {
		return
	}
	var room model.Room
	if err := json.NewDecoder(r.Body).Decode(&room); err != nil {
		http.Error(w, "invalid room JSON", http.StatusBadRequest)
		return
	}

	updated, err := repository.RoomRepo.UpdateRoom(id, room)
	if err != nil {
		http.Error(w, "failed to update room", http.StatusInternalServerError)
		return
	}
	if updated == nil {
		http.NotFound(w, r)
		return
	}
	utils.WriteJSON(w, http.StatusOK, updated)
}

func RemoveRoom(w http.ResponseWriter, r *http.Request) {
	id, ok := utils.ParsePathID(w, r, "room_id")
	if !ok {
		return
	}
	deleted, err := repository.RoomRepo.DeleteRoom(id)
	if err != nil {
		http.Error(w, "failed to delete room", http.StatusInternalServerError)
		return
	}
	if !deleted {
		http.NotFound(w, r)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func GetRoomPrice(w http.ResponseWriter, r *http.Request) {
	id, ok := utils.ParsePathID(w, r, "room_id")
	if !ok {
		return
	}
	room, err := repository.RoomRepo.GetRoomById(id)
	if err != nil {
		http.Error(w, "failed to get room", http.StatusInternalServerError)
		return
	}
	if room == nil {
		http.NotFound(w, r)
		return
	}

	currency := r.URL.Query().Get("currency")
	if currency == "" {
		utils.WriteJSON(w, http.StatusOK, room.Price)
		return
	}

	price, err := repository.ExchangeRateRepo.ConvertPrice(room.Price, "IDR", currency)
	if err != nil {
		http.Error(w, "failed to convert currency", http.StatusInternalServerError)
		return
	}
	utils.WriteJSON(w, http.StatusOK, price)
}
