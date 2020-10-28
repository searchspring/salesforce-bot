package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"reflect"
	"strings"
	"time"

	petname "github.com/dustinkirkland/golang-petname"
	"github.com/kelseyhightower/envconfig"
	"github.com/nlopes/slack"
	"searchspring.com/slack/nextopia"
	"searchspring.com/slack/salesforce"
)

type envVars struct {
	DevMode                string `split_words:"true" required:"true"`
	SlackVerificationToken string `split_words:"true" required:"true"`
	SlackOauthToken        string `split_words:"true" required:"true"`
	SfURL                  string `split_words:"true" required:"true"`
	SfUser                 string `split_words:"true" required:"true"`
	SfPassword             string `split_words:"true" required:"true"`
	SfToken                string `split_words:"true" required:"true"`
	NxUser                 string `split_words:"true" required:"true"`
	NxPassword             string `split_words:"true" required:"true"`
	GdriveFireDocFolderID  string `split_words:"true" required:"true"`
}

var salesForceDAO salesforce.DAO = nil
var nextopiaDAO nextopia.DAO = nil

// Handler - check routing and call correct methods
func Handler(w http.ResponseWriter, r *http.Request) {
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
		log.Printf(err.Error())
	}

	s, err := slack.SlashCommandParse(r)
	if err != nil {
		sendInternalServerError(w, err)
		return
	}

	if !s.ValidateToken(env.SlackVerificationToken) {
		err := errors.New("slack verification failed")
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	nextopiaDAO = nextopia.NewDAO(env.NxUser, env.NxPassword)
	salesForceDAO = salesforce.NewDAO(env.SfURL, env.SfUser, env.SfPassword, env.SfToken)

	w.Header().Set("Content-type", "application/json")
	switch s.Command {
	case "/rep", "/alpha-nebo", "/nebo":
		if strings.TrimSpace(s.Text) == "help" || strings.TrimSpace(s.Text) == "" {
			writeHelpNebo(w)
			return
		}
		if salesForceDAO == nil {
			sendInternalServerError(w, errors.New("missing required Salesforce credentials"))
			return
		}
		responseJSON, err := salesForceDAO.Query(s.Text)
		if err != nil {
			sendInternalServerError(w, err)
			return
		}
		w.Write(responseJSON)
		return

	case "/fire", "/firetest":
		if strings.TrimSpace(s.Text) == "help" {
			writeHelpFire(w)
			return
		}
		fireResponse(env.GdriveFireDocFolderID, s.Text, s.ResponseURL)
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
			sendInternalServerError(w, errors.New("missing required Nextopia credentials"))
			return
		}
		responseJSON, err := nextopiaDAO.Query(s.Text)
		if err != nil {
			sendInternalServerError(w, err)
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
			sendInternalServerError(w, errors.New("missing required Salesforce credentials"))
			return
		}
		responseJSON, err := salesForceDAO.IDQuery(s.Text)
		if err != nil {
			sendInternalServerError(w, err)
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
		sendInternalServerError(w, errors.New("unknown slash command "+s.Command))
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
		Text:         "Fire usage:\n`/fire <fire title>` - generate a fire checklist to handle the fire",
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

func fireResponse(folderID string, title string, responseURL string) {
	checklist := fireChecklist(folderID, title)
	postSlackMessage(responseURL, slack.ResponseTypeInChannel, checklist)
}

func postSlackMessage(responseURL string, responseType string, text string) error {
	msg := &slack.Msg{
		ResponseType: responseType,
		Text:         text,
	}
	json, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	_, err = http.Post(responseURL, "application/json", bytes.NewBuffer(json))
	return err
}

func cleanFireTitle(title string) string {
	title = strings.TrimSpace(title)
	if title == "" {
		title = "New Fire"
	}
	return title
}

func fireChecklist(folderID string, title string) string {
	title = cleanFireTitle(title)
	link := fmt.Sprintf("create new fire doc <https://drive.google.com/drive/folders/%s|here>", folderID)
	text := "...\n:fire:*" + title + "*:fire:\n" +
		"1. Designate fire leader\n" +
		"2. Designate fire doc maintainer and " + link + "\n" +
		"3. Post link to fire doc in this message's thread" +
		"4. If a real fire - make an announcement in the announcements channel \"There is a fire and engineering is investigating, updates will be posted in a thread on this message\"\n" +
		"5. Post a link to the fire document in the announcement channel thread\n" +
		"6. Designate helper(s) to update announcement\n" +
		"7. Fight! " + getMeetLink(title) + "\n" +
		"8. Use `/firedown` when the fire is out"
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

func sendInternalServerError(w http.ResponseWriter, err error) {
	log.Println(err.Error())
	http.Error(w, err.Error(), http.StatusInternalServerError)
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
