module github.com/bdk38/HellPot

go 1.25.8

replace github.com/mitchellh/reflectwalk => github.com/bdk38/reflectwalk v1.0.4 // use fork with HellPot-specific patches

replace github.com/mitchellh/copystructure => github.com/bdk38/copystructure v1.2.2 // use fork with HellPot-specific patches

require (
	github.com/fasthttp/router v1.5.4
	github.com/knadh/koanf/parsers/toml/v2 v2.2.0
	github.com/knadh/koanf/providers/env/v2 v2.0.0
	github.com/knadh/koanf/providers/file v1.2.1
	github.com/knadh/koanf/v2 v2.3.3
	github.com/rs/zerolog v1.34.0
	github.com/valyala/fasthttp v1.69.0
	golang.org/x/term v0.41.0
)

require (
	github.com/andybalholm/brotli v1.2.0 // indirect
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/go-viper/mapstructure/v2 v2.4.0 // indirect
	github.com/klauspost/compress v1.18.2 // indirect
	github.com/knadh/koanf/maps v0.1.2 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/pelletier/go-toml/v2 v2.2.4 // indirect
	github.com/savsgio/gotils v0.0.0-20240704082632-aef3928b8a38 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	golang.org/x/sys v0.42.0 // indirect
)
