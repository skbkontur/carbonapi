package aliasByNode

import (
	"math"
	"testing"
	"time"

	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	th "github.com/go-graphite/carbonapi/tests"

	"github.com/go-graphite/carbonapi/expr/functions/aliasSub"
	"github.com/go-graphite/carbonapi/expr/functions/perSecond"
	"github.com/go-graphite/carbonapi/expr/functions/transformNull"
)

func init() {
	md := New("")
	for _, m := range md {
		metadata.RegisterFunction(m.Name, m.F)
	}
	asFunc := aliasSub.New("")
	for _, m := range asFunc {
		metadata.RegisterFunction(m.Name, m.F)
	}
	tnFunc := transformNull.New("")
	for _, m := range tnFunc {
		metadata.RegisterFunction(m.Name, m.F)
	}
	psFunc := perSecond.New("")
	for _, m := range psFunc {
		metadata.RegisterFunction(m.Name, m.F)
	}

	evaluator := th.EvaluatorFromFuncWithMetadata(metadata.FunctionMD.Functions)
	metadata.SetEvaluator(evaluator)
	helper.SetEvaluator(evaluator)
}

func TestAliasByNode(t *testing.T) {
	now32 := int64(time.Now().Unix())

	tests := []th.EvalTestItem{
		{
			Target: "aliasByNode(aliasSub(transformNull(metric1.foo.bar.ba*, 0), 'baz', 'word'), 2, 3)",
			M: map[parser.MetricRequest][]*types.MetricData{
				{Metric: "metric1.foo.bar.ba*", From: 0, Until: 1}: {types.MakeMetricData("metric1.foo.bar.baz", []float64{1, 2, 3, 4, 5}, 1, now32)},
			},
			Want: []*types.MetricData{types.MakeMetricData("bar.word", []float64{1, 2, 3, 4, 5}, 1, now32)},
		},
		{
			Target: "aliasByNode(metric1.foo.bar.baz,1)",
			M: map[parser.MetricRequest][]*types.MetricData{
				{Metric: "metric1.foo.bar.baz", From: 0, Until: 1}: {types.MakeMetricData("metric1.foo.bar.baz", []float64{1, 2, 3, 4, 5}, 1, now32)},
			},
			Want: []*types.MetricData{types.MakeMetricData("foo", []float64{1, 2, 3, 4, 5}, 1, now32)},
		},
		{
			Target: "aliasByNode(metric1.foo.bar.baz,1,3)",
			M: map[parser.MetricRequest][]*types.MetricData{
				{Metric: "metric1.foo.bar.baz", From: 0, Until: 1}: {types.MakeMetricData("metric1.foo.bar.baz", []float64{1, 2, 3, 4, 5}, 1, now32)},
			},
			Want: []*types.MetricData{types.MakeMetricData("foo.baz",
				[]float64{1, 2, 3, 4, 5}, 1, now32)},
		},
		{
			Target: "aliasByNode(metric1.foo.bar.baz,1,-2)",
			M: map[parser.MetricRequest][]*types.MetricData{
				{Metric: "metric1.foo.bar.baz", From: 0, Until: 1}: {types.MakeMetricData("metric1.foo.bar.baz", []float64{1, 2, 3, 4, 5}, 1, now32)},
			},
			Want: []*types.MetricData{types.MakeMetricData("foo.bar",
				[]float64{1, 2, 3, 4, 5}, 1, now32)},
		},
		{
			Target: `aliasByTags(*, "foo")`,
			M: map[parser.MetricRequest][]*types.MetricData{
				{Metric: "*", From: 0, Until: 1}: {types.MakeMetricData("metric1.foo.bar.baz;foo=bar;baz=bam", []float64{1, 2, 3, 4, 5}, 1, now32)},
			},
			Want: []*types.MetricData{types.MakeMetricData("bar", []float64{1, 2, 3, 4, 5}, 1, now32)},
		},
		{
			Target: `aliasByTags(*, "foo", "name")`,
			M: map[parser.MetricRequest][]*types.MetricData{
				{Metric: "*", From: 0, Until: 1}: {types.MakeMetricData("metric1;foo=bar", []float64{1, 2, 3, 4, 5}, 1, now32)},
			},
			Want: []*types.MetricData{types.MakeMetricData("bar.metric1", []float64{1, 2, 3, 4, 5}, 1, now32)},
		},
		{
			Target: `aliasByTags(*, 2, "blah", "foo", 1)`,
			M: map[parser.MetricRequest][]*types.MetricData{
				{Metric: "*", From: 0, Until: 1}: {types.MakeMetricData("base.metric1;foo=bar;baz=bam", []float64{1, 2, 3, 4, 5}, 1, now32)},
			},
			Want: []*types.MetricData{types.MakeMetricData(".bar.metric1", []float64{1, 2, 3, 4, 5}, 1, now32)},
		},
		{
			Target: `aliasByTags(*, 2, "baz", "foo", 1)`,
			M: map[parser.MetricRequest][]*types.MetricData{
				{Metric: "*", From: 0, Until: 1}: {types.MakeMetricData("base.metric1;foo=bar;baz=bam", []float64{1, 2, 3, 4, 5}, 1, now32)},
			},
			Want: []*types.MetricData{types.MakeMetricData("bam.bar.metric1", []float64{1, 2, 3, 4, 5}, 1, now32)},
		},
		{
			Target: `aliasByTags(perSecond(*), 'name')`,
			M: map[parser.MetricRequest][]*types.MetricData{
				{Metric: "*", From: 0, Until: 1}: {types.MakeMetricData("base.metric1;foo=bar;baz=bam", []float64{1, 2, 3, 4, 5}, 1, now32)},
			},
			Want: []*types.MetricData{types.MakeMetricData("base.metric1", []float64{math.NaN(), 1, 1, 1, 1}, 1, now32)},
		},
		{
			Target: "aliasByNode(metric1.foo.*.*,2)",
			M: map[parser.MetricRequest][]*types.MetricData{
				{Metric: "metric1.foo.*.*", From: 0, Until: 1}: {
					types.MakeMetricData("metric1.foo.Ab1==.lag", []float64{1, 2, 3, 4, 5}, 1, now32),
				},
			},
			Want: []*types.MetricData{types.MakeMetricData("Ab1==", []float64{1, 2, 3, 4, 5}, 1, now32)},
		},
		{
			Target: "aliasByNode(*,0,2)",
			M: map[parser.MetricRequest][]*types.MetricData{
				{Metric: "*", From: 0, Until: 1}: {
					types.MakeMetricData("metric1.foo.bar=;tag1=value1", []float64{1, 2, 3, 4, 5}, 1, now32),
				},
			},
			Want: []*types.MetricData{types.MakeMetricData("metric1.bar=", []float64{1, 2, 3, 4, 5}, 1, now32)},
		},
	}

	for _, tt := range tests {
		testName := tt.Target
		t.Run(testName, func(t *testing.T) {
			th.TestEvalExpr(t, &tt)
		})
	}

}
