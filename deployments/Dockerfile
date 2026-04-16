FROM alpine:3.22

RUN apk --no-cache add ca-certificates tzdata

COPY openvpn-as-exporter /usr/local/bin/openvpn-as-exporter

EXPOSE 9176

ENTRYPOINT ["openvpn-as-exporter"]
