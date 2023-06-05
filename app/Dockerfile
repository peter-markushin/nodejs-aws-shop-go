# syntax=docker/dockerfile:1

FROM golang:1.20-alpine as  build

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY ./ /app/

RUN go build -o /lambdaHandler

FROM build as dev

RUN go install github.com/cespare/reflex@latest

CMD reflex -g '**/*.go' go run main.go

FROM scratch as prod

WORKDIR /

COPY --from=build /lambdaHandler /lambdaHandler

CMD ["/lambdaHandler"]



