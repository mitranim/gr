package gr

import (
	"net/http"
)

/*
Alias of `http.Header`. Provides additional methods. Provides a chainable
builder-style API. All methods support both canonical and non-canonical
versions of header keys, able to find EXISTING non-canonical entries,
canonicalizing them when possible. Freely castable to and from `http.Header`.
*/
type Head http.Header

// If nil, returns a non-nil empty map. If already allocated, returns self.
func (self Head) Init() Head {
	if self == nil {
		return Head{}
	}
	return self
}

// Free cast to `http.Header`.
func (self Head) Header() http.Header { return http.Header(self) }

// Same as `http.Header`. Makes a deep copy.
func (self Head) Clone() Head { return Head(self.Header().Clone()) }

/*
Similar to `http.Header.Get`, but also works if the key is non-canonical.
*/
func (self Head) Get(key string) string {
	if len(self) == 0 {
		return ``
	}

	val, ok := self[key]
	if ok {
		if len(val) > 0 {
			return val[0]
		}
		return ``
	}

	return self.Header().Get(key)
}

/*
Similar to `http.Header.Values`, but also works if the key is non-canonical.
*/
func (self Head) Values(key string) []string {
	if len(self) == 0 {
		return nil
	}
	val, ok := self[key]
	if ok {
		return val
	}
	return self.Header().Values(key)
}

/*
True if the header contains either this exact key or its canonical version.
*/
func (self Head) Has(key string) bool {
	// Avoids pointless key conversion below.
	if len(self) == 0 {
		return false
	}

	_, ok := self[key]
	if ok {
		return true
	}

	_, ok = self[canonKey(key)]
	return ok
}

/*
Deletes both this exact key and its canonical version. Mutates and returns the
receiver.
*/
func (self Head) Del(key string) Head {
	if len(self) == 0 {
		return self
	}

	delete(self, key)
	delete(self, canonKey(key))
	return self
}

/*
Similar to `http.Header.Add`, but also looks for an existing entry under this
EXACT key, as well as an entry for the canonical version of the key, combining
both under the canonical key. Internally calls `append`, which may mutate the
backing array of any existing slices for this key. Mutates and returns the
receiver. If the receiver is nil, allocates and returns a new map. For
correctness, you must always reassign the returned value.
*/
func (self Head) Add(key, val string) Head {
	keyCanon := canonKey(key)
	if self == nil {
		return Head{keyCanon: {val}}
	}

	if key == keyCanon {
		self[key] = append(self[key], val)
		return self
	}

	prev := self[key]
	prevCanon := self[keyCanon]

	if cap(prev) > 0 {
		self[keyCanon] = append(append(prev, prevCanon...), val)
	} else {
		self[keyCanon] = append(prevCanon, val)
	}

	delete(self, key)
	return self
}

/*
Similar to `http.Header.Set`, but also replaces the previous entry under this
exact key, if the key is non-canonical. The resulting entry always has the
canonical key. Mutates and returns the receiver. If the receiver is nil,
allocates and returns a new map. For correctness, you must always reassign the
returned value.
*/
func (self Head) Set(key, val string) Head {
	keyCanon := canonKey(key)
	if self == nil {
		return Head{keyCanon: {val}}
	}

	if key != keyCanon {
		delete(self, key)
	}
	self[keyCanon] = []string{val}
	return self
}

/*
Replaces the given key-value, canonicalizing the key. When called with no vals,
this is identical to `gr.Head.Del`, deleting the previous entry at BOTH this
exact key and its canonical version. When called with some vals, this replaces
the previous entry at the canonical version of this key, while deleting the
entry at this exact key. The received slice is set as-is, allowing you to reuse
a preallocated slice. Mutates and returns the receiver. If the receiver is nil,
allocates and returns a new map. For correctness, you must always reassign the
returned value.
*/
func (self Head) Replace(key string, vals ...string) Head {
	keyCanon := canonKey(key)

	if len(vals) == 0 {
		delete(self, key)
		delete(self, keyCanon)
		return self
	}

	if self == nil {
		return Head{keyCanon: vals}
	}

	if key != keyCanon {
		delete(self, key)
	}
	self[keyCanon] = vals
	return self
}

/*
Applies the patch, using `gr.Head.Replace` for each key-values entry. Mutates
and returns the receiver. If the receiver is nil, allocates and returns a new
map. For correctness, you must always reassign the returned value.

Accepts an "anonymous" type because all alias types such as `gr.Head` and
`http.Header` are automatically castable into it.
*/
func (self Head) Patch(head map[string][]string) Head {
	for key, vals := range head {
		self = self.Replace(key, vals...)
	}
	return self
}
