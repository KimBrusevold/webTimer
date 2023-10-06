FROM golang:1.21-alpine
WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY ./ ./

RUN go build -o webtimer ./cmd/webtimer/webtimer.go

ENV PORT=8080
ENV HOSTURL=http://*:$PORT
ENV DATABASE_URL=file:\web.db
ENV GIN_MODE=release

EXPOSE 8080

ENTRYPOINT ["./webtimer"]