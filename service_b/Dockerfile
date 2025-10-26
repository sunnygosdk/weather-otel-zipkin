FROM golang:1.25

WORKDIR /app

COPY . .

RUN go build -o serviceb ./cmd/main.go

EXPOSE 8081

CMD [ "./serviceb"]