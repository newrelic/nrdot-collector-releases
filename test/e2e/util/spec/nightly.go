package spec

type NightlySystemUnderTest struct {
	TestKeyPattern  string
	ExcludedMetrics []string
	SkipIf          func(testSpec *TestSpec) bool
}
