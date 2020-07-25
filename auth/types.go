package auth

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

type IceServers struct {
	Username string `json:"username"`
	Credential string	`json:"credential"`
	Urls []string	`json:"urls"`
}

type XirsysResponse struct {
	V IceServers `json:"v"`
	S string `json:"s"`
}

type IceResponse struct {
	Ice []IceServers `json:"iceServers"`
}
