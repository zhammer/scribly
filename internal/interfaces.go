package internal

import "context"

type EmailGateway interface {
	SendEmail(ctx context.Context, email Email) error
}

type MessageGateway interface {
	AnnounceUserCreated(ctx context.Context, user User)
	AnnounceTurnTaken(ctx context.Context, story Story)
	AnnounceCowritersAdded(ctx context.Context, story Story)
}

type OpenAIGateway interface {
	PredictText(ctx context.Context, story Story) (string, error)
}
