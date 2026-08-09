all: deps check build
format:
	find . -iname "*.go" -exec gofmt -s -l -w {} \;
check:
	go vet ./...
run:
	go run cmd/HellPot/*.go
deps:
	go mod tidy -v
build:
	go build -trimpath -ldflags "-s -w -X main.version=`git tag --sort=-version:refname | head -n 1`" cmd/HellPot/*.go

# Docker helpers
docker-up:
	docker compose up -d --build

docker-logs:
	docker compose logs -f

docker-down:
	docker compose down
