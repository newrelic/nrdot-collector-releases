# Security

## Reporting Vulnerabilities

As noted in our [security policy](https://github.com/newrelic/nrdot-collector-releases/security/policy),
New Relic is committed to the privacy and security of our customers and their data. If you believe
you have found a security vulnerability in this project, please report it through our
[coordinated disclosure program](https://github.com/newrelic/nrdot-collector-releases/security/policy#coordinated-disclosure-program).

## GPG Signing Key

Linux packages and archives are signed with the following GPG key:

| Field       | Value                                                        |
|-------------|--------------------------------------------------------------|
| Fingerprint | `87768BAEA82E2B136FB75CD61F2D1176E50959B0`                  |
| UID         | opentelemetry (NewRelic) <opentelemetry@newrelic.com>        |

This key was rotated as part of a security review. The previous key
(`8ECAA86AB2C1904FAAC12E34B0EE4ACC08A81CD2`) was used for releases v1.11.1 and earlier and remains
in `nrdot.gpg` for verification of those releases.

To import the key:
```bash
curl -s "https://raw.githubusercontent.com/newrelic/nrdot-collector-releases/refs/tags/${RELEASE}/nrdot.gpg" | gpg --import
```
