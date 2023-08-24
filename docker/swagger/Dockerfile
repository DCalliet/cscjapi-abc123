FROM golang:1.19-alpine

RUN apk update \
  && apk --no-cache --update add build-base

WORKDIR /app
COPY openapi.yaml ./
COPY pkg/go.mod ./
COPY pkg/go.sum ./
COPY pkg/cmd ./cmd
COPY pkg/devtools ./devtools

RUN go mod download
RUN go build -o swagger ./cmd/swagger

CMD [ "./swagger" ]