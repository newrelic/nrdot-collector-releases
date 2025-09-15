package spec

type NightlySystemUnderTest struct {
	HostNamePattern string
	ExcludedMetrics []string
	SkipIf          func(testSpec *TestSpec) bool
}
