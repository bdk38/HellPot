> **Note**: This is a community-maintained fork of HellPot. The original author, [yunginnanet](https://github.com/yunginnanet), has passed away. This fork exists to continue maintaining and supporting this project for the community. All credit for the original creation and design goes to yunginnanet. The original repository can be found at [yunginnanet/HellPot](https://github.com/yunginnanet/HellPot).

---

<div align="center">
  <img src="https://tcp.ac/i/00ctL.gif" alt="HellPot"/>

[![GoDoc](https://godoc.org/github.com/bdk38/HellPot?status.svg)](https://godoc.org/github.com/bdk38/HellPot) [![Go Report Card](https://goreportcard.com/badge/github.com/bdk38/HellPot)](https://goreportcard.com/report/github.com/bdk38/HellPot) [![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://github.com/bdk38/HellPot/blob/main/LICENSE) [![Go Version](https://img.shields.io/github/go-mod/go-version/bdk38/HellPot)](https://github.com/bdk38/HellPot) [![IRC](https://img.shields.io/badge/ircd.chat-%23tcpdirect-blue.svg)](ircs://ircd.chat:6697/#tcpdirect) [![Mentioned in Awesome Honeypots](https://awesome.re/mentioned-badge.svg)](https://github.com/paralax/awesome-honeypots)

</div>

## Summary

HellPot is an endless honeypot based on [Heffalump](https://github.com/carlmjohnson/heffalump) that sends unruly HTTP bots to hell.

Notably it implements a [toml configuration file](https://github.com/knadh/koanf), has [JSON logging](https://github.com/rs/zerolog), and comes with significant performance gains.

## Grave Consequences

Clients (hopefully bots) that disregard `robots.txt` and connect to your instance of HellPot will **suffer eternal consequences**.

HellPot will send an infinite stream of data that is _just close enough_ to being a real website that they might just stick around until their soul is ripped apart and they cease to exist.

Under the hood of this eternal suffering is a markov engine that chucks bits and pieces of [The Birth of Tragedy (Hellenism and Pessimism)](https://www.gutenberg.org/files/51356/51356-h/51356-h.htm) by Friedrich Nietzsche at the client using [fasthttp](https://github.com/valyala/fasthttp).

## Building From Source

HellPot should probably be built with Go version 1.17 or higher.

HellPot uses [go modules](https://go.dev/blog/using-go-modules). This should make it dead simple to build with a stock Go installation. To make it even simpler, we've added a GNU Makefile.

1 ) `git clone https://github.com/bdk38/HellPot`

2 ) `cd HellPot`

4 ) `make`

5 ) _Consider the potential grave consequences of your actions._

## Usage

### Quick Start (User Installation):

1 ) Download a [compiled release](https://github.com/bdk38/HellPot/releases/latest)

2 ) Generate default config:
   ```bash
   ./HellPot --genconfig
   ```

3 ) Edit `config.toml` as desired.

4 ) Run HellPot:
   ```bash
   ./HellPot --config config.toml
   ```

---

### System-Wide Installation:

For production deployments or systemd services:

1 ) Install binary:
   ```bash
   sudo cp HellPot /usr/local/bin/
   ```

2 ) Create config directory and generate config:
   ```bash
   sudo mkdir -p /etc/hellpot
   cd /etc/hellpot
   sudo /usr/local/bin/HellPot --genconfig
   ```

3 ) Edit `/etc/hellpot/config.toml` as desired.

4 ) Run with explicit config path:
   ```bash
   sudo /usr/local/bin/HellPot --config /etc/hellpot/config.toml
   ```

5 ) (Optional) Create dedicated user for better security:
   ```bash
   sudo useradd -r -s /bin/false hellpot
   sudo chown hellpot:hellpot /etc/hellpot/config.toml
   sudo -u hellpot /usr/local/bin/HellPot --config /etc/hellpot/config.toml
   ```

---

### Alternative: Use Repository Config

You can also copy the example config directly from the repository:

```bash
wget https://raw.githubusercontent.com/bdk38/HellPot/main/config.toml
# Edit config.toml as needed
./HellPot --config config.toml
```

---

## Configuration Reference


> [!TIP]
> Configuration values can be overridden with environment variables prefixed with `HELLPOT_`.
> When using this method, replace underscores in configuration keys with two underscores.
>
> e.g:
> to set `http.bind_addr` via env, set `HELLPOT_HTTP_BIND__ADDR="x.x.x.x"`


```toml
[deception]
  # Used as "Server" HTTP header. Note that reverse proxies may hide this.
  server_name = "nginx"

[http]
  # TCP Listener (default)
  bind_addr = "127.0.0.1"
  bind_port = "8080"

  # header name containing clients real IP, for reverse proxy deployments
  real_ip_header = 'X-Real-IP'

  # this contains a list of blacklisted useragent strings. (case sensitive)
  # clients with useragents containing any of these strings will receive "Not found" for any requests.
  uagent_string_blacklist = ["Cloudflare-Traffic-Manager", "curl"]

  # Unix Socket Listener (will override default)
  unix_socket_path = "/var/run/hellpot"
  unix_socket_permissions = "0666"
  use_unix_socket = false

  [http.router]
    # Toggling this to true will cause all GET requests to match. Forces makerobots = false.
    catchall = false
    # Toggling this to false will prevent creation of robots.txt handler.
    makerobots = true
    # Handlers will be created for these paths, as well as robots.txt entries. Only valid if catchall = false.
    paths = ["wp-login.php", "wp-login"]

[logger]
  # verbose (-v)
  debug = true
  # extra verbose (-vv)
  trace = false
  # JSON log files will be stored in the below directory. Defaults to ~/.local/share/HellPot/logs/ if unset.
  # directory = "/path/to/logs/"
  # disable all color in console output. when using Windows this will default to true.
  nocolor = false
  # toggles the use of the current date as the names for new log files.
  use_date_filename = true

[performance]
  # max_workers is only valid if restrict_concurrency is true
  max_workers = 256
  restrict_concurrency = false
```

### Example Reverse Proxy Configs

#### nginx

<details>
  <summary>nginx</summary>

```nginx
location '/robots.txt' {
	proxy_set_header Host $host;
	proxy_set_header X-Real-IP $remote_addr;
	proxy_pass http://127.0.0.1:8080$request_uri;
}

location '/wp-login.php' {
	proxy_set_header Host $host;
	proxy_set_header X-Real-IP $remote_addr;
	proxy_pass http://127.0.0.1:8080$request_uri;
}
```

</details>

#### Apache

<details>
  <summary>apache (mod_proxy + mod_proxy_http)</summary>

All nonexisting URLs are being reverse proxied to a HellPot instance on localhost, which is set to catchall. Traffic served by HellPot is rate limited to 5 KiB/s.

- Create your normal robots.txt and usual content. Also create the fake Errordocument directory and files (files can be empty). In the example, the directory is "/content/"
- A request on a URL with an existing handler (f.e. a file) will be handled by apache
- Requests on nonexisting URLs cause a HTTP Error 404, which content is served by HellPot
- URLs under the "/.well-known/" suffix are excluded.

```apache
<VirtualHost yourserver>
    ErrorDocument 400 "/content/400"
    ErrorDocument 403 "/content/403"
    ErrorDocument 404 "/content/404"
    ErrorDocument 500 "/content/405"
    <Directory "$wwwroot/.well-known/">
        ErrorDocument 400 default
        ErrorDocument 403 default
        ErrorDocument 404 default
        ErrorDocument 500 default
    </Directory>
    /* HTTP Honeypot / HellPot (need mod_proxy, mod_proxy_http) */
    ProxyPreserveHost	on
    ProxyPass         "/content/" "http://localhost:8080/"
    ProxyPassReverse  "/content/" "http://localhost:8080/"

    /* Rate Limit config, need mod_ratelimit */
    <Location "/content/">
        SetOutputFilter RATE_LIMIT
        SetEnv rate-limit 5
    </Location>

    /* Remaining config */

</VirtualHost>
```

</details>

## Related Suffering

- https://github.com/ginger51011/pandoras_pot
  - A HellPot inspired HTTP honeypot to punish and educate unruly web crawlers, written in Rust (🦀)
