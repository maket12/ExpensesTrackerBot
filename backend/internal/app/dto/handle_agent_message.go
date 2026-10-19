package dto

type HandleAgentMessageInput struct {
	UserID int64
	Text   string
}

type HandleAgentMessageOutput struct {
	ResponseText string
}
