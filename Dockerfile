# CardGame-Admin (マルチステージビルド)
FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/admin ./cmd/admin

FROM alpine:3.21
RUN adduser -D app
USER app
COPY --from=build /out/admin /usr/local/bin/admin
EXPOSE 9090
ENTRYPOINT ["admin"]
