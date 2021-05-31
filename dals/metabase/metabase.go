package metabase

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"

	"github.com/grokify/go-metabase/metabase"
	"github.com/grokify/go-metabase/metabaseutil"
	metabaseOAuth "github.com/grokify/oauth2more/metabase"
	"github.com/grokify/simplego/fmt/fmtutil"
	common "github.com/searchspring/nebo/common"
)

type DAO interface {
	QueryAll() ([]byte, error)
	QueryNPS(string) (*NpsInfo, error)
	Query(string) ([]*common.AccountInfo, error)
	StructFromResult(*metabase.DatasetQueryResultsData) (*NpsInfo, error)
	ResultToMessage(string, *metabase.DatasetQueryResultsData) ([]*common.AccountInfo, error)
	GetSearchKey() string
}

type DAOImpl struct {
	Client *metabase.APIClient
	Key string
}

type NpsInfo struct {
	Manager   string
	MRR       float64
	FamilyMRR float64
}

type DomainAndID struct {
	Website string
	SiteId  string
}

const databaseId = 5

const domainFields = "name, trackingCode, active"
const npsFields = "active, mrr, familyMrr, csm, name"
const accountFields = "domainName, csm, active, familyMrr, mrr, platform_smart, integrationType, trackingCode, city, state"

func NewDAO(metabaseURL string, metabaseUser string, metabasePassword string, metabaseToken string) DAO {

	config := metabaseOAuth.Config{
		BaseURL:       metabaseURL,
		Username:      metabaseUser,
		Password:      metabasePassword,
		SessionID:     metabaseToken,
		TLSSkipVerify: true,
	}

	apiClient, _, err := metabaseutil.NewApiClient(config)
	if err != nil {
		log.Println(err.Error())
		return nil
	}

	return &DAOImpl{
		Client: apiClient,
	}
}

func (s *DAOImpl) QueryAll() ([]byte, error) {
	data := []DomainAndID{}

	q := "SELECT " + domainFields + " " + "FROM websites WHERE active"

	info, resp, err := metabaseutil.QuerySQL(s.Client, databaseId, q)
	if err != nil {
		log.Fatal(err)
		return []byte{}, err
	} else if resp.StatusCode >= 300 {
		log.Println(fmt.Sprintf("STATUS_CODE [%v]", resp.StatusCode))
		return []byte{}, err
	} else if info.RowCount == 2000 {
		log.Println("DATABASE HAS SURPASSED QUERY LIMIT")
		return []byte{}, err
	}

	rows := info.Data.Rows

	for _, v := range rows {
		data = append(data, DomainAndID{
			Website: fmt.Sprintf("%s", v[0]),
			SiteId:  fmt.Sprintf("%s", v[1]),
		})
	}

	return json.Marshal(data)
}

func (s *DAOImpl) QueryNPS(search string) (*NpsInfo, error) {
	reg, err := regexp.Compile("[^a-zA-Z0-9_.-]+")
	if err != nil {
		return nil, err
	}

	sanitized := reg.ReplaceAllString(search, "")

	q := "SELECT " + npsFields + " " +
		"FROM websites WHERE active " +
		"AND name LIKE '%" + sanitized + "%' ORDER BY mrr DESC"

	info, resp, err := metabaseutil.QuerySQL(s.Client, databaseId, q)
	if err != nil {
		log.Fatal(err)
		return &NpsInfo{}, err
	} else if resp.StatusCode >= 300 {
		log.Println(fmt.Sprintf("STATUS_CODE [%v]", resp.StatusCode))
		return &NpsInfo{}, err
	}

	return s.StructFromResult(&info.Data)
}

func (s *DAOImpl) Query(search string) ([]*common.AccountInfo, error) {
	reg, err := regexp.Compile("[^a-zA-Z0-9_.-]+")
	if err != nil {
		return nil, err
	}

	sanitized := reg.ReplaceAllString(search, "")

	q := "SELECT " + accountFields + " " +
		"FROM websites WHERE active AND !presales AND !sandbox " +
		"AND (name LIKE '%" + sanitized + "%' OR platform_smart LIKE '%" + sanitized +
		"%' OR trackingCode = '" + sanitized + "') ORDER BY mrr DESC"
	info, resp, err := metabaseutil.QuerySQL(s.Client, databaseId, q)
	if err != nil {
		log.Fatal(err)
		return []*common.AccountInfo{}, err
	} else if resp.StatusCode >= 300 {
		log.Println(fmt.Sprintf("STATUS_CODE [%v]", resp.StatusCode))
		return []*common.AccountInfo{}, err
	}

	return s.ResultToMessage(sanitized, &info.Data)
}

// formatting results

func (s *DAOImpl) StructFromResult(result *metabase.DatasetQueryResultsData) (*NpsInfo, error) {
	account := &NpsInfo{
		MRR: float64(-1),
		FamilyMRR: float64(-1),
		Manager: "Unknown",
	}

	for i := range result.Rows {
		for k, colInfo := range result.Cols {
			value := result.Rows[i][k]
			switch colInfo.Name {
			case "mrr":
				account.MRR = float64(0)
				if value != nil {
					account.MRR = value.(float64)
					fmt.Println("MRR: ", value)
				}
			case "familyMrr":
				account.FamilyMRR = float64(0)
				if value != nil {
					account.FamilyMRR = value.(float64)
				}
			case "csm":
				account.Manager = "Unknown"
				if value != nil {
					account.Manager = fmt.Sprint(value)
				}
			}
		}
		if account.MRR != 0 || account.FamilyMRR != 0 {
			break
		}
	} 

	return account, nil
}

func (s *DAOImpl) ResultToMessage(search string, result *metabase.DatasetQueryResultsData) ([]*common.AccountInfo, error) {
	accounts := []*common.AccountInfo{}
	fmtutil.PrintJSON(result)
	if len(result.Rows) > 0 {
		for i := range result.Rows {
			website := "unknown"
			csm := "unknown"
			active := "Active"
			mrr := float64(-1)
			familymrr := float64(-1)
			platform := "unknown"
			integration := "unknown"
			provider := "Searchspring"
			siteId := ""
			city := "unknown"
			state := ""
			for k, colInfo := range result.Cols {
				value := result.Rows[i][k]
				switch colInfo.Name {
				case "domainName":
					if value != nil {
						website = fmt.Sprint(value)
					}
				case "csm":
					if value != nil {
						csm = fmt.Sprint(value)
					}
				case "mrr":
					if value != nil {
						mrr = value.(float64)
					}
				case "familyMrr":
					if value != nil {
						familymrr = value.(float64)
					}
				case "platform_smart":
					if value != nil {
						platform = fmt.Sprint(value)
					}
				case "integrationType":
					if value != nil {
						integration = fmt.Sprint(value)
					}
				case "trackingCode":
					if value != nil {
						siteId = fmt.Sprint(value)
					}
				case "city":
					if value != nil {
						city = fmt.Sprint(value)
					}
				case "state":
					if value != nil {
						state = fmt.Sprint(value)
					}
				}
			}
			accounts = append(accounts, &common.AccountInfo{
				Website:     website,
				Manager:     csm,
				Active:      active,
				MRR:         mrr,
				FamilyMRR:   familymrr,
				Platform:    platform,
				Integration: integration,
				Provider:    provider,
				SiteId:      siteId,
				City:        city,
				State:       state,
			})
			if i > 20 {
				break
			}
		}
	}

	return accounts, nil
}

// helper functions

func (s *DAOImpl) GetSearchKey() string {
	return s.Key
}
