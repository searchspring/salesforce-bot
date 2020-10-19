package gapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"

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

//CreateFireDoc creates a fire document in the provided Google Drive folder
func CreateFireDoc(email string, privateKey string) {
	client, err := getClientFromEnvVars(email, privateKey, googleScopes...)
	if err != nil {
		log.Fatal(err)
	}

	requestBody, err := json.Marshal(map[string]interface{}{
		"title": fmt.Sprintf("%s Fire Title", time.Now().Format(time.RFC3339)),
	})

	if err != nil {
		log.Fatalln(err)
	}

	resp, err := client.Post("https://docs.googleapis.com/v1/documents", "application/json", bytes.NewBuffer(requestBody))

	if err != nil {
		log.Fatalln(err)
	}

	defer resp.Body.Close()

	// body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Fatalln(err)
	}

	documentID, found := jsonDecode(resp.Body)["documentId"]

	if !found {
		log.Fatalln(err)
	}

	documentIDString, ok := documentID.(string)

	if !ok {
		log.Fatalln(err)
	}

	AssignParentFolder(email, privateKey, documentIDString)
}

//AssignParentFolder assigns the document to the Current Fires folder
func AssignParentFolder(email string, privateKey string, fileID string) {
	client, err := getClientFromEnvVars(email, privateKey, googleScopes...)
	if err != nil {
		log.Fatal(err)
	}

	requestBody, err := json.Marshal(map[string]interface{}{
		// "id": "1CgRBFg2CTbvjLp57yfoUOD_OZlaVxOht", post mortem / current fires folder
		"id": "19p5T5iTuouXMMaHoVdmbZdDjD8EshRxq",
	})

	if err != nil {
		log.Fatalln(err)
	}

	resp, err := client.Post(fmt.Sprintf("https://www.googleapis.com/drive/v2/files/%s/parents", fileID), "application/json", bytes.NewBuffer(requestBody))

	if err != nil {
		log.Fatalln(err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println(string(body))
}
