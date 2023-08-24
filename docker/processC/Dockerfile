FROM golang:1.19-alpine

RUN apk update \
  && apk --no-cache --update add build-base

WORKDIR /app
COPY pkg/go.mod ./
COPY pkg/go.sum ./
COPY pkg/job.go ./
COPY pkg/job_test.go ./
COPY pkg/job_insert.sql ./
COPY pkg/job_select.sql ./
COPY pkg/job_update.sql ./
COPY pkg/cmd ./cmd
COPY pkg/devtools ./devtools

RUN go mod download
RUN go build -o processC ./cmd/processC

CMD [ "./processC" ]