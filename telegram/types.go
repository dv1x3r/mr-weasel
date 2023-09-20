package telegram

import (
	"encoding/json"
	"fmt"
)

// https://core.telegram.org/bots/api

type APIResponse struct {
	Ok          bool                `json:"ok"`
	Result      json.RawMessage     `json:"result,omitempty"`
	ErrorCode   int                 `json:"error_code,omitempty"`
	Description string              `json:"description,omitempty"`
	Parameters  *ResponseParameters `json:"parameters,omitempty"`
}

// Describes why a request was unsuccessful.
type ResponseParameters struct {
	// Optional. The group has been migrated to a supergroup with the specified identifier.
	MigrateToChatID int64 `json:"migrate_to_chat_id,omitempty"`
	// Optional. In case of exceeding flood control, the number of seconds left to wait before the request can be repeated.
	RetryAfter int `json:"retry_after,omitempty"`
}

type APIError struct {
	Code    int
	Message string
	ResponseParameters
}

func (e *APIError) Error() string {
	return fmt.Sprintf("telegram.APIError: %d %s", e.Code, e.Message)
}

type Form = map[string]FormFile

type FormFile struct {
	Name   string
	Path   string
	Delete bool
}

// This object represents an incoming update. At most one of the optional parameters can be present in any given update.
type Update struct {
	// The update's unique identifier
	UpdateID int `json:"update_id"`
	// Optional. New incoming message of any kind — text, photo, sticker, etc.
	Message *Message `json:"message,omitempty"`
	// Optional. New incoming callback query.
	CallbackQuery *CallbackQuery `json:"callback_query,omitempty"`
}

// This object represents a Telegram user or bot.
type User struct {
	// Unique identifier for this user or bot
	ID int64 `json:"id"`
	// True, if this user is a bot.
	IsBot bool `json:"is_bot"`
	// User's or bot's first name.
	FirstName string `json:"first_name"`
	// Optional. User's or bot's last name.
	LastName string `json:"last_name,omitempty"`
	// Optional. User's or bot's username.
	Username string `json:"username,omitempty"`
	// Optional. IETF language tag of the user's language.
	LanguageCode string `json:"language_code,omitempty"`
	// Optional. True, if this user is a Telegram Premium user.
	IsPremium bool `json:"is_premium,omitempty"`
	// Optional. True, if this user added the bot to the attachment menu.
	AddedToAttachmentMenu bool `json:"added_to_attachment_menu,omitempty"`
	// Optional. True, if the bot can be invited to groups. Returned only in getMe.
	CanJoinGroups bool `json:"can_join_groups,omitempty"`
	// Optional. True, if privacy mode is disabled for the bot. Returned only in getMe.
	CanReadAllGroupMessages bool `json:"can_read_all_group_messages,omitempty"`
	// Optional. True, if the bot supports inline queries. Returned only in getMe.
	SupportsInlineQueries bool `json:"supports_inline_queries,omitempty"`
}

// This object represents a chat.
type Chat struct {
	// Unique identifier for this chat
	ID int64 `json:"id"`
	// Type of chat, can be either “private”, “group”, “supergroup” or “channel”.
	Type string `json:"type"`
	// Optional. Title, for supergroups, channels and group chats.
	Title string `json:"title,omitempty"`
}

// This object represents a message.
type Message struct {
	// Unique message identifier inside this chat
	MessageID int `json:"message_id"`
	// Optional. Sender of the message; empty for messages sent to channels. For backward compatibility, the field contains a fake sender user in non-channel chats, if the message was sent on behalf of a chat.
	From *User `json:"from,omitempty"`
	// Date the message was sent in Unix time
	Date int `json:"date"`
	// Conversation the message belongs to
	Chat *Chat `json:"chat"`
	// Optional. For replies, the original message. Note that the Message object in this field will not contain further reply_to_message fields even if it itself is a reply.
	ReplyToMessage *Message `json:"reply_to_message,omitempty"`
	// Optional. For text messages, the actual UTF-8 text of the message.
	Text string `json:"text,omitempty"`
	// Optional. Message is an audio file, information about the file.
	Audio *Audio `json:"audio,omitempty"`
	// Optional. Inline keyboard attached to the message. login_url buttons are represented as ordinary url buttons.
	ReplyMarkup *InlineKeyboardMarkup `json:"reply_markup,omitempty"`
}

// This object represents one special entity in a text message. For example, hashtags, usernames, URLs, etc.
type MessageEntity struct {
	// Type of the entity.
	// "mention", "hashtag", "cashtag", "bot_command", "url", "email", "phone_number" etc.
	Type string `json:"type"`
	// Offset in UTF-16 code units to the start of the entity.
	Offset int `json:"offset"`
	// Length of the entity in UTF-16 code units.
	Length int `json:"length"`
	// Optional. For “text_link” only, URL that will be opened after user taps on the text.
	URL string `json:"url,omitempty"`
	// Optional. For “text_mention” only, the mentioned user.
	User *User `json:"user,omitempty"`
	// Optional. For “pre” only, the programming language of the entity text.
	Language string `json:"language,omitempty"`
	// Optional. For “custom_emoji” only, unique identifier of the custom emoji. Use getCustomEmojiStickers to get full information about the sticker.
	CustomEmojiId string `json:"custom_emoji_id,omitempty"`
}

// This object represents one size of a photo or a file / sticker thumbnail.
type PhotoSize struct {
	// Identifier for this file, which can be used to download or reuse the file.
	FileID string `json:"file_id"`
	// Unique identifier for this file, which is supposed to be the same over time and for different bots. Can't be used to download or reuse the file.
	FileUniqueID string `json:"file_unique_id"`
	// Photo width.
	Width int64 `json:"width"`
	// Photo height.
	Height int64 `json:"height"`
	// Optional. File size in bytes.
	FileSize int64 `json:"file_size,omitempty"`
}

// This object represents an audio file to be treated as music by the Telegram clients.
type Audio struct {
	// Identifier for this file, which can be used to download or reuse the file.
	FileID string `json:"file_id"`
	// Unique identifier for this file, which is supposed to be the same over time and for different bots. Can't be used to download or reuse the file.
	FileUniqueID string `json:"file_unique_id"`
	// Duration of the audio in seconds as defined by sender
	Duration int64 `json:"duration"`
	// Optional. Performer of the audio as defined by sender or by audio tags.
	Performer string `json:"performer,omitempty"`
	// Optional. Title of the audio as defined by sender or by audio tags.
	Title string `json:"title,omitempty"`
	// Optional. Original filename as defined by sender.
	FileName string `json:"file_name,omitempty"`
	// Optional. MIME type of the file as defined by sender.
	MimeType string `json:"mime_type,omitempty"`
	// Optional. File size in bytes. It can be bigger than 2^31 and some programming languages may have difficulty/silent defects in interpreting it.
	// But it has at most 52 significant bits, so a signed 64-bit integer or double-precision float type are safe for storing this value.
	FileSize int64 `json:"file_size,omitempty"`
	// Optional. Thumbnail of the album cover to which the music file belongs.
	Thumbnail *PhotoSize `json:"thumbnail,omitempty"`
}

// This object represents a file ready to be downloaded.
// The file can be downloaded via the link https://api.telegram.org/file/bot<token>/<file_path>.
// It is guaranteed that the link will be valid for at least 1 hour.
// When the link expires, a new one can be requested by calling getFile.
type File struct {
	// Identifier for this file, which can be used to download or reuse the file.
	FileID string `json:"file_id"`
	// Unique identifier for this file, which is supposed to be the same over time and for different bots. Can't be used to download or reuse the file.
	FileUniqueID string `json:"file_unique_id"`
	// Optional. File size in bytes. It can be bigger than 2^31 and some programming languages may have difficulty/silent defects in interpreting it.
	// But it has at most 52 significant bits, so a signed 64-bit integer or double-precision float type are safe for storing this value.
	FileSize int64 `json:"file_size,omitempty"`
	// Optional. File path. Use https://api.telegram.org/file/bot<token>/<file_path> to get the file.
	FilePath string `json:"file_path,omitempty"`
}

// This object represents an inline keyboard that appears right next to the message it belongs to.
type InlineKeyboardMarkup struct {
	// Array of button rows, each represented by an Array of InlineKeyboardButton objects.
	InlineKeyboard [][]InlineKeyboardButton `json:"inline_keyboard"`
}

// This object represents one button of an inline keyboard. You must use exactly one of the optional fields.
type InlineKeyboardButton struct {
	// Label text on the button
	Text string `json:"text"`
	// Optional. Data to be sent in a callback query to the bot when button is pressed, 1-64 bytes.
	CallbackData string `json:"callback_data,omitempty"`
}

// This object represents an incoming callback query from a callback button in an inline keyboard.
// NOTE: After the user presses a callback button, Telegram clients will display a progress bar until you call answerCallbackQuery.
// It is, therefore, necessary to react by calling answerCallbackQuery even if no notification to the user is needed (e.g., without specifying any of the optional parameters).
type CallbackQuery struct {
	// Unique identifier for this query
	ID string `json:"id"`
	// Sender
	From *User `json:"from"`
	// Optional. Message with the callback button that originated the query. Note that message content and message date will not be available if the message is too old.
	Message *Message `json:"message,omitempty"`
	// Optional. Identifier of the message sent via the bot in inline mode, that originated the query.
	InlineMessageID string `json:"inline_message_id,omitempty"`
	// Global identifier, uniquely corresponding to the chat to which the message with the callback button was sent. Useful for high scores in games.
	ChatInstance string `json:"chat_instance"`
	// Optional. Data associated with the callback button. Be aware that the message originated the query can contain no callback buttons with this data.
	Data string `json:"data,omitempty"`
}

// This object represents a bot command.
type BotCommand struct {
	// Text of the command; 1-32 characters. Can contain only lowercase English letters, digits and underscores.
	Command string `json:"command"`
	// Description of the command; 1-256 characters.
	Description string `json:"description"`
}

// This object represents the scope to which bot commands are applied.
type BotCommandScope struct {
	Type   string `json:"type"`
	ChatID int64  `json:"chat_id,omitempty"`
	UserID int64  `json:"user_id,omitempty"`
}

// This object represents the content of a media message to be sent.
// It should be one of: InputMediaAnimation, InputMediaDocument, InputMediaAudio, InputMediaPhoto, InputMediaVideo.
type InputMedia interface {
	SetInputMediaType()
}

// Represents an audio file to be treated as music to be sent.
type InputMediaAudio struct {
	// Type of the result, must be audio.
	Type string `json:"type"`
	// File to send. Pass a file_id to send a file that exists on the Telegram servers (recommended),
	// pass an HTTP URL for Telegram to get a file from the Internet,
	// or pass “attach://<file_attach_name>” to upload a new one using multipart/form-data under <file_attach_name> name.
	Media string `json:"media"`
	// Optional. Thumbnail of the file sent.
	// Thumbnails can't be reused and can be only uploaded as a new file,
	// so you can pass “attach://<file_attach_name>” if the thumbnail was uploaded using multipart/form-data under <file_attach_name>.
	Thumbnail string `json:"thumbnail,omitempty"`
	// Optional. Caption of the audio to be sent, 0-1024 characters after entities parsing.
	Caption string `json:"caption,omitempty"`
	// Optional. Mode for parsing entities in the audio caption. See formatting options for more details.
	ParseMode string `json:"parse_mode,omitempty"`
	// Optional. List of special entities that appear in the caption, which can be specified instead of parse_mode.
	CaptionEntities []MessageEntity `json:"caption_entities,omitempty"`
	// Optional. Duration of the audio in seconds.
	Duration int64 `json:"duration,omitempty"`
	// Optional. Performer of the audio.
	Performer string `json:"performer,omitempty"`
	// Optional. Title of the audio.
	Title string `json:"title,omitempty"`
}

func (im *InputMediaAudio) SetInputMediaType() {
	im.Type = "audio"
}
