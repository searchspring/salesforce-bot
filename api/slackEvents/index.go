package nps

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"reflect"
	"strings"

	"github.com/kelseyhightower/envconfig"
	"github.com/nlopes/slack"
)

type envVars struct {
	DevMode                string `split_words:"true" required:"false"`
	SlackVerificationToken string `split_words:"true" required:"false"`
	SlackOauthToken        string `split_words:"true" required:"false"`
}

type SlackDAO interface {
	sendSlackMessage(token string, attachments slack.Attachment, channel string) error
	getValues() []string
}

type SlackDAOFake struct {
	Recorded []string
}
type SlackDAOReal struct{}

var slackDAO SlackDAO = nil

func (s *SlackDAOFake) sendSlackMessage(token string, attachments slack.Attachment, channel string) error {

	s.Recorded = []string{token, channel}
	return nil
}

func (s *SlackDAOFake) getValues() []string {
	return s.Recorded
}

func (s *SlackDAOReal) getValues() []string {
	return []string{"", ""}
}

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
	Channel Channel `json:"channel"`
}

type Channel struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

func (s *SlackDAOReal) sendSlackMessage(token string, attachments slack.Attachment, channel string) error {
	api := slack.New(token)
	if _, _, err := api.PostMessage(channel, slack.MsgOptionAttachments(attachments)); err != nil {
		return err
	}
	return nil
}

func Handler(w http.ResponseWriter, r *http.Request) {
	slackDAO = &SlackDAOReal{}
	var env envVars
	err := envconfig.Process("", &env)
	if err != nil {
		sendInternalServerError(w, err)
		return
	}

	blanks := findBlankEnvVars(env)
	if len(blanks) > 0 {
		err := fmt.Errorf("the following env vars are blank: %s", strings.Join(blanks, ", "))
		if env.DevMode != "development" {
			sendInternalServerError(w, err)
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

	event := &slack.Event{}
	json.Unmarshal(body, event)
	if event.Type == "url_verification" {
		eventDetails := &ChallengeEvent{}
		w.Write([]byte(eventDetails.Challenge))
		return
	}

	if event.Type == "event_callback" {
		eventDetails := &ChannelEvent{}
		w.Write([]byte("success"))
		json.Unmarshal(body, eventDetails)
		fmt.Printf("Event received: %s\n", eventDetails.Event.Type)
		if eventDetails.Event.Type == "channel_created" {
			slackDAO.sendSlackMessage(env.SlackOauthToken, slack.Attachment{
				AuthorIcon: "https://emoji.slack-edge.com/T024FV14T/slack/7d462d2443.png",
				AuthorName: "Slack Event",
				Text:       fmt.Sprintf("New channel: <#%s>", eventDetails.Event.Channel.ID),
			}, "C01VD4Z343B")
		}
		return
	}
}

func sendInternalServerError(res http.ResponseWriter, err error) {
	log.Println(err.Error())
	http.Error(res, err.Error(), http.StatusInternalServerError)
}

func findBlankEnvVars(env envVars) []string {
	var blanks []string
	valueOfStruct := reflect.ValueOf(env)
	typeOfStruct := valueOfStruct.Type()
	for i := 0; i < valueOfStruct.NumField(); i++ {
		if valueOfStruct.Field(i).Interface() == "" {
			blanks = append(blanks, typeOfStruct.Field(i).Name)
		}
	}
	return blanks
}
