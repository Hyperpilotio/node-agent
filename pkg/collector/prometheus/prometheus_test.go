package prometheus

import (
	"io"
	"math"
	"strings"
	"testing"

	"github.com/hyperpilotio/node-agent/pkg/snap"
	. "github.com/smartystreets/goconvey/convey"
)

type MockMetricsDownloader struct {
}

const TEST_DATA = `
# HELP api_booking_service_request_count Number of requests received.
# TYPE api_booking_service_request_count counter
api_booking_service_request_count{method="assign_to_route"} 3242
api_booking_service_request_count{method="book"} 586
api_booking_service_request_count{method="list_cargos"} 29775
api_booking_service_request_count{method="list_locations"} 29355
api_booking_service_request_count{method="request_routes"} 65306
# HELP api_booking_service_request_latency_microseconds Total duration of requests in microseconds.
# TYPE api_booking_service_request_latency_microseconds summary
api_booking_service_request_latency_microseconds{method="assign_to_route",quantile="0.5"} NaN
api_booking_service_request_latency_microseconds{method="assign_to_route",quantile="0.9"} NaN
api_booking_service_request_latency_microseconds{method="assign_to_route",quantile="0.99"} NaN
api_booking_service_request_latency_microseconds_sum{method="assign_to_route"} 4943.821842667005
api_booking_service_request_latency_microseconds_count{method="assign_to_route"} 3242
api_booking_service_request_latency_microseconds{method="book",quantile="0.5"} 0.507768194
api_booking_service_request_latency_microseconds{method="book",quantile="0.9"} 2.198266745
api_booking_service_request_latency_microseconds{method="book",quantile="0.99"} 3.102522408
api_booking_service_request_latency_microseconds_sum{method="book"} 1116.8336741770001
api_booking_service_request_latency_microseconds_count{method="book"} 586
api_booking_service_request_latency_microseconds{method="list_cargos",quantile="0.5"} 2.202007494
api_booking_service_request_latency_microseconds{method="list_cargos",quantile="0.9"} 3.6868262979999997
api_booking_service_request_latency_microseconds{method="list_cargos",quantile="0.99"} 5.19988122
api_booking_service_request_latency_microseconds_sum{method="list_cargos"} 127949.75269605908
api_booking_service_request_latency_microseconds_count{method="list_cargos"} 29766
api_booking_service_request_latency_microseconds{method="list_locations",quantile="0.5"} 0.306621555
api_booking_service_request_latency_microseconds{method="list_locations",quantile="0.9"} 1.212673664
api_booking_service_request_latency_microseconds{method="list_locations",quantile="0.99"} 2.402411438
api_booking_service_request_latency_microseconds_sum{method="list_locations"} 61888.972835401946
api_booking_service_request_latency_microseconds_count{method="list_locations"} 29350
api_booking_service_request_latency_microseconds{method="request_routes",quantile="0.5"} 0.584821579
api_booking_service_request_latency_microseconds{method="request_routes",quantile="0.9"} 1.601493396
api_booking_service_request_latency_microseconds{method="request_routes",quantile="0.99"} 2.811945445
api_booking_service_request_latency_microseconds_sum{method="request_routes"} 159950.66701443284
api_booking_service_request_latency_microseconds_count{method="request_routes"} 65281
# HELP go_gc_duration_seconds A summary of the GC invocation durations.
# TYPE go_gc_duration_seconds summary
go_gc_duration_seconds{quantile="0"} 6.7705e-05
go_gc_duration_seconds{quantile="0.25"} 0.000114327
go_gc_duration_seconds{quantile="0.5"} 0.000175894
go_gc_duration_seconds{quantile="0.75"} 0.080870819
go_gc_duration_seconds{quantile="1"} 0.090786708
go_gc_duration_seconds_sum 164.272704957
go_gc_duration_seconds_count 6487
# HELP go_goroutines Number of goroutines that currently exist.
# TYPE go_goroutines gauge
go_goroutines 437
# HELP go_memstats_alloc_bytes Number of bytes allocated and still in use.
# TYPE go_memstats_alloc_bytes gauge
go_memstats_alloc_bytes 2.1982296e+07
# HELP go_memstats_alloc_bytes_total Total number of bytes allocated, even if freed.
# TYPE go_memstats_alloc_bytes_total counter
go_memstats_alloc_bytes_total 9.264552748e+10
# HELP go_memstats_buck_hash_sys_bytes Number of bytes used by the profiling bucket hash table.
# TYPE go_memstats_buck_hash_sys_bytes gauge
go_memstats_buck_hash_sys_bytes 1.739178e+06
# HELP go_memstats_frees_total Total number of frees.
# TYPE go_memstats_frees_total counter
go_memstats_frees_total 3.986514233e+09
# HELP go_memstats_gc_sys_bytes Number of bytes used for garbage collection system metadata.
# TYPE go_memstats_gc_sys_bytes gauge
go_memstats_gc_sys_bytes 5.152768e+06
# HELP go_memstats_heap_alloc_bytes Number of heap bytes allocated and still in use.
# TYPE go_memstats_heap_alloc_bytes gauge
go_memstats_heap_alloc_bytes 2.1982296e+07
# HELP go_memstats_heap_idle_bytes Number of heap bytes waiting to be used.
# TYPE go_memstats_heap_idle_bytes gauge
go_memstats_heap_idle_bytes 7.7357056e+07
# HELP go_memstats_heap_inuse_bytes Number of heap bytes that are in use.
# TYPE go_memstats_heap_inuse_bytes gauge
go_memstats_heap_inuse_bytes 2.4977408e+07
# HELP go_memstats_heap_objects Number of allocated objects.
# TYPE go_memstats_heap_objects gauge
go_memstats_heap_objects 373755
# HELP go_memstats_heap_released_bytes_total Total number of heap bytes released to OS.
# TYPE go_memstats_heap_released_bytes_total counter
go_memstats_heap_released_bytes_total 4.7153152e+07
# HELP go_memstats_heap_sys_bytes Number of heap bytes obtained from system.
# TYPE go_memstats_heap_sys_bytes gauge
go_memstats_heap_sys_bytes 1.02334464e+08
# HELP go_memstats_last_gc_time_seconds Number of seconds since 1970 of last garbage collection.
# TYPE go_memstats_last_gc_time_seconds gauge
go_memstats_last_gc_time_seconds 1.4907389694112728e+09
# HELP go_memstats_lookups_total Total number of pointer lookups.
# TYPE go_memstats_lookups_total counter
go_memstats_lookups_total 8421
# HELP go_memstats_mallocs_total Total number of mallocs.
# TYPE go_memstats_mallocs_total counter
go_memstats_mallocs_total 3.986887988e+09
# HELP go_memstats_mcache_inuse_bytes Number of bytes in use by mcache structures.
# TYPE go_memstats_mcache_inuse_bytes gauge
go_memstats_mcache_inuse_bytes 4800
# HELP go_memstats_mcache_sys_bytes Number of bytes used for mcache structures obtained from system.
# TYPE go_memstats_mcache_sys_bytes gauge
go_memstats_mcache_sys_bytes 16384
# HELP go_memstats_mspan_inuse_bytes Number of bytes in use by mspan structures.
# TYPE go_memstats_mspan_inuse_bytes gauge
go_memstats_mspan_inuse_bytes 387296
# HELP go_memstats_mspan_sys_bytes Number of bytes used for mspan structures obtained from system.
# TYPE go_memstats_mspan_sys_bytes gauge
go_memstats_mspan_sys_bytes 1.294336e+06
# HELP go_memstats_next_gc_bytes Number of heap bytes when next garbage collection will take place.
# TYPE go_memstats_next_gc_bytes gauge
go_memstats_next_gc_bytes 2.6971744e+07
# HELP go_memstats_other_sys_bytes Number of bytes used for other system allocations.
# TYPE go_memstats_other_sys_bytes gauge
go_memstats_other_sys_bytes 1.142606e+06
# HELP go_memstats_stack_inuse_bytes Number of bytes in use by the stack allocator.
# TYPE go_memstats_stack_inuse_bytes gauge
go_memstats_stack_inuse_bytes 3.571712e+06
# HELP go_memstats_stack_sys_bytes Number of bytes obtained from system for stack allocator.
# TYPE go_memstats_stack_sys_bytes gauge
go_memstats_stack_sys_bytes 3.571712e+06
# HELP go_memstats_sys_bytes Number of bytes obtained by system. Sum of all system allocations.
# TYPE go_memstats_sys_bytes gauge
go_memstats_sys_bytes 1.15251448e+08
# HELP http_request_duration_microseconds The HTTP request latencies in microseconds.
# TYPE http_request_duration_microseconds summary
http_request_duration_microseconds{handler="prometheus",quantile="0.5"} 103930.902
http_request_duration_microseconds{handler="prometheus",quantile="0.9"} 498765.854
http_request_duration_microseconds{handler="prometheus",quantile="0.99"} 706319.863
http_request_duration_microseconds_sum{handler="prometheus"} 9.684036459999997e+06
http_request_duration_microseconds_count{handler="prometheus"} 21
# HELP http_request_size_bytes The HTTP request sizes in bytes.
# TYPE http_request_size_bytes summary
http_request_size_bytes{handler="prometheus",quantile="0.5"} 84
http_request_size_bytes{handler="prometheus",quantile="0.9"} 84
http_request_size_bytes{handler="prometheus",quantile="0.99"} 84
http_request_size_bytes_sum{handler="prometheus"} 1743
http_request_size_bytes_count{handler="prometheus"} 21
# HELP http_requests_total Total number of HTTP requests made.
# TYPE http_requests_total counter
http_requests_total{code="200",handler="prometheus",method="get"} 21
# HELP http_response_size_bytes The HTTP response sizes in bytes.
# TYPE http_response_size_bytes summary
http_response_size_bytes{handler="prometheus",quantile="0.5"} 1862
http_response_size_bytes{handler="prometheus",quantile="0.9"} 1877
http_response_size_bytes{handler="prometheus",quantile="0.99"} 1880
http_response_size_bytes_sum{handler="prometheus"} 46562
http_response_size_bytes_count{handler="prometheus"} 21
# HELP process_cpu_seconds_total Total user and system CPU time spent in seconds.
# TYPE process_cpu_seconds_total counter
process_cpu_seconds_total 1241.24
# HELP process_max_fds Maximum number of open file descriptors.
# TYPE process_max_fds gauge
process_max_fds 65536
# HELP process_open_fds Number of open file descriptors.
# TYPE process_open_fds gauge
process_open_fds 364
# HELP process_resident_memory_bytes Resident memory size in bytes.
# TYPE process_resident_memory_bytes gauge
process_resident_memory_bytes 7.081984e+07
# HELP process_start_time_seconds Start time of the process since unix epoch in seconds.
# TYPE process_start_time_seconds gauge
process_start_time_seconds 1.49073556736e+09
# HELP process_virtual_memory_bytes Virtual memory size in bytes.
# TYPE process_virtual_memory_bytes gauge
process_virtual_memory_bytes 1.21430016e+08
`

func (downloader MockMetricsDownloader) GetMetricsReader(url string) (io.Reader, error) {
	return strings.NewReader(TEST_DATA), nil
}

func (downloader MockMetricsDownloader) GetEndpoint(config snap.Config) (string, error) {
	return "test", nil
}

func TestPrometheusPlugin(t *testing.T) {
	Convey("Create Prometheus Collector", t, func() {
		collector, _ := New()
		Convey("So Prometheus collector should not be nil", func() {
			So(collector, ShouldNotBeNil)
		})
		Convey("collector.GetConfigPolicy() should return a config policy", func() {
			configPolicy, err := collector.GetConfigPolicy()
			Convey("So config policy should not be nil", func() {
				So(err, ShouldBeNil)
				So(configPolicy, ShouldNotBeNil)
				t.Log(configPolicy)
			})
			Convey("So config policy should be a policy.ConfigPolicy", func() {
				So(configPolicy, ShouldHaveSameTypeAs, snap.ConfigPolicy{})
			})
		})
	})

	Convey("Test parsing metrics", t, func() {
		// collector := New()
		collector := &PrometheusCollector{
			Downloader: &MockMetricsDownloader{},
		}

		Convey("Prometheus collect metrics should succesfully parse test metrics", func() {
			metricTypes, err := collector.GetMetricTypes(snap.Config{})
			So(err, ShouldBeNil)
			metrics, err := collector.CollectMetrics(metricTypes)
			So(err, ShouldBeNil)

			Convey("Prometheus collector should skip NaN metric values", func() {
				for _, metric := range metrics {
					So(math.IsNaN(metric.Data.(float64)), ShouldBeFalse)
				}
			})
		})
	})
}
