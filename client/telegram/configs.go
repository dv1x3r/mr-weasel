package tgclient

// A simple method for testing your bot's authentication token. Requires no parameters.
// Returns basic information about the bot in form of a User object.
type GetMeConfig struct {
}

func (GetMeConfig) Method() string {
	return "getMe"
}

// Use this method to receive incoming updates using long polling.
// Returns an Array of Update objects.
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

// Use this method to send text messages.
// On success, the sent Message is returned.
type SendMessageConfig struct {
	// Unique identifier for the target chat or username.
	ChatId int64 `json:"chat_id"`
	// Unique identifier for the target message thread (topic) of the forum; for forum supergroups only.
	MessageThreadId int `json:"message_thread_id,omitempty"`
	// Text of the message to be sent, 1-4096 characters after entities parsing.
	Text string `json:"text"`
	// Mode for parsing entities in the message text. See formatting options for more details.
	ParseMode string `json:"parse_mode,omitempty"`
	// A JSON-serialized list of special entities that appear in message text, which can be specified instead of parse_mode.
	Entities []MessageEntity `json:"entities,omitempty"`
	// Disables link previews for links in this message.
	DisableWebPagePreview bool `json:"disable_web_page_preview,omitempty"`
	// Sends the message silently. Users will receive a notification with no sound.
	DisableNotification bool `json:"disable_notification,omitempty"`
	// Protects the contents of the sent message from forwarding and saving.
	ProtectContent bool `json:"protect_content,omitempty"`
	// If the message is a reply, ID of the original message.
	ReplyToMessageId int `json:"reply_to_message_id,omitempty"`
	// Pass True if the message should be sent even if the specified replied-to message is not found.
	AllowSendingWithoutReply bool `json:"allow_sending_without_reply,omitempty"`
	// Additional interface options. A JSON-serialized object for an inline keyboard, custom reply keyboard, instructions to remove reply keyboard or to force a reply from the user.
	ReplyMarkup *InlineKeyboardMarkup `json:"reply_markup,omitempty"`
}

func (SendMessageConfig) Method() string {
	return "sendMessage"
}
