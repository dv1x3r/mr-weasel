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
	return fmt.Sprintf("telegram.APIError: %d %s", e.Code, e.Message)
}

// This object represents an incoming update. At most one of the optional parameters can be present in any given update.
type Update struct {
	// The update's unique identifier.
	UpdateID int `json:"update_id"`
	// Optional. New incoming message of any kind — text, photo, sticker, etc.
	Message *Message `json:"message,omitempty"`
	// Optional. New incoming callback query.
	CallbackQuery *CallbackQuery `json:"callback_query,omitempty"`
}

// This object represents a Telegram user or bot.
type User struct {
	// Unique identifier for this user or bot.
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
	// Unique identifier for this chat.
	ID int64 `json:"id"`
	// Type of chat, can be either “private”, “group”, “supergroup” or “channel”.
	Type string `json:"type"`
	// Optional. Title, for supergroups, channels and group chats.
	Title string `json:"title,omitempty"`
}

// This object represents a message.
type Message struct {
	// Unique message identifier inside this chat.
	MessageID int `json:"message_id"`
	// Optional. Sender of the message; empty for messages sent to channels. For backward compatibility, the field contains a fake sender user in non-channel chats, if the message was sent on behalf of a chat.
	From *User `json:"from,omitempty"`
	// Date the message was sent in Unix time.
	Date int `json:"date"`
	// Conversation the message belongs to.
	Chat *Chat `json:"chat"`
	// Optional. For replies, the original message. Note that the Message object in this field will not contain further reply_to_message fields even if it itself is a reply.
	ReplyToMessage *Message `json:"reply_to_message,omitempty"`
	// Optional. For text messages, the actual UTF-8 text of the message.
	Text string `json:"text,omitempty"`
	// Optional. For text messages, special entities like usernames, URLs, bot commands, etc. that appear in the text.
	Entities []MessageEntity `json:"entities,omitempty"`
	// Optional. Message is an audio file, information about the file.
	Audio *Audio `json:"audio,omitempty"`
	// Optional. Service message: a user was shared with the bot.
	UserShared *UserShared `json:"user_shared,omitempty"`
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

// This object contains information about the user whose identifier was shared with the bot using a KeyboardButtonRequestUser button.
type UserShared struct {
	// Identifier of the request.
	RequestID int64 `json:"request_id"`
	// Identifier of the shared user.
	UserID int64 `json:"user_id"`
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

// Describes a Web App.
type WebAppInfo struct {
	// An HTTPS URL of a Web App to be opened with additional data as specified in Initializing Web Apps.
	URL string `json:"url"`
}

// Additional interface options. A JSON-serialized object for an inline keyboard, custom reply keyboard, instructions to remove reply keyboard or to force a reply from the user.
type ReplyMarkup interface {
	ReplyMarkuper()
}

// This object represents a custom keyboard with reply options (see Introduction to bots for details and examples).
type ReplyKeyboardMarkup struct {
	// Array of button rows, each represented by an Array of KeyboardButton objects
	Keyboard [][]KeyboardButton `json:"keyboard"`
	// Optional. Requests clients to always show the keyboard when the regular keyboard is hidden. Defaults to false, in which case the custom keyboard can be hidden and opened with a keyboard icon.
	IsPersistent bool `json:"is_persistent,omitempty"`
	// Optional. Requests clients to resize the keyboard vertically for optimal fit (e.g., make the keyboard smaller if there are just two rows of buttons). Defaults to false, in which case the custom keyboard is always of the same height as the app's standard keyboard.
	ResizeKeyboard bool `json:"resize_keyboard,omitempty"`
	// Optional. Requests clients to hide the keyboard as soon as it's been used. The keyboard will still be available, but clients will automatically display the usual letter-keyboard in the chat - the user can press a special button in the input field to see the custom keyboard again. Defaults to false.
	OneTimeKeyboard bool `json:"one_time_keyboard,omitempty"`
	// Optional. The placeholder to be shown in the input field when the keyboard is active; 1-64 characters.
	InputFieldPlaceholder string `json:"input_field_placeholder,omitempty"`
	// Optional. Use this parameter if you want to show the keyboard to specific users only. Targets: 1) users that are @mentioned in the text of the Message object; 2) if the bot's message is a reply (has reply_to_message_id), sender of the original message.
	Selective bool `json:"selective,omitempty"`
}

func (ReplyKeyboardMarkup) ReplyMarkuper() {}

// This object represents one button of the reply keyboard. For simple text buttons, String can be used instead of this object to specify the button text.
type KeyboardButton struct {
	// Text of the button. If none of the optional fields are used, it will be sent as a message when the button is pressed.
	Text string `json:"text"`
	// Optional. If specified, pressing the button will open a list of suitable users. Tapping on any user will send their identifier to the bot in a “user_shared” service message. Available in private chats only.
	RequestUser *KeyboardButtonRequestUser `json:"request_user,omitempty"`
	// Optional. If specified, pressing the button will open a list of suitable chats. Tapping on a chat will send its identifier to the bot in a “chat_shared” service message. Available in private chats only.
	RequestChat *KeyboardButtonRequestChat `json:"request_chat,omitempty"`
	// Optional. If True, the user's phone number will be sent as a contact when the button is pressed. Available in private chats only.
	RequestContact bool `json:"request_contact,omitempty"`
	// Optional. If True, the user's current location will be sent when the button is pressed. Available in private chats only.
	RequestLocation bool `json:"request_location,omitempty"`
	// Optional. If specified, the user will be asked to create a poll and send it to the bot when the button is pressed. Available in private chats only.
	RequestPoll *KeyboardButtonPollType `json:"request_poll,omitempty"`
	// Optional. If specified, the described Web App will be launched when the button is pressed. The Web App will be able to send a “web_app_data” service message. Available in private chats only.
	WebApp *WebAppInfo `json:"web_app,omitempty"`
}

// This object defines the criteria used to request a suitable user. The identifier of the selected user will be shared with the bot when the corresponding button is pressed.
type KeyboardButtonRequestUser struct {
	// Signed 32-bit identifier of the request, which will be received back in the UserShared object. Must be unique within the message.
	RequestID int64 `json:"request_id"`
	// Optional. Pass True to request a bot, pass False to request a regular user. If not specified, no additional restrictions are applied.
	UserIsBot *bool `json:"user_is_bot,omitempty"`
	// Optional. Pass True to request a premium user, pass False to request a non-premium user. If not specified, no additional restrictions are applied.
	UserIsPremium *bool `json:"user_is_premium,omitempty"`
}

// This object defines the criteria used to request a suitable chat. The identifier of the selected chat will be shared with the bot when the corresponding button is pressed.
type KeyboardButtonRequestChat struct {
	// Signed 32-bit identifier of the request, which will be received back in the ChatShared object. Must be unique within the message.
	RequestID int64 `json:"request_id"`
	// Pass True to request a channel chat, pass False to request a group or a supergroup chat.
	ChatIsChannel bool `json:"chat_is_channel"`
	// Optional. Pass True to request a forum supergroup, pass False to request a non-forum chat. If not specified, no additional restrictions are applied.
	ChatIsForum *bool `json:"chat_is_forum,omitempty"`
	// Optional. Pass True to request a supergroup or a channel with a username, pass False to request a chat without a username. If not specified, no additional restrictions are applied.
	ChatHasUsername *bool `json:"chat_has_username,omitempty"`
	// Optional. Pass True to request a chat owned by the user. Otherwise, no additional restrictions are applied.
	ChatIsCreated bool `json:"chat_is_created,omitempty"`
	// Optional. A JSON-serialized object listing the required administrator rights of the user in the chat. The rights must be a superset of bot_administrator_rights. If not specified, no additional restrictions are applied.
	UserAdministratorRights *ChatAdministratorRights `json:"user_administrator_rights,omitempty"`
	// Optional. A JSON-serialized object listing the required administrator rights of the bot in the chat. The rights must be a subset of user_administrator_rights. If not specified, no additional restrictions are applied.
	BotAdministratorRights *ChatAdministratorRights `json:"bot_administrator_rights,omitempty"`
	// Optional. Pass True to request a chat with the bot as a member. Otherwise, no additional restrictions are applied.
	BotIsMember bool `json:"bot_is_member,omitempty"`
}

// This object represents type of a poll, which is allowed to be created and sent when the corresponding button is pressed.
type KeyboardButtonPollType struct {
	// Optional. If quiz is passed, the user will be allowed to create only polls in the quiz mode. If regular is passed, only regular polls will be allowed. Otherwise, the user will be allowed to create a poll of any type.
	Type string `json:"type,omitempty"`
}

// Upon receiving a message with this object, Telegram clients will remove the current custom keyboard and display the default letter-keyboard. By default, custom keyboards are displayed until a new keyboard is sent by a bot. An exception is made for one-time keyboards that are hidden immediately after the user presses a button.
type ReplyKeyboardRemove struct {
	// Requests clients to remove the custom keyboard (user will not be able to summon this keyboard; if you want to hide the keyboard from sight but keep it accessible, use one_time_keyboard in ReplyKeyboardMarkup).
	RemoveKeyboard bool `json:"remove_keyboard"`
	// Optional. Use this parameter if you want to remove the keyboard for specific users only. Targets: 1) users that are @mentioned in the text of the Message object; 2) if the bot's message is a reply (has reply_to_message_id), sender of the original message.
	Selective bool `json:"selective,omitempty"`
}

func (ReplyKeyboardRemove) ReplyMarkuper() {}

// This object represents an inline keyboard that appears right next to the message it belongs to.
type InlineKeyboardMarkup struct {
	// Array of button rows, each represented by an Array of InlineKeyboardButton objects.
	InlineKeyboard [][]InlineKeyboardButton `json:"inline_keyboard"`
}

func (InlineKeyboardMarkup) ReplyMarkuper() {}

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

// Represents the rights of an administrator in a chat.
type ChatAdministratorRights struct {
	// True, if the user's presence in the chat is hidden
	IsAnonymous bool `json:"is_anonymous"`
	// True, if the administrator can access the chat event log, boost list in channels, see channel members, report spam messages, see anonymous administrators in supergroups and ignore slow mode. Implied by any other administrator privilege
	CanManageChat bool `json:"can_manage_chat"`
	// True, if the administrator can delete messages of other users
	CanDeleteMessages bool `json:"can_delete_messages"`
	// True, if the administrator can manage video chats
	CanManageVideoChats bool `json:"can_manage_video_chats"`
	// True, if the administrator can restrict, ban or unban chat members, or access supergroup statistics
	CanRestrictMembers bool `json:"can_restrict_members"`
	// True, if the administrator can add new administrators with a subset of their own privileges or demote administrators that they have promoted, directly or indirectly (promoted by administrators that were appointed by the user)
	CanPromoteMembers bool `json:"can_promote_members"`
	// True, if the user is allowed to change the chat title, photo and other settings
	CanChangeInfo bool `json:"can_change_info"`
	// True, if the user is allowed to invite new users to the chat
	CanInviteUsers *bool `json:"can_invite_users,omitempty"`
	// Optional. True, if the administrator can post messages in the channel, or access channel statistics; channels only
	CanPostMessages *bool `json:"can_post_messages,omitempty"`
	// Optional. True, if the administrator can edit messages of other users and can pin messages; channels only
	CanEditMessages *bool `json:"can_edit_messages,omitempty"`
	// Optional. True, if the user is allowed to pin messages; groups and supergroups only
	CanPinMessages *bool `json:"can_pin_messages,omitempty"`
	// Optional. True, if the administrator can post stories in the channel; channels only
	CanPostStories *bool `json:"can_post_stories,omitempty"`
	// Optional. True, if the administrator can edit stories posted by other users; channels only
	CanEditStories *bool `json:"can_edit_stories,omitempty"`
	// Optional. True, if the administrator can delete stories posted by other users; channels only
	CanDeleteStories *bool `json:"can_delete_stories,omitempty"`
	// Optional. True, if the user is allowed to create, rename, close, and reopen forum topics; supergroups only
	CanManageTopics *bool `json:"can_manage_topics,omitempty"`
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
