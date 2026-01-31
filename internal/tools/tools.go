// Copyright New Relic, Inc. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

//go:build tools
// +build tools

package tools // import "github.com/newrelic/opentelemetry-collector-components/internal/tools"
import (
	_ "github.com/newrelic/nrdot-collector-components/cmd/nrlicense"
	_ "go.elastic.co/go-licence-detector"
)
