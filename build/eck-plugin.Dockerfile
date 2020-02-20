FROM alpine as certs
RUN apk update && apk add ca-certificates


FROM busybox:glibc

ARG BIN_DIR
ARG PLUGIN_NAME

COPY --from=certs /etc/ssl/certs /etc/ssl/certs
COPY _output/bin/${PLUGIN_NAME} /usr/local/bin/${PLUGIN_NAME}
RUN chmod +x /usr/local/bin/${PLUGIN_NAME}

ENV PLUGIN=${PLUGIN_NAME}

ENTRYPOINT $PLUGIN
