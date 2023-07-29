# syntax=docker/dockerfile:1

FROM public.ecr.aws/docker/library/golang:1.20-alpine as  build

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY ./ /app/

RUN go build -o /bff main.go

FROM public.ecr.aws/docker/library/alpine:latest as prod

RUN apk update \
    && apk upgrade \
    && apk add --no-cache ca-certificates \
    && update-ca-certificates

WORKDIR /

COPY --from=build /bff /bff

EXPOSE 80

CMD ["/bff"]
