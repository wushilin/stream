package stream

import (
	_ "fmt"
)

//Defined a iterator for generic purpose
type Iterator[T any] interface {
	/*
		Next() returns the next element. If an element is available, val is the element and ok is true.
		If an element is not available, val is nil, and ok is set as false
		A common way of iterating:
		for next, ok := iter.Next(); ok; next, ok = iter.Next() {
		   do_some_thing_with(next)
		}
	*/
	Next() (result T, ok bool)
}

/*
Utilty struct for array iteration.
Note: Not thread safe
*/
type arrayIterator[T any] struct {
	target []T
	index  int
	length int
}

/*
Return a new iterator based on array. target must be an array,
or a slice, or a pointer to an array, or a pointer to a slice
It can't be pointer to pointer to an array/slice
Note: Not thread safe
*/
func NewArrayIterator[T any](target []T) Iterator[T] {
	length := len(target)
	return &arrayIterator[T]{target, 0, length}
}

/*
Implements Iterator interface

  Note: Not thread safe
*/
func (v *arrayIterator[T]) Next() (result T, ok bool) {
	if v.index == v.length {
		var zv T
		return zv, false
	} else {
		result := v.target[v.index]
		v.index++
		return result, true
	}
}

// Map entry iteration: Just like java.util.Map.Entry class
type MapEntry[K, V any] struct {
	Key   K
	Value V
}

/*
Utility struct for map key iteration

  Note: Not thread safe
*/
type MapKeyIterator[K comparable, V any] struct {
	target map[K]V
	keys   []K
	index  int
	length int
}

/*
Implements Iterator. Returns a map key on each Next() call
All map keys are obtained when iterator is constructed. Thus no new key added after construction
will be returned.

  Note: Not thread safe
*/
func (v *MapKeyIterator[K, V]) Next() (result K, ok bool) {
	if v.index == v.length {
		var zv K
		return zv, false
	} else {
		k := v.keys[v.index]
		v.index++
		return k, true
	}
}

// Utility struct for map value iteration
//   Note: Not thread safe
type MapValueIterator[K comparable, V any] struct {
	target map[K]V
	keys   []K
	index  int
	length int
}

// Implements Iterator. Returns a map value on each Next() call
// All map keys are obtained when iterator is constructed. Thus no new key's value added after construction
// will be returned. However, values modified after iteration construction, new value will be returned
//   Note: Not thread safe
func (v *MapValueIterator[K, V]) Next() (result V, ok bool) {
	if v.index == v.length {
		var zv V
		return zv, false
	} else {
		k := v.keys[v.index]
		result, ok = v.target[k]
		v.index++
		return result, ok
	}
}

// Utility struct for Map Entry iteration.
// All map keys are obtained when iterator is constructed. Thus no new key's value added after construction
// will be returned. However, values modified after iteration construction, new  key/value will be returned
//   Note: Not thread safe
type MapEntryIterator[K comparable, V any] struct {
	target map[K]V
	keys   []K
	index  int
	length int
}

// Implements Iterator. Will return MapEntry on each call.
//   Note: Not thread safe
func (v *MapEntryIterator[K, V]) Next() (result MapEntry[K, V], ok bool) {
	if v.index == v.length {
		var zv MapEntry[K, V]
		return zv, false
	} else {
		k := v.keys[v.index]
		val := v.target[k]
		v.index++
		return MapEntry[K, V]{k, val}, true
	}
}

// Utility struct to for channel iteration
// Iterate through channels requires explicit channel closure. Failing to do so will result
// indefinite waiting.
//   Note: thread safe
type ChannelIterator[T any] struct {
	target chan T
}

// Implements Iterator. result will be channel's corresponding value type, requires casting
//   Note: thread safe
func (v *ChannelIterator[T]) Next() (result T, ok bool) {
	va, okb := <-v.target
	if okb {
		return va, okb
	} else {
		var zv T
		return zv, okb
	}
}

// Collect at most count elements from Iterator object.
// Returns a slice of object collected.  If count is 0, empty slice is returned.
// If count is less than 0, Maximum number of objects are collected (similar to Integer.MAX_VALUE in java)
//   Note: Collected objects might be less than count if Iterator reached end
//   Note: Not thread safe
func CollectN[T any](iter Iterator[T], count int) []T {
	result := make([]T, 0)
	if count == 0 {
		return result
	}
	for next, ok := iter.Next(); ok; next, ok = iter.Next() {
		result = append(result, next)
		if len(result) >= count && count > 0 {
			return result
		}
	}
	return result
}

// Same as CollecctN, but no limit
func CollectAll[T any](iter Iterator[T]) []T {
	return CollectN(iter, -1)
}

// Create Iterator on a channel.
// Target must be a channel
//   Note: Channel sender need to close channel or Iterator's last Next() call will hang forever
//   Multiple Iterator on same channel is possible since each Next() call will receive from the channel
func NewChannelIterator[T any](target chan T) *ChannelIterator[T] {
	return &ChannelIterator[T]{target}
}

// Create Iterator to iterate over a map's keys.
// Target must be a map or code may panic
// Multiple Iterator on same map is possible
func NewMapKeyIterator[K comparable, V any](target map[K]V) *MapKeyIterator[K, V] {
	keys := make([]K, len(target))
	var i = 0
	for k := range target {
		keys[i] = k
		i++
	}
	length := len(keys)
	return &MapKeyIterator[K, V]{target, keys, 0, length}
}

// Create Iterator to iterate over a map's values.
// Target must be a map
// Multiple Iterator on same map is possible
func NewMapValueIterator[K comparable, V any](target map[K]V) *MapValueIterator[K, V] {
	keys := keys(target)
	length := len(keys)
	return &MapValueIterator[K, V]{target, keys, 0, length}
}

func keys[K comparable, V any](target map[K]V) []K {
	keys := make([]K, len(target))
	var i = 0
	for k := range target {
		keys[i] = k
		i++
	}
	return keys
}

// Create Iterator to iterate over a map's entries
// Target must be a map
// Multiple Iterator on same map is possible.
func NewMapEntryIterator[K comparable, V any](target map[K]V) *MapEntryIterator[K, V] {
	keys := keys(target)
	length := len(keys)
	return &MapEntryIterator[K, V]{target, keys, 0, length}
}

