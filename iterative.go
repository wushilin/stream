package stream

import (
	_ "fmt"
	"reflect"
)
//Defined a iterator for generic purpose
type Iterator interface {
	/*
	Next() returns the next element. If an element is available, val is the element and ok is true.
	If an element is not available, val is nil, and ok is set as false
	A common way of iterating:  
	for next, ok := iter.Next(); ok; next, ok = iter.Next() {
	   do_some_thing_with(next)
	}
	*/
	Next() (val interface{}, ok bool)
}

/* 
Utilty struct for array iteration. 
Note: Not thread safe
*/
type arrayIterator struct {
	target reflect.Value
	index  int
	length int
}

/* 
Return a new iterator based on array. target must be an array, 
or a slice, or a pointer to an array, or a pointer to a slice
It can't be pointer to pointer to an array/slice
Note: Not thread safe
*/
func NewArrayIterator(target interface{}) Iterator {
	tv := reflect.Indirect(reflect.ValueOf(target))
	length := tv.Len()
	return &arrayIterator{tv, 0, length}
}

/*
Implements Iterator interface

  Note: Not thread safe
*/
func (v *arrayIterator) Next() (result interface{}, ok bool) {
	if v.index == v.length {
		return nil, false
	} else {
		result = v.target.Index(v.index).Interface()
		v.index++
		return result, true
	}
}

// Map entry iteration: Just like java.util.Map.Entry class
type MapEntry struct {
	Key   interface{}
	Value interface{}
}

/* 
Utility struct for map key iteration

  Note: Not thread safe
*/
type MapKeyIterator struct {
	target reflect.Value
	keys   []reflect.Value
	index  int
	length int
}

/* 
Implements Iterator. Returns a map key on each Next() call
All map keys are obtained when iterator is constructed. Thus no new key added after construction
will be returned.

  Note: Not thread safe
*/
func (v *MapKeyIterator) Next() (result interface{}, ok bool) {
	if v.index == v.length {
		return nil, false
	} else {
		k := v.keys[v.index]
		v.index++
		return k.Interface(), true
	}
}

// Utility struct for map value iteration
//   Note: Not thread safe
type MapValueIterator struct {
	target reflect.Value
	keys   []reflect.Value
	index  int
	length int
}

// Implements Iterator. Returns a map value on each Next() call
// All map keys are obtained when iterator is constructed. Thus no new key's value added after construction
// will be returned. However, values modified after iteration construction, new value will be returned
//   Note: Not thread safe 
func (v *MapValueIterator) Next() (result interface{}, ok bool) {
	if v.index == v.length {
		return nil, false
	} else {
		k := v.keys[v.index]
		result = v.target.MapIndex(k).Interface()
		v.index++
		return result, true
	}
}

// Utility struct for Map Entry iteration.
// All map keys are obtained when iterator is constructed. Thus no new key's value added after construction
// will be returned. However, values modified after iteration construction, new  key/value will be returned 
//   Note: Not thread safe
type MapEntryIterator struct {
	target reflect.Value
	keys   []reflect.Value
	index  int
	length int
}

// Implements Iterator. Will return MapEntry on each call.
//   Note: Not thread safe
func (v *MapEntryIterator) Next() (result interface{}, ok bool) {
	if v.index == v.length {
		return nil, false
	} else {
		k := v.keys[v.index]
		val := v.target.MapIndex(k).Interface()
		v.index++
		return MapEntry{k, val}, true
	}
}

// Utility struct to for channel iteration
// Iterate through channels requires explicit channel closure. Failing to do so will result 
// indefinite waiting.
//   Note: thread safe
type ChannelIterator struct {
	target reflect.Value
}

// Implements Iterator. result will be channel's corresponding value type, requires casting
//   Note: thread safe
func (v *ChannelIterator) Next() (result interface{}, ok bool) {
	val, okb := v.target.Recv()
	if okb {
		return val.Interface(), okb
	} else {
		return nil, okb
	}
}

// Collect at most count elements from Iterator object.
// Returns a slice of object collected.  If count is 0, empty slice is returned.
// If count is less than 0, Maximum number of objects are collected (similar to Integer.MAX_VALUE in java)
//   Note: Collected objects might be less than count if Iterator reached end 
//   Note: Not thread safe
func CollectN(iter Iterator, count int) ([]interface{}) {
	result := make([]interface{}, 0)
	if count == 0 {
		return result
	}
	for next, ok := iter.Next(); ok; next, ok = iter.Next() {
		result = append(result, next)
		if len(result) >= count && count > 0{
			return result
		}
	}
	return result
}

// Same as CollecctN, but no limit
func CollectAll(iter Iterator) []interface{} {
	return CollectN(iter, -1)
}

// Create Iterator on a channel. 
// Target must be a channel
//   Note: Channel sender need to close channel or Iterator's last Next() call will hang forever
//   Multiple Iterator on same channel is possible since each Next() call will receive from the channel
func NewChannelIterator(target interface{}) *ChannelIterator {
	v := reflect.ValueOf(target)
	
	return &ChannelIterator{v}
}

// Create Iterator to iterate over a map's keys.
// Target must be a map or code may panic
// Multiple Iterator on same map is possible
func NewMapKeyIterator(target interface{}) *MapKeyIterator {
	tv := reflect.Indirect(reflect.ValueOf(target))

	keys := tv.MapKeys()
	length := len(keys)
	return &MapKeyIterator{tv, keys, 0, length}
}

// Create Iterator to iterate over a map's values.
// Target must be a map
// Multiple Iterator on same map is possible
func NewMapValueIterator(target interface{}) *MapValueIterator {
	tv := reflect.Indirect(reflect.ValueOf(target))

	keys := tv.MapKeys()
	length := len(keys)
	return &MapValueIterator{tv, keys, 0, length}
}

// Create Iterator to iterate over a map's entries
// Target must be a map
// Multiple Iterator on same map is possible.
func NewMapEntryIterator(target interface{}) *MapEntryIterator {
	tv := reflect.Indirect(reflect.ValueOf(target))

	keys := tv.MapKeys()
	length := len(keys)
	return &MapEntryIterator{tv, keys, 0, length}
}
