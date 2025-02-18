FROM alpine:3.21 as certs
RUN apk --update add ca-certificates

FROM scratch

ARG USER_UID=10001
USER ${USER_UID}

COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --chmod=755 nrdot-collector-k8s /nrdot-collector-k8s
COPY config-daemonset.yaml /etc/nrdot-collector-k8s/config-daemonset.yaml
COPY config-deployment.yaml /etc/nrdot-collector-k8s/config-deployment.yaml

ENTRYPOINT ["/nrdot-collector-k8s"]
CMD ["--config", "/etc/nrdot-collector-k8s/config-daemonset.yaml"]

# `4137` and `4318`: OTLP
EXPOSE 4317 4318