// Copyright New Relic, Inc. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

//go:build goexperiment.boringcrypto

package main

import (
	"log"
	"os"

	"crypto/boring"
_ "crypto/tls/fipsonly"
)

func init() {
	attestFIPS()
}

func attestFIPS() {
	if boring.Enabled() {
		log.Print("Using BoringSSL and running in FIPS mode")
	} else {
		log.Print("ERROR: not using boringcrypto")
		os.Exit(1)
	}
}
