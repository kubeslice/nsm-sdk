// Code generated by "-output mechanism_map.gen.go -type mechanismMap<string,*github.com/networkservicemesh/api/pkg/api/networkservice.Mechanism> -output mechanism_map.gen.go -type mechanismMap<string,*github.com/networkservicemesh/api/pkg/api/networkservice.Mechanism>"; DO NOT EDIT.
// Install -output mechanism_map.gen.go -type mechanismMap<string,*github.com/networkservicemesh/api/pkg/api/networkservice.Mechanism> by "go get -u github.com/searKing/golang/tools/-output mechanism_map.gen.go -type mechanismMap<string,*github.com/networkservicemesh/api/pkg/api/networkservice.Mechanism>"

package mechanismtranslation

import (
	"sync" // Used by sync.Map.

	"github.com/networkservicemesh/api/pkg/api/networkservice"
)

// Generate code that will fail if the constants change value.
func _() {
	// An "cannot convert mechanismMap literal (type mechanismMap) to type sync.Map" compiler error signifies that the base type have changed.
	// Re-run the go-syncmap command to generate them again.
	_ = (sync.Map)(mechanismMap{})
}

var _nil_mechanismMap_networkservice_Mechanism_value = func() (val *networkservice.Mechanism) { return }()

// Load returns the value stored in the map for a key, or nil if no
// value is present.
// The ok result indicates whether value was found in the map.
func (m *mechanismMap) Load(key string) (*networkservice.Mechanism, bool) {
	value, ok := (*sync.Map)(m).Load(key)
	if value == nil {
		return _nil_mechanismMap_networkservice_Mechanism_value, ok
	}
	return value.(*networkservice.Mechanism), ok
}

// Store sets the value for a key.
func (m *mechanismMap) Store(key string, value *networkservice.Mechanism) {
	(*sync.Map)(m).Store(key, value)
}

// LoadOrStore returns the existing value for the key if present.
// Otherwise, it stores and returns the given value.
// The loaded result is true if the value was loaded, false if stored.
func (m *mechanismMap) LoadOrStore(key string, value *networkservice.Mechanism) (*networkservice.Mechanism, bool) {
	actual, loaded := (*sync.Map)(m).LoadOrStore(key, value)
	if actual == nil {
		return _nil_mechanismMap_networkservice_Mechanism_value, loaded
	}
	return actual.(*networkservice.Mechanism), loaded
}

// LoadAndDelete deletes the value for a key, returning the previous value if any.
// The loaded result reports whether the key was present.
func (m *mechanismMap) LoadAndDelete(key string) (value *networkservice.Mechanism, loaded bool) {
	actual, loaded := (*sync.Map)(m).LoadAndDelete(key)
	if actual == nil {
		return _nil_mechanismMap_networkservice_Mechanism_value, loaded
	}
	return actual.(*networkservice.Mechanism), loaded
}

// Delete deletes the value for a key.
func (m *mechanismMap) Delete(key string) {
	(*sync.Map)(m).Delete(key)
}

// Range calls f sequentially for each key and value present in the map.
// If f returns false, range stops the iteration.
//
// Range does not necessarily correspond to any consistent snapshot of the Map's
// contents: no key will be visited more than once, but if the value for any key
// is stored or deleted concurrently, Range may reflect any mapping for that key
// from any point during the Range call.
//
// Range may be O(N) with the number of elements in the map even if f returns
// false after a constant number of calls.
func (m *mechanismMap) Range(f func(key string, value *networkservice.Mechanism) bool) {
	(*sync.Map)(m).Range(func(key, value interface{}) bool {
		return f(key.(string), value.(*networkservice.Mechanism))
	})
}
