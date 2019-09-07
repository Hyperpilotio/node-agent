package ddagent

import "testing"

func TestAPI(t *testing.T) {
	d, _ := New()

	c := make(chan interface{})

	d.StreamMetrics(nil)

	<-c

}
