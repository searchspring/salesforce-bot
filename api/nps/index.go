package nps

import (
	//"bytes"
	//"encoding/json"
	//"errors"
	"fmt"
	"log"
	"strconv"

	//"math/rand"
	"net/http"
	"net/url"
	"reflect"
	"strings"

	//"time"

	//petname "github.com/dustinkirkland/golang-petname"
	"github.com/kelseyhightower/envconfig"
	"github.com/nlopes/slack"
	//"github.com/searchspring/nebo/nextopia"
	//"github.com/searchspring/nebo/salesforce"
)

type envVars struct {
	DevMode                string `split_words:"true" required:"false"`
	SlackVerificationToken string `split_words:"true" required:"false"`
	SlackOauthToken        string `split_words:"true" required:"false"`
	SfURL                  string `split_words:"true" required:"false"`
	SfUser                 string `split_words:"true" required:"false"`
	SfPassword             string `split_words:"true" required:"false"`
	SfToken                string `split_words:"true" required:"false"`
	NxUser                 string `split_words:"true" required:"false"`
	NxPassword             string `split_words:"true" required:"false"`
	GdriveFireDocFolderID  string `split_words:"true" required:"false"`
}
type slackAttachment struct {
	Text string
}

type slackMessage struct {
	Text string
	Attachments []slackAttachment
}

type SlackDAO interface {
	sendSlackMessage(token string, text string, attachments slack.Attachment, authorID string, channel string) error
	getValues() []string
}

type SlackDAOFake struct {
	Recorded []string
}
type SlackDAOReal struct{}

var slackDAO SlackDAO = nil

func (s *SlackDAOFake) sendSlackMessage(token string, text string, attachments slack.Attachment, authorID string, channel string) error {

	s.Recorded = []string{token, text, authorID, channel}
	return nil
}

func (s *SlackDAOFake) getValues() []string {
	return s.Recorded
}

func (s *SlackDAOReal) getValues() []string {
	return []string{"", ""}
}

func (s *SlackDAOReal) sendSlackMessage(token string, text string, attachments slack.Attachment, authorID string, channel string) error {
	api := slack.New(token)
	channelID, timestamp, err := api.PostMessage(
		channel,
		//slack.MsgOptionText("<@"+authorID+"> requests: "+text, false), 
		slack.MsgOptionAttachments(attachments))
	if err != nil {
		return err
	}
	fmt.Printf("Message successfully sent to channel %s at %s", channelID, timestamp)
	return nil
}

// Handler - check routing and call correct methods
func Handler(w http.ResponseWriter, r *http.Request) {

	var env envVars
	err := envconfig.Process("", &env)
	if err != nil {
		fmt.Println(err.Error())
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
		log.Print(err.Error())
	}

	urlMap, err := parseUrl(r)

	if err != nil {
		sendInternalServerError(w, err)
		return
	}
	fmt.Println(urlMap)

	if _, exists := urlMap["test"]; exists {
		slackDAO = &SlackDAOFake{}
	} else {
		slackDAO = &SlackDAOReal{}
	}

	attachments, err := createSlackAttachment(urlMap)
	if err != nil {
		sendInternalServerError(w, err)
		return
	}

	err = slackDAO.sendSlackMessage(env.SlackOauthToken, "", attachments, "U01R5TH2DK4", "C01TWG8D6CC")
	if err != nil {
		fmt.Println(err.Error())
		sendInternalServerError(w, err)
		return
	}

}

func createSlackAttachment(urlMap map[string][]string) (slack.Attachment, error) {
	attachments := slack.Attachment{
		AuthorName: "New NPS Rating",
		AuthorIcon: "https://avatars.slack-edge.com/2020-01-08/900543610438_6d658dd2df4b32187c53_512.png",
		Fields: []slack.AttachmentField{
			{
				Title: "Name",
				Value: urlMap["name"][0],
				Short: true,
			},
			{
				Title: "Website",
				Value: urlMap["website"][0],
				Short: true,
			},
			{
				Title: "Email",
				Value: urlMap["email"][0],
				Short: true,
			},
		},
	}
	newField := slack.AttachmentField{}
	if _, exists := urlMap["rating"]; exists {
		newField = slack.AttachmentField{
			Title: "Rating",
			Value: urlMap["rating"][0],
			Short: true,
		}

		i, err := strconv.Atoi(urlMap["rating"][0]) 
		if err != nil {
			return slack.Attachment{}, err 
		}
		if i > 8 {
			attachments.Color = "#35a64f"
		} else if i > 6 && i <= 8 {
			attachments.Color = "#b8ba31"
		} else {
			attachments.Color = "#eb0101"
		}
	} else if _, exists := urlMap["feedback"]; exists {
		newField = slack.AttachmentField{
			Title: "Feedback",
			Value: urlMap["feedback"][0],
			Short: true,
		}
	}
	attachments.Fields = append([]slack.AttachmentField{newField}, attachments.Fields...)

	return attachments, nil
}

func parseUrl(r *http.Request) (map[string][]string, error) {
	expectedKeys := map[string]bool{"rating": true, "feedback": true, "name": false, "email": false, "website": false, "test": true}
	u, err := url.Parse(r.URL.String())

	if err != nil {
		return nil, err
	}

	urlParams := u.Query()

	if len(urlParams) < 1 {
		return nil, fmt.Errorf("url params are missing")
	}

	for k := range urlParams {
		//fmt.Printf( "Url Param %s is %v: ", k, string(v[0]))
		_, exists := expectedKeys[k]
		if !exists {
			return nil, fmt.Errorf("field %s does not exist", k)
		}
		expectedKeys[k] = true
	}

	if falseKeys, ok := mapIsTrue(expectedKeys); ok {
		return nil, fmt.Errorf("request is missing keys: %s", falseKeys)
	}

	return urlParams, nil
}

func mapIsTrue(inputMap map[string]bool) (string, bool) {
	falseKeys := ""
	for k, v := range inputMap {
		if !v {
			falseKeys += (k + ", ")
		}
	}
	if len(falseKeys) > 0 {
		return falseKeys, true
	}
	return falseKeys, false

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
