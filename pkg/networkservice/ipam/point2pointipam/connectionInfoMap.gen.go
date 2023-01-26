// Code generated by "-output connectionInfoMap.gen.go -type Map<string,*connectionInfo> -output connectionInfoMap.gen.go -type Map<string,*connectionInfo>"; DO NOT EDIT.
// Install -output connectionInfoMap.gen.go -type Map<string,*connectionInfo> by "go get -u github.com/searKing/golang/tools/-output connectionInfoMap.gen.go -type Map<string,*connectionInfo>"

package point2pointipam

import (
	"sync" // Used by sync.Map.
)

// Generate code that will fail if the constants change value.
func _() {
	// An "cannot convert Map literal (type Map) to type sync.Map" compiler error signifies that the base type have changed.
	// Re-run the go-syncmap command to generate them again.
	_ = (sync.Map)(Map{})
}

var _nil_Map_connectionInfo_value = func() (val *connectionInfo) { return }()

// Load returns the value stored in the map for a key, or nil if no
// value is present.
// The ok result indicates whether value was found in the map.
func (m *Map) Load(key string) (*connectionInfo, bool) {
	value, ok := (*sync.Map)(m).Load(key)
	if value == nil {
		return _nil_Map_connectionInfo_value, ok
	}
	return value.(*connectionInfo), ok
}

// Store sets the value for a key.
func (m *Map) Store(key string, value *connectionInfo) {
	(*sync.Map)(m).Store(key, value)
}

// LoadOrStore returns the existing value for the key if present.
// Otherwise, it stores and returns the given value.
// The loaded result is true if the value was loaded, false if stored.
func (m *Map) LoadOrStore(key string, value *connectionInfo) (*connectionInfo, bool) {
	actual, loaded := (*sync.Map)(m).LoadOrStore(key, value)
	if actual == nil {
		return _nil_Map_connectionInfo_value, loaded
	}
	return actual.(*connectionInfo), loaded
}

// LoadAndDelete deletes the value for a key, returning the previous value if any.
// The loaded result reports whether the key was present.
func (m *Map) LoadAndDelete(key string) (value *connectionInfo, loaded bool) {
	actual, loaded := (*sync.Map)(m).LoadAndDelete(key)
	if actual == nil {
		return _nil_Map_connectionInfo_value, loaded
	}
	return actual.(*connectionInfo), loaded
}

// Delete deletes the value for a key.
func (m *Map) Delete(key string) {
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
func (m *Map) Range(f func(key string, value *connectionInfo) bool) {
	(*sync.Map)(m).Range(func(key, value interface{}) bool {
		return f(key.(string), value.(*connectionInfo))
	})
}
