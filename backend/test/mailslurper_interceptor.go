package test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"runtime"
)

type MailslurperConfiguration struct {
	ServicePort int `json:"servicePort" required:"true"`
}

func NewMailslurperInterceptor() (*mailslurperInterceptor, error) {
	_, b, _, _ := runtime.Caller(0)
	testRoot := path.Dir(b)
	jsonFile, err := os.Open(path.Join(testRoot, "config", "mailslurper.json"))
	if err != nil {
		return nil, fmt.Errorf("failed to open mailslurper config: %w", err)
	}
	defer jsonFile.Close()
	var config MailslurperConfiguration
	if err := json.NewDecoder(jsonFile).Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to decode mailslurper config: %w", err)
	}
	return &mailslurperInterceptor{
		ServiceBaseUrl: fmt.Sprintf("http://localhost:%d", config.ServicePort),
		ServicePort:    config.ServicePort,
	}, nil
}

type mailslurperInterceptor struct {
	ServiceBaseUrl string
	ServicePort    int
}

type GetEmailResponse struct {
	MailItems []GetEmailResponseMailItem `json:"mailItems"`
}

type GetEmailResponseMailItem struct {
	Id          string   `json:"id"`
	DateSent    string   `json:"dateSent"`
	FromAddress string   `json:"fromAddress"`
	ToAddresses []string `json:"toAddresses"`
	Subject     string   `json:"subject"`
	Body        string   `json:"body"`
	ContentType string   `json:"contentType"`
}

func (m *mailslurperInterceptor) GetEmails() (*GetEmailResponse, error) {
	response, err := http.Get(fmt.Sprintf("%s/mail", m.ServiceBaseUrl))
	if err != nil {
		return nil, fmt.Errorf("failed to get emails from mailslurper: %w", err)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	var result GetEmailResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %w", err)
	}
	return &result, nil
}

func (m *mailslurperInterceptor) GetEmailPort() int {
	return m.ServicePort
}
