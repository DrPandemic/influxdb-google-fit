package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var conf *oauth2.Config
var state string
var store sessions.CookieStore;
var googleToken *oauth2.Token;
var currentEmail *string;
var running bool = false;
var runningMutex sync.Mutex;
var nextTick time.Time;

func randToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	store = sessions.NewCookieStore([]byte(os.Getenv("SESSION_SECRET")))

	redirect_url := "http://localhost:3000"
	if env_url := os.Getenv("REDIRECT_URL"); env_url != "" {
		redirect_url = env_url
	}

	conf = &oauth2.Config{
		ClientID:     os.Getenv("CID"),
		ClientSecret: os.Getenv("CSECRET"),
		RedirectURL:  fmt.Sprintf("%s/auth/google/callback", redirect_url),
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/fitness.activity.read",
		},
		Endpoint: google.Endpoint,
	}
}

func indexHandler(c *gin.Context) {
	state = randToken()
	session := sessions.Default(c)
	session.Set("state", state)
	session.Save()
	var expiry *time.Time = nil
	if googleToken != nil {
		expiry = &googleToken.Expiry
	}
	c.HTML(http.StatusOK, "index.tmpl", gin.H{
		"url": conf.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce),
		"expiry": expiry,
		"email": currentEmail,
		"nextTick": nextTick,
	})
}

func validateState(c *gin.Context) bool {
	session := sessions.Default(c)
	retrievedState := session.Get("state")
	if retrievedState != c.Query("state") {
		c.AbortWithError(http.StatusUnauthorized, fmt.Errorf("Invalid session state: %s", retrievedState))
		return false
	}
	return true
}

func authHandler(c *gin.Context) {
	if !validateState(c) {
		return
	}

	token, err := conf.Exchange(oauth2.NoContext, c.Query("code"))
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	googleToken = token

	email, err := fetchEmail(token)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	currentEmail = email

	startBgTask()

    c.Redirect(http.StatusFound, "/")
}

func startBgTask() {
	if running {
		return
	}

	runningMutex.Lock()
	defer runningMutex.Unlock()
	if running {
		return
	}
	running = true

	go func() {
		ticker := time.NewTicker(time.Hour)
		for ; true; <- ticker.C {
			fmt.Println("tick")
			nextTick = time.Now().Add(time.Hour)
			if currentEmail != nil && googleToken != nil {
				if err := fetchAndSaveFitnessData(googleToken, conf); err != nil {
					fmt.Println("Failed to fetch data", err)
				}
			} else {
				fmt.Println("Wasn't ready to fetch data")
			}
		}
	}()
}

type BackfillData struct {
  Start	time.Time `json:"start" binding:"required"`
  End	time.Time `json:"end" binding:"required"`
}

func backfillHandler(c *gin.Context) {
	var input BackfillData
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	fmt.Println(input)

	if err := fetchAndSaveFitnessDataWithDates(googleToken, conf, input.Start, input.End); err != nil {
		fmt.Println("Failed to fetch data", err)
	}
}

func main() {
	router := gin.Default()
	router.Use(sessions.Sessions("fit", store))
	router.Static("/css", "./static/css")
	router.Static("/img", "./static/img")
	router.LoadHTMLGlob("templates/*")

	router.GET("/", indexHandler)
	router.GET("/auth/google/callback", authHandler)
	router.POST("/backfill", backfillHandler)

	router.Run("localhost:3000")
}
