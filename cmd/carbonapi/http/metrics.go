package http

import (
	"fmt"

	"github.com/go-graphite/carbonapi/cache"
	"github.com/go-graphite/carbonapi/cmd/carbonapi/config"
	zipperTypes "github.com/go-graphite/carbonapi/zipper/types"
	"github.com/msaf1980/go-metrics"
	"go.uber.org/zap"
)

var ApiMetrics = struct {
	RenderRequests        metrics.Counter
	RequestCacheHits      metrics.Counter
	RequestCacheMisses    metrics.Counter
	BackendCacheHits      metrics.Counter
	BackendCacheMisses    metrics.Counter
	RenderCacheOverheadNS metrics.Counter
	RequestsH             metrics.Histogram

	FindRequests metrics.Counter

	MemcacheTimeouts metrics.UGauge

	CacheSize  metrics.UGauge
	CacheItems metrics.Gauge
}{
	RenderRequests:        metrics.NewCounter(),
	RequestCacheHits:      metrics.NewCounter(),
	RequestCacheMisses:    metrics.NewCounter(),
	BackendCacheHits:      metrics.NewCounter(),
	BackendCacheMisses:    metrics.NewCounter(),
	RenderCacheOverheadNS: metrics.NewCounter(),

	FindRequests: metrics.NewCounter(),
}

var ZipperMetrics = struct {
	FindRequests metrics.Counter
	FindTimeouts metrics.Counter
	FindErrors   metrics.Counter

	SearchRequests metrics.Counter

	RenderRequests metrics.Counter
	RenderTimeouts metrics.Counter
	RenderErrors   metrics.Counter

	InfoRequests metrics.Counter
	InfoTimeouts metrics.Counter
	InfoErrors   metrics.Counter

	Timeouts metrics.Counter

	CacheMisses metrics.Counter
	CacheHits   metrics.Counter
}{
	FindRequests: metrics.NewCounter(),
	FindTimeouts: metrics.NewCounter(),
	FindErrors:   metrics.NewCounter(),

	SearchRequests: metrics.NewCounter(),

	RenderRequests: metrics.NewCounter(),
	RenderTimeouts: metrics.NewCounter(),
	RenderErrors:   metrics.NewCounter(),

	InfoRequests: metrics.NewCounter(),
	InfoTimeouts: metrics.NewCounter(),
	InfoErrors:   metrics.NewCounter(),

	Timeouts: metrics.NewCounter(),

	CacheHits:   metrics.NewCounter(),
	CacheMisses: metrics.NewCounter(),
}

func ZipperStats(stats *zipperTypes.Stats) {
	if stats == nil {
		return
	}
	ZipperMetrics.Timeouts.Inc(stats.Timeouts)
	ZipperMetrics.FindRequests.Inc(stats.FindRequests)
	ZipperMetrics.FindTimeouts.Inc(stats.FindTimeouts)
	ZipperMetrics.FindErrors.Inc(stats.FindErrors)
	ZipperMetrics.RenderRequests.Inc(stats.RenderRequests)
	ZipperMetrics.RenderTimeouts.Inc(stats.RenderTimeouts)
	ZipperMetrics.RenderErrors.Inc(stats.RenderErrors)
	ZipperMetrics.InfoRequests.Inc(stats.InfoRequests)
	ZipperMetrics.InfoTimeouts.Inc(stats.InfoTimeouts)
	ZipperMetrics.InfoErrors.Inc(stats.InfoErrors)
	ZipperMetrics.SearchRequests.Inc(stats.SearchRequests)
	ZipperMetrics.CacheMisses.Inc(stats.CacheMisses)
	ZipperMetrics.CacheHits.Inc(stats.CacheHits)
}

func SetupMetrics(logger *zap.Logger) {
	switch config.Config.ResponseCacheConfig.Type {
	case "memcache":
		mcache := config.Config.ResponseCache.(*cache.MemcachedCache)

		ApiMetrics.MemcacheTimeouts = metrics.NewFunctionalUGauge(mcache.Timeouts)
	case "mem":
		qcache := config.Config.ResponseCache.(*cache.ExpireCache)

		ApiMetrics.CacheSize = metrics.NewFunctionalUGauge(qcache.Size)
		ApiMetrics.CacheItems = metrics.NewFunctionalGauge(func() int64 {
			return int64(qcache.Items())
		})
	default:
	}

	ApiMetrics.RequestsH = initRequestsHistogram()
}

func initRequestsHistogram() metrics.Histogram {
	if config.Config.Upstreams.SumBuckets {
		if len(config.Config.Upstreams.BucketsWidth) > 0 {
			labels := make([]string, len(config.Config.Upstreams.BucketsWidth)+1)

			for i := 0; i <= len(config.Config.Upstreams.BucketsWidth); i++ {
				if i >= len(config.Config.Upstreams.BucketsLabels) || config.Config.Upstreams.BucketsLabels[i] == "" {
					labels[i] = fmt.Sprintf("_to_%dms", (i+1)*100)
				} else {
					labels[i] = config.Config.Upstreams.BucketsLabels[i]
				}
			}
			return metrics.NewVSumHistogram(config.Config.Upstreams.BucketsWidth, config.Config.Upstreams.BucketsLabels).
				SetLabels(labels).
				SetNameTotal("")
		} else {
			labels := make([]string, config.Config.Upstreams.Buckets+1)

			for i := 0; i <= config.Config.Upstreams.Buckets; i++ {
				labels[i] = fmt.Sprintf("_to_%dms", (i+1)*100)
			}
			return metrics.NewFixedSumHistogram(100, int64(config.Config.Upstreams.Buckets)*100, 100).
				SetLabels(labels).
				SetNameTotal("")
		}
	} else if len(config.Config.Upstreams.BucketsWidth) > 0 {
		labels := make([]string, len(config.Config.Upstreams.BucketsWidth)+1)

		for i := 0; i <= len(config.Config.Upstreams.BucketsWidth); i++ {
			if i >= len(config.Config.Upstreams.BucketsLabels) || config.Config.Upstreams.BucketsLabels[i] == "" {
				labels[i] = fmt.Sprintf("_in_%dms_to_%dms", i*100, (i+1)*100)
			} else {
				labels[i] = config.Config.Upstreams.BucketsLabels[i]
			}
		}
		return metrics.NewVSumHistogram(config.Config.Upstreams.BucketsWidth, config.Config.Upstreams.BucketsLabels).
			SetLabels(labels).
			SetNameTotal("")
	} else {
		labels := make([]string, config.Config.Upstreams.Buckets+1)

		for i := 0; i <= config.Config.Upstreams.Buckets; i++ {
			labels[i] = fmt.Sprintf("_in_%dms_to_%dms", i*100, (i+1)*100)
		}
		return metrics.NewFixedSumHistogram(100, int64(config.Config.Upstreams.Buckets)*100, 100).
			SetLabels(labels).
			SetNameTotal("")
	}
}
