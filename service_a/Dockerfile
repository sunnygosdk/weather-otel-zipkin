FROM golang:1.25

WORKDIR /app

COPY . .

RUN go build -o servicea ./cmd/main.go

EXPOSE 8080

CMD [ "./servicea"]