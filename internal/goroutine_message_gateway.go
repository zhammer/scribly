package internal

import (
	"context"
)

type GoroutineMessageGateway struct {
	scribly *Scribly
}

func (g *GoroutineMessageGateway) AnnounceUserCreated(ctx context.Context, user User) error {
	return g.scribly.SendVerificationEmail(ctx, user.ID)
}

func (g *GoroutineMessageGateway) AnnounceTurnTaken(ctx context.Context, story Story) error {
	return g.scribly.SendTurnEmailNotifications(ctx, story.ID, len(story.Turns))
}

func (g *GoroutineMessageGateway) AnnounceCowritersAdded(ctx context.Context, story Story) error {
	return g.scribly.SendAddedToStoryEmails(ctx, story.ID)
}

// a little cyclical dependency
func (g *GoroutineMessageGateway) Scribly(scribly *Scribly) {
	g.scribly = scribly
}
