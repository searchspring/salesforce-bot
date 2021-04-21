package nps

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
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

type NpsMessage struct {
	Name string `schema:"name,required"`
	Email string `schema:"email,required"`
	Website string `schema:"website,required"`
	Rating *int `schema:"rating"`
	Feedback *string `schema:"feedback"`
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

var decoder = schema.NewDecoder()

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

	var nps NpsMessage

	err := decoder.Decode(&nps, r.URL.Query())
    if err != nil {
        sendInternalServerError(w, err)
		return
    }

	query := strings.Split(nps.Website, " ")[0]
	responseData, err := salesforceApi.NPSQuery(query)
	if err != nil {
		sendInternalServerError(w, err)
		return
	}

	attachments, err := createSlackAttachment(nps, responseData)
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

func createSlackAttachment(nps NpsMessage, salesforceData []*salesforce.AccountInfo) (slack.Attachment, error) {
	mrr, rep := "Unknown", "Unknown"
	if len(salesforceData) > 0 {
		mrr = "$" + humanize.Comma(int64(salesforceData[0].FamilyMRR))
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
				Value: nps.Name,
				Short: true,
			},
			{
				Title: "Website",
				Value: nps.Website,
				Short: true,
			},
			{
				Title: "Email",
				Value: nps.Email,
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
	if nps.Rating != nil {
		newField = slack.AttachmentField{
			Title: "Rating",
			Value: strconv.Itoa(*nps.Rating),
			Short: true,
		}

		if *nps.Rating > 8 {
			attachments.Color = green
			attachments.AuthorIcon = "https://emojipedia-us.s3.dualstack.us-west-1.amazonaws.com/thumbs/240/apple/271/glowing-star_1f31f.png"
		} else if *nps.Rating > 6 && *nps.Rating <= 8 {
			attachments.Color = yellow
			attachments.AuthorIcon = "https://emojipedia-us.s3.dualstack.us-west-1.amazonaws.com/thumbs/240/apple/271/neutral-face_1f610.png"
		} else {
			attachments.Color = red
			attachments.AuthorIcon = "https://emojipedia-us.s3.dualstack.us-west-1.amazonaws.com/thumbs/240/apple/271/pile-of-poo_1f4a9.png"
		}
	} else if nps.Feedback != nil {
		attachments.AuthorName = "New NPS Feedback"
		newField = slack.AttachmentField{
			Title: "Feedback",
			Value: *nps.Feedback,
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
