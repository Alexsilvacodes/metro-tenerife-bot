FROM golang:alpine

RUN apk add --no-cache git

WORKDIR /app/notams

COPY go.mod .

RUN go mod download

COPY . .

ENTRYPOINT ["go", "run", "main.go"]
