package tgclient

type GetMeConfig struct {
}

func (GetMeConfig) Method() string {
	return "getMe"
}

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

type SendMessageConfig struct {
	// Unique identifier for the target chat or username.
	ChatId int64 `json:"chat_id"`
	// Optional. Unique identifier for the target message thread (topic) of the forum; for forum supergroups only.
	MessageThreadId int `json:"message_thread_id,omitempty"`
	// Text of the message to be sent, 1-4096 characters after entities parsing.
	Text string `json:"text"`
	// Optional. Mode for parsing entities in the message text. See formatting options for more details.
	ParseMode string `json:"parse_mode,omitempty"`
	// Optional. A JSON-serialized list of special entities that appear in message text, which can be specified instead of parse_mode.
	Entities []MessageEntity `json:"entities,omitempty"`
	// Optional. Disables link previews for links in this message.
	DisableWebPagePreview bool `json:"disable_web_page_preview,omitempty"`
	// Optional. Sends the message silently. Users will receive a notification with no sound.
	DisableNotification bool `json:"disable_notification,omitempty"`
	// Optional. Protects the contents of the sent message from forwarding and saving.
	ProtectContent bool `json:"protect_content,omitempty"`
	// Optional. If the message is a reply, ID of the original message.
	ReplyToMessageId int `json:"reply_to_message_id,omitempty"`
	// Optional. Pass True if the message should be sent even if the specified replied-to message is not found.
	AllowSendingWithoutReply bool `json:"allow_sending_without_reply,omitempty"`
	// Optional. Additional interface options. A JSON-serialized object for an inline keyboard, custom reply keyboard, instructions to remove reply keyboard or to force a reply from the user.
	ReplyMarkup *InlineKeyboardMarkup `json:"reply_markup,omitempty"`
}

func (SendMessageConfig) Method() string {
	return "sendMessage"
}

type SetMyCommandsConfig struct {
	// A JSON-serialized list of bot commands to be set as the list of the bot's commands. At most 100 commands can be specified.
	Commands []BotCommand `json:"commands"`
	// Optional. A JSON-serialized object, describing scope of users for which the commands are relevant. Defaults to BotCommandScopeDefault.
	Scope *BotCommandScope `json:"scope,omitempty"`
	// Optional. A two-letter ISO 639-1 language code. If empty, commands will be applied to all users from the given scope, for whose language there are no dedicated commands.
	LanguageCode string `json:"language_code,omitempty"`
}

func (SetMyCommandsConfig) Method() string {
	return "setMyCommands"
}
