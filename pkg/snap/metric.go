package snap

import (
	"fmt"
	"strings"
	"time"
)

// Metric contains all info related to a Snap Metric
type Metric struct {
	Namespace   Namespace
	Version     int64
	Config      Config
	Data        interface{}
	Tags        map[string]string
	Timestamp   time.Time
	Unit        string
	Description string
	//Unexported but passed through for legacy reasons
	lastAdvertisedTime time.Time
}

type Namespace []NamespaceElement

// Strings returns an array of strings that represent the elements of the
// namespace.
func (n Namespace) Strings() []string {
	var ns []string
	for _, namespaceElement := range n {
		ns = append(ns, namespaceElement.Value)
	}
	return ns
}

// IsDynamic returns true if there is any element of the namespace which is
// dynamic.  If the namespace is dynamic the second return value will contain
// an array of namespace elements (indexes) where there are dynamic namespace
// elements. A dynamic component of the namespace are those elements that
// contain variable data.
func (n Namespace) IsDynamic() (bool, []int) {
	var idx []int
	ret := false
	for i := range n {
		if n[i].IsDynamic() {
			ret = true
			idx = append(idx, i)
		}
	}
	return ret, idx
}

// Newnamespace takes an array of strings and returns a namespace.  A namespace
// is an array of namespaceElements.  The provided array of strings is used to
// set the corresponding Value fields in the array of namespaceElements.
func NewNamespace(ns ...string) Namespace {
	n := make([]NamespaceElement, len(ns))
	for i, ns := range ns {
		n[i] = NamespaceElement{Value: ns}
	}
	return n
}

// CopyNamespace copies array of namespace elements to new array
func CopyNamespace(src Namespace) Namespace {
	dst := make([]NamespaceElement, len(src))
	copy(dst, src)
	return dst
}

// AddDynamicElement adds a dynamic element to the given Namespace.  A dynamic
// namespaceElement is defined by having a nonempty Name field.
func (n Namespace) AddDynamicElement(name, description string) Namespace {
	nse := NamespaceElement{Name: name, Description: description, Value: "*"}
	return append(n, nse)
}

// AddStaticElement adds a static element to the given Namespace.  A static
// namespaceElement is defined by having an empty Name field.
func (n Namespace) AddStaticElement(value string) Namespace {
	nse := NamespaceElement{Value: value}
	return append(n, nse)
}

// AddStaticElements adds a static elements to the given Namespace.  A static
// namespaceElement is defined by having an empty Name field.
func (n Namespace) AddStaticElements(values ...string) Namespace {
	for _, value := range values {
		n = append(n, NamespaceElement{Value: value})
	}
	return n
}

func (n Namespace) Element(idx int) NamespaceElement {
	if idx >= 0 && idx < len(n) {
		return n[idx]
	}
	return NamespaceElement{}
}

// String returns the string representation of the namespace with "/" joining
// the elements of the namespace.  A leading "/" is added.
func (n Namespace) String() string {
	ns := n.Strings()
	s := n.getSeparator()
	return s + strings.Join(ns, s)
}

// getSeparator returns the highest suitable separator from the nsPriorityList.
// Otherwise the core separator is returned.
func (n Namespace) getSeparator() string {
	var nsPriorityList = []string{"/", "|", "%", ":", "-", ";", "_", "^", ">", "<", "+", "=", "&", "㊽", "Ä", "大", "小", "ᵹ", "☍", "ヒ"}
	var Separator = "\U0001f422"

	smap := map[string]bool{}

	for _, s := range nsPriorityList {
		smap[s] = false
	}

	for _, e := range n {
		// look at each char
		for _, r := range e.Value {
			ch := fmt.Sprintf("%c", r)
			if v, ok := smap[ch]; ok && !v {
				smap[ch] = true
			}
		}
	}

	// Go through our separator list
	for _, s := range nsPriorityList {
		if v, ok := smap[s]; ok && !v {
			return s
		}
	}
	return Separator
}

// namespaceElement provides meta data related to the namespace.
// This is of particular importance when the namespace contains data.
type NamespaceElement struct {
	Value       string
	Description string
	Name        string
}

// NewNamespaceElement tasks a string and returns a namespaceElement where the
// Value field is set to the provided string argument.
func NewNamespaceElement(e string) NamespaceElement {
	if e != "" {
		return NamespaceElement{Value: e}
	}
	return NamespaceElement{}
}

// IsDynamic returns true if the namespace element contains data.  A namespace
// element that has a nonempty Name field is considered dynamic.
func (n *NamespaceElement) IsDynamic() bool {
	if n.Name != "" {
		return true
	}
	return false
}
