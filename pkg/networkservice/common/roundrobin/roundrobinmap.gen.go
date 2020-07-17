// Code generated by "go-syncmap -output roundrobinmap.gen.go -type roundRobinMap<string,*github.com/networkservicemesh/sdk/pkg/tools/algorithm/roundrobin.RoundRobin>"; DO NOT EDIT.

package roundrobin

import (
	"sync"

	"github.com/networkservicemesh/sdk/pkg/tools/algorithm/roundrobin"
)

func _() {
	// An "cannot convert roundRobinMap literal (type roundRobinMap) to type sync.Map" compiler error signifies that the base type have changed.
	// Re-run the go-syncmap command to generate them again.
	_ = (sync.Map)(roundRobinMap{})
}
func (m *roundRobinMap) Store(key string, value *roundrobin.RoundRobin) {
	(*sync.Map)(m).Store(key, value)
}

func (m *roundRobinMap) LoadOrStore(key string, value *roundrobin.RoundRobin) (*roundrobin.RoundRobin, bool) {
	actual, loaded := (*sync.Map)(m).LoadOrStore(key, value)
	if actual == nil {
		return nil, loaded
	}
	return actual.(*roundrobin.RoundRobin), loaded
}

func (m *roundRobinMap) Load(key string) (*roundrobin.RoundRobin, bool) {
	value, ok := (*sync.Map)(m).Load(key)
	if value == nil {
		return nil, ok
	}
	return value.(*roundrobin.RoundRobin), ok
}

func (m *roundRobinMap) Delete(key string) {
	(*sync.Map)(m).Delete(key)
}

func (m *roundRobinMap) Range(f func(key string, value *roundrobin.RoundRobin) bool) {
	(*sync.Map)(m).Range(func(key, value interface{}) bool {
		return f(key.(string), value.(*roundrobin.RoundRobin))
	})
}
