package common

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"reflect"
	"strings"

	"github.com/nlopes/slack"
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