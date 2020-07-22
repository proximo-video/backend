package auth

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

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
		fmt.Fprintf(os.Stdout, "could not parse JSON response: %v", err)
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
		fmt.Fprintf(os.Stdout, "could not parse JSON response: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err := database.DeleteRoom(ctx, client, id, room.RoomId)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
}

func CheckRoom(w http.ResponseWriter, r *http.Request) {
	ok, id := CheckHandler(r)
	if !ok || id == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	var roomId string
	if err := json.NewDecoder(r.Body).Decode(&roomId); err != nil {
		fmt.Fprintf(os.Stdout, "could not parse JSON response: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err := database.CheckRoom(ctx, client, roomId)
	if err.Error() ==  "Room already present: " + roomId {
		w.WriteHeader(http.StatusOK)
		return
	}
	w.WriteHeader(http.StatusNotFound)
	return
}
