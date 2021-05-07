package google

import (
	"encoding/json"

	common "github.com/searchspring/nebo/api/config"
)

type DAO interface {
	CheckUserLoggedIn(token string) (string, error)
}
type DAOImpl struct {
	Client *common.Client
}

func NewDAO(client *common.Client) DAO {
	return &DAOImpl{
		Client: client,
	}
}

func (d *DAOImpl) CheckUserLoggedIn(token string) (string, error) {
	body, err := d.Client.AuthorizedGet(token, "https://www.googleapis.com/oauth2/v2/userinfo?access_token="+token)
	if err != nil {
		return "", err
	}
	type emailHolder struct {
		Email string `json:"email"`
	}
	user := &emailHolder{}
	
	err = json.Unmarshal(body, user)
	if err != nil {
		return "", err
	}
	return user.Email, nil
} 
