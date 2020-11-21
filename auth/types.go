package auth
import "github.com/dgrijalva/jwt-go"


//Token - token
type Token struct {
	Code    string `json:"code"`
	Service string `json:"service"`
}

// OAuthAccessResponse - Response Token
type OAuthAccessResponse struct {
	AccessToken string `json:"access_token"`
}

// GithubUserData - Userdata from github
type GithubUserData struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

//GoogleSource -
type GoogleSource struct {
	ID string `json:"id"`
}

//GoogleMetadata -
type GoogleMetadata struct {
	Sources []GoogleSource `json:"sources"`
}

//GoogleName -
type GoogleName struct {
	DisplayName string `json:"displayName"`
}

// GoogleUserData - Userdata from google
type GoogleUserData struct {
	MetaData GoogleMetadata `json:"metadata"`
	Names    []GoogleName   `json:"names"`
}

// Enviroment variables
type EnvVariables struct {
	GithubClientID     string
	GithubClientSecret string
	GoogleClientID     string
	GoogleClientSecret string
	RedirectURI        string
	IceURLs            []string
	IceTokens          []string
}

type XirsysResponse struct {
	V V      `json:"v"`
	S string `json:"s"`
}

type V struct {
	IceObject IceServers `json:"iceServers"`
}

type IceServers struct {
	Username   string   `json:"username,omitempty"`
	Urls       []string `json:"urls"`
	Credential string   `json:"credential,omitempty"`
}

type IceResponse struct {
	Ice []IceServers `json:"iceServers"`
}

type XirsysPayload struct {
	Format string `json:"format"`
}



//<---------------For One Tap JWT------------------>


type GoogleOneTapResponse struct {
	ClientId string `json:"clientId"`
	Credential string `json:"credential"`
}


type JSONWebKeys struct {
	Keys []JSONWebKey `json:"keys"`
}

type JSONWebKey struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	E   string `json:"e"`
	Alg string `json:"alg"`
	Use string `json:"use"`
	N   string `json:"n"`
}

type GoogleTokenClaims struct {
	Name          string `json:"name"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	jwt.StandardClaims
}
