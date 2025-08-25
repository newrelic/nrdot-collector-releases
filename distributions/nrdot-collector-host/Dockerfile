FROM alpine:3.21 as certs
RUN apk --update add ca-certificates

FROM scratch

ARG USER_UID=10001
ARG DIST_NAME="nrdot-collector-host"
USER ${USER_UID}

COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --chmod=755 ${DIST_NAME} /${DIST_NAME}
COPY config.yaml /etc/${DIST_NAME}/config.yaml
ENTRYPOINT ["/${DIST_NAME}"]
CMD ["--config", "/etc/${DIST_NAME}/config.yaml"]
# `4137` and `4318`: OTLP
EXPOSE 4317 4318
