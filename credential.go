package main

import (
	"encoding/json"
	"github.com/google/uuid"
	"time"
)

const prefix = "cred_"

type Credential struct {
	Id          string `json:"id"`
	AccessToken string `json:"access_token"`
	UserId      string `json:"user_id"`
	CreateDate  string `json:"create_date"`
	ExpiresIn   int    `json:"expires_in"`
	TenantId    string `json:"tenant_id"`
}

func (cred *Credential) ToJsonStr() string {
	bytes, _ := json.Marshal(cred)
	return string(bytes)
}

func NewCredential() *Credential {
	return &Credential{
		Id:         uuid.NewString(),
		CreateDate: time.Now().Format(time.RFC3339),
	}
}

func CredentialExistenceByUserId(userId string, db *Db) bool {
	key := []byte(prefix + userId)
	item, err := db.Get(key)
	if err != nil || item == "" {
		return false
	}
	return true
}

func GetCredentialById(id string, db *Db) *Credential {
	key := []byte(prefix + id)
	item, err := db.Get(key)
	if err != nil || item == "" {
		return nil
	}

	cred := &Credential{}
	_ = json.Unmarshal([]byte(item), cred)
	return cred
}

func GetCredentialIdByUserId(userId string, db *Db) string {
	key := []byte(prefix + userId)
	item, err := db.Get(key)
	if err != nil || item == "" {
		return ""
	}
	return item
}

func (cred *Credential) Save(db *Db) error {
	credJson, _ := json.Marshal(cred)
	ttl := time.Duration(cred.ExpiresIn) * time.Second
	err := db.Set([]byte(prefix+cred.UserId), []byte(cred.Id), ttl)
	if err != nil {
		return err
	}
	err = db.Set([]byte(prefix+cred.Id), credJson, ttl)
	if err != nil {
		return err
	}
	return err
}
