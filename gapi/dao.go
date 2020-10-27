package gapi

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
	"google.golang.org/api/docs/v1"
	"google.golang.org/api/drive/v2"
	"google.golang.org/api/option"
	"searchspring.com/slack/validator"
)

// DAO acts as the gapi DAO
type DAO interface {
	GenerateFireDoc(title string, now time.Time) (string, error)
}

// DAOImpl defines the properties of the DAO
type DAOImpl struct {
	Email        string
	PrivateKey   []byte
	DocsService  *docs.Service
	DriveService *drive.Service
	FolderID     string
}

// NewDAO returns a DAO including a Google API authenticated HTTP client
func NewDAO(email string, privateKey string, folderID string) DAO {
	if validator.ContainsEmptyString(email, privateKey, folderID) {
		return nil
	}
	scopes := []string{"https://www.googleapis.com/auth/drive", "https://www.googleapis.com/auth/documents"}
	conf := &jwt.Config{
		Email:      email,
		PrivateKey: []byte(privateKey),
		Scopes:     scopes,
		TokenURL:   google.JWTTokenURL,
	}
	ctx := context.Background()
	creds := &google.Credentials{TokenSource: conf.TokenSource(ctx)}
	docsService, err := docs.NewService(ctx, option.WithCredentials(creds))
	driveService, err := drive.NewService(ctx, option.WithCredentials(creds))
	if err != nil {
		log.Println(err.Error())
		return nil
	}
	return &DAOImpl{
		Email:        conf.Email,
		PrivateKey:   conf.PrivateKey,
		DocsService:  docsService,
		DriveService: driveService,
		FolderID:     folderID,
	}
}

// GenerateFireDoc creates a new Fire Doc in GDrive as needed
func (d *DAOImpl) GenerateFireDoc(title string, now time.Time) (string, error) {
	documentID, err := d.createFireDoc(title, now)
	if err != nil {
		log.Println(err)
		return "", errors.New("Error creating fire doc")
	}

	err = d.writeDoc(documentID, now)
	if err != nil {
		log.Println(err)
		// In this case we can still use the created doc so there is an error and a documentID returned
		return documentID, errors.New("Unable to write default content to fire doc")
	}

	err = d.assignParentFolder(documentID)
	if err != nil {
		log.Println(err)
		return "", errors.New("Unable to move fire doc into correct folder in GDrive")
	}
	return documentID, nil
}

func (d *DAOImpl) createFireDoc(title string, now time.Time) (string, error) {
	document := &docs.Document{Title: fmt.Sprintf("%s %s", now.Format(time.RFC3339), title)}
	doc, err := d.DocsService.Documents.Create(document).Do()
	if err != nil {
		return "", err
	}
	return doc.DocumentId, nil
}

func (d *DAOImpl) assignParentFolder(documentID string) error {
	file, err := d.DriveService.Files.Get(documentID).Do()
	if err != nil {
		return err
	}

	_, err = d.DriveService.Files.Update(file.Id, file).AddParents(d.FolderID).Do()
	if err != nil {
		return err
	}

	return nil
}

func (d *DAOImpl) writeDoc(documentID string, now time.Time) error {
	zone, _ := now.Zone()
	lines := []string{
		fmt.Sprintf("Timeline (%s)", zone),
		fmt.Sprintf("%s - Fire was called", now.Format("2006-01-02 15:04")),
	}
	requests := &docs.BatchUpdateDocumentRequest{
		Requests: []*docs.Request{
			&docs.Request{
				InsertText: &docs.InsertTextRequest{
					Location: &docs.Location{Index: 1},
					Text:     strings.Join(lines, "\n"),
				},
			},
			&docs.Request{
				UpdateTextStyle: &docs.UpdateTextStyleRequest{
					Range: &docs.Range{
						StartIndex: 1,
						EndIndex:   int64(len(lines[0]) + 1),
					},
					Fields: "*",
					TextStyle: &docs.TextStyle{
						Bold: true,
					},
				},
			},
			&docs.Request{
				CreateParagraphBullets: &docs.CreateParagraphBulletsRequest{
					BulletPreset: "BULLET_DISC_CIRCLE_SQUARE",
					Range: &docs.Range{
						StartIndex: int64(len(lines[0]) + 2),
						EndIndex:   int64(len(lines[1])),
					},
				},
			},
		},
	}
	_, err := d.DocsService.Documents.BatchUpdate(documentID, requests).Do()
	if err != nil {
		return err
	}
	return nil
}
