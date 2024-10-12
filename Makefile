# Get the latest commit branch, hash, and date
TAG=$(shell git describe --tags --abbrev=0 --exact-match 2>/dev/null)
BRANCH=$(if $(TAG),$(TAG),$(shell git rev-parse --abbrev-ref HEAD 2>/dev/null))
HASH=$(shell git rev-parse --short=7 HEAD 2>/dev/null)
TIMESTAMP=$(shell git log -1 --format=%ct HEAD 2>/dev/null | xargs -I{} date -u -d @{} +%Y%m%dT%H%M%S)
GIT_REV=$(shell printf "%s-%s-%s" "$(BRANCH)" "$(HASH)" "$(TIMESTAMP)")
REV=$(if $(filter --,$(GIT_REV)),latest,$(GIT_REV)) # fallback to latest if not in git repo


docker:
	docker build -t systemnick/telegram-system-message-deleter-bot .

race_test:
	go test -race -mod=vendor -timeout=60s -count 1 ./...

build:
	mkdir -p .bin
	go build -ldflags "-X main.revision=$(REV) -s -w" -o .bin/telegram-system-message-deleter-bot.$(BRANCH)
	cp .bin/telegram-system-message-deleter-bot.$(BRANCH) .bin/telegram-system-message-deleter-bot

test:
	go clean -testcache
	go test -race -coverprofile=coverage.out ./...
	grep -v "_mock.go" coverage.out | grep -v mocks > coverage_no_mocks.out
	go tool cover -func=coverage_no_mocks.out
	rm coverage.out coverage_no_mocks.out


.PHONY: docker race_test build test
