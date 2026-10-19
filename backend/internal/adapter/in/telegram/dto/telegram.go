package dto

type ProcessMessageRequest struct {
	UserID int64
	Text   string
}

type ProcessMessageResponse struct {
	Text string
}
