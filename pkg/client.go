package pkg

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	APIKey     string
	BaseURL    string
	ApiVersion int
}

type commandReq struct {
	Name string `json:"name"`
}

type commandResp struct {
	Id     int    `json:"id"`
	Status string `json:"status"`
}

type createdBackup struct {
	id     int
	client *Client
}

func (c *Client) StartBackup(ctx context.Context) (*createdBackup, error) {
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(commandReq{Name: "Backup"})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/v%d/command", c.BaseURL, c.ApiVersion), &buf)
	if err != nil {
		return nil, err
	}

	req = req.WithContext(ctx)
	req.Header.Add("X-Api-Key", c.APIKey)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var commandResp commandResp
	err = json.NewDecoder(resp.Body).Decode(&commandResp)
	if err != nil {
		return nil, err
	}

	return &createdBackup{id: commandResp.Id, client: c}, nil
}

func (b *createdBackup) Wait(ctx context.Context) error {
	timer := time.NewTicker(time.Second)
	defer timer.Stop()
	for {
		select {
		case <-timer.C:
			req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v%d/command/%d", b.client.BaseURL, b.client.ApiVersion, b.id), nil)
			if err != nil {
				return err
			}

			req = req.WithContext(ctx)
			req.Header.Add("X-Api-Key", b.client.APIKey)
			req.Header.Add("Accept", "application/json")
			req.Header.Add("Content-Type", "application/json")

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			var commandResp commandResp
			err = json.NewDecoder(resp.Body).Decode(&commandResp)
			if err != nil {
				return err
			}

			if commandResp.Status == "completed" {
				return nil
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

type backup struct {
	Path   string `json:"path"`
	Type   string `json:"type"`
	Id     int    `json:"id"`
	client *Client
}

func (c *Client) DownloadLatestBackup(ctx context.Context) (io.ReadCloser, *backup, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v%d/system/backup", c.BaseURL, c.ApiVersion), nil)
	if err != nil {
		return nil, nil, err
	}

	req = req.WithContext(ctx)
	req.Header.Add("X-Api-Key", c.APIKey)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	var backups []backup
	err = json.NewDecoder(resp.Body).Decode(&backups)
	if err != nil {
		return nil, nil, err
	}

	var backup *backup
	for _, b := range backups {
		if b.Type == "manual" {
			backup = &b
			break
		}
	}

	if backup != nil && backup.Path == "" {
		return nil, nil, fmt.Errorf("No recent manual backup found")
	}

	resp, err = http.Get(fmt.Sprintf("%s%s", c.BaseURL, backup.Path))
	if err != nil {
		return nil, nil, err
	}
	backup.client = c

	return resp.Body, backup, nil
}

func (b *backup) Delete(ctx context.Context) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/api/v%d/system/backup/%d", b.client.BaseURL, b.client.ApiVersion, b.Id), nil)
	if err != nil {
		return err
	}

	req = req.WithContext(ctx)
	req.Header.Add("X-Api-Key", b.client.APIKey)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}
