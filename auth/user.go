package auth

import (
	"encoding/json"
	"log"
	"net/http"
	"WebRTCConf/database"
)

func GetUser(w http.ResponseWriter, r *http.Request) {
	ok, id := CheckHandler(r)
	if !ok || id == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	user, err := database.GetUser(ctx, client, id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	js, err := json.Marshal(user)
	if err != nil {
		log.Printf("Marshalling Error in GetUser: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func DeleteUser(w http.ResponseWriter, r *http.Request) {
	ok, id := CheckHandler(r)
	if !ok || id == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	err := LogoutHandler(w, r)
	if err != nil {
		log.Printf("Error in LogoutHandler: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = database.DeleteUser(ctx, client, id)
	if err != nil {
		log.Printf("Error in DeleteUser: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func NewRoom(w http.ResponseWriter, r *http.Request) {
	ok, id := CheckHandler(r)
	if !ok || id == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	var room database.Room
	if err := json.NewDecoder(r.Body).Decode(&room); err != nil {
		log.Printf("Could not parse JSON response in NewRoom: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err := database.NewRoom(ctx, client, id, room)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func DeleteRoom(w http.ResponseWriter, r *http.Request) {
	ok, id := CheckHandler(r)
	if !ok || id == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	var room database.Room
	if err := json.NewDecoder(r.Body).Decode(&room); err != nil {
		log.Printf("Could not parse JSON response in DeleteRoom: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err := database.DeleteRoom(ctx, client, id, room.RoomId)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
}

// CheckRoom checks whether room with roomId is present in the database or not
func CheckRoom(w http.ResponseWriter, r *http.Request) {
	var roomId string
	if err := json.NewDecoder(r.Body).Decode(&roomId); err != nil {
		log.Printf("Could not parse JSON response in CheckRoom: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	doc, err := database.CheckRoom(ctx, client, roomId)
	if err != nil {
		log.Printf("Error in checkRoom handler: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if doc != nil {
		w.WriteHeader(http.StatusOK)
		return
	}
	w.WriteHeader(http.StatusNotFound)
	return
}

func ToggleRoomLock(w http.ResponseWriter, r *http.Request) {
	ok, id := CheckHandler(r)
	if !ok || id == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	var roomId string
	if err := json.NewDecoder(r.Body).Decode(&roomId); err != nil {
		log.Printf("Could not parse JSON response in ToggleRoomLock: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err := database.ToggleRoomLock(ctx, client, id, roomId)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
	return
}
