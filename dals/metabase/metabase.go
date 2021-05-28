package metabase

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"sort"
	"strings"

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
	account := &NpsInfo{}

	if len(result.Rows) > 0 {
		for i, colInfo := range result.Cols {
			value := result.Rows[0][i]
			switch colInfo.Name {
			case "mrr":
				account.MRR = float64(0)
				if value != nil {
					account.MRR = value.(float64)
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
	} else {
		account.MRR = float64(-1)
		account.FamilyMRR = float64(-1)
		account.Manager = "No company found"
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

	accounts = cleanAccounts(accounts)
	if !isPlatformSearch(search) {
		accounts = sortAccounts(accounts, "website")
	}
	accounts = truncateAccounts(accounts)

	accounts = sortAccounts(accounts, "mrr")
	fmtutil.PrintJSON(accounts)
	//msg := formatcommon.AccountInfos(accounts, search)
	return accounts, nil
}

// cleaning account arrays

func truncateAccounts(accounts []*common.AccountInfo) []*common.AccountInfo {
	truncated := []*common.AccountInfo{}
	for i, account := range accounts {
		if i == 20 {
			break
		}
		truncated = append(truncated, account)
	}
	return truncated
}

func isPlatformSearch(search string) bool {
	for _, platform := range common.Platforms {
		if strings.EqualFold(search, platform) {
			return true
		}
	}
	return false
}

func cleanAccounts(accounts []*common.AccountInfo) []*common.AccountInfo {
	for _, account := range accounts {
		w := account.Website
		if strings.HasPrefix(w, "http://") || strings.HasPrefix(w, "https://") {
			w = w[strings.Index(w, ":")+3:]
		}
		if strings.HasPrefix(w, "www.") {
			w = w[4:]
		}
		if strings.HasSuffix(w, "/") {
			w = w[0 : len(w)-1]
		}
		account.Website = w
	}
	return accounts
}

func sortAccounts(accounts []*common.AccountInfo, sortType string) []*common.AccountInfo {
	sort.Slice(accounts, func(i, j int) bool {
		if sortType == "website" {
			return len(accounts[i].Website) < len(accounts[j].Website)
		} else {
			return accounts[i].MRR > accounts[j].MRR
		}
	})
	return accounts
}

// helper functions

func (s *DAOImpl) GetSearchKey() string {
	return ""
}
