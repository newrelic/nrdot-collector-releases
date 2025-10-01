# NRDOT FIPS Compliance

## What is FIPS?

FIPS (Federal Information Processing Standards) are a set of computer security standards developed by [NIST (National Institute of Standards and Technology)](https://csrc.nist.gov/projects/cryptographic-module-validation-program/certificate/4953) and used by non-military government agencies and contractors. For encryption guidance, look [here](https://newrelic.atlassian.net/wiki/spaces/STAN/pages/3500179508/Encryption+-+New+FY25).

## How does NRDOT achieve FIPS compliance?

NRDOT achieves FIPS compliance by having dependencies that are also FIPS compliant.
Go 1.24 is not verified as FIPS compliant, but we can force our distro to only use cryptographic functions from BoringCrypto, which is an [approved encryption library](https://newrelic.atlassian.net/wiki/spaces/STAN/pages/2788884481/Approved+Encryption+Libraries).

## Which distributions are FIPS compliant?

Where a given non-compliant distribution may be named something like:

```
nrdot-collector-host_linux_amb64_v1
```

The corresponding FIPS-compliant distribution would have `fips` added in the name:

```
nrdot-collector-host-fips_linux_amb64_v1
```

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
# TLS Handshake
