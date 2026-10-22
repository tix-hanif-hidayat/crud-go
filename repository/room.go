package repository

import (
	"errors"

	m "example.com/demo/model"
	"gorm.io/gorm"
)

type RoomStore interface {
	GetAllRooms(hotelID *int) ([]m.Room, error)
	GetRoomById(id int) (*m.Room, error)
	CreateRoom(room m.Room) (*m.Room, error)
	UpdateRoom(id int, room m.Room) (*m.Room, error)
	DeleteRoom(id int) (bool, error)
}

type RoomRepository struct {
	db *gorm.DB
}

var RoomRepo RoomStore

func NewRoomRepository(database *gorm.DB) *RoomRepository {
	return &RoomRepository{db: database}
}

func (r *RoomRepository) GetAllRooms(hotelID *int) ([]m.Room, error) {
	rooms := make([]m.Room, 0)
	query := r.db
	if hotelID != nil {
		query = query.Where("hotel_id = ?", *hotelID)
	}
	if err := query.Find(&rooms).Error; err != nil {
		return nil, err
	}
	return rooms, nil
}

func (r *RoomRepository) GetRoomById(id int) (*m.Room, error) {
	var room m.Room
	err := r.db.First(&room, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &room, nil
}

func (r *RoomRepository) CreateRoom(room m.Room) (*m.Room, error) {
	room.Id = 0
	if err := r.db.Create(&room).Error; err != nil {
		return nil, err
	}
	return &room, nil
}

func (r *RoomRepository) UpdateRoom(id int, room m.Room) (*m.Room, error) {
	result := r.db.Model(&m.Room{}).Where("id = ?", id).Updates(map[string]any{
		"hotel_id": room.HotelId,
		"name":     room.Name,
		"price":    room.Price,
	})
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}
	return r.GetRoomById(id)
}

func (r *RoomRepository) DeleteRoom(id int) (bool, error) {
	result := r.db.Delete(&m.Room{}, id)
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}
