package main

import (
	"os"
	"strings"
	"time"

	"github.com/go-graphite/carbonapi/cmd/carbonapi/config"
	"github.com/go-graphite/carbonapi/cmd/carbonapi/http"

	"github.com/msaf1980/go-metrics"
	"github.com/msaf1980/go-metrics/graphite"
	"go.uber.org/zap"
)

func setupGraphiteMetrics(logger *zap.Logger) {
	var host string
	if envhost := os.Getenv("GRAPHITEHOST") + ":" + os.Getenv("GRAPHITEPORT"); envhost != ":" || config.Config.Graphite.Host != "" {
		switch {
		case envhost != ":" && config.Config.Graphite.Host != "":
			host = config.Config.Graphite.Host
		case envhost != ":":
			host = envhost
		case config.Config.Graphite.Host != "":
			host = config.Config.Graphite.Host
		}
	}

	logger.Info("starting carbonapi",
		zap.String("build_version", BuildVersion),
		zap.Any("config", config.Config),
	)

	if host != "" {
		hostname, _ := os.Hostname()
		hostname = strings.ReplaceAll(hostname, ".", "_")

		prefix := config.Config.Graphite.Prefix

		pattern := config.Config.Graphite.Pattern
		pattern = strings.ReplaceAll(pattern, "{prefix}", prefix)
		pattern = strings.ReplaceAll(pattern, "{fqdn}", hostname)

		// register our metrics with graphite
		graphite := graphite.New(config.Config.Graphite.Interval, pattern, host, 10*time.Second)

		metrics.Register("request_cache_hits", http.ApiMetrics.RequestCacheHits)
		metrics.Register("request_cache_misses", http.ApiMetrics.RequestCacheMisses)
		metrics.Register("request_cache_overhead_ns", http.ApiMetrics.RenderCacheOverheadNS)
		metrics.Register("backend_cache_hits", http.ApiMetrics.BackendCacheHits)
		metrics.Register("backend_cache_misses", http.ApiMetrics.BackendCacheMisses)

		// requests histogram
		metrics.Register("requests", http.ApiMetrics.RequestsH)

		metrics.Register("find_requests", http.ApiMetrics.FindRequests)
		metrics.Register("render_requests", http.ApiMetrics.RenderRequests)

		if http.ApiMetrics.MemcacheTimeouts != nil {
			metrics.Register("memcache_timeouts", http.ApiMetrics.MemcacheTimeouts)
		}

		if http.ApiMetrics.CacheSize != nil {
			metrics.Register("cache_size", http.ApiMetrics.CacheSize)
			metrics.Register("cache_items", http.ApiMetrics.CacheItems)
		}

		metrics.Register("zipper.find_requests", http.ZipperMetrics.FindRequests)
		metrics.Register("zipper.find_errors", http.ZipperMetrics.FindErrors)

		metrics.Register("zipper.render_requests", http.ZipperMetrics.RenderRequests)
		metrics.Register("zipper.render_errors", http.ZipperMetrics.RenderErrors)

		metrics.Register("zipper.info_requests", http.ZipperMetrics.InfoRequests)
		metrics.Register("zipper.info_errors", http.ZipperMetrics.InfoErrors)

		metrics.Register("zipper.timeouts", http.ZipperMetrics.Timeouts)

		metrics.Register("zipper.cache_hits", http.ZipperMetrics.CacheHits)
		metrics.Register("zipper.cache_misses", http.ZipperMetrics.CacheMisses)

		metrics.RegisterRuntimeMemStats(nil)
		go metrics.CaptureRuntimeMemStats(config.Config.Graphite.Interval)

		// go mstats.Start(config.Config.Graphite.Interval)
		// metrics.Register("alloc", &mstats.Alloc)
		// metrics.Register("total_alloc", &mstats.TotalAlloc)
		// metrics.Register("num_gc", &mstats.NumGC)
		// metrics.Register("pause_ns", &mstats.PauseNS)

		graphite.Start(nil)
	}
}
