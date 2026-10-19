package mapper

import (
	appdto "backend/internal/app/dto"
	tgdto "backend/internal/adapter/in/telegram/dto"
)

func ToAgentInput(req tgdto.ProcessMessageRequest) appdto.HandleAgentMessageInput {
	return appdto.HandleAgentMessageInput{
		UserID: req.UserID,
		Text:   req.Text,
	}
}

func ToTelegramResponse(out appdto.HandleAgentMessageOutput) tgdto.ProcessMessageResponse {
	return tgdto.ProcessMessageResponse{
		Text: out.ResponseText,
	}
}
