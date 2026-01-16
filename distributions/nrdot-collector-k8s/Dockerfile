FROM alpine:3.21 AS certs
RUN apk --update add ca-certificates

FROM scratch

ARG USER_UID=10001
ARG DIST_NAME="nrdot-collector-k8s"
USER ${USER_UID}

COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --chmod=755 ${DIST_NAME} /nrdot-collector-k8s
COPY LICENSE_APACHE_nrdot-collector-k8s /usr/share/doc/LICENSE_APACHE_nrdot-collector-k8s
COPY THIRD_PARTY_NOTICES.md /usr/share/doc/THIRD_PARTY_NOTICES.md
ENTRYPOINT ["/nrdot-collector-k8s"]

# `4137` and `4318`: OTLP
EXPOSE 4317 4318
