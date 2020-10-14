package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var scopes = []string{"https://www.googleapis.com/auth/drive", "https://www.googleapis.com/auth/documents"}

func getClientFromCredentialsJSON(filename string, scopes ...string) (*http.Client, error) {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	creds, err := google.CredentialsFromJSON(context.Background(), b, scopes...)
	if err != nil {
		return nil, err
	}

	client := oauth2.NewClient(context.Background(), creds.TokenSource)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func jsonDecode(body io.ReadCloser) map[string]interface{} {
	data := make(map[string]interface{})
	json.NewDecoder(body).Decode(&data)
	return data
}

func main() {
	client, err := getClientFromCredentialsJSON("nebo.json", scopes...)
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
