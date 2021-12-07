# syntax=docker/dockerfile:1

FROM golang:1.16-alpine AS build

WORKDIR /poopchat

COPY . ./

RUN go mod download

RUN CGO_ENABLED=0 go build -o /poopchat_server cmd/server/server.go

FROM scratch

COPY --from=build /poopchat_server .

EXPOSE 8080

CMD [ "/poopchat_server" ]
