FROM alpine:3.21 AS certs
RUN apk --update add ca-certificates

FROM scratch

ARG USER_UID=10001
USER ${USER_UID}

COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --chmod=755 nrdot-collector-gateway /nrdot-collector-gateway
COPY config.yaml /etc/nrdot-collector-gateway/config.yaml
ENTRYPOINT ["/nrdot-collector-gateway"]
CMD ["--config", "/etc/nrdot-collector-gateway/config.yaml"]
# `4137` and `4318`: OTLP
EXPOSE 4317 4318
