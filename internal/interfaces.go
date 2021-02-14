package internal

import "context"

type EmailGateway interface {
	SendEmail(ctx context.Context, email Email) error
}
