package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"WebRTCConf/database"

	"github.com/gorilla/sessions"
)

var Ctx = context.Background()
var Env EnvVariables
var Store *sessions.CookieStore
var Client = database.CreateDatabaseClient(Ctx)

func GetSession(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, post-check=0, pre-check=0")
	if ok, _ := CheckHandler(r); !ok {
		w.WriteHeader(http.StatusUnauthorized)
	}
}

func Auth(w http.ResponseWriter, r *http.Request) {
	// First, we need to get the value of the `code` query param
	var code Token
	if err := json.NewDecoder(r.Body).Decode(&code); err != nil {
		fmt.Fprintf(os.Stdout, "could not parse JSON response: %v", err)
		w.WriteHeader(http.StatusBadRequest)
	}
	// Next, lets for the HTTP request to call the github oauth enpoint
	// to get our access token
	var user database.User
	if code.Service == "github" {
		user = authGithub(w, r, code.Code)
		MyHandler(w, r, user.Id)
	} else if code.Service == "google" {
		user = authGoogle(w, r, code.Code)
		MyHandler(w, r, user.Id)
	}
	err := database.CheckUser(Ctx, Client, user.Id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	database.SaveUser(Ctx, Client, user.Id, user)
}

func authGithub(w http.ResponseWriter, r *http.Request, code string) database.User {
	var user database.User
	httpClient := http.Client{}
	reqURL := fmt.Sprintf("https://github.com/login/oauth/access_token?client_id=%s&client_secret=%s&code=%s", Env.GithubClientID, Env.GithubClientSecret, code)
	req, err := http.NewRequest(http.MethodPost, reqURL, nil)
	if err != nil {
		fmt.Fprintf(os.Stdout, "could not create HTTP request: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return user
	}
	// We set this header since we want the response
	// as JSON
	req.Header.Set("accept", "application/json")

	// Send out the HTTP request
	res1, err := httpClient.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stdout, "could not send HTTP request: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return user
	}
	defer res1.Body.Close()

	// Parse the request body into the `OAuthAccessResponse` struct
	var t OAuthAccessResponse
	if err := json.NewDecoder(res1.Body).Decode(&t); err != nil {
		fmt.Fprintf(os.Stdout, "could not parse JSON response: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return user
	}
	if t.AccessToken == "" {
		w.WriteHeader(http.StatusBadRequest)
		return user
	}
	req, err = http.NewRequest(http.MethodGet, "https://api.github.com/user", nil)
	req.Header.Add("Authorization", "Token "+t.AccessToken)
	res2, err := httpClient.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stdout, "could not send HTTP request: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return user
	}
	defer res2.Body.Close()
	var ud GithubUserData
	if err := json.NewDecoder(res2.Body).Decode(&ud); err != nil {
		fmt.Fprintf(os.Stdout, "could not parse JSON response: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return user
	}
	user.Id = strconv.Itoa(ud.ID)
	user.Name = ud.Name
	return user
}

func authGoogle(w http.ResponseWriter, r *http.Request, code string) database.User {
	var user database.User
	httpClient := http.Client{}
	reqURL := "https://oauth2.googleapis.com/token"
	data := url.Values{}
	data.Set("code", code)
	data.Set("client_id", Env.GoogleClientID)
	data.Set("client_secret", Env.GoogleClientSecret)
	data.Set("grant_type", "authorization_code")
	data.Set("redirect_uri", Env.RedirectURI)
	req, err := http.NewRequest(http.MethodPost, reqURL, bytes.NewBufferString(data.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	if err != nil {
		fmt.Fprintf(os.Stdout, "could not create HTTP request: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return user
	}
	// We set this header since we want the response
	// as JSON
	req.Header.Set("accept", "application/json")

	// Send out the HTTP request
	res1, err := httpClient.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stdout, "could not send HTTP request: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return user
	}
	defer res1.Body.Close()
	// Parse the request body into the `OAuthAccessResponse` struct
	var t OAuthAccessResponse
	if err := json.NewDecoder(res1.Body).Decode(&t); err != nil {
		fmt.Fprintf(os.Stdout, "could not parse JSON response: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return user
	}
	if t.AccessToken == "" {
		w.WriteHeader(http.StatusBadRequest)
		return user
	}
	req, err = http.NewRequest(http.MethodGet, "https://people.googleapis.com/v1/people/me?personFields=emailAddresses,names,metadata", nil)
	req.Header.Add("Authorization", "Bearer "+t.AccessToken)
	res2, err := httpClient.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stdout, "could not send HTTP request: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return user
	}
	defer res2.Body.Close()
	var ud GoogleUserData
	if err := json.NewDecoder(res2.Body).Decode(&ud); err != nil {
		fmt.Fprintf(os.Stdout, "could not parse JSON response: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return user
	}
	user.Id = ud.MetaData.Sources[0].ID
	user.Name = ud.Names[0].DisplayName
	return user
}
