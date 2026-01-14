package internal

import "context"

type EmailGateway interface {
	SendEmail(ctx context.Context, email Email) error
}

type MessageGateway interface {
	AnnounceUserCreated(ctx context.Context, user User) error
	AnnounceTurnTaken(ctx context.Context, story Story) error
	AnnounceCowritersAdded(ctx context.Context, story Story) error
}

type OpenAIGateway interface {
	GenerateTurnText(ctx context.Context, story Story) (string, error)
}
