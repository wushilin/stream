/*
Define stream just like java
*/
package stream

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

// Comparator returns -1 if arg1 < arg2, returns 0 if arg1 == arg2 and returns 1 if arg1 > arg2
type Comparator[T any] func(arg1, arg2 T) int

// Optional value may or may not have a value
type Optional[T any] interface {
	// Return the value and if value doesn't exist, false.
	// Caller need to check value exist before the value can be trusted
	Value() (T, bool)

	// Return another optional value, when call Value() on the new optional,
	// return this object's value if present
	// If not, return other's Value()
	Or(other Optional[T]) Optional[T]

	// Return this object's value if present, or the defaultValue supplied
	OrValue(defaultV T) T
}

func Map[F any, T any](src Stream[F], f func(in F) T) Stream[T] {
	return &baseStream[T]{&mapIterWrapper[F, T]{src.Iterator(), f}, func() {
		src.Close()
	}}
}

// Stream defines java like stream magics
type Stream[T any] interface {
	// Due to Golang generic limitation, the map only takes interface{} as result
	// If you want to have the type parameterized, use stream.Map(src Stream[F], func(F) T) Stream[T]
	// For each element of the stream, apply a map function, and return the
	// mapped result stream. It is lazy so only operated when you use a terminal operator
	// e.g.
	//   stream.Range(0, 10).Map(func(a interface{}) interface{} {
	//     return a.(int) + 1
	//   }
	// will return a lazy stream, from 1~10 (not range(0,10) is 0~9)
	Map(f func(in T) interface{}) Stream[interface{}]

	// Use a reduce function to reduce elements in the stream and return reduced resumt.
	// e.g.
	//   stream.Range(0, 10).Reduce(func(a, b interface{}) interface{} {
	//     if a.(int) > b.(int) {
	//       return a
	//     }
	//     return b
	//   }
	// will return an optional, value is 9. This essentially reduce using Max(a,b) function
	// This is a terminal operator
	Reduce(f func(arg1, arg2 T) T) Optional[T]

	// Lazily limit the elements to process to number
	Limit(number int) Stream[T]

	// Count elements in stream.
	// This is terminal operator
	Count() int

	// Similar to Map, but skip if the predict func did not pass the test.
	// Return a lazy stream
	// e.g.
	//   stream.Range(0, 10).Filer(func(a interface{}) bool) {
	//      return a.(int) > 5
	//   }
	// will return a stream from [6 ~ 9]
	Filter(f func(arg T) bool) Stream[T]

	// For each of the stream, do the function
	// e.g.
	// var sum := 0
	//   stream.Of(1, 2, 3).Each(func(i interface{}) {
	//     sum = sum + i.(int)
	//   })
	// will return 6
	//   Note this is a terminal operator
	Each(f func(T))

	// Same as Each, but depend on the bool value to continue
	// returns true => move to next
	// returns false => aborts
	EachCondition(func(T) bool)
	// Close the stream. If your stream if from files, you have to close it.
	// Any OnClose handler previously attached will be called
	Close()

	// Attach another close handler to the stream.
	// The handler will be called after the previous handler had been called
	OnClose(closefunc func()) Stream[T]

	// Return an iterator of elements from the stream
	Iterator() Iterator[T]

	// Concat the other stream to end of current stream and return the new stream.
	// Both original stream and other stream itself are not modified.
	Concat(other Stream[T]) Stream[T]

	// Collect elements to target. Target must be array/slice of same type as the elements.
	// The max number if elements to collect is len(target).
	// Returns the number of elements collected
	// When number of elements collected equal to slice/array cap, there might be more elemnts to be collected
	// otherwise, it is guaranteed no more elements left in stream
	CollectTo(target []T) int

	// Get max in the stream using a less comparator function
	MaxCmp(cmp Comparator[T]) Optional[T]

	// Get min in the stream using a less comparator function
	MinCmp(cmp Comparator[T]) Optional[T]

	// Skip N elements in the stream, returning a new stream
	Skip(number int) Stream[T]

	// Return a new stream of itself, but when elements are consumed, the peek function is called
	// e.g.
	//   sum := 0
	//   fmt.Println(stream.Range(0, 10).Peek(func(i interface) {
	//     sum = sum + i.(int)
	//   }).Count())
	//   fmt.Println("Sum is", sum)
	// This will count the elements, as well as add a sum
	Peek(f func(T)) Stream[T]

	// Stream all elements to channel. May block the caller if
	// the channel is full
	SendTo(chan T)
}

// GenFunc is a function generate values, It is also a ProducerFunc
type GenFunc[T any] func() T

type iterIter[T any] struct {
	seed         T
	f            func(in T) T
	initReturned bool
}

func (v *iterIter[T]) Next() (T, bool) {
	if !v.initReturned {
		v.initReturned = true
		return v.seed, true
	} else {
		v.seed = v.f(v.seed)
		return v.seed, true
	}
}

// Generate a infinite stream, using seed as first element, then
// use MapFun and the previously returned value to generate the new value
// e.g.
//   func add1(i interface{}) interface{} {
//     return i.(int) + 1
//   }
//   stream.Iterate(1, add1).Limit(3) <= will produce the same as
//   stream.Of(1, 2, 3)
func Iterate[T any](seed T, f func(T) T) Stream[T] {
	return &baseStream[T]{&iterIter[T]{seed, f, false}, nil}
}

type genIter[T any] struct {
	f func() T
}

func (v *genIter[T]) Next() (T, bool) {
	return v.f(), true
}

// Generate will use the GenFunc to generate a infinite stream
// e.g.
//   stream.Generate(func() interface{} {
//     return 5
//   }
// will generate a infinite stream of 5.
// Note it is lazy so do not count infinite stream, it will not complete
// Similarly, do not Reduce infinite stream
// You can limit first before reducing
func Generate[T any](f func() T) Stream[T] {
	return &baseStream[T]{&genIter[T]{f}, nil}
}

// Return stream from iterator. stream's reduce function will consume all
// items in iterator
func FromIterator[T any](it Iterator[T]) Stream[T] {
	return &baseStream[T]{it, nil}
}

// Return stream from array. It is safe to call multiple FromArray on same array
// The call doesn't modify source array
func FromArray[T any](it []T) Stream[T] {
	ai := NewArrayIterator(it)
	return &baseStream[T]{ai, nil}
}

// Return stream from channel. If you use same channel create multiple streams
// all streams will share the channel and may see part of the data
// Sender of channel must close or stream's reduce function/map function
// may not terminate
func FromChannel[T any](it chan T) Stream[T] {
	ai := NewChannelIterator(it)
	return &baseStream[T]{ai, nil}
}

// Create stream from map's keys. Note the iterator is a snapshot of map
// subsequent modification after Stream is created won't be visiable to stream
func FromMapKeys[K comparable, V any](it map[K]V) Stream[K] {
	ai := NewMapKeyIterator(it)
	return &baseStream[K]{ai, nil}
}

// Create stream from map's values. Note the iterator is a snapshot of map
// subsequent modification after Stream is created won't be visiable to stream
func FromMapValues[K comparable, V any](it map[K]V) Stream[V] {
	ai := NewMapValueIterator(it)
	return &baseStream[V]{ai, nil}
}

// Create stream from map's key value pairs. Note the iterator is a snapshot of map
// subsequent modification after Stream is created won't be visiable to stream
func FromMapEntries[K comparable, V any](it map[K]V) Stream[MapEntry[K, V]] {
	ai := NewMapEntryIterator(it)
	return &baseStream[MapEntry[K, V]]{ai, nil}
}

// Return stream's max, using supplied less than comparator.
// Note since stream might be empty, the value is Optional.
// Caller must use
//   val, ok := result.Value()
//   if ok {
//     do_something_with(val)
//   }
func (v *baseStream[T]) MaxCmp(f Comparator[T]) Optional[T] {
	return v.Reduce(func(arg1, arg2 T) T {
		if f(arg1, arg2) >= 0 {
			return arg1
		}
		return arg2
	})
}

func (v *baseStream[T]) Map(f func(in T) interface{}) Stream[interface{}] {
	return Map[T](v, f)
}

// Return stream's min, using natural comparison. Support number and string
// Note since stream might be empty, the value is Optional.
// Caller must use
//   val, ok := result.Value()
//   if ok {
//     do_something_with(val)
//   }
func (v *baseStream[T]) MinCmp(f Comparator[T]) Optional[T] {
	return v.Reduce(func(arg1, arg2 T) T {
		if f(arg1, arg2) <= 0 {
			return arg1
		}
		return arg2
	})
}

type baseStream[T any] struct {
	src       Iterator[T]
	closefunc func()
}

type mapIterWrapper[F, T any] struct {
	src Iterator[F]
	f   func(F) T
}

type filterIterWrapper[T any] struct {
	src Iterator[T]
	f   func(T) bool
}

type limitIterWrapper[T any] struct {
	src   Iterator[T]
	limit int
	count int
}

func (v *limitIterWrapper[T]) Next() (T, bool) {
	if v.limit <= v.count {
		var zv T
		return zv, false
	}

	nextv, ok := v.src.Next()
	if ok {
		v.count++
		return nextv, ok
	} else {
		var zv T
		return zv, ok
	}
}

func (v *filterIterWrapper[T]) Next() (T, bool) {
	for {
		next, ok := v.src.Next()
		if ok {
			if v.f(next) {
				return next, ok
			}
		} else {
			var zv T
			return zv, false
		}
	}
}

func (v *mapIterWrapper[F, T]) Next() (T, bool) {
	val, ok := v.src.Next()
	if !ok {
		var zv T
		return zv, false
	} else {
		return v.f(val), ok
	}
}

type baseOptional[T any] struct {
	value   T
	present bool
}

func (v *baseOptional[T]) Value() (T, bool) {
	if v.present {
		return v.value, true
	} else {
		var zv T
		return zv, false
	}
}

func (v *baseOptional[T]) Or(other Optional[T]) Optional[T] {
	if v.present {
		return v
	} else {
		return other
	}
}

func (v *baseOptional[T]) OrValue(dv T) T {
	if v.present {
		return v.value
	} else {
		return dv
	}
}

type Number interface {
	~int8 | ~int16 | ~int32 | ~int64 | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~string | ~int | ~float32 | ~float64
}

func Sum[T Number](v Stream[T]) Optional[T] {
	return v.Reduce(func(a, b T) T {
		return a + b
	})
}

func (v *baseStream[T]) Close() {
	if v.closefunc != nil {
		v.closefunc()
	}
}

func (v *baseStream[T]) Count() int {
	var count int = 0
	for _, ok := v.src.Next(); ok; _, ok = v.src.Next() {
		count++
	}
	return count
}

func (v *baseStream[T]) Each(f func(T)) {
	for val, ok := v.src.Next(); ok; val, ok = v.src.Next() {
		f(val)
	}
}

func (v *baseStream[T]) EachCondition(f func(T) bool) {
	for val, ok := v.src.Next(); ok; val, ok = v.src.Next() {
		if !f(val) {
			break
		}
	}
}

func (v *baseStream[T]) Filter(f func(T) bool) Stream[T] {
	iter := v.src
	dest := &filterIterWrapper[T]{iter, f}
	return &baseStream[T]{dest, nil}
}

func (v *baseStream[T]) Limit(limit int) Stream[T] {
	iter := v.src
	dest := &limitIterWrapper[T]{iter, limit, 0}
	return &baseStream[T]{dest, nil}
}

func (v *baseStream[T]) OnClose(f func()) Stream[T] {
	oldclosefunc := v.closefunc
	v.closefunc = func() {
		if oldclosefunc != nil {
			oldclosefunc()
		}
		if f != nil {
			f()
		}
	}
	return v
}

func NewEmptyOptional[T any]() Optional[T] {
	var zv T
	return &baseOptional[T]{zv, false}
}

// Return an optional has a value
// call to result.Value() will return val, true
func NewOptional[T any](val T) Optional[T] {
	return &baseOptional[T]{val, true}
}

func (v *baseStream[T]) Reduce(f func(arg1, arg2 T) T) Optional[T] {

	last_val, ok := v.src.Next()
	if !ok {
		return NewEmptyOptional[T]()
	}

	for nv, ok := v.src.Next(); ok; nv, ok = v.src.Next() {
		last_val = f(last_val, nv)
	}
	return NewOptional(last_val)
}

func (v *baseStream[T]) Iterator() Iterator[T] {
	return v.src
}

type concatIter[T any] struct {
	first, second                   Iterator[T]
	firstExhausted, secondExhausted bool
}

func (v *concatIter[T]) Next() (T, bool) {
	if v.firstExhausted && v.secondExhausted {
		var zv T
		return zv, false
	}
	if !v.firstExhausted {
		val, ok := v.first.Next()
		if ok {
			return val, ok
		} else {
			v.firstExhausted = true
		}
	}

	if !v.secondExhausted {
		val, ok := v.second.Next()
		if ok {
			return val, ok
		} else {
			v.secondExhausted = true
		}
	}
	var zv T
	return zv, false
}

func (v *baseStream[T]) Concat(other Stream[T]) Stream[T] {
	iter := v.src
	dest := &concatIter[T]{iter, other.Iterator(), false, false}
	return &baseStream[T]{dest, nil}
}

func (v *baseStream[T]) CollectTo(dest []T) (count int) {
	max := len(dest)
	count = 0
	for val, ok := v.src.Next(); ok; val, ok = v.src.Next() {
		dest[count] = val
		count++
		if count >= max {
			break
		}
	}
	return
}

type fileLineIter struct {
	src *bufio.Scanner
}

func (v *fileLineIter) Next() (string, bool) {
	ok := v.src.Scan()
	if !ok {
		return "", false
	} else {
		return v.src.Text(), true
	}
}

// Return a stream from file lines. Caller must call close
// on return stream. If file open fails, the stream will
// be nil and error will be returned.
func FromFileLines(filepath string) (Stream[string], error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	iter := &fileLineIter{scanner}
	result := FromIterator[string](iter)
	return result.OnClose(func() {
		fmt.Println("Closing file")
		file.Close()
	}), nil
}

// Create new stream from a reader, with the delimiter for tokenizer
func FromReader(r io.Reader, delimeter rune) Stream[string] {
	scanner := bufio.NewScanner(r)
	scanner.Split(bufio.ScanLines)
	iter := &fileLineIter{scanner}
	return FromIterator[string](iter)
}

// Return Stream of range from low (inclusive) to high(exclusive).
// values are all int.
// Again, it is lazy, don't worry about the pre-allocation of memory.
// It is safe to do stream.Range(0, 10000000)
func Range(low, high int) Stream[int] {
	return Iterate(low, func(i int) int {
		return i + 1
	}).Limit(high - low)
}

type skipIter[T any] struct {
	src     Iterator[T]
	skipN   int
	skipped int
}

func (v *skipIter[T]) Next() (T, bool) {
	if v.skipped >= v.skipN {
		return v.src.Next()
	} else {
		var i int
		for i = 0; i < v.skipN; i++ {
			_, ok := v.src.Next()
			if !ok {
				var zv T
				return zv, false
			} else {
				v.skipped++
			}
		}
		return v.src.Next()
	}
}

func (v *baseStream[T]) Skip(number int) Stream[T] {
	return &baseStream[T]{&skipIter[T]{v.src, number, 0}, nil}
}

func (v *baseStream[T]) SendTo(c chan T) {
	v.Each(func(i T) {
		c <- i
	})
}

type peekIter[T any] struct {
	src Iterator[T]
	f   func(T)
}

func (v *peekIter[T]) Next() (T, bool) {
	next, ok := v.src.Next()
	if !ok {
		return next, ok
	} else {
		v.f(next)
		return next, ok
	}
}
func (v *baseStream[T]) Peek(f func(T)) Stream[T] {
	return &baseStream[T]{&peekIter[T]{v.src, f}, nil}
}

// Create a stream from a set of values.
// e.g.
//   stream.Of(1,2,3,4,5).Sum().Value() => 15, true
// 15 is the result, true means the optional actually have a value
//   stream.Of().Sum().Value() => nil, false
// nil is default result, false means the stream is empty, so sum
// is non-existent
func Of[T any](vars ...T) Stream[T] {
	return FromIterator(NewArrayIterator(vars))
}

type NumberedItem[T any] struct {
	Index int
	Item  T
}

// Allow stream to be accessed with Index.
// Usage:
// new_stream := stream.WithIndex(stream.Of("1","2","3"))
// new_stream.Each(func(i stream.NumberedItem) {
//    fmt.Println(i.Index, i.Item)
// })
// This is actually implemented using the Map function internally.
// func WithIndex(in Stream) Stream {
//   index := 0
//   return in.Map(func(i interface{}) NumberedItem {
//     new_index := index
//     index++
//     return NumberedItem{new_index, i}
//   })
// }

func WithIndex[T any](v Stream[T]) Stream[NumberedItem[T]] {
	index := 0
	return Map(v, func(i T) NumberedItem[T] {
		new_index := index
		index++
		return NumberedItem[T]{new_index, i}
	})
}
