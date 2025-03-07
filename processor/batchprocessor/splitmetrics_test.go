// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package batchprocessor

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"go.opentelemetry.io/collector/internal/testdata"
	"go.opentelemetry.io/collector/model/pdata"
)

func TestSplitMetrics_noop(t *testing.T) {
	td := testdata.GenerateMetricsManyMetricsSameResource(20)
	splitSize := 40
	split := splitMetrics(splitSize, td)
	assert.Equal(t, td, split)

	i := 0
	td.ResourceMetrics().At(0).InstrumentationLibraryMetrics().At(0).Metrics().RemoveIf(func(_ pdata.Metric) bool {
		i++
		return i > 5
	})
	assert.EqualValues(t, td, split)
}

func TestSplitMetrics(t *testing.T) {
	md := testdata.GenerateMetricsManyMetricsSameResource(20)
	metrics := md.ResourceMetrics().At(0).InstrumentationLibraryMetrics().At(0).Metrics()
	dataPointCount := metricDPC(metrics.At(0))
	for i := 0; i < metrics.Len(); i++ {
		metrics.At(i).SetName(getTestMetricName(0, i))
		assert.Equal(t, dataPointCount, metricDPC(metrics.At(i)))
	}
	cp := pdata.NewMetrics()
	cpMetrics := cp.ResourceMetrics().AppendEmpty().InstrumentationLibraryMetrics().AppendEmpty().Metrics()
	cpMetrics.EnsureCapacity(5)
	md.ResourceMetrics().At(0).InstrumentationLibraryMetrics().At(0).InstrumentationLibrary().CopyTo(
		cp.ResourceMetrics().At(0).InstrumentationLibraryMetrics().At(0).InstrumentationLibrary())
	md.ResourceMetrics().At(0).Resource().CopyTo(
		cp.ResourceMetrics().At(0).Resource())
	metrics.At(0).CopyTo(cpMetrics.AppendEmpty())
	metrics.At(1).CopyTo(cpMetrics.AppendEmpty())
	metrics.At(2).CopyTo(cpMetrics.AppendEmpty())
	metrics.At(3).CopyTo(cpMetrics.AppendEmpty())
	metrics.At(4).CopyTo(cpMetrics.AppendEmpty())

	splitMetricCount := 5
	splitSize := splitMetricCount * dataPointCount
	split := splitMetrics(splitSize, md)
	assert.Equal(t, splitMetricCount, split.MetricCount())
	assert.Equal(t, cp, split)
	assert.Equal(t, 15, md.MetricCount())
	assert.Equal(t, "test-metric-int-0-0", split.ResourceMetrics().At(0).InstrumentationLibraryMetrics().At(0).Metrics().At(0).Name())
	assert.Equal(t, "test-metric-int-0-4", split.ResourceMetrics().At(0).InstrumentationLibraryMetrics().At(0).Metrics().At(4).Name())

	split = splitMetrics(splitSize, md)
	assert.Equal(t, 10, md.MetricCount())
	assert.Equal(t, "test-metric-int-0-5", split.ResourceMetrics().At(0).InstrumentationLibraryMetrics().At(0).Metrics().At(0).Name())
	assert.Equal(t, "test-metric-int-0-9", split.ResourceMetrics().At(0).InstrumentationLibraryMetrics().At(0).Metrics().At(4).Name())

	split = splitMetrics(splitSize, md)
	assert.Equal(t, 5, md.MetricCount())
	assert.Equal(t, "test-metric-int-0-10", split.ResourceMetrics().At(0).InstrumentationLibraryMetrics().At(0).Metrics().At(0).Name())
	assert.Equal(t, "test-metric-int-0-14", split.ResourceMetrics().At(0).InstrumentationLibraryMetrics().At(0).Metrics().At(4).Name())

	split = splitMetrics(splitSize, md)
	assert.Equal(t, 5, md.MetricCount())
	assert.Equal(t, "test-metric-int-0-15", split.ResourceMetrics().At(0).InstrumentationLibraryMetrics().At(0).Metrics().At(0).Name())
	assert.Equal(t, "test-metric-int-0-19", split.ResourceMetrics().At(0).InstrumentationLibraryMetrics().At(0).Metrics().At(4).Name())
}

func TestSplitMetricsMultipleResourceSpans(t *testing.T) {
	md := testdata.GenerateMetricsManyMetricsSameResource(20)
	metrics := md.ResourceMetrics().At(0).InstrumentationLibraryMetrics().At(0).Metrics()
	dataPointCount := metricDPC(metrics.At(0))
	for i := 0; i < metrics.Len(); i++ {
		metrics.At(i).SetName(getTestMetricName(0, i))
		assert.Equal(t, dataPointCount, metricDPC(metrics.At(i)))
	}
	// add second index to resource metrics
	testdata.GenerateMetricsManyMetricsSameResource(20).
		ResourceMetrics().At(0).CopyTo(md.ResourceMetrics().AppendEmpty())
	metrics = md.ResourceMetrics().At(1).InstrumentationLibraryMetrics().At(0).Metrics()
	for i := 0; i < metrics.Len(); i++ {
		metrics.At(i).SetName(getTestMetricName(1, i))
	}

	splitMetricCount := 5
	splitSize := splitMetricCount * dataPointCount
	split := splitMetrics(splitSize, md)
	assert.Equal(t, splitMetricCount, split.MetricCount())
	assert.Equal(t, 35, md.MetricCount())
	assert.Equal(t, "test-metric-int-0-0", split.ResourceMetrics().At(0).InstrumentationLibraryMetrics().At(0).Metrics().At(0).Name())
	assert.Equal(t, "test-metric-int-0-4", split.ResourceMetrics().At(0).InstrumentationLibraryMetrics().At(0).Metrics().At(4).Name())
}

func TestSplitMetricsMultipleResourceSpans_SplitSizeGreaterThanMetricSize(t *testing.T) {
	td := testdata.GenerateMetricsManyMetricsSameResource(20)
	metrics := td.ResourceMetrics().At(0).InstrumentationLibraryMetrics().At(0).Metrics()
	dataPointCount := metricDPC(metrics.At(0))
	for i := 0; i < metrics.Len(); i++ {
		metrics.At(i).SetName(getTestMetricName(0, i))
		assert.Equal(t, dataPointCount, metricDPC(metrics.At(i)))
	}
	// add second index to resource metrics
	testdata.GenerateMetricsManyMetricsSameResource(20).
		ResourceMetrics().At(0).CopyTo(td.ResourceMetrics().AppendEmpty())
	metrics = td.ResourceMetrics().At(1).InstrumentationLibraryMetrics().At(0).Metrics()
	for i := 0; i < metrics.Len(); i++ {
		metrics.At(i).SetName(getTestMetricName(1, i))
	}

	splitMetricCount := 25
	splitSize := splitMetricCount * dataPointCount
	split := splitMetrics(splitSize, td)
	assert.Equal(t, splitMetricCount, split.MetricCount())
	assert.Equal(t, 40-splitMetricCount, td.MetricCount())
	assert.Equal(t, 1, td.ResourceMetrics().Len())
	assert.Equal(t, "test-metric-int-0-0", split.ResourceMetrics().At(0).InstrumentationLibraryMetrics().At(0).Metrics().At(0).Name())
	assert.Equal(t, "test-metric-int-0-19", split.ResourceMetrics().At(0).InstrumentationLibraryMetrics().At(0).Metrics().At(19).Name())
	assert.Equal(t, "test-metric-int-1-0", split.ResourceMetrics().At(1).InstrumentationLibraryMetrics().At(0).Metrics().At(0).Name())
	assert.Equal(t, "test-metric-int-1-4", split.ResourceMetrics().At(1).InstrumentationLibraryMetrics().At(0).Metrics().At(4).Name())
}

func TestSplitMetricsUneven(t *testing.T) {
	md := testdata.GenerateMetricsManyMetricsSameResource(10)
	metrics := md.ResourceMetrics().At(0).InstrumentationLibraryMetrics().At(0).Metrics()
	dataPointCount := 2
	for i := 0; i < metrics.Len(); i++ {
		metrics.At(i).SetName(getTestMetricName(0, i))
		assert.Equal(t, dataPointCount, metricDPC(metrics.At(i)))
	}

	splitSize := 9
	split := splitMetrics(splitSize, md)
	assert.Equal(t, 5, split.MetricCount())
	assert.Equal(t, 6, md.MetricCount())
	assert.Equal(t, "test-metric-int-0-0", split.ResourceMetrics().At(0).InstrumentationLibraryMetrics().At(0).Metrics().At(0).Name())
	assert.Equal(t, "test-metric-int-0-4", split.ResourceMetrics().At(0).InstrumentationLibraryMetrics().At(0).Metrics().At(4).Name())

	split = splitMetrics(splitSize, md)
	assert.Equal(t, 5, split.MetricCount())
	assert.Equal(t, 1, md.MetricCount())
	assert.Equal(t, "test-metric-int-0-4", split.ResourceMetrics().At(0).InstrumentationLibraryMetrics().At(0).Metrics().At(0).Name())
	assert.Equal(t, "test-metric-int-0-8", split.ResourceMetrics().At(0).InstrumentationLibraryMetrics().At(0).Metrics().At(4).Name())

	split = splitMetrics(splitSize, md)
	assert.Equal(t, 1, split.MetricCount())
	assert.Equal(t, "test-metric-int-0-9", split.ResourceMetrics().At(0).InstrumentationLibraryMetrics().At(0).Metrics().At(0).Name())
}

func TestSplitMetricsAllTypes(t *testing.T) {
	md := testdata.GeneratMetricsAllTypesWithSampleDatapoints()
	dataPointCount := 2
	metrics := md.ResourceMetrics().At(0).InstrumentationLibraryMetrics().At(0).Metrics()
	for i := 0; i < metrics.Len(); i++ {
		metrics.At(i).SetName(getTestMetricName(0, i))
		assert.Equal(t, dataPointCount, metricDPC(metrics.At(i)))
	}

	splitSize := 2
	// Start with 6 metric types, and 2 points per-metric. Split out the first,
	// and then split by 2 for the rest so that each metric is split in half.
	// Verify that descriptors are preserved for all data types across splits.

	split := splitMetrics(1, md)
	assert.Equal(t, 1, split.MetricCount())
	assert.Equal(t, 6, md.MetricCount())
	gaugeInt := split.ResourceMetrics().At(0).InstrumentationLibraryMetrics().At(0).Metrics().At(0)
	assert.Equal(t, 1, gaugeInt.Gauge().DataPoints().Len())
	assert.Equal(t, "test-metric-int-0-0", gaugeInt.Name())

	split = splitMetrics(splitSize, md)
	assert.Equal(t, 2, split.MetricCount())
	assert.Equal(t, 5, md.MetricCount())
	gaugeInt = split.ResourceMetrics().At(0).InstrumentationLibraryMetrics().At(0).Metrics().At(0)
	gaugeDouble := split.ResourceMetrics().At(0).InstrumentationLibraryMetrics().At(0).Metrics().At(1)
	assert.Equal(t, 1, gaugeInt.Gauge().DataPoints().Len())
	assert.Equal(t, "test-metric-int-0-0", gaugeInt.Name())
	assert.Equal(t, 1, gaugeDouble.Gauge().DataPoints().Len())
	assert.Equal(t, "test-metric-int-0-1", gaugeDouble.Name())

	split = splitMetrics(splitSize, md)
	assert.Equal(t, 2, split.MetricCount())
	assert.Equal(t, 4, md.MetricCount())
	gaugeDouble = split.ResourceMetrics().At(0).InstrumentationLibraryMetrics().At(0).Metrics().At(0)
	sumInt := split.ResourceMetrics().At(0).InstrumentationLibraryMetrics().At(0).Metrics().At(1)
	assert.Equal(t, 1, gaugeDouble.Gauge().DataPoints().Len())
	assert.Equal(t, "test-metric-int-0-1", gaugeDouble.Name())
	assert.Equal(t, 1, sumInt.Sum().DataPoints().Len())
	assert.Equal(t, pdata.MetricAggregationTemporalityCumulative, sumInt.Sum().AggregationTemporality())
	assert.Equal(t, true, sumInt.Sum().IsMonotonic())
	assert.Equal(t, "test-metric-int-0-2", sumInt.Name())

	split = splitMetrics(splitSize, md)
	assert.Equal(t, 2, split.MetricCount())
	assert.Equal(t, 3, md.MetricCount())
	sumInt = split.ResourceMetrics().At(0).InstrumentationLibraryMetrics().At(0).Metrics().At(0)
	sumDouble := split.ResourceMetrics().At(0).InstrumentationLibraryMetrics().At(0).Metrics().At(1)
	assert.Equal(t, 1, sumInt.Sum().DataPoints().Len())
	assert.Equal(t, pdata.MetricAggregationTemporalityCumulative, sumInt.Sum().AggregationTemporality())
	assert.Equal(t, true, sumInt.Sum().IsMonotonic())
	assert.Equal(t, "test-metric-int-0-2", sumInt.Name())
	assert.Equal(t, 1, sumDouble.Sum().DataPoints().Len())
	assert.Equal(t, pdata.MetricAggregationTemporalityCumulative, sumDouble.Sum().AggregationTemporality())
	assert.Equal(t, true, sumDouble.Sum().IsMonotonic())
	assert.Equal(t, "test-metric-int-0-3", sumDouble.Name())

	split = splitMetrics(splitSize, md)
	assert.Equal(t, 2, split.MetricCount())
	assert.Equal(t, 2, md.MetricCount())
	sumDouble = split.ResourceMetrics().At(0).InstrumentationLibraryMetrics().At(0).Metrics().At(0)
	doubleHistogram := split.ResourceMetrics().At(0).InstrumentationLibraryMetrics().At(0).Metrics().At(1)
	assert.Equal(t, 1, sumDouble.Sum().DataPoints().Len())
	assert.Equal(t, pdata.MetricAggregationTemporalityCumulative, sumDouble.Sum().AggregationTemporality())
	assert.Equal(t, true, sumDouble.Sum().IsMonotonic())
	assert.Equal(t, "test-metric-int-0-3", sumDouble.Name())
	assert.Equal(t, 1, doubleHistogram.Histogram().DataPoints().Len())
	assert.Equal(t, pdata.MetricAggregationTemporalityCumulative, doubleHistogram.Histogram().AggregationTemporality())
	assert.Equal(t, "test-metric-int-0-4", doubleHistogram.Name())

	split = splitMetrics(splitSize, md)
	assert.Equal(t, 2, split.MetricCount())
	assert.Equal(t, 1, md.MetricCount())
	doubleHistogram = split.ResourceMetrics().At(0).InstrumentationLibraryMetrics().At(0).Metrics().At(0)
	doubleSummary := split.ResourceMetrics().At(0).InstrumentationLibraryMetrics().At(0).Metrics().At(1)
	assert.Equal(t, 1, doubleHistogram.Histogram().DataPoints().Len())
	assert.Equal(t, pdata.MetricAggregationTemporalityCumulative, doubleHistogram.Histogram().AggregationTemporality())
	assert.Equal(t, "test-metric-int-0-4", doubleHistogram.Name())
	assert.Equal(t, 1, doubleSummary.Summary().DataPoints().Len())
	assert.Equal(t, "test-metric-int-0-5", doubleSummary.Name())

	split = splitMetrics(splitSize, md)
	doubleSummary = split.ResourceMetrics().At(0).InstrumentationLibraryMetrics().At(0).Metrics().At(0)
	assert.Equal(t, 1, doubleSummary.Summary().DataPoints().Len())
	assert.Equal(t, "test-metric-int-0-5", doubleSummary.Name())
}

func TestSplitMetricsBatchSizeSmallerThanDataPointCount(t *testing.T) {
	md := testdata.GenerateMetricsManyMetricsSameResource(2)
	metrics := md.ResourceMetrics().At(0).InstrumentationLibraryMetrics().At(0).Metrics()
	dataPointCount := 2
	for i := 0; i < metrics.Len(); i++ {
		metrics.At(i).SetName(getTestMetricName(0, i))
		assert.Equal(t, dataPointCount, metricDPC(metrics.At(i)))
	}

	splitSize := 1
	split := splitMetrics(splitSize, md)
	splitMetric := split.ResourceMetrics().At(0).InstrumentationLibraryMetrics().At(0).Metrics().At(0)
	assert.Equal(t, 1, split.MetricCount())
	assert.Equal(t, 2, md.MetricCount())
	assert.Equal(t, pdata.MetricAggregationTemporalityCumulative, splitMetric.Sum().AggregationTemporality())
	assert.Equal(t, true, splitMetric.Sum().IsMonotonic())
	assert.Equal(t, "test-metric-int-0-0", splitMetric.Name())

	split = splitMetrics(splitSize, md)
	splitMetric = split.ResourceMetrics().At(0).InstrumentationLibraryMetrics().At(0).Metrics().At(0)
	assert.Equal(t, 1, split.MetricCount())
	assert.Equal(t, 1, md.MetricCount())
	assert.Equal(t, pdata.MetricAggregationTemporalityCumulative, splitMetric.Sum().AggregationTemporality())
	assert.Equal(t, true, splitMetric.Sum().IsMonotonic())
	assert.Equal(t, "test-metric-int-0-0", splitMetric.Name())

	split = splitMetrics(splitSize, md)
	splitMetric = split.ResourceMetrics().At(0).InstrumentationLibraryMetrics().At(0).Metrics().At(0)
	assert.Equal(t, 1, split.MetricCount())
	assert.Equal(t, 1, md.MetricCount())
	assert.Equal(t, pdata.MetricAggregationTemporalityCumulative, splitMetric.Sum().AggregationTemporality())
	assert.Equal(t, true, splitMetric.Sum().IsMonotonic())
	assert.Equal(t, "test-metric-int-0-1", splitMetric.Name())

	split = splitMetrics(splitSize, md)
	splitMetric = split.ResourceMetrics().At(0).InstrumentationLibraryMetrics().At(0).Metrics().At(0)
	assert.Equal(t, 1, split.MetricCount())
	assert.Equal(t, 1, md.MetricCount())
	assert.Equal(t, pdata.MetricAggregationTemporalityCumulative, splitMetric.Sum().AggregationTemporality())
	assert.Equal(t, true, splitMetric.Sum().IsMonotonic())
	assert.Equal(t, "test-metric-int-0-1", splitMetric.Name())
}

func TestSplitMetricsMultipleILM(t *testing.T) {
	md := testdata.GenerateMetricsManyMetricsSameResource(20)
	metrics := md.ResourceMetrics().At(0).InstrumentationLibraryMetrics().At(0).Metrics()
	dataPointCount := metricDPC(metrics.At(0))
	for i := 0; i < metrics.Len(); i++ {
		metrics.At(i).SetName(getTestMetricName(0, i))
		assert.Equal(t, dataPointCount, metricDPC(metrics.At(i)))
	}
	// add second index to ilm
	md.ResourceMetrics().At(0).InstrumentationLibraryMetrics().At(0).
		CopyTo(md.ResourceMetrics().At(0).InstrumentationLibraryMetrics().AppendEmpty())

	// add a third index to ilm
	md.ResourceMetrics().At(0).InstrumentationLibraryMetrics().At(0).
		CopyTo(md.ResourceMetrics().At(0).InstrumentationLibraryMetrics().AppendEmpty())
	metrics = md.ResourceMetrics().At(0).InstrumentationLibraryMetrics().At(2).Metrics()
	for i := 0; i < metrics.Len(); i++ {
		metrics.At(i).SetName(getTestMetricName(2, i))
	}

	splitMetricCount := 40
	splitSize := splitMetricCount * dataPointCount
	split := splitMetrics(splitSize, md)
	assert.Equal(t, splitMetricCount, split.MetricCount())
	assert.Equal(t, 20, md.MetricCount())
	assert.Equal(t, "test-metric-int-0-0", split.ResourceMetrics().At(0).InstrumentationLibraryMetrics().At(0).Metrics().At(0).Name())
	assert.Equal(t, "test-metric-int-0-4", split.ResourceMetrics().At(0).InstrumentationLibraryMetrics().At(0).Metrics().At(4).Name())
}

func BenchmarkSplitMetrics(b *testing.B) {
	md := pdata.NewMetrics()
	rms := md.ResourceMetrics()
	for i := 0; i < 20; i++ {
		testdata.GenerateMetricsManyMetricsSameResource(20).ResourceMetrics().MoveAndAppendTo(md.ResourceMetrics())
		ms := rms.At(rms.Len() - 1).InstrumentationLibraryMetrics().At(0).Metrics()
		for i := 0; i < ms.Len(); i++ {
			ms.At(i).SetName(getTestMetricName(1, i))
		}
	}

	if b.N > 100000 {
		b.Skipf("SKIP: b.N too high, set -benchtime=<n>x with n < 100000")
	}

	dataPointCount := metricDPC(md.ResourceMetrics().At(0).InstrumentationLibraryMetrics().At(0).Metrics().At(0))
	clones := make([]pdata.Metrics, b.N)
	for n := 0; n < b.N; n++ {
		clones[n] = md.Clone()
	}
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		cloneReq := clones[n]
		split := splitMetrics(128*dataPointCount, cloneReq)
		if split.MetricCount() != 128 || cloneReq.MetricCount() != 400-128 {
			b.Fail()
		}
	}
}
