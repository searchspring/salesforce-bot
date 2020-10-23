package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	petname "github.com/dustinkirkland/golang-petname"
	"github.com/kelseyhightower/envconfig"
	"github.com/nlopes/slack"
	"searchspring.com/slack/gapi"
	"searchspring.com/slack/nextopia"
	"searchspring.com/slack/salesforce"
)

type envVars struct {
	SlackVerificationToken      string `required:"true" split_words:"true"`
	SlackOauthToken             string `required:"true" split_words:"true"`
	SalesforceURL               string `split_words:"true"`
	SalesforceUser              string `split_words:"true"`
	SalesforcePassword          string `split_words:"true"`
	SalesforceToken             string `split_words:"true"`
	NextopiaUser                string `split_words:"true"`
	NextopiaPassword            string `split_words:"true"`
	GcpServiceAccountEmail      string `split_words:"true"`
	GcpServiceAccountPrivateKey string `split_words:"true"`
	GdriveFireDocFolderID       string `split_words:"true"`
}

var salesForceDAO salesforce.DAO = nil
var nextopiaDAO nextopia.DAO = nil
var gapiDAO gapi.DAO = nil

func containsEmptyString(vars ...string) bool {
	for _, v := range vars {
		if v == "" {
			return true
		}
	}
	return false
}

// Handler - check routing and call correct methods
func Handler(w http.ResponseWriter, r *http.Request) {
	var env envVars
	err := envconfig.Process("", &env)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s, err := slack.SlashCommandParse(r)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !s.ValidateToken(env.SlackVerificationToken) {
		err := errors.New("slack verification failed")
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	salesForceVars := []string{env.SalesforceURL, env.SalesforceUser, env.SalesforcePassword, env.SalesforceToken}
	if salesForceDAO == nil && !containsEmptyString(salesForceVars...) {
		salesForceDAO, err = salesforce.NewDAO(env.SalesforceURL, env.SalesforceUser, env.SalesforcePassword, env.SalesforceToken)
		if err != nil {
			log.Println(err.Error())
			http.Error(w, "salesforce client was not created successfully: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	nextopiaVars := []string{env.NextopiaUser, env.NextopiaPassword}
	if nextopiaDAO == nil && !containsEmptyString(nextopiaVars...) {
		nextopiaDAO = nextopia.NewDAO(env.NextopiaUser, env.NextopiaPassword)
	}

	gapiVars := []string{env.GcpServiceAccountEmail, env.GcpServiceAccountPrivateKey, env.GdriveFireDocFolderID}
	if gapiDAO == nil && !containsEmptyString(gapiVars...) {
		gapiDAO = gapi.NewDAO(env.GcpServiceAccountEmail, env.GcpServiceAccountPrivateKey, env.GdriveFireDocFolderID)
	}

	w.Header().Set("Content-type", "application/json")
	switch s.Command {
	case "/rep", "/alpha-nebo", "/nebo":
		if strings.TrimSpace(s.Text) == "help" || strings.TrimSpace(s.Text) == "" {
			writeHelpNebo(w)
			return
		}
		if salesForceDAO == nil {
			err := errors.New("missing required Salesforce credentials")
			log.Println(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		responseJSON, err := salesForceDAO.Query(s.Text)
		if err != nil {
			log.Println(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(responseJSON)
		return

	case "/fire":
	case "/firetest":
		if strings.TrimSpace(s.Text) == "help" {
			writeHelpFire(w)
			return
		}
		if gapiDAO == nil {
			err := errors.New("missing required Google API credentials")
			log.Println(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// We only have 3 seconds to initially respond
		// https://api.slack.com/interactivity/slash-commands#responding_to_commands
		// So we ACK before doing our work because Google APIs can be slow enough
		// that slack will drop our connection before we finish doing everything and responding
		fireTitle := cleanFireTitle(s.Text)
		w.Write([]byte("New Fire: :fire:*" + fireTitle + "*:fire:\nCreating fire doc & checklist now...\n"))
		go fireResponse(gapiDAO, env.GdriveFireDocFolderID, fireTitle, s.ResponseURL)
		return

	case "/firedown":
		w.Write(fireDownResponse())
		return

	case "/neboidnx", "/neboid":
		if strings.TrimSpace(s.Text) == "help" || strings.TrimSpace(s.Text) == "" {
			writeHelpNeboid(w)
			return
		}
		if nextopiaDAO == nil {
			err := errors.New("missing required Nextopia credentials")
			log.Println(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		responseJSON, err := nextopiaDAO.Query(s.Text)
		if err != nil {
			log.Println(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(responseJSON)
		return

	case "/neboidss":
		if strings.TrimSpace(s.Text) == "help" || strings.TrimSpace(s.Text) == "" {
			writeHelpNeboid(w)
			return
		}
		if salesForceDAO == nil {
			err := errors.New("missing required Salesforce credentials")
			log.Println(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		responseJSON, err := salesForceDAO.IDQuery(s.Text)
		if err != nil {
			log.Println(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(responseJSON)
		return

	case "/feature":
		if strings.TrimSpace(s.Text) == "help" || strings.TrimSpace(s.Text) == "" {
			writeHelpFeature(w)
			return
		}
		sendSlackMessage(env.SlackOauthToken, s.Text, s.UserID)
		responseJSON := featureResponse(s.Text)
		w.Write(responseJSON)
		return

	case "/meet":
		if strings.TrimSpace(s.Text) == "help" {
			writeHelpMeet(w)
			return
		}
		responseJSON := meetResponse(s.Text)
		w.Write(responseJSON)
		return

	case "/meettest":
		if strings.TrimSpace(s.Text) == "help" {
			writeHelpMeet(w)
			return
		}
		responseJSON := meetResponse(s.Text)
		w.Write(responseJSON)
		return

	default:
		err := errors.New("unknown slash command " + s.Command)
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func writeHelpFeature(w http.ResponseWriter) {
	msg := &slack.Msg{
		ResponseType: slack.ResponseTypeEphemeral,
		Text:         "Feature usage:\n`/feature description of feature required` - submits a feature to the product team\n`/feature help` - this message",
	}
	json, _ := json.Marshal(msg)
	w.Write(json)
}

func writeHelpFire(w http.ResponseWriter) {
	msg := &slack.Msg{
		ResponseType: slack.ResponseTypeEphemeral,
		Text:         "Fire usage:\n`/fire Fire Title` - start a new fire doc named 'Fire Title' and generate a fire checklist to handle the fire",
	}
	json, _ := json.Marshal(msg)
	w.Write(json)
}

func writeHelpNebo(w http.ResponseWriter) {
	platformsJoined := strings.ToLower(strings.Join(salesforce.Platforms, ", "))
	msg := &slack.Msg{
		ResponseType: slack.ResponseTypeEphemeral,
		Text:         "Nebo usage:\n`/nebo shoes` - find all customers with shoe in the name\n`/nebo shopify` - show {" + platformsJoined + "} clients sorted by MRR\n`/nebo help` - this message",
	}
	json, _ := json.Marshal(msg)
	w.Write(json)
}
func writeHelpNeboid(w http.ResponseWriter) {
	msg := &slack.Msg{
		ResponseType: slack.ResponseTypeEphemeral,
		Text:         "Neboid usage:\n`/neboid <id prefix>` - find all customers with an id that starts with this prefix\n`/neboid help` - this message",
	}
	json, _ := json.Marshal(msg)
	w.Write(json)
}

func writeHelpMeet(w http.ResponseWriter) {
	msg := &slack.Msg{
		ResponseType: slack.ResponseTypeEphemeral,
		Text:         "Meet usage:\n`/meet` - generate a random meet\n`/meet name` - generate a meet with a name\n`/meet help` - this message",
	}
	json, _ := json.Marshal(msg)
	w.Write(json)
}

func sendSlackMessage(token string, text string, authorID string) {
	api := slack.New(token)
	channelID, timestamp, err := api.PostMessage("G013YLWL3EX", slack.MsgOptionText("<@"+authorID+"> requests: "+text, false))
	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}
	fmt.Printf("Message successfully sent to channel %s at %s", channelID, timestamp)
}

func featureResponse(search string) []byte {
	msg := &slack.Msg{
		ResponseType: slack.ResponseTypeEphemeral,
		Text:         "feature request submitted, we'll be in touch!",
	}
	json, _ := json.Marshal(msg)
	return json
}

func meetResponse(search string) []byte {
	msg := &slack.Msg{
		ResponseType: slack.ResponseTypeInChannel,
		Text:         getMeetLink(search),
	}
	json, _ := json.Marshal(msg)
	return json
}

func getMeetLink(search string) string {
	name := search
	name = strings.ReplaceAll(name, " ", "-")
	if strings.TrimSpace(search) == "" {
		rand.Seed(time.Now().UnixNano())
		name = petname.Generate(3, "-")
	}
	return "g.co/meet/" + name
}

func fireResponse(d gapi.DAO, folderID string, title string, responseURL string) {
	documentID, err := d.GenerateFireDoc(title)
	msg := &slack.Msg{
		ResponseType: slack.ResponseTypeInChannel,
		Text:         fireChecklist(folderID, documentID, title, err),
	}
	json, _ := json.Marshal(msg)
	http.Post(responseURL, "application/json", bytes.NewBuffer(json))
}

func cleanFireTitle(title string) string {
	title = strings.Trim(title, " ")
	if title == "" {
		title = "Untitled Fire"
	}
	return title
}

func fireChecklist(fireDocFolderID string, documentID string, title string, err error) string {
	fireDocLink := fmt.Sprintf("<https://docs.google.com/document/d/%s/edit|%s>", documentID, title)
	if documentID == "" {
		fireDocLink = fmt.Sprintf("Create new fire doc <https://drive.google.com/drive/folders/%s|here>", fireDocFolderID)
	}
	text := "1. Designate Fire Leader\n" +
		"2. Designate Fire Doc Maintainer: " + fireDocLink + "\n" +
		"3. If a real fire - make an announcement in the annoucements channel \"There is a fire and engineering is investigating, updates will be posted in a thread on this message\"\n" +
		"4. Post a link to the fire document in the announcement channel thread\n" +
		"5. Designate helper(s) to update announcement\n" +
		"6. Fight! " + getMeetLink(title) + "\n" +
		"7. Use `/firedown` when the fire is out"

	if err != nil {
		text += "\n*Warning* - Encountered error:\n"
		text += "- " + err.Error()
	}
	return text
}

func fireDownResponse() []byte {
	msg := &slack.Msg{
		ResponseType: slack.ResponseTypeInChannel,
		Text: "1. Ask if there are any cleanup tasks to do\n" +
			"2. Update announcements channel\n" +
			"3. If applicable, schedule post mortem\n",
	}
	json, _ := json.Marshal(msg)
	return json
}

func mustGetEnv(key string) string {
	if v, ok := os.LookupEnv(key); !ok {
		panic(fmt.Sprintf("Variable %s is not set", key))
	} else if v == "" {
		panic(fmt.Sprintf("Variable %s is blank", key))
	} else {
		return v
	}
}

func getEnvironmentValues() (string, string, string, string, string, string, string, string, string) {
	return os.Getenv("SF_URL"),
		os.Getenv("SF_USER"),
		os.Getenv("SF_PASSWORD"),
		os.Getenv("SF_TOKEN"),
		os.Getenv("NX_USER"),
		os.Getenv("NX_PASSWORD"),
		os.Getenv("GCP_SERVICE_ACCOUNT_EMAIL"),
		os.Getenv("GCP_SERVICE_ACCOUNT_PRIVATE_KEY"),
		os.Getenv("GDRIVE_FIRE_DOC_FOLDER_ID")
}
