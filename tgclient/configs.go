package tgclient

type Config interface {
	Method() string
}

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
	ChatID int64 `json:"chat_id"`
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
	ReplyMarkup *ReplyMarkup `json:"reply_markup,omitempty"`
}

func (SendMessageConfig) Method() string {
	return "sendMessage"
}

type SendAudioConfig struct {
	// Unique identifier for the target chat or username of the target channel.
	ChatID int64 `json:"chat_id"`
	// Unique identifier for the target message thread (topic) of the forum; for forum supergroups only.
	MessageThreadID int64 `json:"message_thread_id,omitempty"`
	// Audio file to send.
	// Pass a file_id as String to send an audio file that exists on the Telegram servers (recommended),
	// pass an HTTP URL as a String for Telegram to get an audio file from the Internet,
	// or upload a new one using multipart/form-data.
	Audio string `json:"audio,omitempty"`
	// Audio caption, 0-1024 characters after entities parsing.
	Caption string `json:"caption,omitempty"`
	// Mode for parsing entities in the audio caption.
	ParseMode string `json:"parse_mode,omitempty"`
	// A JSON-serialized list of special entities that appear in the caption, which can be specified instead of parse_mode.
	CaptionEntities []MessageEntity `json:"caption_entities,omitempty"`
	// Duration of the audio in seconds.
	Duration int64 `json:"duration,omitempty"`
	// Performer.
	Performer string `json:"performer,omitempty"`
	// Track name.
	Title string `json:"title,omitempty"`
	// Thumbnail of the file sent; can be ignored if thumbnail generation for the file is supported server-side.
	Thumbnail string `json:"thumbnail,omitempty"`
	// Sends the message silently. Users will receive a notification with no sound.
	DisableNotification bool `json:"disable_notification,omitempty"`
	// Protects the contents of the sent message from forwarding and saving.
	ProtectContent bool `json:"protect_content,omitempty"`
	// If the message is a reply, ID of the original message.
	ReplyToMessageID int64 `json:"reply_to_message_id,omitempty"`
	// Pass True if the message should be sent even if the specified replied-to message is not found.
	AllowSendingWithoutReply bool `json:"allow_sending_without_reply,omitempty"`
	// Additional interface options. A JSON-serialized object for an inline keyboard, custom reply keyboard, instructions to remove reply keyboard or to force a reply from the user.
	ReplyMarkup *ReplyMarkup `json:"reply_markup,omitempty"`
}

func (SendAudioConfig) Method() string {
	return "sendAudio"
}

type SendMediaGroupConfig struct {
	// Unique identifier for the target chat or username of the target channel.
	ChatID int64 `json:"chat_id"`
	// Optional. Unique identifier for the target message thread (topic) of the forum; for forum supergroups only.
	MessageThreadID int64 `json:"message_thread_id,omitempty"`
	// A JSON-serialized array describing messages to be sent, must include 2-10 items.
	Media []InputMedia `json:"media"`
	// Optional. Sends messages silently. Users will receive a notification with no sound.
	DisableNotification bool `json:"disable_notification,omitempty"`
	// Optional. Protects the contents of the sent messages from forwarding and saving.
	ProtectContent bool `json:"protect_content,omitempty"`
	// Optional. If the messages are a reply, ID of the original message.
	ReplyToMessageID int64 `json:"reply_to_message_id,omitempty"`
	// Optional. Pass True if the message should be sent even if the specified replied-to message is not found.
	AllowSendingWithoutReply bool `json:"allow_sending_without_reply,omitempty"`
}

func (SendMediaGroupConfig) Method() string {
	return "sendMediaGroup"
}

type EditMessageTextConfig struct {
	// Optional. Required if inline_message_id is not specified. Unique identifier for the target chat or username of the target channel.
	ChatID int64 `json:"chat_id,omitempty"`
	// Optional. Required if inline_message_id is not specified. Identifier of the message to edit.
	MessageID int `json:"message_id,omitempty"`
	// Optional. Required if chat_id and message_id are not specified. Identifier of the inline message.
	InlineMessageID string `json:"inline_message_id"`
	// New text of the message, 1-4096 characters after entities parsing.
	Text string `json:"text"`
	// Optional. Mode for parsing entities in the message text. See formatting options for more details.
	ParseMode string `json:"parse_mode,omitempty"`
	// Optional. A JSON-serialized list of special entities that appear in message text, which can be specified instead of parse_mode.
	Entities []MessageEntity `json:"entities,omitempty"`
	// Optional. Disables link previews for links in this message.
	DisableWebPagePreview bool `json:"disable_web_page_preview,omitempty"`
	// Optional. A JSON-serialized object for an inline keyboard.
	ReplyMarkup *InlineKeyboardMarkup `json:"reply_markup,omitempty"`
}

func (EditMessageTextConfig) Method() string {
	return "editMessageText"
}

type AnswerCallbackQueryConfig struct {
	// Unique identifier for the query to be answered.
	CallbackQueryID string `json:"callback_query_id"`
	// Optional. Text of the notification. If not specified, nothing will be shown to the user, 0-200 characters.
	Text string `json:"text"`
	// Optional. If True, an alert will be shown by the client instead of a notification at the top of the chat screen. Defaults to false.
	ShowAlert bool `json:"show_alert"`
	// Optional. URL that will be opened by the user's client.
	URL string `json:"url"`
	// Optional. The maximum amount of time in seconds that the result of the callback query may be cached client-side.
	CacheTime int `json:"cache_time"`
}

func (AnswerCallbackQueryConfig) Method() string {
	return "answerCallbackQuery"
}

type GetFileConfig struct {
	// File identifier to get information about.
	FileID string `json:"file_id"`
}

func (GetFileConfig) Method() string {
	return "getFile"
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
