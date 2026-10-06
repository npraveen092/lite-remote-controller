FROM golang:1.24-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY internal ./internal
COPY cmd ./cmd
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/signaling ./cmd/signaling

FROM alpine:3.22
COPY --from=build /out/signaling /usr/local/bin/signaling
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/signaling"]
