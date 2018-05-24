FROM golang
WORKDIR /go/src/github.com/Pungyeon/coffee
COPY main.go .
RUN go build main.go
EXPOSE 8080
ENTRYPOINT ["./main"]