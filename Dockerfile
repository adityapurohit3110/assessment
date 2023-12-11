FROM golang:1.21.5

WORKDIR  /app

COPY go.mod ./

RUN go mod download

COPY *.go ./

RUN CGO_ENABLE=0 GOOS=linux go build -o /task

EXPOSE 8000

CMD ["/task"]