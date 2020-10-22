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
	"github.com/nlopes/slack"
	"searchspring.com/slack/gapi"
	"searchspring.com/slack/nextopia"
	"searchspring.com/slack/salesforce"
)

var salesForceDAO salesforce.DAO = nil
var nextopiaDAO nextopia.DAO = nil

func containsEmptyString(vars ...string) bool {
	for _, v := range vars {
		if v == "" {
			return true
		}
	}
	return false
}

// Handler - check routing and call correct methods
func Handler(res http.ResponseWriter, req *http.Request) {
	slackVerificationCode := mustGetEnv("SLACK_VERIFICATION_TOKEN")
	slackOauthToken := mustGetEnv("SLACK_OAUTH_TOKEN")
	sfURL, sfUser, sfPassword, sfToken, nxUser, nxPassword, gcpEmail, gcpPrivateKey, fireDocFolderID := getEnvironmentValues()

	s, err := slack.SlashCommandParse(req)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		res.Write([]byte(err.Error()))
		return
	}

	if !s.ValidateToken(slackVerificationCode) {
		res.WriteHeader(http.StatusUnauthorized)
		res.Write([]byte("slack verification failed"))
		return
	}

	salesForceVars := []string{sfURL, sfUser, sfPassword, sfToken}
	if salesForceDAO == nil && !containsEmptyString(salesForceVars...) {
		salesForceDAO, err = salesforce.NewDAO(sfURL, sfUser, sfPassword, sfToken)
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			res.Write([]byte("salesforce client was not created successfully: " + err.Error()))
			return
		}
	}

	nextopiaVars := []string{nxUser, nxPassword}
	if nextopiaDAO == nil && !containsEmptyString(nextopiaVars...) {
		nextopiaDAO = nextopia.NewDAO(nxUser, nxPassword)
	}

	res.Header().Set("Content-type", "application/json")
	switch s.Command {
	case "/rep", "/alpha-nebo", "/nebo":
		if strings.TrimSpace(s.Text) == "help" || strings.TrimSpace(s.Text) == "" {
			writeHelpNebo(res)
			return
		}
		if salesForceDAO == nil {
			res.WriteHeader(http.StatusInternalServerError)
			res.Write([]byte("Missing required Salesforce credentials."))
			return
		}
		responseJSON, err := salesForceDAO.Query(s.Text)
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			res.Write([]byte(err.Error()))
			return
		}
		res.Write(responseJSON)
		return

	case "/fire":
		if strings.TrimSpace(s.Text) == "help" {
			writeHelpFire(res)
			return
		}
		// We only have 3 seconds to initially respond
		// https://api.slack.com/interactivity/slash-commands#responding_to_commands
		// So we ACK before doing our work because Google APIs can be slow enough
		// that slack will drop our connection before we finish doing everything and responding
		fireTitle := cleanFireTitle(s.Text)
		res.Write([]byte("New Fire: :fire:*" + fireTitle + "*:fire:\nCreating fire doc & checklist now...\n"))
		go fireResponse(gcpEmail, gcpPrivateKey, fireDocFolderID, fireTitle, s.ResponseURL)
		return

	case "/firetest":
		if strings.TrimSpace(s.Text) == "help" {
			writeHelpFire(res)
			return
		}
		fireTitle := cleanFireTitle(s.Text)
		res.Write([]byte("New Fire: :fire:*" + fireTitle + "*:fire:\nCreating fire doc & checklist now...\n"))
		go fireResponse(gcpEmail, gcpPrivateKey, fireDocFolderID, fireTitle, s.ResponseURL)
		return

	case "/firedown":
		res.Write(fireDownResponse())
		return

	case "/neboidnx", "/neboid":
		if strings.TrimSpace(s.Text) == "help" || strings.TrimSpace(s.Text) == "" {
			writeHelpNeboid(res)
			return
		}
		if nextopiaDAO == nil {
			res.WriteHeader(http.StatusInternalServerError)
			res.Write([]byte("Missing required Nextopia credentials."))
			return
		}
		responseJSON, err := nextopiaDAO.Query(s.Text)
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			res.Write([]byte(err.Error()))
			return
		}
		res.Write(responseJSON)
		return

	case "/neboidss":
		if strings.TrimSpace(s.Text) == "help" || strings.TrimSpace(s.Text) == "" {
			writeHelpNeboid(res)
			return
		}
		if salesForceDAO == nil {
			res.WriteHeader(http.StatusInternalServerError)
			res.Write([]byte("Missing required Salesforce credentials."))
			return
		}
		responseJSON, err := salesForceDAO.IDQuery(s.Text)
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			res.Write([]byte(err.Error()))
			return
		}
		res.Write(responseJSON)
		return

	case "/feature":
		if strings.TrimSpace(s.Text) == "help" || strings.TrimSpace(s.Text) == "" {
			writeHelpFeature(res)
			return
		}
		sendSlackMessage(slackOauthToken, s.Text, s.UserID)
		responseJSON := featureResponse(s.Text)
		res.Write(responseJSON)
		return

	case "/meet":
		if strings.TrimSpace(s.Text) == "help" {
			writeHelpMeet(res)
			return
		}
		responseJSON := meetResponse(s.Text)
		res.Write(responseJSON)
		return

	case "/meettest":
		if strings.TrimSpace(s.Text) == "help" {
			writeHelpMeet(res)
			return
		}
		responseJSON := meetResponse(s.Text)
		res.Write(responseJSON)
		return

	default:
		res.WriteHeader(http.StatusInternalServerError)
		res.Write([]byte("unknown slash command " + s.Command))
		return
	}
}

func writeHelpFeature(res http.ResponseWriter) {
	msg := &slack.Msg{
		ResponseType: slack.ResponseTypeEphemeral,
		Text:         "Feature usage:\n`/feature description of feature required` - submits a feature to the product team\n`/feature help` - this message",
	}
	json, _ := json.Marshal(msg)
	res.Write(json)
}

func writeHelpFire(res http.ResponseWriter) {
	msg := &slack.Msg{
		ResponseType: slack.ResponseTypeEphemeral,
		Text:         "Fire usage:\n`/fire Fire Title` - start a new fire doc named 'Fire Title' and generate a fire checklist to handle the fire",
	}
	json, _ := json.Marshal(msg)
	res.Write(json)
}

func writeHelpNebo(res http.ResponseWriter) {
	platformsJoined := strings.ToLower(strings.Join(salesforce.Platforms, ", "))
	msg := &slack.Msg{
		ResponseType: slack.ResponseTypeEphemeral,
		Text:         "Nebo usage:\n`/nebo shoes` - find all customers with shoe in the name\n`/nebo shopify` - show {" + platformsJoined + "} clients sorted by MRR\n`/nebo help` - this message",
	}
	json, _ := json.Marshal(msg)
	res.Write(json)
}
func writeHelpNeboid(res http.ResponseWriter) {
	msg := &slack.Msg{
		ResponseType: slack.ResponseTypeEphemeral,
		Text:         "Neboid usage:\n`/neboid <id prefix>` - find all customers with an id that starts with this prefix\n`/neboid help` - this message",
	}
	json, _ := json.Marshal(msg)
	res.Write(json)
}

func writeHelpMeet(res http.ResponseWriter) {
	msg := &slack.Msg{
		ResponseType: slack.ResponseTypeEphemeral,
		Text:         "Meet usage:\n`/meet` - generate a random meet\n`/meet name` - generate a meet with a name\n`/meet help` - this message",
	}
	json, _ := json.Marshal(msg)
	res.Write(json)
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

func fireResponse(gcpEmail string, gcpPrivateKey string, fireDocFolderID string, title string, responseURL string) {
	documentID, err := generateFireDoc(gcpEmail, gcpPrivateKey, fireDocFolderID, title)
	msg := &slack.Msg{
		ResponseType: slack.ResponseTypeInChannel,
		Text:         fireChecklist(documentID, title, fireDocFolderID, err),
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

func generateFireDoc(gcpEmail string, gcpPrivateKey string, fireDocFolderID string, title string) (string, error) {
	client := gapi.GetGoogleAPIClient(gcpEmail, gcpPrivateKey, gapi.Scopes...)

	documentID, err := gapi.CreateFireDoc(client, title)
	if err != nil {
		log.Println(err)
		return "", errors.New("Error creating fire doc")
	}

	err = gapi.AssignParentFolder(client, documentID, fireDocFolderID)
	if err != nil {
		log.Println(err)
		return "", errors.New("Unable to move fire doc into correct folder in GDrive")
	}

	err = gapi.WriteDoc(client, documentID)
	if err != nil {
		log.Println(err)
		// In this case we can still use the created doc so there is an error and a documentID returned
		return documentID, errors.New("Unable to write default content to fire doc")
	}
	return documentID, nil
}

func fireChecklist(documentID string, title string, fireDocFolderID string, err error) string {
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
