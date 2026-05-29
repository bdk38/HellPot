package http

import (
	"github.com/bdk38/HellPot/internal/config"
	"github.com/valyala/fasthttp"
)

// robotsBody is built once at startup; config.Paths is fixed after init.
var robotsBody []byte

// initRobots pre-builds the robots.txt response body.
// Must be called from Serve() before the router is configured.
func initRobots() {
	var b []byte
	b = append(b, "User-agent: *\r\n"...)
	for _, p := range config.Paths {
		b = append(b, "Disallow: /"...)
		b = append(b, p...)
		b = append(b, "\r\n"...)
	}
	b = append(b, "\r\n"...)
	robotsBody = b
}

func robotsTXT(ctx *fasthttp.RequestCtx) {
	slog := alog.With().
		Str("USERAGENT", string(ctx.UserAgent())).
		Str("REMOTE_ADDR", getRealRemote(ctx)).
		Str("URL", string(ctx.RequestURI())).Logger()

	ctx.SetContentType("text/plain; charset=utf-8")

	slog.Log().
		Strs("PATHS", config.Paths).
		Msg("SERVE_ROBOTS")

	if _, err := ctx.Write(robotsBody); err != nil {
		slog.Log().Err(err).Msg("SERVE_ROBOTS_ERROR")
		// Surface write errors in the system log so they are visible
		// alongside other operational errors.
		log.Error().Err(err).
			Str("REMOTE_ADDR", getRealRemote(ctx)).
			Msg("SERVE_ROBOTS_ERROR")
	}
}
