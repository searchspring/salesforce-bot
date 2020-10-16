package api

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	petname "github.com/dustinkirkland/golang-petname"
	"github.com/nlopes/slack"
	"searchspring.com/slack/google"
	"searchspring.com/slack/nextopia"
	"searchspring.com/slack/salesforce"
)

var salesForceDAO salesforce.DAO = nil
var nextopiaDAO nextopia.DAO = nil

// Handler - check routing and call correct methods
func Handler(res http.ResponseWriter, req *http.Request) {
	slackVerificationCode, slackOauthToken, sfURL, sfUser, sfPassword, sfToken, nxUser, nxPassword, err := getEnvironmentValues()
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		res.Write([]byte(err.Error()))
		return
	}

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

	if salesForceDAO == nil {
		salesForceDAO, err = salesforce.NewDAO(sfURL, sfUser, sfPassword, sfToken)
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			res.Write([]byte("salesforce client was not created successfully: " + err.Error()))
			return
		}
	}

	if nextopiaDAO == nil {
		nextopiaDAO = nextopia.NewDAO(nxUser, nxPassword)
	}

	res.Header().Set("Content-type", "application/json")
	switch s.Command {
	case "/rep", "/alpha-nebo", "/nebo":
		if strings.TrimSpace(s.Text) == "help" || strings.TrimSpace(s.Text) == "" {
			writeHelpNebo(res)
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
		res.Write(fireResponse())
		return

	case "/firedown":
		res.Write(fireDownResponse())
		return

	case "/neboidnx", "/neboid":
		if strings.TrimSpace(s.Text) == "help" || strings.TrimSpace(s.Text) == "" {
			writeHelpNeboid(res)
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

func fireResponse() []byte {
	google.GetDoc()
	msg := &slack.Msg{
		ResponseType: slack.ResponseTypeInChannel,
		Text: "1. Create Fire document - https://docs.google.com/document/create?usp=drive_web&ouid=117735186481765666461&folder=1CgRBFg2CTbvjLp57yfoUOD_OZlaVxOht\n" +
			"2. Designate Fire Leader\n" +
			"3. If a real fire - make an announcement in the annoucements channel \"There is a fire and engineering is investigating, updates will be posted in a thread on this message\"\n" +
			"4. Post a link to the fire document in the announcement channel thread\n" +
			"5. Designate helper(s) to update document\n" +
			"6. Designate helper(s) to update announcement\n" +
			"7. Fight!\n" +
			getMeetLink(""),
	}
	json, _ := json.Marshal(msg)
	return json
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

func getEnvironmentValues() (string, string, string, string, string, string, string, string, error) {
	if os.Getenv("SLACK_VERIFICATION_TOKEN") == "" {
		return "", "", "", "", "", "", "", "", fmt.Errorf("Must set: SLACK_VERIFICATION_TOKEN")
	}
	if os.Getenv("SLACK_OAUTH_TOKEN") == "" {
		return "", "", "", "", "", "", "", "", fmt.Errorf("Must set: SLACK_OAUTH_TOKEN")
	}
	if os.Getenv("SF_URL") == "" {
		return "", "", "", "", "", "", "", "", fmt.Errorf("Must set: SF_URL")
	}
	if os.Getenv("SF_USER") == "" {
		return "", "", "", "", "", "", "", "", fmt.Errorf("Must set: SF_USER")
	}
	if os.Getenv("SF_PASSWORD") == "" {
		return "", "", "", "", "", "", "", "", fmt.Errorf("Must set: SF_PASSWORD")
	}
	if os.Getenv("SF_TOKEN") == "" {
		return "", "", "", "", "", "", "", "", fmt.Errorf("Must set: SF_TOKEN")
	}
	if os.Getenv("NX_USER") == "" {
		return "", "", "", "", "", "", "", "", fmt.Errorf("Must set: NX_USER")
	}
	if os.Getenv("NX_PASSWORD") == "" {
		return "", "", "", "", "", "", "", "", fmt.Errorf("Must set: NX_PASSWORD")
	}
	return os.Getenv("SLACK_VERIFICATION_TOKEN"),
		os.Getenv("SLACK_OAUTH_TOKEN"),
		os.Getenv("SF_URL"),
		os.Getenv("SF_USER"),
		os.Getenv("SF_PASSWORD"),
		os.Getenv("SF_TOKEN"),
		os.Getenv("NX_USER"),
		os.Getenv("NX_PASSWORD"),
		nil
}
