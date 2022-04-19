# build stage
FROM golang:1.18-alpine AS builder
WORKDIR /code

RUN apk add --no-cache git

# download and cache go dependencies
COPY go.mod .
COPY go.sum .
RUN go mod download

# then copy source code as the last step
COPY . .

RUN go build ./cmd/serverd

# final stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates

COPY --from=builder /code/serverd /serverd
ENTRYPOINT ["./serverd"]
EXPOSE 8080
