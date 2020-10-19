package gapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
)

var fireDocFolderID = "19p5T5iTuouXMMaHoVdmbZdDjD8EshRxq"

// Scopes is the Google API scopes list
var Scopes = []string{"https://www.googleapis.com/auth/drive", "https://www.googleapis.com/auth/documents"}

// GetGoogleAPIClient returns a Google API authenticated HTTP client
func GetGoogleAPIClient(email string, privateKey string, scopes ...string) *http.Client {
	conf := &jwt.Config{
		Email:      email,
		PrivateKey: []byte(privateKey),
		Scopes:     scopes,
		TokenURL:   google.JWTTokenURL,
	}
	client := conf.Client(oauth2.NoContext)
	return client
}

func jsonDecode(body io.ReadCloser) map[string]interface{} {
	data := make(map[string]interface{})
	json.NewDecoder(body).Decode(&data)
	return data
}

// CreateFireDoc creates a fire document and returns the document ID
func CreateFireDoc(client *http.Client) (string, error) {
	now := time.Now().Format(time.RFC3339)
	requestBody, err := json.Marshal(map[string]interface{}{"title": fmt.Sprintf("%s Fire Title", now)})
	if err != nil {
		return "", err
	}

	resp, err := client.Post("https://docs.googleapis.com/v1/documents", "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	documentID, found := jsonDecode(resp.Body)["documentId"]
	if !found {
		return "", err
	}

	documentIDString, ok := documentID.(string)
	if !ok {
		return "", err
	}

	return documentIDString, nil
}

// AssignParentFolder assigns the document to the Current Fires folder
func AssignParentFolder(client *http.Client, fileID string) error {
	// "id": "1CgRBFg2CTbvjLp57yfoUOD_OZlaVxOht", post mortem / current fires folder
	requestBody, err := json.Marshal(map[string]interface{}{"id": fireDocFolderID})
	if err != nil {
		return err
	}

	resp, err := client.Post(fmt.Sprintf("https://www.googleapis.com/drive/v2/files/%s/parents", fileID), "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Error moving document %s to folder %s:\n%s", fileID, fireDocFolderID, body)
	}
	return nil
}
