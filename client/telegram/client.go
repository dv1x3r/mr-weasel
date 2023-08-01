package tgclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

const apiEndpoint = "https://api.telegram.org/bot%s/%s"

type Client struct {
	client *http.Client
	token  string
	debug  bool
}

func Connect(token string, debug bool) (*Client, error) {
	c := &Client{
		client: &http.Client{Timeout: 100 * time.Second},
		token:  token,
		debug:  debug,
	}

	me, err := c.GetMe(context.Background(), GetMeConfig{})
	if err != nil {
		log.Println("[ERROR] Failed to start the bot")
	} else {
		log.Printf("[INFO] Logged in as [%s]", me.Username)
	}

	return c, err
}

func MustConnect(token string, debug bool) *Client {
	c, err := Connect(token, debug)
	if err != nil {
		panic(err)
	}
	return c
}

// A simple method for testing your bot's authentication token. Requires no parameters. Returns basic information about the bot in form of a User object.
func (c *Client) GetMe(ctx context.Context, cfg GetMeConfig) (User, error) {
	return executeMethod[User](ctx, c, cfg)
}

// Use this method to receive incoming updates using long polling. Returns an Array of Update objects.
func (c *Client) GetUpdates(ctx context.Context, cfg GetUpdatesConfig) ([]Update, error) {
	return executeMethod[[]Update](ctx, c, cfg)
}

// Use this method to send text messages. On success, the sent Message is returned.
func (c *Client) SendMessage(ctx context.Context, cfg SendMessageConfig) (Message, error) {
	return executeMethod[Message](ctx, c, cfg)
}

// Use this method to edit text and game messages. On success, if the edited message is not an inline message, the edited Message is returned, otherwise True is returned.
func (c *Client) EditMessageText(ctx context.Context, cfg EditMessageTextConfig) (Message, error) {
	return executeMethod[Message](ctx, c, cfg)
}

// Use this method to change the list of the bot's commands. See this manual for more details about bot commands. Returns True on success.
func (c *Client) SetMyCommands(ctx context.Context, cfg SetMyCommandsConfig) (bool, error) {
	return executeMethod[bool](ctx, c, cfg)
}

// Use this method to receive incoming updates using long polling. Starts a background goroutine, and returns a Channel with Update objects.
func (c *Client) GetUpdatesChan(ctx context.Context, cfg GetUpdatesConfig, chanSize int) <-chan Update {
	ch := make(chan Update, chanSize)
	log.Println("[INFO]", "Goroutine GetUpdatesChan started")

	go func() {
		for {
			select {
			case <-ctx.Done():
				close(ch)
				return
			default:
			}

			updates, err := c.GetUpdates(ctx, cfg)
			if err != nil {
				if errors.Is(err, context.Canceled) {
					log.Println("[INFO]", "Goroutine GetUpdatesChan closed")
					close(ch)
					return
				} else {
					log.Println("[WARN]", "Failed to get updates, retrying in 3 seconds...")
					time.Sleep(time.Second * 3)
					continue
				}
			}

			for _, update := range updates {
				if update.UpdateID >= cfg.Offset {
					cfg.Offset = update.UpdateID + 1
				}
				ch <- update
			}
		}
	}()

	return ch
}

func executeMethod[T any](ctx context.Context, client *Client, cfg APICaller) (T, error) {
	value := new(T)

	res, err := client.makeRequest(ctx, cfg)
	if err != nil {
		return *value, err
	}

	err = json.Unmarshal(res.Result, value)
	return *value, err
}

func (c *Client) makeRequest(ctx context.Context, cfg APICaller) (*APIResponse, error) {
	url := fmt.Sprintf(apiEndpoint, c.token, cfg.Method())

	params := new(bytes.Buffer)
	if cfg != nil {
		err := json.NewEncoder(params).Encode(cfg)
		if err != nil {
			return nil, err
		}
	}

	if c.debug {
		log.Println("[DEBUG] Request", cfg.Method(), strings.TrimSpace(string(params.Bytes())))
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, params)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	apiRes := new(APIResponse)
	err = json.NewDecoder(res.Body).Decode(apiRes)
	if err != nil {
		return nil, fmt.Errorf("[ERROR] makeRequest Decode: %w", err)
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

	if c.debug {
		log.Println("[DEBUG] Response", cfg.Method(), strings.TrimSpace(string(apiRes.Result)))
	}

	return apiRes, nil
}
