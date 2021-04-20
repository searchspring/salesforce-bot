package nps

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/kelseyhightower/envconfig"
	"github.com/nlopes/slack"
	"github.com/searchspring/nebo/salesforce"
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

type SlackDAO interface {
	sendSlackMessage(token string, attachments slack.Attachment, channel string) error
	getValues() []string
}

type SlackDAOFake struct {
	Recorded []string
}
type SlackDAOReal struct{}

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

func (s *SlackDAOReal) sendSlackMessage(token string, attachments slack.Attachment, channel string) error {
	api := slack.New(token)
	channelID, timestamp, err := api.PostMessage(
		channel,
		slack.MsgOptionAttachments(attachments))
	if err != nil {
		return err
	}
	fmt.Printf("Message successfully sent to channel %s at %s", channelID, timestamp)
	return nil
}

var router *mux.Router
var env envVars

func Handler(w http.ResponseWriter, r *http.Request) {
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

	log.Println(r.Method, r.URL.Path)
	if router == nil {
		r, err := CreateRouter()
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		router = r
	}
	router.ServeHTTP(w, r)
}

func CreateRouter() (*mux.Router, error) {
	router := mux.NewRouter()
	salesforceDAOReal := salesforce.NewDAO(env.SfURL, env.SfUser, env.SfPassword, env.SfToken)
	router.HandleFunc("/nps", wrapSendNPSMessage(SendNPSMessage, &SlackDAOReal{}, salesforceDAOReal)).Methods(http.MethodGet, http.MethodOptions)
	router.Use(mux.CORSMethodMiddleware(router))
	return router, nil
}

func wrapSendNPSMessage(apiRequest func(w http.ResponseWriter, r *http.Request, slackApi SlackDAO, salesforceDAOReal salesforce.DAO), slackApi SlackDAO, salesforceDAOReal salesforce.DAO) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		if r.Method == http.MethodOptions {
			return
		}
		apiRequest(w, r, slackApi, salesforceDAOReal)
	}
}

func SendNPSMessage(w http.ResponseWriter, r *http.Request, slackApi SlackDAO, salesforceApi salesforce.DAO) {

	urlMap, err := parseUrl(r)
	if err != nil {
		sendInternalServerError(w, err)
		return
	}

	fmt.Println("SF API: ", salesforceApi)
	query := strings.Split(urlMap["website"][0], " ")[0]
	responseData, err := salesforceApi.NPSQuery(query)
	if err != nil {
		sendInternalServerError(w, err)
		return
	}

	attachments, err := createSlackAttachment(urlMap, responseData)
	if err != nil {
		sendInternalServerError(w, err)
		return
	}

	err = slackApi.sendSlackMessage(env.SlackOauthToken, attachments, os.Getenv("CHANNEL_ID"))
	if err != nil {
		sendInternalServerError(w, err)
		return
	}
}

func createSlackAttachment(urlMap map[string][]string, salesforceData []*salesforce.AccountInfo) (slack.Attachment, error) {
	mrr, rep := "Unknown", "Unknown"
	if len(salesforceData) > 0 {
		mrr = "$" + formatInt(int(salesforceData[0].FamilyMRR))
		rep = salesforceData[0].Manager
	} 
	red := "#eb0101"
	yellow := "#b8ba31"
	green := "#35a64f"
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
			{
				Title: "Family MRR",
				Value: mrr,
				Short: true,
			},
			{
				Title: "Customer Success Manager",
				Value: rep,
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
			attachments.Color = green
			attachments.AuthorIcon = "https://emojipedia-us.s3.dualstack.us-west-1.amazonaws.com/thumbs/240/apple/271/glowing-star_1f31f.png"
		} else if i > 6 && i <= 8 {
			attachments.Color = yellow
			attachments.AuthorIcon = "https://emojipedia-us.s3.dualstack.us-west-1.amazonaws.com/thumbs/240/apple/271/neutral-face_1f610.png"
		} else {
			attachments.Color = red
			attachments.AuthorIcon = "https://emojipedia-us.s3.dualstack.us-west-1.amazonaws.com/thumbs/240/apple/271/pile-of-poo_1f4a9.png"
		}
	} else if _, exists := urlMap["feedback"]; exists {
		attachments.AuthorName = "New NPS Feedback"
		newField = slack.AttachmentField{
			Title: "Feedback",
			Value: urlMap["feedback"][0],
		}
	} else {
		attachments.AuthorName = "Error"
		newField = slack.AttachmentField{
			Title: "Error",
			Value: "No rating or feedback was given",
		}
	}

	attachments.Fields = append([]slack.AttachmentField{newField}, attachments.Fields...)
	return attachments, nil
}

func formatInt(number int) string {
    output := strconv.Itoa(number)
    startOffset := 3
    if number < 0 {
        startOffset++
    }
    for outputIndex := len(output); outputIndex > startOffset; {
        outputIndex -= 3
        output = output[:outputIndex] + "," + output[outputIndex:]
    }
    return output
}

func parseUrl(r *http.Request) (map[string][]string, error) {
	expectedKeys := map[string]bool{"rating": true, "feedback": true, "name": false, "email": false, "website": false}
	u, err := url.Parse(r.URL.String())

	if err != nil {
		return nil, err
	}

	urlParams := u.Query()

	if len(urlParams) < 1 {
		return nil, fmt.Errorf("url params are missing")
	}

	for k := range urlParams {
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
