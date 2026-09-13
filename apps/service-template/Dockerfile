FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/service-template ./cmd/server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=build /out/service-template /usr/local/bin/service-template
EXPOSE 50053
ENTRYPOINT ["/usr/local/bin/service-template"]
