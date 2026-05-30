package http

import (
	"github.com/bdk38/HellPot/internal/config"
	"github.com/valyala/fasthttp"
)

// robotsBody is built once at startup; config.HTTP.Router.Paths is fixed after init.
var robotsBody []byte

// initRobots pre-builds the robots.txt response body.
// Must be called from Serve() before the router is configured.
func initRobots() {
	var b []byte
	b = append(b, "User-agent: *\r\n"...)
	for _, p := range config.HTTP.Router.Paths {
		b = append(b, "Disallow: /"...)
		b = append(b, p...)
		b = append(b, "\r\n"...)
	}
	b = append(b, "\r\n"...)
	robotsBody = b
}

func robotsTXT(ctx *fasthttp.RequestCtx) {
	slog := log.With().
		Str("USERAGENT", string(ctx.UserAgent())).
		Str("REMOTE_ADDR", getRealRemote(ctx)).
		Str("URL", string(ctx.RequestURI())).Logger()

	ctx.SetContentType("text/plain; charset=utf-8")

	slog.Debug().
		Strs("PATHS", config.HTTP.Router.Paths).
		Msg("SERVE_ROBOTS")

	if _, err := ctx.Write(robotsBody); err != nil {
		slog.Error().Err(err).Msg("SERVE_ROBOTS_ERROR")
	}
}
