FROM golang:1.19

WORKDIR /go/src/app
COPY . .

WORKDIR /go/src/app/example

RUN go get -d -v ./...
RUN go install -v ./...

CMD [ "go", "run", "main.go" ]




