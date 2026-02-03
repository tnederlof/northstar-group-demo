package demo

import (
	"net/url"

	"github.com/getfider/fider/app/models/query"
	"github.com/getfider/fider/app/pkg/bus"
	"github.com/getfider/fider/app/pkg/env"
	"github.com/getfider/fider/app/pkg/web"
	webutil "github.com/getfider/fider/app/pkg/web/util"
)

// personaEmails maps persona slugs to seeded user emails (from Northstar demo seed data)
var personaEmails = map[string]string{
	"alex":     "alex.rivera@northstar.io",     // Administrator (Platform Eng Lead)
	"sarah":    "sarah.chen@northstar.io",      // Collaborator (PM, Digital)
	"marcus":   "marcus.wright@northstar.io",   // Visitor (Ops Analyst, Logistics)
	"jennifer": "jennifer.patel@northstar.io",  // Visitor (Finance Systems Lead)
}

// Login handles demo-mode authentication bypass for seeded users
// Route: GET /__demo/login/:persona?key=...
func Login() web.HandlerFunc {
	return func(c *web.Context) error {
		// If demo mode is not enabled or key is missing/invalid, return 404
		// (demo mode should be invisible in production)
		if !env.IsDemoMode() {
			return c.NotFound()
		}

		key := c.QueryParam("key")
		if key == "" || key != env.Config.Demo.LoginKey {
			return c.NotFound()
		}

		persona := c.Param("persona")
		email, ok := personaEmails[persona]
		if !ok {
			return c.NotFound()
		}

		// Query user by email
		userByEmail := &query.GetUserByEmail{Email: email}
		err := bus.Dispatch(c, userByEmail)
		if err != nil {
			// User not found in database (seed data not loaded?)
			return c.NotFound()
		}

		// Set auth cookie
		webutil.AddAuthUserCookie(c, userByEmail.Result)

		// Redirect to / or sanitized redirect path
		redirectPath := c.QueryParam("redirect")
		if redirectPath == "" {
			redirectPath = "/"
		} else {
			// Sanitize redirect path to prevent open redirects
			parsed, err := url.Parse(redirectPath)
			if err != nil || parsed.Host != "" || parsed.Scheme != "" {
				redirectPath = "/"
			}
		}

		return c.Redirect(redirectPath)
	}
}
