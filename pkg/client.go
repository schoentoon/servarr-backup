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
	APIKey  string
	BaseURL string
}

type commandReq struct {
	Name string `json:"name"`
}

type commandResp struct {
	Id     int    `json:"id"`
	Status string `json:"status"`
}

type backup struct {
	id     int
	client *Client
}

func (c *Client) StartBackup(ctx context.Context) (*backup, error) {
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(commandReq{Name: "Backup"})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/v3/command", c.BaseURL), &buf)
	if err != nil {
		return nil, err
	}

	req = req.WithContext(ctx)
	req.Header.Add("X-Api-Key", c.APIKey)

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

	return &backup{id: commandResp.Id, client: c}, nil
}

func (b *backup) Wait(ctx context.Context) error {
	timer := time.NewTicker(time.Second)
	defer timer.Stop()
	for {
		select {
		case <-timer.C:
			req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v3/command/%d", b.client.BaseURL, b.id), nil)
			if err != nil {
				return err
			}

			req = req.WithContext(ctx)
			req.Header.Add("X-Api-Key", b.client.APIKey)

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

type backups struct {
	Path string `json:"path"`
	Type string `json:"type"`
}

func (c *Client) DownloadLatestBackup(ctx context.Context) (io.ReadCloser, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v3/system/backup", c.BaseURL), nil)
	if err != nil {
		return nil, err
	}

	req = req.WithContext(ctx)
	req.Header.Add("X-Api-Key", c.APIKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var backups []backups
	err = json.NewDecoder(resp.Body).Decode(&backups)
	if err != nil {
		return nil, err
	}

	path := ""
	for _, backup := range backups {
		if backup.Type == "manual" {
			path = backup.Path
			break
		}
	}

	if path == "" {
		return nil, fmt.Errorf("No recent manual backup found")
	}

	resp, err = http.Get(fmt.Sprintf("%s%s", c.BaseURL, path))
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}
