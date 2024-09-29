FROM golang:1.22-alpine3.19 as build

WORKDIR /var/backend

COPY . .
COPY server.crt /
COPY server.key /
COPY .env /

COPY go.mod go.mod
COPY go.sum go.sum

RUN go mod tidy
RUN go build -o /main ./cmd/app/main.go

FROM alpine:edge as prod

COPY --from=build /main /main
COPY --from=build server.key .
COPY --from=build server.crt .
COPY --from=build .env .

EXPOSE 8000
EXPOSE 8080

CMD ["./main"]