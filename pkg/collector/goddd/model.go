package goddd

// Summary a container to encode the summary type's metric from goddd / go-kit
type Summary struct {
	SampleCount uint64         `json:"sampleCount"`
	SampleSum   float64        `json:"sampleSum"`
	Quantile050 float64        `json:"quantile050"`
	Quantile090 float64        `json:"quantile090"`
	Quantile099 float64        `json:"quantile099"`
	Label       []*LabelStruct `json:"label"`
}

// LabelStruct smallest unit of label of metric
type LabelStruct struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// MetricList list of metrics from go-kit
var MetricList = []string{
	"api_booking_service_request_count",
	"api_booking_service_request_latency_microseconds",
	"go_gc_duration_seconds",
	"go_goroutines",
	"go_memstats_alloc_bytes",
	"go_memstats_alloc_bytes_total",
	"go_memstats_buck_hash_sys_bytes",
	"go_memstats_frees_total",
	"go_memstats_gc_sys_bytes",
	"go_memstats_heap_alloc_bytes",
	"go_memstats_heap_idle_bytes",
	"go_memstats_heap_inuse_bytes",
	"go_memstats_heap_objects",
	"go_memstats_heap_released_bytes_total",
	"go_memstats_heap_sys_bytes",
	"go_memstats_last_gc_time_seconds",
	"go_memstats_lookups_total",
	"go_memstats_mallocs_total",
	"go_memstats_mcache_inuse_bytes",
	"go_memstats_mcache_sys_bytes",
	"go_memstats_mspan_inuse_bytes",
	"go_memstats_mspan_sys_bytes",
	"go_memstats_next_gc_bytes",
	"go_memstats_other_sys_bytes",
	"go_memstats_stack_inuse_bytes",
	"go_memstats_stack_sys_bytes",
	"go_memstats_sys_bytes",
	"http_request_duration_microseconds",
	"http_request_size_bytes",
	"http_requests_total",
	"http_response_size_bytes",
	"process_cpu_seconds_total",
	"process_max_fds",
	"process_open_fds",
	"process_resident_memory_bytes",
	"process_start_time_seconds",
	"process_virtual_memory_bytes",
}

// MultiGroupsMetricList list of metrics that needs an extra "total" tag
var MultiGroupsMetricList = []string{
	"api_booking_service_request_count",
	"api_booking_service_request_latency_microseconds",
}

// Cache the structure of json file for cache
type CacheType struct {
	CounterType map[string]CounterCache `json:"counterType"`
}

// CounterCache struct to store counter type data
type CounterCache struct {
	Pre float64 `json:"pre"`
}
