FROM golang:latest as GO
COPY . /build
WORKDIR /build

ENV VERIFICATION_URL=http://hwr-discord-verification.srv01.lh-info.eu/?token={{.}}
RUN go build discord-verification.go mail.go user-verification.go

ENTRYPOINT [ "./discord-verification" ]