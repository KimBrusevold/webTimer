FROM golang:1.22-alpine
WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY ./ ./

RUN rm ./.env
RUN go build -o webtimer ./cmd/webtimer/webtimer.go

ENV GIN_MODE=release
ENV PORT=8080

EXPOSE 8080

ENTRYPOINT ["./webtimer"]