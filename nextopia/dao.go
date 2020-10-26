package nextopia

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/nlopes/slack"
	"searchspring.com/slack/validator"
)

// DAO acts as the nextopia DAO
type DAO interface {
	Query(query string) ([]byte, error)
}

// DAOImpl defines the properties of the DAO
type DAOImpl struct {
	Client    *http.Client
	User      string
	Password  string
	Customers map[string][]string
}

// NewDAO returns the nextopia DAO
func NewDAO(nxUser string, nxPassword string) DAO {
	if validator.ContainsEmptyString(nxUser, nxPassword) {
		return nil
	}
	return &DAOImpl{
		User:     nxUser,
		Password: nxPassword,
		Client:   http.DefaultClient,
	}
}

// {"result":"success","data":[["50ae9d89c8d2879b028227bad4ad0220","54762cbb0dc2475aa35485a26c79cf41","","INACTIVE","","Trial","n\/a","32-bit","legacy","0000-00-00 00:00:00"],["00b5a6084631611ae5ff7e6d037c7a1e","b913c134faf624e8e26b2f841a346352","ec_101inkscom","INACTIVE","101inks.com","Trial","n\/a","unset","legacy","2015-10-07 14:23:28"],["ee33869e9bdf9371963dca152444c212","6130a8c8e4e4543953af4118186b145f","ec_123djcom","ACTIVE","123dj.com","Professional","n\/a","unset","v1.5.1","2020-06-17 18:25:59"],["3502dc102d967598693d671cd0a82d68","7213b73fa377d8572ae0731e6aa0d3f1","ec_123healthshopcouk","INACTIVE","123healthshop.co.uk","Trial","n\/a","unset","v2.0","2018-06-16 10:37:52"],["c3f3888a9c554f58ccd420a6491284a4","cec4a2a2d6680cf83c3cf8685176e6c5","ec_123securityproductscom","ACTIVE","123securityproducts.com","Professional","n\/a","watson","v2.0","2020-06-11 14:19:40"],
type resultData struct {
	Data [][]string `json:"data"`
}

// Query queries the nextopia client report DB using provided query string
func (d *DAOImpl) Query(query string) ([]byte, error) {
	if d.Customers == nil {
		res, err := d.Client.Get("http://" + d.User + ":" + d.Password + "@client-report.nxtpd.com/api/data-table.php?table=accounts&_=1592606239141")
		if err != nil {
			return nil, err
		}
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}
		resultData := &resultData{}

		err = json.Unmarshal(body, resultData)
		if err != nil {
			return nil, err
		}
		d.Customers = map[string][]string{}
		for _, row := range resultData.Data {
			d.Customers[row[0]] = row
		}
	}
	msg := d.findMatch(query)
	return json.Marshal(msg)
}

const ID1 = 0
const ID2 = 1
const NAME = 2
const URL = 4
const TYPE = 5
const VERSION = 7
const SYSTEM = 8

func (d *DAOImpl) findMatch(query string) *slack.Msg {
	msg := &slack.Msg{
		ResponseType: slack.ResponseTypeInChannel,
		Text:         "matches",
		Attachments:  []slack.Attachment{},
	}
	for _, value := range d.Customers {
		if matches(value, query) {
			color := "3A23AD" // Searchspring purple
			text := "URL: " + value[URL] +
				"\nID 1: " + value[ID1] +
				"\nID 2: " + value[ID2] +
				"\nType: " + value[TYPE] +
				"\nVersion: " + value[VERSION] + ", System: " + value[SYSTEM]

			msg.Attachments = append(msg.Attachments, slack.Attachment{
				Color:      "#" + color,
				Text:       text,
				AuthorName: value[NAME],
			})
		}
		if len(msg.Attachments) > 100 {
			break
		}
	}
	if len(msg.Attachments) == 0 {
		msg.Text = "No Matches :("
	}
	return msg
}

func matches(customer []string, query string) bool {
	if strings.HasPrefix(customer[ID1], query) {
		return true
	}
	if strings.HasPrefix(customer[ID2], query) {
		return true
	}
	if strings.Contains(customer[NAME], query) {
		return true
	}
	return false
}
