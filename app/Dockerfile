# syntax=docker/dockerfile:1

FROM golang:1.20-alpine as  build

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY ./ /app/

RUN go build -o /productsHandler handlers/productsHandler/main.go

FROM build as dev

RUN go install github.com/cespare/reflex@latest

ENTRYPOINT reflex -g '**/*.go' go run /app/handlers/productsHandler/main.go

FROM scratch as prod

WORKDIR /

COPY --from=build /productsHandler /productsHandler

CMD ["/productsHandler"]



