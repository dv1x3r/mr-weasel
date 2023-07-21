package tgclient

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
	return fmt.Sprintf("Telegram Error: %d %s", e.Code, e.Message)
}

type APICaller interface {
	Method() string
}

type GetMeConfig struct {
}

func (GetMeConfig) Method() string {
	return "getMe"
}

// This object contains information required by GetUpdates request.
type GetUpdatesConfig struct {
	// Optional. Identifier of the first update to be returned. Must be greater by one than the highest among the identifiers of previously received updates.
	// An update is considered confirmed as soon as getUpdates is called with an offset higher than its update_id.
	Offset int `json:"offset,omitempty"`
	// Optional. Limits the number of updates to be retrieved. Values between 1-100 are accepted. Defaults to 100.
	Limit int `json:"limit,omitempty"`
	// Optional. Timeout in seconds for long polling. Defaults to 0, i.e. usual short polling. Should be positive, short polling should be used for testing purposes only.
	Timeout int `json:"timeout,omitempty"`
	// Optional. A JSON-serialized list of the update types you want your bot to receive. For example, specify [“message”, “edited_channel_post”, “callback_query”] to only receive updates of these types.
	AllowedUpdates []string `json:"allowed_updates,omitempty"`
}

func (GetUpdatesConfig) Method() string {
	return "getUpdates"
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
	// Optional. Inline keyboard attached to the message. login_url buttons are represented as ordinary url buttons.
	ReplyMarkup *InlineKeyboardMarkup `json:"reply_markup,omitempty"`
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
	CallbackData *string `json:"callback_data,omitempty"`
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
