FROM ghcr.io/umputun/baseimage/buildgo:latest as build

ARG GIT_BRANCH
ARG GITHUB_SHA
ARG CI

ENV GOFLAGS="-mod=vendor"

ADD . /build
WORKDIR /build

RUN go version

RUN apk add --update git openssh-client ca-certificates; \
    if [ -z "$CI" ] ; then \
    echo "runs outside of CI" && version=$(git rev-parse --abbrev-ref HEAD)-$(git log -1 --format=%h)-$(date +%Y%m%dT%H:%M:%S); \
    else version=${GIT_BRANCH}-${GITHUB_SHA:0:7}-$(date +%Y%m%dT%H:%M:%S); fi && \
    echo "version=$version" && \
    go build -o /build/telegram-system-message-deleter-bot -ldflags "-X main.revision=${version} -s -w"


FROM scratch
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /build/telegram-system-message-deleter-bot /srv/telegram-system-message-deleter-bot
#RUN adduser -s /bin/sh -D -u 1000 app && chown -R app:app /home/app

#USER app
WORKDIR /srv
ENTRYPOINT ["/srv/telegram-system-message-deleter-bot"]
