package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

func main() {
	flag.Parse()
	db := InitDb()

	inputChannel := make(chan string)
	go listenForInput(inputChannel)
	go func() {
		for {
			select {
			case cmd := <-inputChannel:
				go handleCommand(cmd, db)
			}
			time.Sleep(50 * time.Millisecond)
		}
	}()

	r := gin.Default()

	r.POST("/ac/session/create", SessionCreate(db))

	r.POST("/ac/session/validate", SessionValidate(db))

	r.POST("/ac/webhook/keycloak/event", EventWebhook(db))

	err := r.Run("0.0.0.0:29706")
	if err != nil {
		panic(err)
	}
}

func SessionCreate(db *Db) func(c *gin.Context) {
	return func(c *gin.Context) {
		type Req struct {
			Token        *Token `json:"token"`
			SessionState string `json:"session_state"`
		}
		req := &Req{}
		_ = c.ShouldBind(req)

		idToken := req.Token.DecodeIdToken()
		sub := GetSubFromIdToken(idToken)

		//check if the token exists
		exists := CredentialExistenceByUserId(sub, db)

		//if token exists, redirect to the app
		if exists {
			credId := GetCredentialIdByUserId(sub, db)
			cred := GetCredentialById(credId, db)
			log.Printf("Credential already exists for user %s", sub)
			c.JSON(http.StatusOK, gin.H{"credential": cred})
			return
		}

		//if token does not exist, save it and redirect to the app
		log.Printf("Credential does not exist for user %s", sub)
		cred := NewCredential()
		cred.UserId = sub
		cred.SessionState = req.SessionState
		cred.AccessToken = req.Token.AccessToken
		cred.ExpiresIn = req.Token.ExpiresIn
		_ = cred.Save(db)
		log.Printf("Saved credential for user %s", cred.UserId)

		c.JSON(http.StatusOK, gin.H{"credential": cred})
	}
}

func SessionValidate(db *Db) func(c *gin.Context) {
	return func(c *gin.Context) {
		id, err := c.Cookie("session_id")
		if err != nil {
			log.Printf("Can not get session id from cookie")
			id = c.PostForm("session_id")
			if id == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "session id not found"})
				c.Abort()
				return
			}
			log.Printf("Got session id from url: %s", id)
		}

		cred := GetCredentialById(id, db)
		if cred != nil {
			c.JSON(http.StatusOK, gin.H{"active": true, "credential": cred})
			log.Printf("Credential found for user %s", cred.UserId)
		} else {
			c.JSON(http.StatusOK, gin.H{"active": false, "credential": cred})
			log.Printf("Credential not found for session id %s", id)
		}
	}
}

func deleteSession(sessionState string, db *Db) (bool, error) {
	cred := GetCredentialBySessionState(sessionState, db)
	if cred == nil {
		log.Printf("Credential not found for session state %s", sessionState)
		return false, errors.New("credential not found")
	}
	b, err := cred.Delete(db)
	if err != nil {
		return false, errors.New("error deleting credential")
	}
	return b, nil
}

func EventWebhook(db *Db) func(c *gin.Context) {
	return func(c *gin.Context) {
		log.Printf("Received event webhook")
		type Event struct {
			Id            string `json:"id"`
			Time          int64  `json:"time"`
			RealmId       string `json:"realmId"`
			ResourcePath  string `json:"resourcePath"`
			Error         string `json:"error"`
			ResourceType  string `json:"resourceType"`
			OperationType string `json:"operationType"`
		}
		event := &Event{}
		_ = c.ShouldBind(event)
		log.Printf("Event received: %v", event)
		if event.ResourceType == "USER_SESSION" && event.OperationType == "DELETE" {
			go func() {
				sessionState := strings.Replace(event.ResourcePath, "sessions/", "", 1)
				_, _ = deleteSession(sessionState, db)
			}()
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
}

func listenForInput(inputChannel chan string) {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		inputChannel <- scanner.Text()
	}
	if scanner.Err() != nil {
		// handle error
		fmt.Println("Error reading from input:", scanner.Err())
	}
}

func handleCommand(cmd string, db *Db) {
	if cmd == "show -all" {
		db.Iterator()
	}
}
