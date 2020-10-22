package gapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
)

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
func CreateFireDoc(client *http.Client, title string) (string, error) {
	now := time.Now().Format(time.RFC3339)
	requestBody, err := json.Marshal(map[string]string{"title": fmt.Sprintf("%s %s", now, title)})
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

// AssignParentFolder assigns the document's parent to the provided folder
func AssignParentFolder(client *http.Client, documentID string, fireDocFolderID string) error {
	requestBody, err := json.Marshal(map[string]string{"id": fireDocFolderID})
	if err != nil {
		return err
	}

	resp, err := client.Post(fmt.Sprintf("https://www.googleapis.com/drive/v2/files/%s/parents", documentID), "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Error moving document %s to folder %s:\n%s", documentID, fireDocFolderID, body)
	}
	return nil
}

// WriteDoc writes initial content to the fire doc
func WriteDoc(client *http.Client, documentID string) error {
	lines, textRequest := fireDocTextToInsert()
	requests := map[string][]interface{}{"requests": {
		textRequest,
		fireDocStyleToApply(1, len(lines[0])+1),
		fireDocBulletsToApply(len(lines[0])+2, len(lines[1])),
	}}
	requestBody, err := json.Marshal(requests)
	if err != nil {
		return err
	}

	resp, err := client.Post(fmt.Sprintf("https://docs.googleapis.com/v1/documents/%s:batchUpdate", documentID), "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Error writing to document %s:\n%s", documentID, body)
	}
	return nil
}

func fireDocStyleToApply(start int, end int) map[string]interface{} {
	textStyle := map[string]bool{"bold": true}
	styleRange := map[string]interface{}{"segmentId": "", "startIndex": start, "endIndex": end}
	updateTextStyle := map[string]interface{}{"textStyle": textStyle, "fields": "*", "range": styleRange}
	request := map[string]interface{}{"updateTextStyle": updateTextStyle}
	return request
}

func fireDocBulletsToApply(start int, end int) map[string]interface{} {
	bulletPreset := "BULLET_DISC_CIRCLE_SQUARE"
	bulletRange := map[string]interface{}{"segmentId": "", "startIndex": start, "endIndex": end}
	createParagraphBullets := map[string]interface{}{"range": bulletRange, "bulletPreset": bulletPreset}
	request := map[string]interface{}{"createParagraphBullets": createParagraphBullets}
	return request
}

func fireDocTextToInsert() ([]string, map[string]interface{}) {
	now := time.Now()
	zone, _ := now.Zone()
	endOfSegmentLocation := map[string]interface{}{"segmentId": ""}
	lines := []string{
		fmt.Sprintf("Timeline (%s)", zone),
		fmt.Sprintf("%s - Fire was called", now.Format("2006-01-02 15:04")),
	}
	insertText := map[string]interface{}{"text": strings.Join(lines, "\n"), "endOfSegmentLocation": endOfSegmentLocation}
	request := map[string]interface{}{"insertText": insertText}
	return lines, request
}
