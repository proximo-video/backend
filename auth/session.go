package auth

import (
	"fmt"
	"net/http"
	"os"
)

// Note: Don't store your key in your source code. Pass it via an
// environmental variable, or flag (or both), and don't accidentally commit it
// alongside your code. Ensure your key is sufficiently random - i.e. use Go's
// crypto/rand or securecookie.GenerateRandomKey(32) and persist the result.
// Ensure SESSION_KEY exists in the environment, or sessions will fail.

// MyHandler - Handle
func MyHandler(w http.ResponseWriter, r *http.Request, id string) {
	// Get a session. Get() always returns a session, even if empty.
	session, err := Store.Get(r, "session-name")
	if err != nil {
		fmt.Fprintf(os.Stdout, "%v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Set some session values.
	session.Values["authenticated"] = true
	session.Values["id"] = id
	// Save it before we write to the response/return from the handler.
	err = session.Save(r, w)
	if err != nil {
		fmt.Fprintf(os.Stdout, "%v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// CheckHandler - Check
func CheckHandler(r *http.Request) (bool, string) {
	session, err := Store.Get(r, "session-name")
	if err != nil {
		return false, ""
	}
	if session.Values["authenticated"] != nil && session.Values["authenticated"] != false {
		return true, session.Values["id"].(string)
	}
	return false, ""
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) error {
	session, err := Store.Get(r, "session-name")
	if err != nil {
		return err
	}
	session.Values["authenticated"] = false
	err = session.Save(r, w)
	if err != nil {
		fmt.Fprintf(os.Stdout, "%v", err)
		return err
	}
	return nil
}
