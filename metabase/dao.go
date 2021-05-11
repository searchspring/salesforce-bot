package metabase

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/grokify/go-metabase/metabase"
	"github.com/grokify/go-metabase/metabaseutil"
	mo "github.com/grokify/oauth2more/metabase"
)

type DAO interface {
	QueryAll() ([]byte, error)
}

type DAOImpl struct {
	Client *metabase.APIClient
}

type domainAndID struct {
	Website string
	SiteId string
}

const domainFields = "name, trackingCode, active"

func NewDAO(mbURL string, mbUser string, mbPassword string, mbToken string) (DAO, error) {

	config := mo.Config{
		BaseURL:       mbURL,
		Username:      mbUser,
		Password:      mbPassword,
		SessionID:     mbToken,
		TLSSkipVerify: true,
	}

	apiClient, _, err := metabaseutil.NewApiClient(config)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	return &DAOImpl{
		Client: apiClient,
	}, nil
}

func (s *DAOImpl) QueryAll() ([]byte, error) {
	var databaseId int64 = 5
	data := []domainAndID{}

	q := "SELECT " + domainFields + " " + "FROM websites WHERE active = true"

	info, resp, err := metabaseutil.QuerySQL(s.Client, databaseId, q)
	if err != nil {
		log.Fatal(err)
		return []byte{}, nil
	} else if resp.StatusCode >= 300 {
		log.Println(fmt.Sprintf("STATUS_CODE [%v]", resp.StatusCode))
		return []byte{}, err
	} else if info.RowCount == 2000 {
		log.Println("DATABASE HAS SURPASSED QUERY LIMIT")
		return []byte{}, err
	}

	rows := info.Data.Rows
	
	for _, v := range rows {
		data = append(data, domainAndID{
			Website: fmt.Sprintf("%s", v[0]),
			SiteId: fmt.Sprintf("%s", v[1]),
		})
	}

	return json.Marshal(data)
}
