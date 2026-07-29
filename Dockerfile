FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /server ./cmd/server

FROM alpine:3.21
COPY --from=build /server /server
COPY web/build /static
ENV STATIC_DIR=/static
EXPOSE 8080
ENTRYPOINT ["/server"]
