package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"WebRTCConf/auth"

	"WebRTCConf/signaling"

	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
)

func init() {
	// loads values from .env into the system
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
	auth.Env.GithubClientID = os.Getenv("GITHUB_CLIENT_ID")
	auth.Env.GithubClientSecret = os.Getenv("GITHUB_CLIENT_SECRET")
	auth.Env.GoogleClientID = os.Getenv("GOOGLE_CLIENT_ID")
	auth.Env.GoogleClientSecret = os.Getenv("GOOGLE_CLIENT_SECRET")
	auth.Env.RedirectURI = os.Getenv("REDIRECT_URI")
	iceUrls := os.Getenv("ICE_URLS")
	auth.Env.IceURLS = strings.Split(iceUrls, "#")
	iceTokens := os.Getenv("ICE_TOKENS")
	auth.Env.IceTokens = strings.Split(iceTokens, "#")
	auth.Store = sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY")))
}

func main() {
	go signaling.RManager.HandleChannels()
	mux := http.NewServeMux()
	mux.HandleFunc("/getSession", auth.GetSession)
	mux.HandleFunc("/auth", auth.Auth)
	mux.HandleFunc("/getUser", auth.GetUser)
	mux.HandleFunc("/deleteUser", auth.DeleteUser)
	mux.HandleFunc("/newRoom", auth.NewRoom)
	mux.HandleFunc("/deleteRoom", auth.DeleteRoom)
	mux.HandleFunc("/checkRoom", auth.CheckRoom)
	mux.HandleFunc("/toggle", auth.ToggleRoomLock)
	mux.HandleFunc("/iceserver", auth.IceServer)
	mux.HandleFunc("/ws", signaling.WebSocketHandler)

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:8000", "https://localhost:8000"},
		AllowCredentials: true,
		// Enable Debugging for testing, consider disabling in production
		Debug: false,
	})
	handler := c.Handler(mux)

	log.Println("server started port " + os.Getenv("PORT"))
	http.ListenAndServe(":"+os.Getenv("PORT"), handler)

}
