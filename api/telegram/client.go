package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"
)

const apiEndpoint = "https://api.telegram.org/bot%s/%s"

type Client struct {
	client *http.Client
	token  string
	Me     User
}

func New(token string) (*Client, error) {
	tg := &Client{
		client: &http.Client{Timeout: 120},
		token:  token,
	}

	me, err := tg.GetMe(context.Background(), GetMeConfig{})
	if err != nil {
		log.Println("Failed to start the bot")
	} else {
		log.Printf("Logged in as [%s]", me.Username)
	}
	tg.Me = me

	return tg, err
}

// A simple method for testing your bot's authentication token. Requires no parameters. Returns basic information about the bot in form of a User object.
func (tg *Client) GetMe(ctx context.Context, cfg GetMeConfig) (User, error) {
	res, err := tg.makeRequest(ctx, cfg)
	if err != nil {
		return User{}, err
	}

	err = json.Unmarshal(res.Result, &tg.Me)
	if err != nil {
		return User{}, fmt.Errorf("GetMe Unmarshal: %w", err)
	}

	return tg.Me, nil
}

// Use this method to receive incoming updates using long polling. Returns an Array of Update objects.
func (tg *Client) GetUpdates(ctx context.Context, cfg GetUpdatesConfig) ([]Update, error) {
	res, err := tg.makeRequest(ctx, cfg)
	if err != nil {
		return nil, err
	}

	updates := []Update{}
	err = json.Unmarshal(res.Result, &updates)
	if err != nil {
		return nil, fmt.Errorf("Updates Unmarshal: %w", err)
	}

	return updates, nil
}

// Use this method to receive incoming updates using long polling. Returns a Channel with Update objects.
func (tg *Client) GetUpdatesChan(ctx context.Context, cfg GetUpdatesConfig, chanSize int) <-chan Update {
	c := make(chan Update, chanSize)
	log.Println("Goroutine GetUpdatesChan started")

	go func() {
		for {
			select {
			case <-ctx.Done():
				close(c)
				return
			default:
			}

			updates, err := tg.GetUpdates(ctx, cfg)
			if err != nil {
				if errors.Is(err, context.Canceled) {
					log.Println("Goroutine GetUpdatesChan closed")
					close(c)
					return
				} else {
					log.Println("Failed to get updates, retrying in 3 seconds...")
					time.Sleep(time.Second * 3)
					continue
				}
			}

			for _, update := range updates {
				cfg.Offset = update.UpdateID + 1
				c <- update
			}
		}
	}()

	return c
}

func (tg *Client) makeRequest(ctx context.Context, cfg APICaller) (*APIResponse, error) {
	url := fmt.Sprintf(apiEndpoint, tg.token, cfg.Method())

	params := new(bytes.Buffer)
	if cfg != nil {
		err := json.NewEncoder(params).Encode(cfg)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, params)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := tg.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	apiRes := new(APIResponse)
	err = json.NewDecoder(res.Body).Decode(apiRes)
	if err != nil {
		return nil, fmt.Errorf("makeRequest Decode: %w", err)
	}

	if !apiRes.Ok {
		var parameters ResponseParameters

		if apiRes.Parameters != nil {
			parameters = *apiRes.Parameters
		}

		return apiRes, &APIError{
			Code:               apiRes.ErrorCode,
			Message:            apiRes.Description,
			ResponseParameters: parameters,
		}
	}

	return apiRes, nil
}
