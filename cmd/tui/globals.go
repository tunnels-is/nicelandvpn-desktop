package main

import (
	"time"

	"github.com/tunnels-is/nicelandvpn-desktop/core"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var user *User
var FR core.FORWARD_REQUEST

// Device token struct need for the login respons from user scruct
type DEVICE_TOKEN struct {
	DT      string    `bson:"DT"`
	N       string    `bson:"N"`
	Created time.Time `bson:"C"`
}

// use struct you get from the login request
type User struct {
	ID               primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	APIKey           string             `bson:"AK" json:"APIKey"`
	Email            string             `bson:"E"`
	TwoFactorEnabled bool               `json:"TwoFactorEnabled" bson:"TFE"`
	Disabled         bool               `bson:"D" json:"Disabled"`
	Tokens           []*DEVICE_TOKEN    `json:"Tokens" bson:"T"`
	DeviceToken      *DEVICE_TOKEN      `json:",omitempty" bson:"-"`

	CashCode      int       `bson:"CSC" json:"CashCode"`
	Affiliate     string    `bson:"AF"`
	SubLevel      int       `bson:"SUL"`
	SubExpiration time.Time `bson:"SE"`
	TrialStarted  time.Time `bson:"TrialStarted" json:"TrialStarted"`

	CancelSub bool `json:"CancelSub" bson:"CS"`

	Version string `json:"Version" bson:"-"`
}
