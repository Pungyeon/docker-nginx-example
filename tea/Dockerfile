FROM golang
WORKDIR /go/src/github.com/Pungyeon/tea
COPY main.go .
RUN go build main.go
EXPOSE 8080
ENTRYPOINT ["./main"]