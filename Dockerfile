FROM golang:1.16 as builder

COPY . /go/src/github.com/w-k-s/simple-budget-tracker

WORKDIR /go/src/github.com/w-k-s/simple-budget-tracker

RUN go get ./...

RUN GOOS=linux GOARCH=arm go build -o app github.com/w-k-s/simple-budget-tracker/cmd/server

FROM alpine:latest

RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /go/src/github.com/w-k-s/simple-budget-tracker/app .

ENTRYPOINT ["./app", "-file=s3://com.wks.budget/config.toml"]

