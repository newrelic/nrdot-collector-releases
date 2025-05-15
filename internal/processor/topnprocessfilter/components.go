package topnprocessfilter

import (
	"go.opentelemetry.io/collector/component"
)

func Components() []component.Factory {
	return []component.Factory{
		NewFactory(),
	}
}
