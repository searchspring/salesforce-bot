package common

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"reflect"
	"strings"

	"github.com/nlopes/slack"
	"github.com/searchspring/nebo/models"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

type EnvVars struct {
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
	MetabaseUser           string `split_words:"true" required:"false"`
	MetabasePassword       string `split_words:"true" required:"false"`
}

// Platforms is a list of platforms in salesforce
var Platforms = []string{
	"3dcart",
	"BigCommerce",
	"CommerceV3",
	"Custom",
	"Magento",
	"Miva",
	"Netsuite",
	"Other",
	"Shopify",
	"Shopify Plus",
	"Yahoo",
}

type SlackDAO interface {
	SendSlackMessage(token string, attachments slack.Attachment, channel string) error
	GetValues() []string
}

type SlackDAOImpl struct{}

func (s *SlackDAOImpl) GetValues() []string {
	return []string{"", ""}
}

func (s *SlackDAOImpl) SendSlackMessage(token string, attachments slack.Attachment, channel string) error {
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

func SendInternalServerError(res http.ResponseWriter, err error) {
	log.Println(err.Error())
	http.Error(res, err.Error(), http.StatusInternalServerError)
}

func FindBlankEnvVars(env EnvVars) []string {
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

// ContainsEmptyString returns true if any of the string variables provided are blank
func ContainsEmptyString(vars ...string) bool {
	for _, v := range vars {
		if v == "" {
			return true
		}
	}
	return false
}

// HTTP Google Client Common Code

// HTTPClient interface
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type Client struct {
	httpClient HTTPClient
	cache      map[string][]byte
}

// Create a new Client
func NewClient(client HTTPClient) *Client {
	return &Client{
		httpClient: client,
		cache:      map[string][]byte{},
	}
}

// AuthorizedGet make a secure request out to the googs
func (c *Client) AuthorizedGet(token string, url string) ([]byte, error) {
	return c.AuthorizedGetWithCache(token, url, true)
}

// AuthorizedGetNoCache make a secure request out to the googs and always hit the live service
func (c *Client) AuthorizedGetNoCache(token string, url string) ([]byte, error) {
	return c.AuthorizedGetWithCache(token, url, false)
}

// AuthorizedGetWithCache make a secure request out to the googs and possibly use a cache.
func (c *Client) AuthorizedGetWithCache(token string, url string, useCache bool) ([]byte, error) {

	if body, ok := c.cache[url]; ok && useCache {
		return body, nil
	}
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, fmt.Errorf("authorization failed - no authorization header")
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("authorization", "Bearer "+token)
	response, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed connecting to %s with error %s", url, err.Error())
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed read response: %s", err.Error())
	}
	if response.StatusCode < 200 || response.StatusCode > 299 {
		log.Println("error from server", string(body))
		return nil, fmt.Errorf("failed to reading from URL - status code: %d - error: %s", response.StatusCode, string(body))
	}
	c.cache[url] = body
	return body, nil
}

// formats AccountInfo into Slack Message

// example formatting here: https://api.slack.com/reference/messaging/attachments
func FormatAccountInfos(accountInfos []*models.AccountInfo, search string) *slack.Msg {
	initialText := "Reps for search: " + search
	if len(accountInfos) == 0 {
		initialText = "No results for: " + search
	}

	p := message.NewPrinter(language.English)

	msg := &slack.Msg{
		ResponseType: slack.ResponseTypeInChannel,
		Text:         initialText,
		Attachments:  []slack.Attachment{},
	}
	globalFamilyMrr := float64(0)
	for _, ai := range accountInfos {
		color := "3A23AD" // Searchspring purple
		if ai.Manager == "unknown" {
			color = "FF0000" // red
		}
		mrr := "unknown"
		if ai.MRR > 0 {
			mrr = p.Sprintf("$%.2f", ai.MRR)
		}
		familymrr := "unknown"
		if ai.FamilyMRR > 0 {
			globalFamilyMrr = ai.FamilyMRR
			familymrr = p.Sprintf("$%.2f", ai.FamilyMRR)
		} else {
			familymrr = p.Sprintf("$%.2f", globalFamilyMrr)
		}

		mrr = mrr + " (Family MRR: " + familymrr + ")"
		loc := ai.City
		if ai.State != "unknown" {
			loc += ", " + ai.State
		}
		text := "Rep: " + ai.Manager + "\nMRR: " + mrr + "\nPlatform: " + ai.Platform + "\nIntegration: " + ai.Integration + "\nProvider: " + ai.Provider + "\nLocation: " + loc
		msg.Attachments = append(msg.Attachments, slack.Attachment{
			Color:      "#" + color,
			Text:       text,
			AuthorName: ai.Website + " (" + ai.Active + ") (SiteId: " + ai.SiteId + ")",
		})
	}
	return msg
}

func FormatPartnerInfos(partnerInfos []*models.PartnerInfo, search string) *slack.Msg {
	initialText := "Partners for search: " + search
	if len(partnerInfos) == 0 {
		initialText = "No results for: " + search
	}

	msg := &slack.Msg{
		ResponseType: slack.ResponseTypeInChannel,
		Text:         initialText,
		Attachments:  []slack.Attachment{},
	}

	for _, partner := range partnerInfos {

		terms := "\nPartner Terms: "
		if partner.PartnerTerms == "unknown" && partner.PartnerTermsNotes == "unknown" {
			terms = ""
		} else {
			if partner.PartnerTerms != "unknown" {
				terms = terms + partner.PartnerTerms + " "
			}
			if partner.PartnerTermsNotes != "unknown" {
				terms = terms + partner.PartnerTermsNotes
			}
		}

		text := "Name: " + partner.Name + "\nType: " + partner.Type + "\nStatus: " + partner.Status +
			"\nOwnerID: " + partner.OwnerID + "\nPartner Type: " + partner.Type + "\nSupported Platforms: " + partner.SupportedPlatforms + terms
		msg.Attachments = append(msg.Attachments, slack.Attachment{
			Color:      "#3A23AD",
			Text:       text,
			AuthorName: partner.Name + " (" + partner.Type + ")",
		})
	}
	return msg
}
