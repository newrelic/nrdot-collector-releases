// Copyright 2020 New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0
package spec

import (
	"testing"
)

func TestRenderWhereClause(t *testing.T) {
	testCaseSpec := LoadTestCaseSpec("host")
	actual := testCaseSpec.RenderWhereClause(map[string]string{
		"hostName": "nrdot-collector-foobar",
	})
	if actual != "WHERE host.name like 'nrdot-collector-foobar'" {
		t.Fatalf("unexpected where clause: %s", actual)
	}
}

func TestRenderWhereClauseFailsIfExpectedVarMissing(t *testing.T) {
	testCaseSpec := LoadTestCaseSpec("host")
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic if var not present")
		}
	}()
	testCaseSpec.RenderWhereClause(map[string]string{
		"hostName1": "nrdot-collector-foobar",
	})

}
