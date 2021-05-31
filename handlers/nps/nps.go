package nps

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"github.com/kelseyhightower/envconfig"
	"github.com/nlopes/slack"
	"github.com/searchspring/nebo/common"
	"github.com/searchspring/nebo/dals/metabase"
)

type NpsMessage struct {
	Name     string  `schema:"name,required"`
	Email    string  `schema:"email,required"`
	Website  string  `schema:"website,required"`
	Rating   *int    `schema:"rating"`
	Feedback *string `schema:"feedback"`
}

var router *mux.Router
var env common.EnvVars

var decoder = schema.NewDecoder()

func Handler(w http.ResponseWriter, r *http.Request) {
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
	metabaseDAO := metabase.NewDAO("https://metabase.kube.searchspring.io/", env.MetabaseUser, env.MetabasePassword, "")
	router.HandleFunc("/nps", wrapSendNPSMessage(SendNPSMessage, &common.SlackDAOImpl{}, metabaseDAO)).Methods(http.MethodGet, http.MethodOptions)
	router.Use(mux.CORSMethodMiddleware(router))
	return router, nil
}

func wrapSendNPSMessage(apiRequest func(w http.ResponseWriter, r *http.Request, slackApi common.SlackDAO, metabaseDAO metabase.DAO), slackApi common.SlackDAO, metabaseDAO metabase.DAO) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		if r.Method == http.MethodOptions {
			return
		}
		apiRequest(w, r, slackApi, metabaseDAO)
	}
}

func SendNPSMessage(w http.ResponseWriter, r *http.Request, slackApi common.SlackDAO, metabaseDAO metabase.DAO) {

	var nps NpsMessage

	err := decoder.Decode(&nps, r.URL.Query())
	if err != nil {
		common.SendInternalServerError(w, err)
		return
	}

	query := strings.Split(nps.Website, ".")[0]
	responseData, err := metabaseDAO.QueryNPS(query)
	if err != nil {
		common.SendInternalServerError(w, err)
		return
	}

	attachments, err := createSlackAttachment(nps, responseData)
	if err != nil {
		common.SendInternalServerError(w, err)
		return
	}

	err = slackApi.SendSlackMessage(env.SlackOauthToken, attachments, os.Getenv("CHANNEL_ID"))
	if err != nil {
		common.SendInternalServerError(w, err)
		return
	}
}

func createSlackAttachment(nps NpsMessage, metabaseData *metabase.NpsInfo) (slack.Attachment, error) {
	mrr, rep := "Unknown", "Unknown"
	if metabaseData != nil {
		mrr = "$" + humanize.Comma(int64(metabaseData.FamilyMRR))
		rep = metabaseData.Manager
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
