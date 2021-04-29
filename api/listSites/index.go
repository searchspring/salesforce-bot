package listSites

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/kelseyhightower/envconfig"

	common "github.com/searchspring/nebo/api/config"
	"github.com/searchspring/nebo/nextopia"
	"github.com/searchspring/nebo/salesforce"
)

var salesForceDAO salesforce.DAO = nil
var nextopiaDAO nextopia.DAO = nil

var router *mux.Router

func Handler(w http.ResponseWriter, r *http.Request) {
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
	router.HandleFunc("/listSites", wrapGetSitesList(GetSitesList)).Methods(http.MethodGet, http.MethodOptions)
	router.Use(mux.CORSMethodMiddleware(router))
	return router, nil
}

func wrapGetSitesList(apiRequest func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			return
		}
		apiRequest(w, r)
	}
}

// Handler - check routing and call correct methods
func GetSitesList(w http.ResponseWriter, r *http.Request) {
	var env common.EnvVars
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
		log.Print(err.Error())
	}

	nextopiaDAO = nextopia.NewDAO(env.NxUser, env.NxPassword)
	salesForceDAO = salesforce.NewDAO(env.SfURL, env.SfUser, env.SfPassword, env.SfToken)

	listOfSites, err := salesForceDAO.DomainQuery()
	if err != nil {
		common.SendInternalServerError(w, err)
		return
	}

	//fmt.Fprintln(w, "Domains: ", siteIdAndDomains)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(listOfSites)

}
