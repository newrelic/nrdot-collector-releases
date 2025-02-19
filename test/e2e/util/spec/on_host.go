package spec

import (
	"test/e2e/util/assert"
)

type TestCase struct {
	Name       string
	Metric     assert.NrMetric
	Assertions []assert.NrAssertion
}

// TODO: eventually this should be generated based on a 'test specification' file
var testCases = []TestCase{
	{
		Name: "host receiver cpu.utilization user",
		Metric: assert.NrMetric{
			Name:        "system.cpu.utilization",
			WhereClause: "WHERE state='user'"},
		Assertions: []assert.NrAssertion{
			{AggregationFunction: "max", ComparisonOperator: ">", Threshold: 0},
		}},
	{
		Name: "host receiver cpu.utilization idle",
		Metric: assert.NrMetric{
			Name:        "system.cpu.utilization",
			WhereClause: "WHERE state='idle'"},
		Assertions: []assert.NrAssertion{
			{AggregationFunction: "max", ComparisonOperator: ">", Threshold: 0},
		}},
	{
		Name: "host receiver cpu.utilization wait",
		Metric: assert.NrMetric{
			Name:        "system.cpu.utilization",
			WhereClause: "WHERE state='wait'"},
		Assertions: []assert.NrAssertion{
			{AggregationFunction: "max", ComparisonOperator: ">", Threshold: 0},
		}},
	{
		Name: "host receiver cpu.utilization steal",
		Metric: assert.NrMetric{
			Name:        "system.cpu.utilization",
			WhereClause: "WHERE state='steal'"},
		Assertions: []assert.NrAssertion{
			{AggregationFunction: "max", ComparisonOperator: ">=", Threshold: 0},
		}},
	{
		Name: "host receiver disk.io read",
		Metric: assert.NrMetric{
			Name:        "system.disk.io",
			WhereClause: "WHERE direction='read'"},
		Assertions: []assert.NrAssertion{
			{AggregationFunction: "max", ComparisonOperator: ">=", Threshold: 0},
		}},
	{
		Name: "host receiver disk.io read",
		Metric: assert.NrMetric{
			Name:        "system.disk.io",
			WhereClause: "WHERE direction='write'"},
		Assertions: []assert.NrAssertion{
			{AggregationFunction: "max", ComparisonOperator: ">", Threshold: 0},
		}},
	{
		Name: "host receiver disk.io_time",
		Metric: assert.NrMetric{
			Name: "system.disk.io_time",
		},
		Assertions: []assert.NrAssertion{
			{AggregationFunction: "max", ComparisonOperator: ">", Threshold: 0},
		}},
	{
		Name: "host receiver disk.operation_time read",
		Metric: assert.NrMetric{
			Name:        "system.disk.operation_time",
			WhereClause: "WHERE direction='read'"},
		Assertions: []assert.NrAssertion{
			{AggregationFunction: "max", ComparisonOperator: ">=", Threshold: 0},
		}},
	{
		Name: "host receiver disk.operation_time read",
		Metric: assert.NrMetric{
			Name:        "system.disk.operation_time",
			WhereClause: "WHERE direction='write'"},
		Assertions: []assert.NrAssertion{
			{AggregationFunction: "max", ComparisonOperator: ">", Threshold: 0},
		}},
	{
		Name: "host receiver disk.operations read",
		Metric: assert.NrMetric{
			Name:        "system.disk.operations",
			WhereClause: "WHERE direction='read'"},
		Assertions: []assert.NrAssertion{
			{AggregationFunction: "max", ComparisonOperator: ">=", Threshold: 0},
		}},
	{
		Name: "host receiver disk.operations write",
		Metric: assert.NrMetric{
			Name:        "system.disk.operations",
			WhereClause: "WHERE direction='write'"},
		Assertions: []assert.NrAssertion{
			{AggregationFunction: "max", ComparisonOperator: ">", Threshold: 0},
		}},
	{
		Name: "host receiver system.load 1m",
		Metric: assert.NrMetric{
			Name: "system.cpu.load_average.1m",
		},
		Assertions: []assert.NrAssertion{
			{AggregationFunction: "max", ComparisonOperator: ">", Threshold: 0},
		}},
	{
		Name: "host receiver system.load 5m",
		Metric: assert.NrMetric{
			Name: "system.cpu.load_average.5m",
		},
		Assertions: []assert.NrAssertion{
			{AggregationFunction: "max", ComparisonOperator: ">", Threshold: 0},
		}},
	{
		Name: "host receiver system.load 15m",
		Metric: assert.NrMetric{
			Name: "system.cpu.load_average.15m",
		},
		Assertions: []assert.NrAssertion{
			{AggregationFunction: "max", ComparisonOperator: ">=", Threshold: 0},
		}},
	{
		Name: "host receiver memory.usage cached",
		Metric: assert.NrMetric{
			Name:        "system.memory.usage",
			WhereClause: "WHERE state='cached'"},
		Assertions: []assert.NrAssertion{
			{AggregationFunction: "max", ComparisonOperator: ">", Threshold: 0},
		}},
	{
		Name: "host receiver memory.usage free",
		Metric: assert.NrMetric{
			Name:        "system.memory.usage",
			WhereClause: "WHERE state='free'"},
		Assertions: []assert.NrAssertion{
			{AggregationFunction: "max", ComparisonOperator: ">", Threshold: 0},
		}},
	{
		Name: "host receiver memory.usage slab_reclaimable",
		Metric: assert.NrMetric{
			Name:        "system.memory.usage",
			WhereClause: "WHERE state='slab_reclaimable'"},
		Assertions: []assert.NrAssertion{
			{AggregationFunction: "max", ComparisonOperator: ">", Threshold: 0},
		}},
	{
		Name: "host receiver memory.usage buffered",
		Metric: assert.NrMetric{
			Name:        "system.memory.usage",
			WhereClause: "WHERE state='buffered'"},
		Assertions: []assert.NrAssertion{
			{AggregationFunction: "max", ComparisonOperator: ">", Threshold: 0},
		}},
	{
		Name: "host receiver memory.usage used",
		Metric: assert.NrMetric{
			Name:        "system.memory.usage",
			WhereClause: "WHERE state='used'"},
		Assertions: []assert.NrAssertion{
			{AggregationFunction: "max", ComparisonOperator: ">", Threshold: 0},
		}},
	{
		Name: "host receiver memory.utilization free",
		Metric: assert.NrMetric{
			Name:        "system.memory.utilization",
			WhereClause: "WHERE state='free'"},
		Assertions: []assert.NrAssertion{
			{AggregationFunction: "max", ComparisonOperator: ">", Threshold: 0},
		}},
	{
		Name: "host receiver memory.utilization used",
		Metric: assert.NrMetric{
			Name:        "system.memory.utilization",
			WhereClause: "WHERE state='used'"},
		Assertions: []assert.NrAssertion{
			{AggregationFunction: "max", ComparisonOperator: ">", Threshold: 0},
		}},
	{
		Name: "host receiver system.paging.operations out",
		Metric: assert.NrMetric{
			Name:        "system.paging.operations",
			WhereClause: "WHERE direction='page_out'"},
		Assertions: []assert.NrAssertion{
			{AggregationFunction: "max", ComparisonOperator: ">=", Threshold: 0},
		}},
	{
		Name: "host receiver system.paging.operations in",
		Metric: assert.NrMetric{
			Name:        "system.paging.operations",
			WhereClause: "WHERE direction='page_in'"},
		Assertions: []assert.NrAssertion{
			{AggregationFunction: "max", ComparisonOperator: ">=", Threshold: 0},
		}},
	{
		Name: "host receiver paging.usage used",
		Metric: assert.NrMetric{
			Name:        "system.paging.usage",
			WhereClause: "WHERE state='used'"},
		Assertions: []assert.NrAssertion{
			{AggregationFunction: "max", ComparisonOperator: ">=", Threshold: 0},
		}},
	{
		Name: "host receiver paging.usage free",
		Metric: assert.NrMetric{
			Name:        "system.paging.usage",
			WhereClause: "WHERE state='free'"},
		Assertions: []assert.NrAssertion{
			{AggregationFunction: "max", ComparisonOperator: ">", Threshold: 0},
		}},
	{
		Name: "host receiver inodes.usage free",
		Metric: assert.NrMetric{
			Name:        "system.filesystem.inodes.usage",
			WhereClause: "WHERE state='free'"},
		Assertions: []assert.NrAssertion{
			{AggregationFunction: "max", ComparisonOperator: ">", Threshold: 0},
		}},
	{
		Name: "host receiver inodes.usage used",
		Metric: assert.NrMetric{
			Name:        "system.filesystem.inodes.usage",
			WhereClause: "WHERE state='used'"},
		Assertions: []assert.NrAssertion{
			{AggregationFunction: "max", ComparisonOperator: ">", Threshold: 0},
		}},
	{
		Name: "host receiver filesystem.usage used",
		Metric: assert.NrMetric{
			Name:        "system.filesystem.usage",
			WhereClause: "WHERE state='used'"},
		Assertions: []assert.NrAssertion{
			{AggregationFunction: "max", ComparisonOperator: ">", Threshold: 0},
		}},
	{
		Name: "host receiver filesystem.usage free",
		Metric: assert.NrMetric{
			Name:        "system.filesystem.usage",
			WhereClause: "WHERE state='free'"},
		Assertions: []assert.NrAssertion{
			{AggregationFunction: "max", ComparisonOperator: ">", Threshold: 0},
		}},
	{
		Name: "host receiver filesystem.utilization",
		Metric: assert.NrMetric{
			Name:        "system.filesystem.utilization",
			WhereClause: "WHERE type != 'squashfs'",
		},
		Assertions: []assert.NrAssertion{
			{AggregationFunction: "max", ComparisonOperator: ">", Threshold: 0},
		}},
	{
		Name: "host receiver network dropped receive",
		Metric: assert.NrMetric{
			Name:        "system.network.dropped",
			WhereClause: "WHERE direction='receive'",
		},
		Assertions: []assert.NrAssertion{
			{AggregationFunction: "max", ComparisonOperator: ">=", Threshold: 0},
		}},
	{
		Name: "host receiver network dropped transmit",
		Metric: assert.NrMetric{
			Name:        "system.network.dropped",
			WhereClause: "WHERE direction='transmit'",
		},
		Assertions: []assert.NrAssertion{
			{AggregationFunction: "max", ComparisonOperator: ">=", Threshold: 0},
		}},
	{
		Name: "host receiver network errors receive",
		Metric: assert.NrMetric{
			Name:        "system.network.errors",
			WhereClause: "WHERE direction='receive'",
		},
		Assertions: []assert.NrAssertion{
			{AggregationFunction: "max", ComparisonOperator: ">=", Threshold: 0},
		}},
	{
		Name: "host receiver network errors transmit",
		Metric: assert.NrMetric{
			Name:        "system.network.errors",
			WhereClause: "WHERE direction='transmit'",
		},
		Assertions: []assert.NrAssertion{
			{AggregationFunction: "max", ComparisonOperator: ">=", Threshold: 0},
		}},
	{
		Name: "host receiver network io receive",
		Metric: assert.NrMetric{
			Name:        "system.network.io",
			WhereClause: "WHERE direction='receive'",
		},
		Assertions: []assert.NrAssertion{
			{AggregationFunction: "max", ComparisonOperator: ">", Threshold: 0},
		}},
	{
		Name: "host receiver network io transmit",
		Metric: assert.NrMetric{
			Name:        "system.network.io",
			WhereClause: "WHERE direction='transmit'",
		},
		Assertions: []assert.NrAssertion{
			{AggregationFunction: "max", ComparisonOperator: ">", Threshold: 0},
		}},
	{
		Name: "host receiver network packets receive",
		Metric: assert.NrMetric{
			Name:        "system.network.packets",
			WhereClause: "WHERE direction='receive'",
		},
		Assertions: []assert.NrAssertion{
			{AggregationFunction: "max", ComparisonOperator: ">", Threshold: 0},
		}},
	{
		Name: "host receiver network packets transmit",
		Metric: assert.NrMetric{
			Name:        "system.network.packets",
			WhereClause: "WHERE direction='transmit'",
		},
		Assertions: []assert.NrAssertion{
			{AggregationFunction: "max", ComparisonOperator: ">", Threshold: 0},
		}},
}

func GetOnHostTestCases() []TestCase {
	return testCases
}

func GetOnHostTestCasesWithout(excludedMetric []string) []TestCase {
	var filteredTestCases []TestCase
	for _, testCase := range testCases {
		included := true
		for _, skippedMetric := range excludedMetric {
			if testCase.Metric.Name == skippedMetric {
				included = false
			}
		}
		if included {
			filteredTestCases = append(filteredTestCases, testCase)
		}
	}
	return filteredTestCases
}
