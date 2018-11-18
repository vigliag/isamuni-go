FROM golang:alpine
RUN apk add git gcc musl-dev sqlite bash

WORKDIR /go/src/github.com/vigliag/isamuni-go

# download and cache big dependencies
RUN go get -v github.com/blevesearch/bleve

COPY . .
RUN go get -d -v ./...
RUN go install -v ./...

# save database and index in data
# cwd to that directory to ease running backup scripts
VOLUME ["/data"]
WORKDIR "/data"

CMD ["isamuni-go", "serve"]