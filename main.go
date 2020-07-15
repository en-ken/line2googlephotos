package p

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/line/line-bot-sdk-go/linebot"
)

type entry struct {
	Message  string   `json:"message"`
	Severity severity `json:"severity,omitempty"`
}

type severity string

const (
	severityInfo  severity = "INFO"
	severityError severity = "ERROR"
)

func outputErrorLog(w http.ResponseWriter, statusCode int, format string, v ...interface{}) {
	e := entry{
		Severity: severityError,
		Message:  fmt.Sprintf(format, v...),
	}
	out, err := json.Marshal(e)
	if err != nil {
		log.Printf("json.Marshal: %v", err)
	}
	fmt.Println(string(out))
	w.WriteHeader(statusCode)
	w.Write(out)
}

func outputLog(format string, v ...interface{}) {
	e := entry{
		Severity: severityInfo,
		Message:  fmt.Sprintf(format, v...),
	}
	out, err := json.Marshal(e)
	if err != nil {
		log.Printf("json.Marshal: %v", err)
	}
	fmt.Println(string(out))
}

func mustGetEnv(envName string) string {
	env := os.Getenv(envName)
	if env == "" {
		log.Fatal("")
	}
	return env
}

type config struct {
	LineChannelSecret      string
	LineChannelAccessToken string
	Oauth2ClientID         string
	Oauth2ClientSecret     string
	Oauth2RefreshToken     string
	AlbumID                string
}

func newConfig() *config {
	return &config{
		LineChannelSecret:      mustGetEnv("LINE_CHANNEL_SECRET"),
		LineChannelAccessToken: mustGetEnv("LINE_CHANNEL_ACCESS_TOKEN"),
		Oauth2ClientID:         mustGetEnv("OAUTH2_CLIENT_ID"),
		Oauth2ClientSecret:     mustGetEnv("OAUTH2_CLIENT_SECRET"),
		Oauth2RefreshToken:     mustGetEnv("OAUTH2_REFRESH_TOKEN"),
		AlbumID:                mustGetEnv("ALBUM_ID"),
	}

}

// MessageReceived fires when Line events received
func MessageReceived(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	config := newConfig()

	// Parse events
	events, err := linebot.ParseRequest(config.LineChannelSecret, r)
	if err != nil {
		outputErrorLog(w, http.StatusBadRequest, fmt.Sprintf("Failed to parse request: %+v", err))
		return
	}

	outputLog("Number of events = %d", len(events))
	for _, ev := range events {
		if ev.Type != linebot.EventTypeMessage {
			outputLog("Event ignored: type=%v", ev.Type)
			return
		}

		msg, ok := ev.Message.(*linebot.ImageMessage)
		if !ok {
			outputLog("Message event ignored: event=%v", ev)
			return
		}

		// Get image
		bot, err := linebot.New(config.LineChannelSecret, config.LineChannelAccessToken)
		if err != nil {
			outputErrorLog(w, http.StatusInternalServerError, fmt.Sprintf("Failed to create linebot client: %+v", err))
			return
		}
		content, err := bot.GetMessageContent(msg.ID).Do()
		if err != nil {
			outputErrorLog(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get message content: %+v", err))
			return
		}
		defer content.Content.Close()

		name := "参加者の誰か"
		profile, err := bot.GetProfile(ev.Source.UserID).Do()
		if err == nil {
			name = profile.DisplayName
		}

		uploader, err := NewGooglePhotoUploader(config)
		if err != nil {
			outputErrorLog(w, http.StatusInternalServerError, fmt.Sprintf("Failed to create uploader: %+v", err))
			return
		}

		if err := uploader.upload(ctx, name, content.Content); err != nil {
			outputErrorLog(w, http.StatusInternalServerError, fmt.Sprintf("Failed to upload: %+v", err))
			return
		}

		outputLog("Upload success: %s", msg.ID)
	}

	outputLog("Finished")

	return
}
