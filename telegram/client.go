package telegram

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

	"mr-weasel/utils"
)

const apiEndpoint = "https://api.telegram.org/bot%s/%s"

type Client struct {
	client *http.Client
	token  string
	debug  bool
}

func NewClient(token string, debug bool) *Client {
	return &Client{
		client: &http.Client{Timeout: 100 * time.Second},
		token:  token,
		debug:  debug,
	}
}

func (c *Client) Connect() (*Client, error) {
	const op = "telegram.Client.Connect"
	me, err := c.GetMe(context.Background(), GetMeConfig{})
	if err != nil {
		log.Println("[ERROR] Failed to start the bot")
	} else {
		log.Printf("[INFO] Logged in as [%s]\n", me.Username)
	}
	return c, utils.WrapIfErr(op, err)
}

func (c *Client) MustConnect() *Client {
	if _, err := c.Connect(); err != nil {
		panic(err)
	}
	return c
}

// A simple method for testing your bot's authentication token. Requires no parameters. Returns basic information about the bot in form of a User object.
func (c *Client) GetMe(ctx context.Context, cfg GetMeConfig) (User, error) {
	const op = "telegram.Client.GetMe"
	value, err := executeMethod[User](ctx, c, cfg)
	return value, utils.WrapIfErr(op, err)
}

// Use this method to receive incoming updates using long polling. Returns an Array of Update objects.
func (c *Client) GetUpdates(ctx context.Context, cfg GetUpdatesConfig) ([]Update, error) {
	const op = "telegram.Client.GetUpdates"
	value, err := executeMethod[[]Update](ctx, c, cfg)
	return value, utils.WrapIfErr(op, err)
}

// Use this method to send text messages. On success, the sent Message is returned.
func (c *Client) SendMessage(ctx context.Context, cfg SendMessageConfig) (Message, error) {
	const op = "telegram.Client.SendMessage"
	value, err := executeMethod[Message](ctx, c, cfg)
	return value, utils.WrapIfErr(op, err)
}

// Use this method to edit text and game messages. On success, if the edited message is not an inline message, the edited Message is returned, otherwise True is returned.
func (c *Client) EditMessageText(ctx context.Context, cfg EditMessageTextConfig) (Message, error) {
	const op = "telegram.Client.EditMessageText"
	value, err := executeMethod[Message](ctx, c, cfg)
	return value, utils.WrapIfErr(op, err)
}

// Use this method to send answers to callback queries sent from inline keyboards. The answer will be displayed to the user as a notification at the top of the chat screen or as an alert. On success, True is returned.
func (c *Client) AnswerCallbackQuery(ctx context.Context, cfg AnswerCallbackQueryConfig) (bool, error) {
	const op = "telegram.Client.AnswerCallbackQuery"
	value, err := executeMethod[bool](ctx, c, cfg)
	return value, utils.WrapIfErr(op, err)
}

func (c *Client) GetFile(ctx context.Context, cfg GetFileConfig) (File, error) {
	const op = "telegram.Client.GetFile"
	value, err := executeMethod[File](ctx, c, cfg)
	return value, utils.WrapIfErr(op, err)
}

func (c *Client) GetFileURL(ctx context.Context, cfg GetFileConfig) (string, error) {
	const op = "telegram.Client.GetFileURL"
	file, err := c.GetFile(ctx, cfg)
	return file.FilePath, utils.WrapIfErr(op, err)
}

// Use this method to change the list of the bot's commands. See this manual for more details about bot commands. Returns True on success.
func (c *Client) SetMyCommands(ctx context.Context, cfg SetMyCommandsConfig) (bool, error) {
	const op = "telegram.Client.SetMyCommands"
	value, err := executeMethod[bool](ctx, c, cfg)
	return value, utils.WrapIfErr(op, err)
}

// Use this method to receive incoming updates using long polling. Starts a background goroutine, and returns a Channel with Update objects.
func (c *Client) GetUpdatesChan(ctx context.Context, cfg GetUpdatesConfig, chanSize int) <-chan Update {
	ch := make(chan Update, chanSize)
	log.Println("[INFO] Goroutine GetUpdatesChan started")

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
					log.Println("[INFO] Goroutine GetUpdatesChan closed")
					close(ch)
					return
				} else {
					log.Println("[WARN]", err)
					log.Println("[WARN] Failed to get updates, retrying in 3 seconds...")
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

func executeMethod[T any](ctx context.Context, client *Client, cfg Config) (T, error) {
	var value T

	res, err := client.makeRequest(ctx, cfg)
	if err != nil {
		return value, err
	}

	err = json.Unmarshal(res.Result, &value)
	return value, err
}

func (c *Client) makeRequest(ctx context.Context, cfg Config) (*APIResponse, error) {
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
		return nil, err
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
