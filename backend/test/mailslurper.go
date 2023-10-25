package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"io"
	"net/http"
	"net/smtp"
	"time"
)

type TestMailslurper struct {
	Id       string
	pool     *dockertest.Pool
	resource *dockertest.Resource
	httpUrl  string
	SmtpPort string
}

func StartMailslurper() (*TestMailslurper, error) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		return nil, fmt.Errorf("could not construct pool: %w", err)
	}

	err = pool.Client.Ping()
	if err != nil {
		return nil, fmt.Errorf("could not connect to docker: %w", err)
	}

	id := uuid.New().String()
	options := getMailslurperOptions(id)
	if err != nil {
		return nil, fmt.Errorf("could not create docker run options: %w", err)
	}

	resource, err := pool.RunWithOptions(options, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		return nil, fmt.Errorf("could not start resource: %w", err)
	}

	hostAndPort := resource.GetHostPort("8085/tcp")
	dsn := fmt.Sprintf("http://%s", hostAndPort)

	_ = resource.Expire(120)

	pool.MaxWait = 30 * time.Second
	if err = pool.Retry(func() error {
		_, err = http.Get(fmt.Sprintf("%s/mail", dsn))
		return err
	}); err != nil {
		return nil, fmt.Errorf("could not connect to docker: %w", err)
	}

	smtpPort := resource.GetPort("2500/tcp")
	smtpHostAndPort := resource.GetHostPort("2500/tcp")
	if err = pool.Retry(func() error {
		client, err := smtp.Dial(fmt.Sprintf("%s", smtpHostAndPort))
		defer client.Close()
		return err
	}); err != nil {
		return nil, fmt.Errorf("could not connect to SMTP port: %w", err)
	}

	return &TestMailslurper{
		Id:       id,
		pool:     pool,
		resource: resource,
		httpUrl:  dsn,
		SmtpPort: smtpPort,
	}, nil
}

func PurgeMailslurper(instance *TestMailslurper) error {
	if instance == nil {
		return nil
	}
	if err := instance.pool.Purge(instance.resource); err != nil {
		return fmt.Errorf("could not purge resource: %w", err)
	}
	return nil
}

func getMailslurperOptions(id string) *dockertest.RunOptions {
	return &dockertest.RunOptions{
		Repository: "marcopas/docker-mailslurper",
		Tag:        "latest",
		ExposedPorts: []string{
			"8085/tcp",
			"2500/tcp",
		},
		Name: fmt.Sprintf("mailslurper-%s", id),
	}
}

type GetEmailResponse struct {
	MailItems    []GetEmailResponseMailItem `json:"mailItems"`
	TotalRecords int                        `json:"totalRecords"`
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

func (m *TestMailslurper) GetEmails() (*GetEmailResponse, error) {
	m.addBuffer()

	response, err := http.Get(fmt.Sprintf("%s/mail", m.httpUrl))
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

type DeleteEmailsRequest struct {
	PruneCode string `json:"pruneCode" required:"true"`
}

func (m *TestMailslurper) DeleteEmails() error {
	m.addBuffer()

	request := DeleteEmailsRequest{
		PruneCode: "all",
	}

	bodyJson, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	httpReq, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/mail", m.httpUrl), bytes.NewReader(bodyJson))
	httpReq.Header.Set("Content-Type", "application/json")

	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	client := http.Client{}

	response, err := client.Do(httpReq)
	defer response.Body.Close()
	if err != nil {
		return fmt.Errorf("failed to delete emails from mailslurper: %w", err)
	}

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to delete emails from mailslurper: %w", err)
	}

	return nil
}

func (m *TestMailslurper) addBuffer() {
	// Allow buffer time for HTTP server to catch up with received emails
	time.Sleep(15 * time.Millisecond)
}
