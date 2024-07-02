package testcommon

import "math/rand"

type ValueIndex[V any] struct {
	v V
	i int
}

type RandomKeyMap[K comparable, V any] struct {
	rand *rand.Rand
	m    map[K]ValueIndex[V]
	s    []K
}

// RandomKeyMap is a map that is O(1) for insertion, deletion, and random key selection
func NewRandomKeyMap[K comparable, V any](r *rand.Rand) *RandomKeyMap[K, V] {
	return &RandomKeyMap[K, V]{
		rand: r,
		m:    make(map[K]ValueIndex[V]),
		s:    []K{},
	}
}

// Get returns an element from the map
func (rkm *RandomKeyMap[K, V]) Get(k K) (V, bool) {
	valueIndex, ok := rkm.m[k]
	return valueIndex.v, ok
}

// Upsert element into the map
func (rkm *RandomKeyMap[K, V]) Upsert(k K, v V) {
	if valueIndex, ok := rkm.m[k]; ok {
		rkm.m[k] = ValueIndex[V]{v, valueIndex.i}
		return
	}
	rkm.m[k] = ValueIndex[V]{i: len(rkm.s), v: v}
	rkm.s = append(rkm.s, k)
}

// Remove element from the map by swapping in the last element in its place.
func (rkm *RandomKeyMap[K, V]) Delete(k K) {
	valueIndexOfK, ok := rkm.m[k]
	if !ok {
		return
	}
	indexOfK := valueIndexOfK.i
	lastElementKey := rkm.s[len(rkm.s)-1]
	lastElementValue := rkm.m[lastElementKey].v
	// set the slice position of the deleted element to the last element
	rkm.s[indexOfK] = lastElementKey
	// chop off the last element of the slice
	rkm.s = rkm.s[:len(rkm.s)-1]
	// update the index of the last element in the map to its new slice position
	rkm.m[lastElementKey] = ValueIndex[V]{i: indexOfK, v: lastElementValue}
	// delete the element from the map
	delete(rkm.m, k)
}

// Get a random key from the map
func (rkm *RandomKeyMap[K, V]) RandomKey() K {
	if len(rkm.s) == 0 {
		panic("RandomKey called on empty M")
	}
	return rkm.s[rkm.rand.Intn(len(rkm.s))]
}

// length of the map & slice
func (rkm *RandomKeyMap[K, V]) Len() int {
	return len(rkm.s)
}
