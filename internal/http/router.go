package http

import (
	"bufio"
	"net"
	"runtime"
	"strings"
	"time"

	"github.com/fasthttp/router"
	"github.com/rs/zerolog"
	"github.com/valyala/fasthttp"

	"github.com/bdk38/HellPot/heffalump"
	"github.com/bdk38/HellPot/internal/config"
)

var log *zerolog.Logger

// getRealRemote returns the real remote address, preferring the configured
// forwarded-IP header when present. The header value is validated with
// net.ParseIP before use — an invalid or spoofed value falls back to the
// direct connection IP so that log entries always contain a well-formed
// address. ip.String() is used rather than the raw header value to normalize
// the representation (canonical IPv6 form, no leading zeros, no injected
// content).
func getRealRemote(ctx *fasthttp.RequestCtx) string {
	xrealip := strings.TrimSpace(string(ctx.Request.Header.Peek(config.HeaderName)))
	if xrealip != "" {
		if ip := net.ParseIP(xrealip); ip != nil {
			return ip.String()
		}
	}
	return ctx.RemoteIP().String()
}

// getSrv constructs and returns the fasthttp server. Concurrency resolution
// and logger initialization are handled by Serve before this is called —
// getSrv has no side effects beyond building the struct.
func getSrv(r *router.Router) fasthttp.Server {
	return fasthttp.Server{
		// Server name sent in the Server response header.
		Name: config.FakeServerName,

		// Cap the time allowed to read the full request including headers.
		ReadTimeout: 5 * time.Second,

		// Limit the request body size. Our handlers never read the body, but
		// this prevents fasthttp from buffering large payloads sent by bots
		// using POST, PUT, or other body-carrying methods before the handler
		// is invoked. Bots sending bodies larger than this receive a 413 from
		// fasthttp before reaching any of our code.
		MaxRequestBodySize: 4096,

		// Limit connections per IP to reduce the impact of a single aggressive
		// scanner or bot net node.
		MaxConnsPerIP: 10,

		// With keepalive disabled, each connection serves exactly one request.
		MaxRequestsPerConn: 1,

		Concurrency: config.MaxWorkers,

		// GetOnly has been removed. Method enforcement is now handled at the
		// router level so that non-GET/HEAD attempts can be logged before
		// rejection. See denyHandler and the MethodNotAllowed/NotFound
		// assignments in Serve.

		// Keep hijacked connections alive through graceful shutdown so that
		// active traps are not released prematurely when HellPot restarts.
		KeepHijackedConns: true,

		CloseOnShutdown: true,

		// Each trapped connection is its own indefinite streaming response;
		// keepalive would never be reached in normal operation.
		DisableKeepalive: true,

		Handler: r.Handler,
		Logger:  log,
	}
}

// Serve registers all routes, builds the server, and begins listening.
func Serve() error {
	log = config.GetLogger()

	// Resolve the worker count before building the server config.
if config.MaxWorkers <= 0 {
    config.MaxWorkers = fasthttp.DefaultConcurrency
}
	// Pre-lowercase the UA blacklist once at startup so the hot path can
	// compare against a single lowercased UA string without allocating
	// per-entry ToLower conversions on every request.
	loweredMatchers := make([]string, len(config.UseragentBlacklistMatchers))
	for i, m := range config.UseragentBlacklistMatchers {
		loweredMatchers[i] = strings.ToLower(m)
	}

	// hellPotHandler is the core trap. It serves an infinite stream of
	// Markov-generated HTML to GET requests that pass the UA blacklist.
	// Content-Type is set to text/html so the response is indistinguishable
	// from a real page to content-type sniffers and browser-based bots.
	//
	// HEAD requests share this route registration; fasthttp suppresses the
	// response body for HEAD at the network level automatically.
	hellPotHandler := func(ctx *fasthttp.RequestCtx) {
		remoteAddr := getRealRemote(ctx)

		// Convert the UA to string once — it's used for both logging and the
		// blacklist check, and each []byte→string conversion allocates.
		ua := string(ctx.UserAgent())

		slog := log.With().
			Str("USERAGENT", ua).
			Str("REMOTE_ADDR", remoteAddr).
			Str("URL", string(ctx.RequestURI())).
			Logger()

		// UA blacklist — lowercase the inbound UA once and compare against the
		// pre-lowercased matchers list. Case-insensitive matching prevents
		// trivial bypasses via casing variations.
		uaLower := strings.ToLower(ua)
		for _, denied := range loweredMatchers {
			if strings.Contains(uaLower, denied) {
				slog.Trace().Msg("IGNORED_UA")
				ctx.Error("Not found", fasthttp.StatusNotFound)
				return
			}
		}

		if config.Trace {
			path, _ := ctx.UserValue("path").(string)
			if path == "" {
				path = "/"
			}
			slog = slog.With().Str("caller", path).Logger()
		}

		slog.Info().Msg("NEW")

		s := time.Now()
		var n int64

		ctx.SetContentType("text/html; charset=utf-8")
		ctx.SetBodyStreamWriter(func(bw *bufio.Writer) {
			wn, err := heffalump.DefaultHeffalump.WriteHell(bw)
			n += wn
			if err != nil {
				slog.Trace().Err(err).Msg("END_ON_ERR")
			}

			slog.Info().
				Int64("BYTES", n).
				Dur("DURATION", time.Since(s)).
				Msg("FINISH")
		})
	}

	// headHandler responds to HEAD requests with matching status and headers
	// but no body. It does not call WriteHell — the stream writer would run
	// indefinitely generating Markov text even though fasthttp suppresses the
	// body for HEAD, wasting CPU for no benefit.
	headHandler := func(ctx *fasthttp.RequestCtx) {
		ctx.SetContentType("text/html; charset=utf-8")
		ctx.SetStatusCode(fasthttp.StatusOK)
	}

	// denyHandler rejects all non-GET/HEAD methods with a 404 and logs the
	// attempt. The request body is never read regardless of method — headers,
	// method, path, and remote IP are the only fields accessed.
	//
	// Methods handled here include all standard HTTP methods (POST, PUT,
	// PATCH, DELETE, OPTIONS, TRACE, CONNECT) and all WebDAV extension methods
	// (ACL, COPY, LOCK, MKCOL, MOVE, PROPFIND, PROPPATCH, UNLOCK). FastHTTP
	// parses all of these as plain method strings so no special cases are
	// needed at the parsing level.
	//
	// 404 is returned rather than 405 (Method Not Allowed) to avoid leaking
	// information about which paths exist or which methods HellPot understands.
	// A 405 would confirm to a scanner that the path is valid; a 404 reveals
	// nothing.
	//
	// CONNECT is logged with a distinct PROXY_ABUSE_ATTEMPT message since it
	// indicates an attempt to use HellPot as a proxy relay rather than a
	// standard scan or crawl probe.
	denyHandler := func(ctx *fasthttp.RequestCtx) {
		method := string(ctx.Method())

		e := log.Warn().
			Str("METHOD", method).
			Str("USERAGENT", string(ctx.UserAgent())).
			Str("REMOTE_ADDR", getRealRemote(ctx)).
			Str("URL", string(ctx.RequestURI()))

		if method == "CONNECT" {
			e.Msg("PROXY_ABUSE_ATTEMPT")
		} else {
			e.Msg("DENIED_METHOD")
		}

		ctx.Error("Not found", fasthttp.StatusNotFound)
	}

	r := router.New()

	// MethodNotAllowed fires when a registered path receives a method that has
	// no handler. NotFound fires for any path/method combination not registered
	// at all. Both are routed to denyHandler so that every non-GET/HEAD probe
	// — regardless of path or method — is logged and returned as 404.
	r.MethodNotAllowed = denyHandler
	r.NotFound = denyHandler

	initRobots()
	if config.MakeRobots && !config.CatchAll {
		r.GET("/robots.txt", robotsTXT)
	}

	if !config.CatchAll {
		for _, p := range config.Paths {
			log.Trace().Str("caller", "router").Msgf("Add route: %s", p)
			r.GET("/"+p, hellPotHandler)
			r.HEAD("/"+p, headHandler)
		}
	} else {
		log.Trace().Msg("Catch-all mode enabled")
		r.GET("/{path:*}", hellPotHandler)
		r.HEAD("/{path:*}", headHandler)
	}

	srv := getSrv(r)

	//goland:noinspection GoBoolExpressions
	if !config.UseUnixSocket || runtime.GOOS == "windows" {
		l := config.HTTPBind + ":" + config.HTTPPort
		log.Info().Str("caller", l).Msg("Listening and serving HTTP...")
		return srv.ListenAndServe(l)
	}

	if len(config.UnixSocketPath) < 1 {
		log.Fatal().Msg("unix_socket_path configuration directive appears to be empty")
	}

	log.Info().Str("caller", config.UnixSocketPath).Msg("Listening and serving HTTP...")
	return listenOnUnixSocket(config.UnixSocketPath, r)
}
