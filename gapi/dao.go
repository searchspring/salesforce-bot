package gapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
	"searchspring.com/slack/validator"
)

// DAO acts as the gapi DAO
type DAO interface {
	GenerateFireDoc(title string) (string, error)
}

// DAOImpl defines the properties of the DAO
type DAOImpl struct {
	Email      string
	PrivateKey []byte
	Client     *http.Client
	FolderID   string
}

// NewDAO returns a DAO including a Google API authenticated HTTP client
func NewDAO(vars map[string]string) (DAO, error) {
	blanks := validator.FindBlankVals(vars)
	if len(blanks) > 0 {
		return nil, fmt.Errorf("the following env vars are not set: %s", strings.Join(blanks, ", "))
	}
	scopes := []string{"https://www.googleapis.com/auth/drive", "https://www.googleapis.com/auth/documents"}
	conf := &jwt.Config{
		Email:      vars["GCP_SERVICE_ACCOUNT_EMAIL"],
		PrivateKey: []byte(vars["GCP_SERVICE_ACCOUNT_PRIVATE_KEY"]),
		Scopes:     scopes,
		TokenURL:   google.JWTTokenURL,
	}
	client := conf.Client(oauth2.NoContext)
	return &DAOImpl{
		Email:      conf.Email,
		PrivateKey: conf.PrivateKey,
		Client:     client,
		FolderID:   vars["GDRIVE_FIRE_DOC_FOLDER_ID"],
	}, nil
}

func jsonDecode(body io.ReadCloser) map[string]interface{} {
	data := make(map[string]interface{})
	json.NewDecoder(body).Decode(&data)
	return data
}

// GenerateFireDoc creates a new Fire Doc in GDrive as needed
func (d *DAOImpl) GenerateFireDoc(title string) (string, error) {
	documentID, err := d.createFireDoc(title)
	if err != nil {
		log.Println(err)
		return "", errors.New("Error creating fire doc")
	}

	err = d.assignParentFolder(documentID)
	if err != nil {
		log.Println(err)
		return "", errors.New("Unable to move fire doc into correct folder in GDrive")
	}

	err = d.writeDoc(documentID)
	if err != nil {
		log.Println(err)
		// In this case we can still use the created doc so there is an error and a documentID returned
		return documentID, errors.New("Unable to write default content to fire doc")
	}
	return documentID, nil
}

func (d *DAOImpl) createFireDoc(title string) (string, error) {
	now := time.Now().Format(time.RFC3339)
	requestBody, err := json.Marshal(map[string]string{"title": fmt.Sprintf("%s %s", now, title)})
	if err != nil {
		return "", err
	}

	resp, err := d.Client.Post("https://docs.googleapis.com/v1/documents", "application/json", bytes.NewBuffer(requestBody))
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

func (d *DAOImpl) assignParentFolder(documentID string) error {
	requestBody, err := json.Marshal(map[string]string{"id": d.FolderID})
	if err != nil {
		return err
	}

	resp, err := d.Client.Post(fmt.Sprintf("https://www.googleapis.com/drive/v2/files/%s/parents", documentID), "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Error moving document %s to folder %s:\n%s", documentID, d.FolderID, body)
	}
	return nil
}

func (d *DAOImpl) writeDoc(documentID string) error {
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

	resp, err := d.Client.Post(fmt.Sprintf("https://docs.googleapis.com/v1/documents/%s:batchUpdate", documentID), "application/json", bytes.NewBuffer(requestBody))
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
