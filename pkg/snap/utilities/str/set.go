package str

// StringSet is set implementation on top of map
type StringSet struct {
	set map[string]bool
}

// Add adds new item to set
func (set *StringSet) Add(element string) bool {
	_, found := set.set[element]
	set.set[element] = true
	return !found
}

// Delete removes element from set
func (set *StringSet) Delete(element string) bool {
	_, found := set.set[element]
	if found {
		delete(set.set, element)
	}
	return found
}

// Elements returns list of set elements
func (set *StringSet) Elements() []string {
	iter := []string{}
	for k := range set.set {
		iter = append(iter, k)
	}
	return iter
}

// Size returns number of elements in set
func (set *StringSet) Size() int {
	return len(set.set)
}

// InitSet initializes sets internal map
func InitSet() StringSet {
	set := StringSet{}
	set.set = map[string]bool{}
	return set
}
