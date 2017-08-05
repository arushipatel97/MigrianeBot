package main

type ThreadUnsafeSet map[string]struct{}

func New(a ...string) ThreadUnsafeSet {
	set := make(map[string]struct{})
	for _, i := range a {
		set[i] = struct{}{}
	}
	return set
}

func (set *ThreadUnsafeSet) Add(i string) bool {
	_, found := (*set)[i]
	(*set)[i] = struct{}{}
	return !found //False if it existed already
}

func (set *ThreadUnsafeSet) Contains(i ...string) bool {
	for _, val := range i {
		if _, ok := (*set)[val]; !ok {
			return false
		}
	}
	return true
}

func (set *ThreadUnsafeSet) IsSubset(other *ThreadUnsafeSet) bool {
	for elem := range *set {
		if !other.Contains(elem) {
			return false
		}
	}
	return true
}

func (set *ThreadUnsafeSet) IsSuperset(other *ThreadUnsafeSet) bool {
	return other.IsSubset(set)
}

func (set *ThreadUnsafeSet) Union(other *ThreadUnsafeSet) *ThreadUnsafeSet {
	// o := other.(*ThreadUnsafeSet)

	unionedSet := New()

	for elem := range *set {
		unionedSet.Add(elem)
	}
	for elem := range *other {
		unionedSet.Add(elem)
	}
	return &unionedSet
}

func (set *ThreadUnsafeSet) Intersect(other *ThreadUnsafeSet) *ThreadUnsafeSet {
	intersection := New()
	// loop over smaller set
	if set.Cardinality() < other.Cardinality() {
		for elem := range *set {
			if other.Contains(elem) {
				intersection.Add(elem)
			}
		}
	} else {
		for elem := range *other {
			if set.Contains(elem) {
				intersection.Add(elem)
			}
		}
	}
	return &intersection
}

func (set *ThreadUnsafeSet) Difference(other *ThreadUnsafeSet) *ThreadUnsafeSet {
	difference := New()
	for elem := range *set {
		if !other.Contains(elem) {
			difference.Add(elem)
		}
	}
	return &difference
}

func (set *ThreadUnsafeSet) SymmetricDifference(other *ThreadUnsafeSet) *ThreadUnsafeSet {
	aDiff := set.Difference(other)
	bDiff := other.Difference(set)
	return aDiff.Union(bDiff)
}

func (set *ThreadUnsafeSet) Clear() {
	*set = New()
}

func (set *ThreadUnsafeSet) Remove(i string) {
	delete(*set, i)
}

func (set *ThreadUnsafeSet) Cardinality() int {
	return len(*set)
}

func (set *ThreadUnsafeSet) Equal(other *ThreadUnsafeSet) bool {
	if set.Cardinality() != other.Cardinality() {
		return false
	}
	for elem := range *set {
		if !other.Contains(elem) {
			return false
		}
	}
	return true
}

func (set *ThreadUnsafeSet) Clone() *ThreadUnsafeSet {
	clonedSet := New()
	for elem := range *set {
		clonedSet.Add(elem)
	}
	return &clonedSet
}

func (set *ThreadUnsafeSet) ToSlice() []string {
	keys := make([]string, 0, set.Cardinality())
	for elem := range *set {
		keys = append(keys, elem)
	}

	return keys
}
