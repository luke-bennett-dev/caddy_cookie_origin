package caddy_cookie_origin

import (
	"net/http"
	"strings"

	"bytes"

	"github.com/caddyserver/caddy/caddyfile"
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
)

func init() {
	caddy.RegisterModule(CookieModifier{})
}

type CookieModifier struct {
	FromDomain string `json:"from_domain"`
	ToDomain   string `json:"to_domain"`
}

func (CookieModifier) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.caddy_cookie_origin",
		New: func() caddy.Module { return new(CookieModifier) },
	}
}

func (cm *CookieModifier) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	if !d.Args(&cm.FromDomain, &cm.ToDomain) {
		return d.ArgErr()
	}
	return nil
}

func (cm CookieModifier) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	// Create a buffer to hold the response body
	buf := new(bytes.Buffer)

	// Define a function that decides when to buffer the response body
	shouldBufferFunc := func(status int, header http.Header) bool {
		// Example: buffer if the response is not a streaming response
		return status < 300 || status >= 400
	}

	// Create a response recorder with a buffer and a function to decide if buffering is necessary
	rec := caddyhttp.NewResponseRecorder(w, buf, shouldBufferFunc)

	// Execute the next handler in the chain and capture the response
	err := next.ServeHTTP(rec, r)
	if err != nil {
		return err
	}

	// Modify the Set-Cookie headers
	cookies := rec.Header()["Set-Cookie"]
	for i, cookie := range cookies {
		if strings.Contains(cookie, "Domain="+cm.FromDomain) {
			cookies[i] = strings.ReplaceAll(cookie, "Domain="+cm.FromDomain, "Domain="+cm.ToDomain)
		}
	}
	rec.Header()["Set-Cookie"] = cookies

	// Write the modified response
	rec.WriteResponse()

	return nil
}
