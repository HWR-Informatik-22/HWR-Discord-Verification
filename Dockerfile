FROM golang:latest as GO
COPY . /build
WORKDIR /build
RUN go build discord-verification.go mail.go user-verification.go

ENTRYPOINT [ "./discord-verification" ]