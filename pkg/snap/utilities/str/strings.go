package str

// ForEach applies handler function to each elements of slice
// Handler function must transform string to string (string -> string)
func ForEach(slice []string, handler func(string) string) {
	for i, element := range slice {
		slice[i] = handler(element)
	}
}

// Contains checks if 'lookup' argument is available in given 'slice'.
// Lookup element must be of type string.
// Returns true if there is at least one such element.
func Contains(slice []string, lookup string) bool {
	for _, element := range slice {
		if element == lookup {
			return true
		}
	}
	return false
}

// Filter filters slice to elements which fulfill handler function condition.
// It returns new, filtered slice.
func Filter(slice []string, handler func(string) bool) []string {
	filtered := []string{}
	for _, element := range slice {
		if handler(element) {
			filtered = append(filtered, element)
		}
	}
	return filtered
}

// Any checks if there is at least one element in slice which fulfills handler function condition.
// Returns true if there is at lease one such element.
func Any(slice []string, handler func(string) bool) bool {
	for _, element := range slice {
		if handler(element) {
			return true
		}
	}
	return false
}

// All checks if all elements in slice fulfill handler function condition.
// Returns false if there is at lease one element which does not.
func All(slice []string, handler func(string) bool) bool {
	for _, element := range slice {
		if !handler(element) {
			return false
		}
	}
	return true
}
