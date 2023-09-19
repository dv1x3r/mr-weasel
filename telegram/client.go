package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"

	"mr-weasel/utils"
)

const apiEndpoint = "https://api.telegram.org/bot%s/%s"
const apiFileEndpoint = "https://api.telegram.org/file/bot%s/%s"

type Client struct {
	httpClient *http.Client
	token      string
	debug      bool
}

func Connect(token string, debug bool) (*Client, error) {
	const op = "telegram.Client.Connect"
	client := &Client{
		httpClient: &http.Client{Timeout: 100 * time.Second},
		token:      token,
		debug:      debug,
	}
	me, err := client.GetMe(context.Background(), GetMeConfig{})
	if err != nil {
		log.Println("[ERROR] Failed to start the bot")
	} else {
		log.Printf("[INFO] Logged in as [%s]\n", me.Username)
	}
	return client, utils.WrapIfErr(op, err)
}

func MustConnect(token string, debug bool) *Client {
	client, err := Connect(token, debug)
	if err != nil {
		panic(err)
	}
	return client
}

// A simple method for testing your bot's authentication token. Requires no parameters. Returns basic information about the bot in form of a User object.
func (c *Client) GetMe(ctx context.Context, cfg GetMeConfig) (User, error) {
	const op = "telegram.Client.GetMe"
	value, err := executeMethod[User](ctx, c, cfg, nil)
	return value, utils.WrapIfErr(op, err)
}

// Use this method to receive incoming updates using long polling. Returns an Array of Update objects.
func (c *Client) GetUpdates(ctx context.Context, cfg GetUpdatesConfig) ([]Update, error) {
	const op = "telegram.Client.GetUpdates"
	value, err := executeMethod[[]Update](ctx, c, cfg, nil)
	return value, utils.WrapIfErr(op, err)
}

// Use this method to send text messages. On success, the sent Message is returned.
func (c *Client) SendMessage(ctx context.Context, cfg SendMessageConfig) (Message, error) {
	const op = "telegram.Client.SendMessage"
	value, err := executeMethod[Message](ctx, c, cfg, nil)
	return value, utils.WrapIfErr(op, err)
}

// Use this method to send audio files, if you want Telegram clients to display them in the music player.
func (c *Client) SendAudio(ctx context.Context, cfg SendAudioConfig, media Form) (Message, error) {
	const op = "telegram.Client.SendAudio"
	value, err := executeMethod[Message](ctx, c, cfg, media)
	return value, utils.WrapIfErr(op, err)
}

// Use this method to send a group of photos, videos, documents or audios as an album. Documents and audio files can be only grouped in an album with messages of the same type. On success, an array of Messages that were sent is returned.
func (c *Client) SendMediaGroup(ctx context.Context, cfg SendMediaGroupConfig, media Form) ([]Message, error) {
	const op = "telegram.Client.SendMediaGroup"
	for _, media := range cfg.Media {
		media.SetInputMediaType()
	}
	value, err := executeMethod[[]Message](ctx, c, cfg, media)
	return value, utils.WrapIfErr(op, err)
}

// Use this method to edit text and game messages. On success, if the edited message is not an inline message, the edited Message is returned, otherwise True is returned.
func (c *Client) EditMessageText(ctx context.Context, cfg EditMessageTextConfig) (Message, error) {
	const op = "telegram.Client.EditMessageText"
	value, err := executeMethod[Message](ctx, c, cfg, nil)
	return value, utils.WrapIfErr(op, err)
}

// Use this method to send answers to callback queries sent from inline keyboards. The answer will be displayed to the user as a notification at the top of the chat screen or as an alert. On success, True is returned.
func (c *Client) AnswerCallbackQuery(ctx context.Context, cfg AnswerCallbackQueryConfig) (bool, error) {
	const op = "telegram.Client.AnswerCallbackQuery"
	value, err := executeMethod[bool](ctx, c, cfg, nil)
	return value, utils.WrapIfErr(op, err)
}

// Use this method to get basic information about a file and prepare it for downloading. On success, a File object is returned.
func (c *Client) GetFile(ctx context.Context, cfg GetFileConfig) (File, error) {
	const op = "telegram.Client.GetFile"
	value, err := executeMethod[File](ctx, c, cfg, nil)
	return value, utils.WrapIfErr(op, err)
}

// Use this method to get basic information about a file and prepare it for downloading. On success, a File URL is returned (which contains the bot token).
func (c *Client) GetFileURL(ctx context.Context, cfg GetFileConfig) (string, error) {
	const op = "telegram.Client.GetFileURL"
	file, err := c.GetFile(ctx, cfg)
	fileURL := fmt.Sprintf(apiFileEndpoint, c.token, file.FilePath)
	return fileURL, utils.WrapIfErr(op, err)
}

// Use this method to change the list of the bot's commands. See this manual for more details about bot commands. Returns True on success.
func (c *Client) SetMyCommands(ctx context.Context, cfg SetMyCommandsConfig) (bool, error) {
	const op = "telegram.Client.SetMyCommands"
	value, err := executeMethod[bool](ctx, c, cfg, nil)
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

func writeMultipart(body *bytes.Buffer, cfg Config, media Form) (string, error) {
	writer := multipart.NewWriter(body)
	defer writer.Close()

	data, err := json.Marshal(cfg)
	if err != nil {
		return "", err
	}

	var raw map[string]json.RawMessage
	err = json.Unmarshal(data, &raw)
	if err != nil {
		return "", err
	}

	for fieldName, fieldValue := range raw {
		writer.WriteField(fieldName, string(fieldValue))
	}

	for partName, partFile := range media {
		file, err := os.Open(partFile.Path)
		if err != nil {
			return "", err
		}
		defer file.Close()

		part, err := writer.CreateFormFile(partName, partFile.Name)
		if err != nil {
			return "", err
		}

		io.Copy(part, file)
	}

	return writer.FormDataContentType(), nil
}

func executeMethod[T any](ctx context.Context, client *Client, cfg Config, media Form) (T, error) {
	var value T
	var err error

	body := new(bytes.Buffer)
	contentType := "application/json"

	if cfg != nil && media == nil {
		// application/json response
		err = json.NewEncoder(body).Encode(cfg)
		if err != nil {
			return value, err
		}
	} else if cfg != nil && media != nil {
		// multitype/form-data response
		contentType, err = writeMultipart(body, cfg, media)
		if err != nil {
			return value, err
		}
	}

	if client.debug {
		log.Printf("[DEBUG] Request %s %s %+v %+v\b", contentType, cfg.Method(), cfg, media)
	}

	url := fmt.Sprintf(apiEndpoint, client.token, cfg.Method())
	res, err := client.makeRequest(ctx, url, contentType, body)
	if err != nil {
		return value, err
	}

	if client.debug {
		log.Println("[DEBUG] Response", cfg.Method(), strings.TrimSpace(string(res.Result)))
	}

	err = json.Unmarshal(res.Result, &value)
	return value, err
}

func (c *Client) makeRequest(ctx context.Context, url string, contentType string, body *bytes.Buffer) (*APIResponse, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", contentType)

	res, err := c.httpClient.Do(req)
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

	return apiRes, nil
}
