package listSites

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/kelseyhightower/envconfig"

	common "github.com/searchspring/nebo/api/config"
	"github.com/searchspring/nebo/google"
	"github.com/searchspring/nebo/metabase"
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
	metabaseDAO, err := metabase.NewDAO("https://metabase.kube.searchspring.io/", env.MetabaseUser, env.MetabasePassword, "")
	if err != nil {
		fmt.Println("Metabase Error: ", err)
	}
	router.HandleFunc("/listSites", wrapWithAuthorizedCheck(googleDAO.CheckUserLoggedIn, GetSitesList, metabaseDAO)).Methods(http.MethodGet, http.MethodOptions)
	router.Use(mux.CORSMethodMiddleware(router))
	return router, nil
}

func wrapWithAuthorizedCheck(checkUserLoggedIn func(authorizationToken string) (string, error), apiRequest func(w http.ResponseWriter, r *http.Request, metabaseDAOReal metabase.DAO), metabaseDAOReal metabase.DAO) func(w http.ResponseWriter, r *http.Request) {
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
			
			apiRequest(w, r, metabaseDAOReal)
		}
	}
}

func GetSitesList(w http.ResponseWriter, r *http.Request, metabaseAPI metabase.DAO) {
	data, err := metabaseAPI.QueryAll()
	if err != nil {
		common.SendInternalServerError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(data)
}