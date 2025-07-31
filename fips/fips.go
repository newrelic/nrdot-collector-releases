//go:build goexperiment.boringcrypto

package fips

import (
	"log"

"crypto/boring"
_ "crypto/tls/fipsonly"
)

func init() {
	attestFIPS()
}

func attestFIPS() {
	if boring.Enabled() {
		log.Print("Using BoringSSL and running in FIPS mode")
	}
}
