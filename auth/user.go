package auth

import (
	"WebRTCConf/database"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
)

func GetUser(w http.ResponseWriter, r *http.Request) {
	ok, id := CheckHandler(r)
	if !ok || id == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	user, err := database.GetUser(Ctx, Client, id)
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
	err = database.DeleteUser(Ctx, Client, id)
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
	err := database.NewRoom(Ctx, Client, id, room)
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
	err := database.DeleteRoom(Ctx, Client, id, room.RoomId)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
}

// CheckRoom checks whether room with roomId is present in the database or not
func CheckRoom(w http.ResponseWriter, r *http.Request) {
	var room database.Room
	if err := json.NewDecoder(r.Body).Decode(&room); err != nil {
		log.Printf("Could not parse JSON response in CheckRoom: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	doc, err := database.CheckRoom(Ctx, Client, room.RoomId)
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
	err := database.ToggleRoomLock(Ctx, Client, id, roomId)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
	return
}

func IceServer(w http.ResponseWriter, r *http.Request) {
	var iceResponse IceResponse
	var sturnServers IceServers
	sturnServers.Urls = append(sturnServers.Urls, "stun:ss-turn1.xirsys.com")
	sturnServers.Urls = append(sturnServers.Urls, "stun:stun.l.google.com:19302")
	iceResponse.Ice = append(iceResponse.Ice, sturnServers)
	httpClient := http.Client{}
	marshalled, err := json.Marshal(XirsysPayload{
		Format: "urls",
	})
	if err != nil {
		log.Printf("Marshalling error: %v", err)
	}
	for i, iceUrl := range Env.IceURLs {
		// log.Printf("IceUrl: %v", iceUrl)

		req, err := http.NewRequest(http.MethodPut, iceUrl, bytes.NewReader(marshalled))
		if err != nil {
			log.Printf("IceServer: Error in Ice Server NewRequest: %v", err)
			continue
		}
		// log.Printf("IceToken: %v", Env.IceTokens[i])
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Basic "+base64.URLEncoding.EncodeToString([]byte(Env.IceTokens[i])))
		res, err := httpClient.Do(req)
		if err != nil {
			log.Printf("IceServer: Error in Ice Server Do: %v", err)
			continue
		}
		// log.Printf("Response body: %v", res.Body)
		var xi XirsysResponse
		if err := json.NewDecoder(res.Body).Decode(&xi); err != nil {
			log.Printf("could not parse JSON response: %v", err)
			continue
		}
		if xi.S != "ok" {
			log.Printf("Error in Credentials response: %v", err)
			continue
		}
		// log.Printf("XiResponse V: %v %v", reflect.TypeOf(xi.V), xi.S)
		// log.Printf("XiIceObject V: %v", xi.V.IceObject)
		iceResponse.Ice = append(iceResponse.Ice, xi.V.IceObject)
	}
	marshalled, err = json.Marshal(iceResponse)
	if err != nil {
		log.Printf("IceServer: Error in marshalling: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
	} else if len(iceResponse.Ice) == 0 {
		w.WriteHeader(http.StatusNotFound)
	} else {
		w.Write(marshalled)
	}
}
