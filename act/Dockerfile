FROM catthehacker/ubuntu:act-latest
COPY certs/ /usr/local/share/ca-certificates/
RUN update-ca-certificates
# some actions (like azure setup helm) uses node.js, which utilizes its own packaged CA unless otherwise specified.
RUN cat /usr/local/share/ca-certificates/*.crt > /usr/local/share/ca-certificates/all-certs.crt
ENV NODE_EXTRA_CA_CERTS=/usr/local/share/ca-certificates/all-certs.crt