package hue

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/joho/godotenv/autoload"

	"github.com/dghubble/sling"
	"github.com/go-redis/redis"
	"github.com/skratchdot/open-golang/open"
	"golang.org/x/oauth2"
)

var oauthConfig *oauth2.Config
var callbackPath = "/oauth2/callback"
var oauthStateString = randomHex(10)
var port = 4567
var RedisClient = redis.NewClient(&redis.Options{})
var token oauth2.Token
var base *sling.Sling

// Client stores the auth'd HTTP client
var Client *http.Client

// Username stores the client's auth'd username
var Username string

const tokenKey = "HUE_TOKEN"
const usernameKey = "HUE_USERNAME"

func init() {
	oauthConfig = makeConfig()
	token = getToken()

	if (token == oauth2.Token{}) {
		return
	}

	Client = makeClient()
	base = makeBaseRequest()
	Username = getUsername()
}

// AuthDance Guides the user through the OAuth dance to get an OAuth token
func AuthDance() {
	url := oauthConfig.AuthCodeURL(oauthStateString)
	log.Printf("Opening %s in your browser", url)
	open.Run(url)

	http.HandleFunc(callbackPath, handleCallback)

	log.Printf("Server listening on http://localhost:%d", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}

// Reset resets all OAuth creds
func Reset() {
	RedisClient.FlushAll()
	fmt.Print("Reset all OAuth creds")
	os.Exit(0)
}

func randomHex(n int) string {
	bytes := make([]byte, n)
	_, err := rand.Read(bytes)

	if err != nil {
		log.Fatal(err)
	}

	return hex.EncodeToString(bytes)
}

func handleCallback(w http.ResponseWriter, r *http.Request) {
	token := buildToken(r.FormValue("state"), r.FormValue("code"))

	tokenJSON, err := json.Marshal(token)
	if err != nil {
		log.Fatal(err)
	}

	err = RedisClient.Set(tokenKey, tokenJSON, 0).Err()
	if err != nil {
		log.Fatal(err)
	}

	log.Print("Got token")
}

func buildToken(state string, code string) *oauth2.Token {
	if state != oauthStateString {
		log.Fatal("Invalid OAuth State")
	}

	token, err := oauthConfig.Exchange(oauth2.NoContext, code)
	if err != nil {
		log.Fatalf("Code exchange failed: %s", err.Error())
	}

	if token.TokenType == "BearerToken" {
		token.TokenType = "bearer"
	}

	return token
}

// https://github.com/Q42Philips/hue-remote-api-debugger
func makeConfig() *oauth2.Config {
	return &oauth2.Config{
		RedirectURL:  fmt.Sprintf("http://localhost:%d%s", port, callbackPath),
		ClientID:     os.Getenv("HUE_CLIENT_ID"),
		ClientSecret: os.Getenv("HUE_CLIENT_SECRET"),
		Scopes:       []string{},
		Endpoint: oauth2.Endpoint{
			AuthURL:  fmt.Sprintf("%s/oauth2/auth?appid=%s&deviceid=%s&devicename=browser", Endpoint, AppID, AppID),
			TokenURL: fmt.Sprintf("%s/oauth2/token", Endpoint),
		},
	}
}

func makeClient() *http.Client {
	ctx := context.Background()
	return oauthConfig.Client(ctx, &token)
}

func makeBaseRequest() *sling.Sling {
	base = sling.New().Base(Endpoint)
	base.Set("Content-Type", "application/json")
	base.Client(Client)
	return base
}

func pushButtion() {
	type putPayload struct {
		LinkButton bool `json:"linkbutton"`
	}

	data := putPayload{LinkButton: true}
	_ = MakeRequest("/bridge/0/config", http.MethodPut, data, nil)
}

// MakeRequest makes a request to the Hue API
func MakeRequest(path string, method string, data interface{}, success interface{}) *http.Response {

	s := base.New().BodyJSON(data)

	switch method {
	case http.MethodPost:
		s.Post(path)
	case http.MethodPut:
		s.Put(path)
	}

	req, err := s.Request()

	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Making %s request to %s with body %#v", req.Method, req.URL, data)
	resp, err := s.ReceiveSuccess(success)

	if err != nil {
		log.Fatalf("%s request to %s failed: %s", method, path, err)
	}

	if resp.StatusCode > 299 {
		log.Fatal(resp)
	}

	log.Printf("Got back %#v, with response code %d", success, resp.StatusCode)

	return resp
}

func getUsername() string {
	val, err := RedisClient.Get(usernameKey).Result()

	if err == redis.Nil {
		log.Print("Username not cached")
	} else if err != nil {
		log.Fatal(err)
	} else {
		log.Print("Got username from cache")
		Username = val
		return val
	}

	pushButtion()

	type postPayload struct {
		DeviceType string `json:"devicetype"`
	}

	data := postPayload{DeviceType: AppID}
	type postResponse struct {
		Success struct {
			Username string
		}
	}

	success := []postResponse{}
	_ = MakeRequest("/bridge", http.MethodPost, data, &success)

	Username = success[0].Success.Username
	RedisClient.Set(usernameKey, Username, 0)

	log.Print("Got username")

	return Username
}

func getToken() oauth2.Token {
	val, err := RedisClient.Get(tokenKey).Result()

	if err != nil {
		log.Printf("Could not find cached token")
		return oauth2.Token{}
	} else {
		log.Printf("Found cached token")
	}

	err = json.Unmarshal([]byte(val), &token)

	if err != nil {
		log.Fatal(err)
	}

	return token
}
