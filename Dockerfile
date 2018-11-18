FROM golang:alpine
RUN apk add git gcc musl-dev sqlite bash

WORKDIR /go/src/github.com/vigliag/isamuni-go
COPY . .

RUN go get -d -v ./...
RUN go install -v ./...

VOLUME ["/data"]

CMD ["isamuni-go", "serve"]