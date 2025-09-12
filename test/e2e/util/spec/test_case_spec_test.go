package spec

import (
	"testing"
)

func TestRenderWhereClause(t *testing.T) {
	testCaseSpec := LoadTestCaseSpec("host")
	actual := testCaseSpec.RenderWhereClause(map[string]string{
		"testKey": "nrdot-collector-host-foobar",
	})
	if actual != "WHERE testKey like 'nrdot-collector-host-foobar'" {
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
		"hostName1": "nrdot-collector-host-foobar",
	})

}
