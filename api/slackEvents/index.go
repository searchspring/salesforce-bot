package slackEvents

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/kelseyhightower/envconfig"
	"github.com/nlopes/slack"
	common "github.com/searchspring/nebo/api/config"
)

type ChallengeEvent struct {
	Token     string `json:"token"`
	Challenge string `json:"challenge"`
	Type      string `json:"type"`
}
type ChannelEvent struct {
	Token     string `json:"token"`
	Challenge string `json:"challenge"`
	Type      string `json:"type"`
	Event     Event  `json:"event"`
}

type Event struct {
	Type    string  `json:"type"`
	Token   string  `json:"token"`
	Channel Channel `json:"channel"`
}

type Channel struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

func Handler(w http.ResponseWriter, r *http.Request) {
	slackDAO := &common.SlackDAOImpl{}
	var env common.EnvVars
	err := envconfig.Process("", &env)
	if err != nil {
		common.SendInternalServerError(w, err)
		return
	}

	blanks := common.FindBlankEnvVars(env)
	if len(blanks) > 0 {
		err := fmt.Errorf("the following env vars are blank: %s", strings.Join(blanks, ", "))
		if env.DevMode != "development" {
			common.SendInternalServerError(w, err)
			return
		}
		log.Println(err.Error())
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading body: %v", err)
		http.Error(w, "can't read body", http.StatusBadRequest)
		return
	}

	event := &Event{}
	json.Unmarshal(body, event)
	if event.Type == "url_verification" {
		eventDetails := &ChallengeEvent{}
		json.Unmarshal(body, eventDetails)
		w.Write([]byte(eventDetails.Challenge))
		return
	}

	if event.Token != env.SlackVerificationToken {
		log.Printf("Error reading body: %v", err)
		http.Error(w, "Invalid Verification Token", http.StatusBadRequest)
		return
	}

	if event.Type == "event_callback" {
		eventDetails := &ChannelEvent{}
		w.Write([]byte("success"))
		json.Unmarshal(body, eventDetails)
		fmt.Printf("Event received: %s\n", eventDetails.Event.Type)
		if eventDetails.Event.Type == "channel_created" {
			slackDAO.SendSlackMessage(env.SlackOauthToken, slack.Attachment{
				AuthorIcon: "https://emoji.slack-edge.com/T024FV14T/slack/7d462d2443.png",
				AuthorName: "Slack Event",
				Text:       fmt.Sprintf("New channel: <#%s>", eventDetails.Event.Channel.ID),
			}, "C01VD4Z343B")
		}
		return
	}
}
