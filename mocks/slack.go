package mocks

import "github.com/nlopes/slack"

type SlackDAO struct {
	Recorded []string
}

func (s *SlackDAO) SendSlackMessage(token string, attachments slack.Attachment, channel string) error {
	s.Recorded = []string{token, channel}
	return nil
}

func (s *SlackDAO) GetValues() []string {
	return s.Recorded
}
