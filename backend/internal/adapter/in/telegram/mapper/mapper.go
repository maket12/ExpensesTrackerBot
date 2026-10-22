package mapper

import (
	tgdto "backend/internal/adapter/in/telegram/dto"
	appdto "backend/internal/app/dto"
)

func ToAgentInput(req tgdto.ProcessMessageRequest) appdto.HandleAgentMessageInput {
	return appdto.HandleAgentMessageInput{
		UserID: req.UserID,
		Text:   req.Text,
	}
}

func ToTelegramResponse(out *appdto.HandleAgentMessageOutput) tgdto.ProcessMessageResponse {
	if out == nil {
		return tgdto.ProcessMessageResponse{}
	}
	return tgdto.ProcessMessageResponse{
		Text: FormatTelegramHTML(out.ResponseText),
	}
}
