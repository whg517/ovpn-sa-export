FROM alpine:3.22

RUN apk --no-cache add ca-certificates tzdata

COPY ovpn-sa-export /usr/local/bin/ovpn-sa-export

EXPOSE 9176

ENTRYPOINT ["ovpn-sa-export"]
