# NRDOT FIPS Compliance

## What is FIPS?

FIPS (Federal Information Processing Standards) are a set of computer security standards developed by [NIST (National Institute of Standards and Technology)](https://csrc.nist.gov/projects/cryptographic-module-validation-program/certificate/4953) and used by non-military government agencies and contractors. For encryption guidance, look [here](https://newrelic.atlassian.net/wiki/spaces/STAN/pages/3500179508/Encryption+-+New+FY25).

We are currently targeting FIPS version 1.40.2 .

## How does NRDOT achieve FIPS compliance?

NRDOT achieves FIPS compliance by having dependencies that are also FIPS compliant.
Go 1.24 is not verified as FIPS compliant, but we can force our distro to only use cryptographic functions from BoringCrypto, which is an approved encryption library.

## Which distributions are FIPS compliant?

Where a given non-compliant distribution may be named something like:

```
nrdot-collector-host_linux_amb64_v1
```

The corresponding FIPS-compliant distribution would have `fips` added in the name:

```
nrdot-collector-host-fips_linux_amb64_v1
```

_Note: FIPS-compliant distributions are only available for linux_

## Validation

### Use of BoringCrypto

If you run the following command, you can verify that BoringCrypto functions are being used.


```
$ go tool nm nrdot-collector-host-fips_linux_amb64_v1 | grep '_Cfunc__goboringcrypto_'
 5053220 T _cgo_39a3e70c2c46_Cfunc__goboringcrypto_AES_cbc_encrypt
 5053250 T _cgo_39a3e70c2c46_Cfunc__goboringcrypto_AES_ctr128_encrypt
 5053280 T _cgo_39a3e70c2c46_Cfunc__goboringcrypto_AES_decrypt
 50532a0 T _cgo_39a3e70c2c46_Cfunc__goboringcrypto_AES_encrypt
```
## TLS Handshake
First, make sure you install nmap. 

```
brew install nmap
```

Next, generate your certificate.

``` 
openssl genrsa -out server.key 2048 2>/dev/null
openssl req -new -x509 -key server.key -out server.crt -days 365 -sha256 \
    -subj "/C=US/O=Test/CN=localhost" \
    -addext "subjectAltName=DNS:localhost,DNS:host.docker.internal,IP:127.0.0.1" 2>/dev/null
```

After that, create a `config.yaml` file and populate it with the following configurations.

``` 
receivers:
  otlp:
    protocols:
      http:
        endpoint: 0.0.0.0:4318
        tls:
          cert_file: /certs/server.crt
          key_file: /certs/server.key
  hostmetrics:
    collection_interval: 1s
    scrapers:
      memory:

exporters:
  otlphttp:
    endpoint: https://localhost:8443
    tls:
      insecure_skip_verify: true

service:
  pipelines:
    metrics:
      receivers: [otlp, hostmetrics]
      exporters: [otlphttp]
```

Run the server.

``` 
openssl s_server \
     -accept 0.0.0.0:8443 \
     -cert server.crt \
     -key server.key \
     -trace \
     -state \
     -www
```

Then nmap.

``` 
nmap -sV --script ssl-enum-ciphers -p 8443 localhost
```

And finally, run docker.

``` 
docker run -d --name "$CONTAINER_NAME" --network host \
        -v "$CERT_DIR:/certs:ro" -v "$CONFIG_FILE:/config.yaml:ro" \
        "$DOCKER_IMAGE" --config=/config.yaml >/dev/null
```

You should get an output similar to the following: 

```
Starting Nmap 7.98 ( https://nmap.org ) at 2025-10-02 16:04 -0400
Nmap scan report for localhost (127.0.0.1)
Host is up (0.00011s latency).
Other addresses for localhost (not scanned): ::1

PORT     STATE SERVICE       VERSION
8443/tcp open  ssl/https-alt
|_http-trane-info: Problem with XML parsing of /evox/about
| fingerprint-strings:
|   FourOhFourRequest, GetRequest:
|     HTTP/1.0 200 ok
|     Content-type: text/html
|     <HTML><BODY BGCOLOR="#ffffff">
|     <pre>
|     s_server -accept 0.0.0.0:8443 -cert server.crt -key server.key -trace -state -www
|     This TLS version forbids renegotiation.
|     Ciphers supported in s_server binary
|     TLSv1.3 :TLS_AES_256_GCM_SHA384 TLSv1.3 :TLS_CHACHA20_POLY1305_SHA256
|     TLSv1.3 :TLS_AES_128_GCM_SHA256 TLSv1.2 :ECDHE-ECDSA-AES256-GCM-SHA384
|     TLSv1.2 :ECDHE-RSA-AES256-GCM-SHA384 TLSv1.2 :DHE-RSA-AES256-GCM-SHA384
|     TLSv1.2 :ECDHE-ECDSA-CHACHA20-POLY1305 TLSv1.2 :ECDHE-RSA-CHACHA20-POLY1305
|     TLSv1.2 :DHE-RSA-CHACHA20-POLY1305 TLSv1.2 :ECDHE-ECDSA-AES128-GCM-SHA256
|     TLSv1.2 :ECDHE-RSA-AES128-GCM-SHA256 TLSv1.2 :DHE-RSA-AES128-GCM-SHA256
|     TLSv1.2 :ECDHE-ECDSA-AES256-SHA384 TLSv1.2 :ECDHE-RSA-AES256-SHA384
|     TLSv1.2 :DHE-RSA-AES256-SHA256 TLSv1.2 :ECDHE-ECDSA-AES128-SHA256
|_    TLSv1.2 :ECDHE-RSA
| ssl-enum-ciphers:
|   TLSv1.2:
|     ciphers:
|       TLS_DHE_RSA_WITH_AES_128_CBC_SHA (dh 2048) - A
|       TLS_DHE_RSA_WITH_AES_128_CBC_SHA256 (dh 2048) - A
|       TLS_DHE_RSA_WITH_AES_128_GCM_SHA256 (dh 2048) - A
|       TLS_DHE_RSA_WITH_AES_256_CBC_SHA (dh 2048) - A
|       TLS_DHE_RSA_WITH_AES_256_CBC_SHA256 (dh 2048) - A
|       TLS_DHE_RSA_WITH_AES_256_GCM_SHA384 (dh 2048) - A
|       TLS_DHE_RSA_WITH_CHACHA20_POLY1305_SHA256 (dh 2048) - A
|       TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA (secp256r1) - A
|       TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256 (secp256r1) - A
|       TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256 (secp256r1) - A
|       TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA (secp256r1) - A
|       TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA384 (secp256r1) - A
|       TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384 (secp256r1) - A
|       TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256 (secp256r1) - A
|       TLS_RSA_WITH_AES_128_CBC_SHA (rsa 2048) - A
|       TLS_RSA_WITH_AES_128_CBC_SHA256 (rsa 2048) - A
|       TLS_RSA_WITH_AES_128_GCM_SHA256 (rsa 2048) - A
|       TLS_RSA_WITH_AES_256_CBC_SHA (rsa 2048) - A
|       TLS_RSA_WITH_AES_256_CBC_SHA256 (rsa 2048) - A
|       TLS_RSA_WITH_AES_256_GCM_SHA384 (rsa 2048) - A
|     compressors:
|       NULL
|     cipher preference: client
|   TLSv1.3:
|     ciphers:
|       TLS_AKE_WITH_AES_128_GCM_SHA256 (X25519MLKEM768) - A
|       TLS_AKE_WITH_AES_256_GCM_SHA384 (X25519MLKEM768) - A
|       TLS_AKE_WITH_CHACHA20_POLY1305_SHA256 (X25519MLKEM768) - A
|     cipher preference: client
|_  least strength: A
```