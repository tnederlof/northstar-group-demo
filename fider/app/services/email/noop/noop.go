package noop

import (
	"context"

	"github.com/getfider/fider/app/models/cmd"
	"github.com/getfider/fider/app/models/dto"
	"github.com/getfider/fider/app/models/query"
	"github.com/getfider/fider/app/pkg/bus"
	"github.com/getfider/fider/app/pkg/env"
	"github.com/getfider/fider/app/pkg/log"
)

func init() {
	bus.Register(Service{})
}

type Service struct{}

func (s Service) Name() string {
	return "Noop"
}

func (s Service) Category() string {
	return "email"
}

func (s Service) Enabled() bool {
	return env.Config.Email.Type == "none"
}

func (s Service) Init() {
	bus.AddListener(sendMail)
	bus.AddHandler(fetchRecentSupressions)
}

func fetchRecentSupressions(ctx context.Context, c *query.FetchRecentSupressions) error {
	// No-op: return empty list
	c.EmailAddresses = []string{}
	return nil
}

func sendMail(ctx context.Context, c *cmd.SendMail) {
	// Log the email that would have been sent (for debugging) but don't actually send
	for _, to := range c.To {
		log.Debugf(ctx, "NOOP email to @{Address} with template @{TemplateName} (not sent - EMAIL=none)", dto.Props{
			"Address":      to.Address,
			"TemplateName": c.TemplateName,
		})
	}
}
