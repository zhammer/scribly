package internal

import (
	"context"
	"fmt"
	"time"
)

type GoroutineMessageGateway struct {
	scribly *Scribly
}

func (g *GoroutineMessageGateway) AnnounceUserCreated(ctx context.Context, user User) {
	g.dispatch(func() error {
		return g.scribly.SendVerificationEmail(context.Background(), user.ID)
	})
}

func (g *GoroutineMessageGateway) AnnounceTurnTaken(ctx context.Context, story Story) {
	g.dispatch(func() error {
		return g.scribly.SendTurnEmailNotifications(context.Background(), story.ID, len(story.Turns))
	})
}

func (g *GoroutineMessageGateway) AnnounceCowritersAdded(ctx context.Context, story Story) {
	g.dispatch(func() error {
		return g.scribly.SendAddedToStoryEmails(context.Background(), story.ID)
	})
}

// a little cyclical dependency
func (g *GoroutineMessageGateway) Scribly(scribly *Scribly) {
	g.scribly = scribly
}

func (g *GoroutineMessageGateway) dispatch(f func() error) {
	go func() {
		for i := 0; i < 3; i++ {
			// if there's an error, retry
			if err := f(); err != nil {
				fmt.Printf("function failed with err: %s\n", err)
				time.Sleep(1 * time.Second)
				continue
			}
			// if success, we're done
			return
		}
		fmt.Println("failed to successfully call function after 3 retries")
	}()
}
