FROM alpine:latest as alpine
RUN apk --no-cache add tzdata zip ca-certificates
WORKDIR /usr/share/zoneinfo
# -0 means no compression.  Needed because go's
# tz loader doesn't handle compressed data.
RUN zip -r -0 /zoneinfo.zip .

FROM scratch

ENV ZONEINFO /zoneinfo.zip
COPY --from=alpine /zoneinfo.zip /

ARG bin
ENV bin ${bin}

COPY ${bin} /${bin}
