package main

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/tunnels-is/nicelandvpn-desktop/core"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

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

func TimedUIUpdate(MONITOR chan int) {
	defer func() {
		time.Sleep(3 * time.Second)
		MONITOR <- 3
	}()
	defer core.RecoverAndLogToFile()

	for {
		time.Sleep(3 * time.Second)
		TUI.Send(&tea.KeyMsg{
			Type: 0,
		})
	}
}

func GetLogs() []string {
	var logs []string
	LR, err := core.GetLogsForCLI()
	if LR != nil && err == nil {
		logs = LR.Content
		for i := range LR.Content {
			if LR.Content[i] != "" {
				logs[i] = fmt.Sprint(LR.Time[i], "||", LR.Function[i], "||", LR.Content[i]+"\n")
			}
		}
	}

	return logs
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
