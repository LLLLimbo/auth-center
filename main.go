package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"os"
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

	err := r.Run("0.0.0.0:29706")
	if err != nil {
		panic(err)
	}
}

func SessionCreate(db *Db) func(c *gin.Context) {
	return func(c *gin.Context) {
		token := &Token{}
		_ = c.ShouldBind(token)
		idToken := token.DecodeIdToken()
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
		cred.AccessToken = token.AccessToken
		cred.ExpiresIn = token.ExpiresIn
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
