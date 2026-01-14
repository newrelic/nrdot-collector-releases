FROM alpine:3.21 AS certs
RUN apk --update add ca-certificates

FROM scratch

ARG USER_UID=10001
ARG DIST_NAME="nrdot-collector-experimental"
USER ${USER_UID}

COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --chmod=755 ${DIST_NAME} /nrdot-collector-experimental
COPY LICENSE_APACHE_nrdot-collector-experimental /usr/share/doc/LICENSE_APACHE_nrdot-collector-experimental
COPY THIRD_PARTY_NOTICES.md /usr/share/doc/THIRD_PARTY_NOTICES.md
ENTRYPOINT ["/nrdot-collector-experimental"]
# `4137` and `4318`: OTLP
EXPOSE 4317 4318
