FROM golang:1.25

WORKDIR /usr/src/app

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build ALL packages, output single binary
RUN go build -v -o /usr/local/bin/app ./cmd

CMD ["app"]