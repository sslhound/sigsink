# syntax=docker/dockerfile:experimental
FROM golang:1.13-alpine3.11 as slim-build

WORKDIR /src

RUN apk add --no-cache --update \
        openssh-client \
        git \
        curl \
        build-base \
        ca-certificates

RUN mkdir -p -m 0600 ~/.ssh && ssh-keyscan -t rsa github.com >> ~/.ssh/known_hosts

RUN set -o pipefail && ssh-keygen -F github.com -l -E sha256 \
        | grep -q "SHA256:nThbg6kXUpJWGl7E1IGOCspRomTxdCARLviKw6E5SY8"

ENV PATH "/go/bin/:$PATH"

ARG GIT_COMMIT
ARG RELEASE_CODE
ARG BUILD_TIME

COPY ./go.mod ./go.sum ./
RUN --mount=type=cache,target=/root/.cache/go-build --mount=type=ssh mkdir -p /var/ssh && GIT_SSH_COMMAND="ssh -o \"ControlMaster auto\" -o \"ControlPersist 300\" -o \"ControlPath /var/ssh/%r@%h:%p\"" go mod download

COPY . .

RUN --mount=type=cache,target=/root/.cache/go-build \
        GOOS=linux GOARCH=amd64 go install \
        -ldflags "-w -s -extldflags '-static' -X main.GitCommit=$GIT_COMMIT -X main.ReleaseCode=$RELEASE_CODE -X 'main.BuildTime=$BUILD_TIME'" \
        github.com/sslhound/sigsink/...

FROM alpine:3.11 as sigsink
RUN apk add --no-cache --update ca-certificates
RUN mkdir -p /sigsink
WORKDIR /sigsink
COPY --from=slim-build /go/bin/sigsink /go/bin/
EXPOSE 7000
ENTRYPOINT ["/go/bin/sigsink"]
CMD []
