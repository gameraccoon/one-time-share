package main

import (
	"bytes"
	"fmt"
	"github.com/gameraccoon/one-time-share/database"
	"github.com/google/uuid"
	"log"
	"net/http"
	"os"
	"time"
)

// index.html that we are going to show to the user
var globalStaticData StaticData

type UserLimits struct {
	// how long the message is going to be available (zero means no limit)
	RetentionLimitMinutes int
	// max size of the message in bytes (zero means no limit)
	MaxMessageSizeBytes int
	// how often a new message can be created (zero means no limit)
	MessageCreationLimitMinutes int
}

type StaticData struct {
	// index.html that we are going to show to the user
	indexHtml string
	limits    UserLimits
	database  *database.OneTimeShareDb
}

func readUserLimits() error {
	if !globalStaticData.database.DoesUserExist("default") {
		globalStaticData.database.SetUserLimits("default", 10080, 1000, 1)
	}

	retentionLimitMinutes, maxSizeBytes, messageCreationLimitMinutes := globalStaticData.database.GetUserLimits("default")
	globalStaticData.limits = UserLimits{
		RetentionLimitMinutes:       retentionLimitMinutes,
		MaxMessageSizeBytes:         maxSizeBytes,
		MessageCreationLimitMinutes: messageCreationLimitMinutes,
	}

	return nil
}

func setupStaticPages() error {
	// read the index.html file
	indexHtml, err := os.ReadFile("index.html")
	if err != nil {
		log.Fatal("Error while reading index.html: ", err)
		return err
	}

	indexHtml = bytes.ReplaceAll(indexHtml, []byte("{{.MessageLimitBytes}}"), []byte(fmt.Sprintf("%d", globalStaticData.limits.MaxMessageSizeBytes)))
	indexHtml = bytes.ReplaceAll(indexHtml, []byte("{{.RetentionLimitMinutes}}"), []byte(fmt.Sprintf("%d", globalStaticData.limits.RetentionLimitMinutes)))

	globalStaticData.indexHtml = string(indexHtml)

	return nil
}

func homePage(w http.ResponseWriter, r *http.Request) {
	// check if the request is a GET request
	if r.Method != "GET" {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	_, err := fmt.Fprintf(w, globalStaticData.indexHtml)
	if err != nil {
		log.Println("Error while writing response: ", err)
		return
	}
}

func createNewMessage(w http.ResponseWriter, r *http.Request) {
	// check if the request is a POST request
	if r.Method != "POST" {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// read user_token from the request
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Can't parse form", http.StatusBadRequest)
		return
	}

	userToken := r.Form.Get("user_token")
	if userToken == "" {
		http.Error(w, "user_token is empty", http.StatusBadRequest)
		return
	}

	if !globalStaticData.database.DoesUserExist(userToken) {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	retentionLimitMinutes, maxSizeBytes, messageCreationLimitMinutes := globalStaticData.database.GetUserLimits(userToken)

	if messageCreationLimitMinutes > 0 {
		// check if the user can create a new message
		lastCreationTime := globalStaticData.database.GetUserLastMessageCreationTime(userToken)
		// if there was a message created before
		if lastCreationTime > 0 {
			timePassedFromLastCreation := time.Now().Sub(time.Unix(lastCreationTime, 0))
			if timePassedFromLastCreation.Minutes() < float64(messageCreationLimitMinutes) {
				minutesLeft := messageCreationLimitMinutes - int(timePassedFromLastCreation.Minutes())
				http.Error(w, "Message creation limit reached. Wait for "+fmt.Sprintf("%d", minutesLeft)+" minutes and repeat", http.StatusBadRequest)
				return
			}
		}
	}

	messageData := r.Form.Get("message_data")
	if messageData == "" {
		http.Error(w, "message_data is empty", http.StatusBadRequest)
		return
	}

	if maxSizeBytes > 0 && len(messageData) > maxSizeBytes {
		http.Error(w, "Message is too big", http.StatusBadRequest)
		return
	}

	requestedRetentionLimitText := r.Form.Get("retention")
	requestedRetentionLimitMinutes := -1
	if requestedRetentionLimitText != "" {
		requestedRetentionLimitMinutes, err = fmt.Sscanf(requestedRetentionLimitText, "%d", &requestedRetentionLimitMinutes)
		if err != nil {
			http.Error(w, "Can't parse retention limit", http.StatusBadRequest)
			return
		}
	}

	if requestedRetentionLimitMinutes < 0 {
		http.Error(w, "Invalid retention limit", http.StatusBadRequest)
	}

	if requestedRetentionLimitMinutes == 0 && retentionLimitMinutes > 0 {
		http.Error(w, "Can't set unlimited retention limit, not allowed", http.StatusBadRequest)
	}

	if requestedRetentionLimitMinutes > 0 && retentionLimitMinutes > 0 && requestedRetentionLimitMinutes > retentionLimitMinutes {
		http.Error(w, "Requested retention limit is bigger than allowed", http.StatusBadRequest)
		return
	}

	expireTimestamp := time.Now().Add(time.Duration(requestedRetentionLimitMinutes) * time.Minute).Unix()

	globalStaticData.database.SetUserLastMessageCreationTime(userToken, time.Now().Unix())

	// generate GUID for the message
	messageToken := uuid.New().String()

	err = globalStaticData.database.SaveMessage(messageToken, expireTimestamp, messageData)
	if err != nil {
		log.Println("Error while saving message: ", err)
		http.Error(w, "Can't save message. Try again", http.StatusInternalServerError)
		return
	}

	// for now just send the message back
	_, err = fmt.Fprintf(w, messageData)
	if err != nil {
		log.Println("Error while writing response: ", err)
		return
	}
}

func tryConsumeExistingMessage(w http.ResponseWriter, r *http.Request) {
	_, err := fmt.Fprintf(w, "Welcome to the HomePage!")
	if err != nil {
		return
	}
	fmt.Println("Endpoint Hit: homePage")
}

func getLimits(w http.ResponseWriter, r *http.Request) {
	// check if the request is a GET request
	if r.Method != "GET" {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// write messageLimitBytes and retentionLimitMinutes to the JSON response
	messageLimitBytes := globalStaticData.limits.MaxMessageSizeBytes
	retentionLimitMinutes := globalStaticData.limits.RetentionLimitMinutes

	_, err := fmt.Fprintf(w, `{"messageLimitBytes": %d, "retentionLimitMinutes": %d}`, messageLimitBytes, retentionLimitMinutes)
	if err != nil {
		log.Println("Error while writing response: ", err)
		return
	}
}

func handleRequests() {
	http.HandleFunc("/", homePage)
	http.HandleFunc("/save", createNewMessage)
	http.HandleFunc("/consume", tryConsumeExistingMessage)
	http.HandleFunc("/get-limits", getLimits)
	log.Fatal(http.ListenAndServe(":10000", nil))
}

func main() {
	db, err := database.ConnectDb("./one-time-share.db")
	defer db.Disconnect()

	if err != nil {
		log.Fatal("Can't connect to database: ", err)
		return
	}

	database.UpdateVersion(db)
	globalStaticData.database = db

	err = readUserLimits()
	if err != nil {
		log.Fatal("Error while reading user limits: ", err)
		return
	}

	err = setupStaticPages()
	if err != nil {
		log.Fatal("Error while setting up static pages: ", err)
		return
	}

	handleRequests()
}
