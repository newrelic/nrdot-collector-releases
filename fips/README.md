# NRDOT FIPS Compliance

## What is FIPS?

FIPS (Federal Information Processing Standards) are a set of computer security standards developed by NIST (National Institute of Standards and Technology) and used by non-military government agencies and contractors.

We are currently targeting [FIPS version 140-3](https://csrc.nist.gov/pubs/fips/140-3/final).

## How does NRDOT achieve FIPS compliance?

NRDOT achieves FIPS 140-3 compliance by instructing the golang compiler to replace the standard cryptographic library with BoringSSL and ensuring that TLS uses [approved ciphers](https://github.com/newrelic/nrdot-collector-releases/blob/main/fips/validation/validate.sh#L27).

The following demonstrates the complete chain from NRDOT to the official NIST FIPS certificate:

1. **NRDOT Collector**: The FIPS-compliant distributions (`-fips` suffix) are built from this repository
2. **Go Compiler**: Built using [Go 1.26](https://github.com/newrelic/nrdot-collector-releases/blob/main/.github/workflows/ci-base.yaml#L71) with [`GOEXPERIMENT=boringcrypto`](https://github.com/newrelic/nrdot-collector-releases/blob/main/distributions/nrdot-collector/.goreleaser-fips.yaml#L26) flag enabled
3. **BoringSSL Module**: The Go 1.26 runtime embeds BoringSSL commit [`0c6f40132b828e92ba365c6b7680e32820c63fa7`](https://github.com/golang/go/blob/go1.26.0/src/crypto/internal/boring/Dockerfile#L67), which corresponds to the [fips-20220613](https://boringssl.googlesource.com/boringssl/+/refs/tags/fips-20220613) tag
4. **NIST Certificate**: This BoringSSL version is FIPS 140-3 validated under [NIST Certificate #4735](https://csrc.nist.gov/projects/cryptographic-module-validation-program/certificate/4735), which is valid until 2029

Note: Once [golang itself is successfully FIPS 140-3 certified](https://go.dev/doc/security/fips140#in-process-module-versions), we will transition to using golang's native implementation.

## Which distributions are FIPS compliant?

Compliant artifacts have a `-fips` suffix added to the version string, e.g. `1.5.0-fips`.

_Note: FIPS-compliant distributions are only available for linux_

## Validation

### Use of BoringCrypto

If you run the following command, you can verify that BoringCrypto functions are being used.

```
docker build --progress=plain -t fips-analyzer - << 'EOF'
FROM golang:1.26
COPY --from=newrelic/nrdot-collector:latest-fips /nrdot-collector /nrdot-collector
RUN go tool nm /nrdot-collector | grep goboringcrypto
EOF
```

The resulting output should be similar to the following: 

``` 
5053220 T _cgo_39a3e70c2c46_Cfunc__goboringcrypto_AES_cbc_encrypt
5053250 T _cgo_39a3e70c2c46_Cfunc__goboringcrypto_AES_ctr128_encrypt
5053280 T _cgo_39a3e70c2c46_Cfunc__goboringcrypto_AES_decrypt
50532a0 T _cgo_39a3e70c2c46_Cfunc__goboringcrypto_AES_encrypt
```

### Setup for Cipher Verification
First, make sure you install a security scanner. We use nmap. 

Please refer to install instructions [here](https://nmap.org/book/install.html)

Next, generate your certificate.

``` 
openssl genrsa -out server.key 2048 2>/dev/null
openssl req -new -x509 -key server.key -out server.crt -days 365 -sha256 \
    -subj "/C=US/O=Test/CN=localhost" \
    -addext "subjectAltName=DNS:localhost,DNS:host.docker.internal,IP:127.0.0.1" 2>/dev/null
```

Run the openssl server which will accept the otlpexporter's requests. The output is redirected to a log file to capture the TLS handshake details for client cipher verification.

```
openssl s_server \
     -accept 0.0.0.0:8443 \
     -cert server.crt \
     -key server.key \
     -trace \
     -state \
     -www \
     > openssl-server.log 2>&1 &
```

Save the server PID so you can stop it later:
```
echo $! > openssl-server.pid
```

After that, create a `config.yaml` file and populate it with the following configurations.
Note: If testing this on MacOS, you will need to replace `localhost` with `host.docker.internal`

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
    # replace with https://host.docker.internal:8443 when testing on MacOS
    endpoint: https://localhost:8443
    tls:
      insecure_skip_verify: true

service:
  pipelines:
    metrics:
      receivers: [otlp, hostmetrics]
      exporters: [otlphttp]
```

Then run NRDOT via docker.

``` 
docker run -d --name "nrdot-fips" -p 4318:4318 \
        -v "./:/certs:ro" -v "./config.yaml:/config.yaml:ro" \
        newrelic/nrdot-collector:latest-fips --config=/config.yaml >/dev/null
```

### Server Cipher Verification

Configure nmap to scan NRDOT's server (port 4318) to verify which ciphers it accepts:

```
nmap -sV --script ssl-enum-ciphers -p 4318 localhost
```

You should get an output similar to the following (showing only FIPS-compliant ciphers).

For the authoritative list of FIPS-compliant cipher suites, refer to the [`is_cipher_fips_compliant`](https://github.com/newrelic/nrdot-collector-releases/blob/main/fips/validation/validate.sh#L30) function in the automated validation script.

_Note: nmap displays TLS 1.3 cipher names as `TLS_AKE_WITH_...` even though the official cipher names would drop the `AKE_WITH`, see [this issue](https://github.com/nmap/nmap/issues/2883).

```
Starting Nmap 7.99 ( https://nmap.org ) at 2026-04-15 17:52 -0700
Nmap scan report for localhost (127.0.0.1)
Host is up (0.00015s latency).
Other addresses for localhost (not scanned): ::1

PORT     STATE SERVICE  VERSION
4318/tcp open  ssl/http Golang net/http server (Go-IPFS json-rpc or InfluxDB API)
| ssl-enum-ciphers:
|   TLSv1.2:
|     ciphers:
|       TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256 (secp256r1) - A
|       TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384 (secp256r1) - A
|     compressors:
|       NULL
|     cipher preference: server
|   TLSv1.3:
|     ciphers:
|       TLS_AKE_WITH_AES_128_GCM_SHA256 (secp256r1) - A
|       TLS_AKE_WITH_AES_256_GCM_SHA384 (secp256r1) - A
|     cipher preference: server
|_  least strength: A
```


### Client Cipher Verification

To verify which ciphers NRDOT offers when acting as a client (making outbound connections), wait a few seconds for NRDOT to connect to the OpenSSL server, then check the captured log:

```
# Wait for NRDOT to attempt connections
sleep 5

# Check if ClientHello was captured
grep -q "ClientHello" openssl-server.log && echo "✓ Client connection captured" || echo "✗ No client connection found"

# View the cipher suites offered by NRDOT client
grep -A 50 "ClientHello" openssl-server.log | grep -E "TLS_|cipher" | head -20
```

You should get an output similar to the following (showing only FIPS-compliant cipher suites):

```
      cipher_suites (len=12)
        {0xC0, 0x2B} TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256
        {0xC0, 0x2F} TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256
        {0xC0, 0x2C} TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384
        {0xC0, 0x30} TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384
        {0x13, 0x01} TLS_AES_128_GCM_SHA256
        {0x13, 0x02} TLS_AES_256_GCM_SHA384
```

For the authoritative list of FIPS-compliant cipher suites, refer to the [`is_cipher_fips_compliant`](https://github.com/newrelic/nrdot-collector-releases/blob/main/fips/validation/validate.sh#L30) function in the automated validation script.

### Cleanup

Stop and remove the containers and OpenSSL server:

```
docker stop nrdot-fips && docker rm nrdot-fips
kill $(cat openssl-server.pid)
```