package listSites

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/grokify/simplego/fmt/fmtutil"
	"github.com/kelseyhightower/envconfig"

	common "github.com/searchspring/nebo/api/config"
	"github.com/searchspring/nebo/google"
	"github.com/searchspring/nebo/metabase"
	"github.com/searchspring/nebo/salesforce"
)

var router *mux.Router
var env common.EnvVars

func Handler(w http.ResponseWriter, r *http.Request) {
	err := envconfig.Process("", &env)
	if err != nil {
		common.SendInternalServerError(w, err)
		return
	}

	blanks := common.FindBlankEnvVars(env)
	if len(blanks) > 0 {
		err := fmt.Errorf("the following env vars are blank: %s", strings.Join(blanks, ", "))
		if env.DevMode != "development" {
			common.SendInternalServerError(w, err)
			return
		}
		log.Println(err.Error())
	}

	if router == nil {
		r, err := CreateRouter()
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		router = r
	}
	router.ServeHTTP(w, r)
}

func CreateRouter() (*mux.Router, error) {
	router := mux.NewRouter()
	googleDAO := google.NewDAO(common.NewClient(&http.Client{}))
	salesforceDAOReal := salesforce.NewDAO(env.SfURL, env.SfUser, env.SfPassword, env.SfToken)
	metabaseDAOReal, authInfo, err := metabase.NewDAO("https://metabase.kube.searchspring.io/", env.MbUser, env.MbPassword, "")
	fmtutil.PrintJSON(authInfo)
	if err != nil {
		fmt.Println("Meta Error: ", err)
	}
	router.HandleFunc("/listSites", wrapWithAuthorizedCheck(googleDAO.CheckUserLoggedIn, GetSitesList, salesforceDAOReal, metabaseDAOReal)).Methods(http.MethodGet, http.MethodOptions)
	router.Use(mux.CORSMethodMiddleware(router))
	return router, nil
}

func wrapWithAuthorizedCheck(checkUserLoggedIn func(authorizationToken string) (string, error), apiRequest func(w http.ResponseWriter, r *http.Request, salesforceDAOReal salesforce.DAO, metabaseDAOReal metabase.DAO), salesforceDAOReal salesforce.DAO, metabaseDAOReal metabase.DAO) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			return
		}
		
		authorization := r.Header.Get("Authorization")
		if email, err := checkUserLoggedIn(authorization); err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		} else {
			if !strings.HasSuffix(email, "@searchspring.com") {
				http.Error(w, "must have searchspring.com email address to use this systsem", http.StatusForbidden)
				return
			}
			
			apiRequest(w, r, salesforceDAOReal, metabaseDAOReal)
		}
	}
}

// Handler - check routing and call correct methods
func GetSitesList(w http.ResponseWriter, r *http.Request, salesforceApi salesforce.DAO, metabaseApi metabase.DAO) {

	listOfSites, err := salesforceApi.DomainQuery()
	if err != nil {
		common.SendInternalServerError(w, err)
		return
	}
	
	data, err := metabaseApi.QueryAll()
	if err != nil {
		common.SendInternalServerError(w, err)
		return
	}
	fmtutil.PrintJSON(data)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(listOfSites)

}
