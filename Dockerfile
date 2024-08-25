FROM golang:1.22-alpine3.20 AS build_deps

RUN apk add --no-cache git

WORKDIR /workspace
ENV GO111MODULE=on
ENV GOPATH="/workspace/.go"

COPY go.mod .
COPY go.sum .

RUN go mod download

FROM build_deps AS build

COPY . .

RUN CGO_ENABLED=0 go build -C ./cmd/cert-manager-webhook-duckdns -o webhook -ldflags '-w -extldflags "-static"'

FROM alpine:3.20

RUN apk add --no-cache ca-certificates

COPY --from=build /workspace/cmd/cert-manager-webhook-duckdns/webhook /usr/local/bin/webhook

USER nobody:nobody

ENTRYPOINT ["webhook"]
