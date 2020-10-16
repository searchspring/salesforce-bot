package gapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"io"
	"log"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
)

var googleScopes = []string{"https://www.googleapis.com/auth/drive", "https://www.googleapis.com/auth/documents"}

func getClientFromEnvVars(email string, privateKey string, scopes ...string) (*http.Client, error) {
	conf := &jwt.Config{
		Email:      email,
		PrivateKey: []byte(privateKey),
		Scopes:     scopes,
		TokenURL:   google.JWTTokenURL,
	}
	client := conf.Client(oauth2.NoContext)
	return client, nil
}

func jsonDecode(body io.ReadCloser) map[string]interface{} {
	data := make(map[string]interface{})
	json.NewDecoder(body).Decode(&data)
	return data
}

//GetDoc is a test func to get doc
func getDoc(email string, privateKey string) {
	client, err := getClientFromEnvVars(email, privateKey, googleScopes...)
	if err != nil {
		log.Fatal(err)
	}

	// TODO - Instead of GET, figure out how to POST a new doc that uses desired format in desired location
	// Google Doc Create: https://developers.google.com/docs/api/reference/rest/v1/documents/create
	// Should be created here: https://drive.google.com/drive/folders/1CgRBFg2CTbvjLp57yfoUOD_OZlaVxOht
	resp, err := client.Get("https://docs.googleapis.com/v1/documents/1vb1gofwabN4Mml-li4O9R82VwD3DPHSQ5NAgs1md6Ik")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(jsonDecode(resp.Body))
}
