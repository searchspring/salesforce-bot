package sapi

import (
	"time"

	"github.com/nlopes/slack"
	"searchspring.com/slack/validator"
)

// DAO acts as the slack DAO
type DAO interface {
	GetUserNow(userID string) (time.Time, error)
}

// DAOImpl defines the properties of the DAO
type DAOImpl struct {
	VerificationToken string
	OAuthToken        string
	Client            *slack.Client
}

// NewDAO returns a DAO including a Google API authenticated HTTP client
func NewDAO(verificationToken string, oauthToken string) DAO {
	if validator.ContainsEmptyString(verificationToken, oauthToken) {
		return nil
	}
	client := slack.New(oauthToken)
	return &DAOImpl{
		VerificationToken: verificationToken,
		OAuthToken:        oauthToken,
		Client:            client,
	}
}

func (d *DAOImpl) getUserTZ(userID string) (string, error) {
	user, err := d.Client.GetUserInfo(userID)
	if err != nil {
		return "", err
	}
	return user.TZ, nil
}

// GetUserNow gets current local time for the slack user or UTC on error
func (d *DAOImpl) GetUserNow(userID string) (time.Time, error) {
	tz, err := d.getUserTZ(userID)
	if err != nil {
		return time.Now().UTC(), err
	}
	loc, err := time.LoadLocation(tz)
	if err != nil {
		return time.Now().UTC(), err
	}
	return time.Now().In(loc), nil
}
