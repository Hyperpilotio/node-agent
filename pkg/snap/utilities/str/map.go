package str

// StringMap is custom wrapper for golang map
// It provides additional capabilities like list of keys, list of values, Empty
type StringMap map[string]string

// AddMap adds all key/value pairs from map m to StringMap
// If key already exists, it will be overwritten by new value
func (sm StringMap) AddMap(m map[string]string) {
	if sm.Size() == 0 {
		sm = m
	} else {
		for k, v := range m {
			sm[k] = v
		}
	}
}

// RemoveAll deletes all key/value pairs from map
func (sm StringMap) RemoveAll() {
	sm = map[string]string{}
}

// Size return number of key/value pairs
func (sm StringMap) Size() int {
	return len(sm)
}

// Empty checks if map contains any key/value pair
func (sm StringMap) Empty() bool {
	return len(sm) == 0
}

// Keys returns slice of map keys
func (sm StringMap) Keys() []string {
	keys := []string{}
	for k := range sm {
		keys = append(keys, k)
	}
	return keys
}

// Values returns slice of map values
func (sm StringMap) Values() []string {
	values := []string{}
	for _, v := range sm {
		values = append(values, v)
	}
	return values
}

// HasKey checks if key exists in map
func (sm StringMap) HasKey(key string) bool {
	_, found := sm[key]
	return found
}
