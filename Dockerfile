FROM golang:1.17 as builder

COPY . /go/src/github.com/w-k-s/simple-budget-tracker

WORKDIR /go/src/github.com/w-k-s/simple-budget-tracker

RUN go get ./...

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o app github.com/w-k-s/simple-budget-tracker/cmd/server

FROM alpine:latest

RUN apk --no-cache add ca-certificates
WORKDIR /root/

RUN mkdir -p /root/.budget/migrations.d
COPY --from=builder /go/src/github.com/w-k-s/simple-budget-tracker/app .
COPY --from=builder /go/src/github.com/w-k-s/simple-budget-tracker/migrations .budget/migrations.d

ENTRYPOINT ["./app", "-file=s3://com.wks.budget/config.toml"]