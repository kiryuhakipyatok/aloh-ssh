FROM golang:1.26-alpine AS builder

WORKDIR /usr/local/src

COPY go.mod go.sum ./
RUN go mod download

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    go install github.com/pressly/goose/v3/cmd/goose@latest

COPY . .

ARG TARGETARCH

RUN CGO_ENABLED=0 GOOS=linux GOARCH=$TARGETARCH go build \
    -trimpath \
    -ldflags="-s -w" \
    -p 4 \
    -o /app-binary cmd/app/main.go

RUN apk add --no-cache file && file /app-binary

FROM alpine AS runner

RUN apk add --no-cache tzdata make
ENV TZ=Europe/Minsk
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

WORKDIR /app

COPY --from=builder /app-binary ./app
COPY --from=builder /go/bin/goose /usr/local/bin/goose

RUN chmod +x ./app /usr/local/bin/goose

COPY configs /configs
COPY internal/migrations /migrations

EXPOSE 3333

CMD ["./app"]