/*
Define stream just like java
*/
package stream

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"reflect"
)

// A void func is a func that takes no argument, and returns nothing
// Note many other func can be easily converted to VoidFunc
// For example
//   func add(i, j int) {return i + j}
// =>
//   func(){
//     add(4, 5)
//   }
type VoidFunc func()

// A func returns a slice of values
type ValueFunc func() []interface{}

// ProducerFunc is a func that produces a result
// If your signature doesn't match, you can wrap it with inline func
type ProducerFunc func() interface{}

// A map function is a transform function
//   func(a interface{}) interface{} {
//     return a
//   }
// is a dummy map function.
// The type of x and result should be the same
type MapFunc func(x interface{}) interface{}

// A reduce function is a function that takes 2 element, but return only one.
// Sum(int, int) int is a reduce function
// The type of x,y and result should be the same
type ReduceFunc func(x, y interface{}) interface{}

// A LessThanCMpFunc takes two elements, return true if first is smaller than second logically
// The type of x,y should be the same
type LessThanCmpFunc func(x, y interface{}) bool

// Predict function is a filter/test func. Test whether an object match the predict.
//   a -> a>=5 is a predict function
type PredictFunc func(x interface{}) bool

// It is like a map function, but produce no result
//   a -> fmt.Println(a) is a MapCallFunc
type MapCallFunc func(x interface{})

// Natural LessThanCmp function. Only supports number and strings
// Panics if trying with other types
var NaturalCmpFunc LessThanCmpFunc = func(i, j interface{}) bool {
	t1 := reflect.TypeOf(i).Kind()
	t2 := reflect.TypeOf(j).Kind()
	if t1 != t2 {
		panic("Can't compare of different types")
	}

	switch t1 {
	case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int8, reflect.Int64:
		return reflect.ValueOf(i).Int() < reflect.ValueOf(j).Int()
	case reflect.Uint, reflect.Uint16, reflect.Uint32, reflect.Uint8, reflect.Uint64:
		return reflect.ValueOf(i).Uint() < reflect.ValueOf(j).Uint()
	case reflect.Float32, reflect.Float64:
		return reflect.ValueOf(i).Float() < reflect.ValueOf(j).Float()
	case reflect.String:
		return reflect.ValueOf(i).String() < reflect.ValueOf(i).String()
	default:
		panic("Can't compare of non int/float/string variable, please implement yourself")
	}
	return false
}

// Int Sum function. It is also a reduce function
var SumInt = func(i, j interface{}) interface{} {
	return i.(int) + j.(int)
}

// Int32 Sum function. It is also a reduce function
var SumInt32 = func(i, j interface{}) interface{} {
	return i.(int32) + j.(int32)
}

// int64 Sum function. It is also a reduce function
var SumInt64 = func(i, j interface{}) interface{} {
	return i.(int64) + j.(int64)
}

// uint32 Sum function. It is also a reduce function
var SumUint32 = func(i, j interface{}) interface{} {
	return i.(uint32) + j.(uint32)
}

// uint64 Sum function. It is also a reduce function
var SumUint64 = func(i, j interface{}) interface{} {
	return i.(uint64) + j.(uint64)
}

// float32 Sum function. It is also a reduce function
var SumFloat32 = func(i, j interface{}) interface{} {
	return i.(float32) + j.(float32)
}

// float64 Sum function. It is also a reduce function
var SumFloat64 = func(i, j interface{}) interface{} {
	return i.(float64) + j.(float64)
}

// Optional value may or may not have a value
type Optional interface {
	// Return the value and if value doesn't exist, false.
	// Caller need to check value exist before the value can be trusted
	Value() (interface{}, bool)

	// Return another optional value, when call Value() on the new optional,
	// return this object's value if present
	// If not, return other's Value()
	Or(other Optional) Optional

	// Return this object's value if present, or the defaultValue supplied
	OrValue(defaultV interface{}) interface{}
}

// Stream defines java like stream magics
type Stream interface {
	// For each element of the stream, apply a map function, and return the
	// mapped result stream. It is lazy so only operated when you use a terminal operator
	// e.g.
	//   stream.Range(0, 10).Map(func(a interface{}) interface{} {
	//     return a.(int) + 1
	//   }
	// will return a lazy stream, from 1~10 (not range(0,10) is 0~9)
	Map(f interface{}) Stream

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
	Reduce(f interface{}) Optional

	// Lazily limit the elements to process to number
	Limit(number int) Stream

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
	Filter(f interface{}) Stream

	// For each of the stream, do the function
	// e.g.
	// var sum := 0
	//   stream.Of(1, 2, 3).Each(func(i interface{}) {
	//     sum = sum + i.(int)
	//   })
	// will return 6
	//   Note this is a terminal operator
	Each(f interface{})

	// Close the stream. If your stream if from files, you have to close it.
	// Any OnClose handler previously attached will be called
	Close()

	// Attach another close handler to the stream.
	// The handler will be called after the previous handler had been called
	OnClose(closefunc func()) Stream

	// Return an iterator of elements from the stream
	Iterator() Iterator

	// Concat the other stream to end of current stream and return the new stream.
	// Both original stream and other stream itself are not modified.
	Concat(other Stream) Stream

	// Collect elements to target. Target must be array/slice of same type as the elements.
	// The max number if elements to collect is len(target).
	// Returns the number of elements collected
	// When number of elements collected equal to slice/array cap, there might be more elemnts to be collected
	// otherwise, it is guaranteed no more elements left in stream
	CollectTo(target interface{}) int

	// Get max in the stream using a less comparator function
	MaxCmp(f interface{}) Optional

	// Get min in the stream using a less comparator function
	MinCmp(f interface{}) Optional

	// Get max in the stream, using natural comparator. Only supports numbers and strings
	Max() Optional

	// Get min in the stream, using natural comparator. Only supports numbers and strings
	Min() Optional

	// Skip N elements in the stream, returning a new stream
	Skip(number int) Stream

	// Return a new stream of itself, but when elements are consumed, the peek function is called
	// e.g.
	//   sum := 0
	//   fmt.Println(stream.Range(0, 10).Peek(func(i interface) {
	//     sum = sum + i.(int)
	//   }).Count())
	//   fmt.Println("Sum is", sum)
	// This will count the elements, as well as add a sum
	Peek(f interface{}) Stream

	// Return sum of the elements, only supports int, int32, int64, uint32, uint64
	// float32, float64
	// call
	//   val, ok := result.Value()
	// if ok => val can be used
	// otherwise, the val can't be used
	Sum() Optional

	// Stream all elements to channel. May block the caller if
	// the channel is full
	SendTo(channel interface{})
}

// GenFunc is a function generate values, It is also a ProducerFunc
type GenFunc func() interface{}

type iterIter struct {
	seed         interface{}
	f            MapFunc
	initReturned bool
}

func (v *iterIter) Next() (interface{}, bool) {
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
func Iterate(seed interface{}, f interface{}) Stream {
	myf := toMapFunc(f)
	return &baseStream{&iterIter{seed, myf, false}, nil}
}

type genIter struct {
	f GenFunc
}

func (v *genIter) Next() (interface{}, bool) {
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
func Generate(f interface{}) Stream {
	myf := toGenFunc(f)
	return &baseStream{&genIter{myf}, nil}
}

// Return stream from iterator. stream's reduce function will consume all
// items in iterator
func FromIterator(it Iterator) Stream {
	return &baseStream{it, nil}
}

// Return stream from array. It is safe to call multiple FromArray on same array
// The call doesn't modify source array
func FromArray(it interface{}) Stream {
	ai := NewArrayIterator(it)
	return &baseStream{ai, nil}
}

// Return stream from channel. If you use same channel create multiple streams
// all streams will share the channel and may see part of the data
// Sender of channel must close or stream's reduce function/map function
// may not terminate
func FromChannel(it interface{}) Stream {
	ai := NewChannelIterator(it)
	return &baseStream{ai, nil}
}

// Create stream from map's keys. Note the iterator is a snapshot of map
// subsequent modification after Stream is created won't be visiable to stream
func FromMapKeys(it interface{}) Stream {
	ai := NewMapKeyIterator(it)
	return &baseStream{ai, nil}
}

// Create stream from map's values. Note the iterator is a snapshot of map
// subsequent modification after Stream is created won't be visiable to stream
func FromMapValues(it interface{}) Stream {
	ai := NewMapValueIterator(it)
	return &baseStream{ai, nil}
}

// Create stream from map's key value pairs. Note the iterator is a snapshot of map
// subsequent modification after Stream is created won't be visiable to stream
func FromMapEntries(it interface{}) Stream {
	ai := NewMapEntryIterator(it)
	return &baseStream{ai, nil}
}

// Return stream's mas, using natural comparison. Support number and string
// Note since stream might be empty, the value is Optional.
// Caller must use
//   val, ok := result.Value()
//   if ok {
//     do_something_with(val)
//   }
func (v *baseStream) Max() Optional {
	return v.MaxCmp(NaturalCmpFunc)
}

// Return stream's min, using natural comparison. Support number and string
// Note since stream might be empty, the value is Optional.
// Caller must use
//   val, ok := result.Value()
//   if ok {
//     do_something_with(val)
//   }
func (v *baseStream) Min() Optional {
	return v.MinCmp(NaturalCmpFunc)
}

// Return stream's max, using supplied less than comparator.
// Note since stream might be empty, the value is Optional.
// Caller must use
//   val, ok := result.Value()
//   if ok {
//     do_something_with(val)
//   }
func (v *baseStream) MaxCmp(f interface{}) Optional {
	cmpf := toLessThanCmpFunc(f)
	return v.Reduce(func(i, j interface{}) interface{} {
		if cmpf(i, j) {
			return j
		} else {
			return i
		}
	})
}

// Return stream's min, using supplied less than comparator.
// Note since stream might be empty, the value is Optional.
// Caller must use
//   val, ok := result.Value()
//   if ok {
//     do_something_with(val)
//   }
func (v *baseStream) MinCmp(f interface{}) Optional {
	cmpf := toLessThanCmpFunc(f)
	return v.Reduce(func(i, j interface{}) interface{} {
		if cmpf(i, j) {
			return i
		} else {
			return j
		}
	})
}

type baseStream struct {
	src       Iterator
	closefunc func()
}

type mapIterWrapper struct {
	src Iterator
	f   MapFunc
}

type filterIterWrapper struct {
	src Iterator
	f   PredictFunc
}

type limitIterWrapper struct {
	src   Iterator
	limit int
	count int
}

func (v *limitIterWrapper) Next() (interface{}, bool) {
	if v.limit <= v.count {
		return nil, false
	}

	nextv, ok := v.src.Next()
	if ok {
		v.count++
		return nextv, ok
	} else {
		return nil, ok
	}
}

func (v *filterIterWrapper) Next() (interface{}, bool) {
	for {
		next, ok := v.src.Next()
		if ok {
			if v.f(next) {
				return next, ok
			}
		} else {
			return nil, false
		}
	}
}

func (v *mapIterWrapper) Next() (interface{}, bool) {
	val, ok := v.src.Next()
	if !ok {
		return nil, false
	} else {
		return v.f(val), ok
	}
}

func (v *baseStream) Map(f interface{}) Stream {
	mf := toMapFunc(f)
	iter := v.src
	dest := &mapIterWrapper{iter, mf}
	return &baseStream{dest, nil}
}

type baseOptional struct {
	value   interface{}
	present bool
}

func (v *baseOptional) Value() (interface{}, bool) {
	if v.present {
		return v.value, true
	} else {
		return nil, false
	}
}

func (v *baseOptional) Or(other Optional) Optional {
	if v.present {
		return v
	} else {
		return other
	}
}

func (v *baseOptional) OrValue(dv interface{}) interface{} {
	if v.present {
		return v.value
	} else {
		return dv
	}
}

func (v *baseStream) Sum() Optional {
	iter := v.src
	next, ok := iter.Next()
	if !ok {
		return NewEmptyOptional()
	} else {
		switch next.(type) {
		case int:
			resulto := v.Reduce(SumInt)
			result, rok := resulto.Value()
			if rok {
				return NewOptional(SumInt(result, next))
			} else {
				return NewOptional(next)
			}
		case int32:
			resulto := v.Reduce(SumInt32)
			result, rok := resulto.Value()
			if rok {
				return NewOptional(SumInt32(result, next))
			} else {
				return NewOptional(next)
			}
		case int64:
			resulto := v.Reduce(SumInt64)
			result, rok := resulto.Value()
			if rok {
				return NewOptional(SumInt64(result, next))
			} else {
				return NewOptional(next)
			}
		case uint32:
			resulto := v.Reduce(SumUint32)
			result, rok := resulto.Value()
			if rok {
				return NewOptional(SumUint32(result, next))
			} else {
				return NewOptional(next)
			}
		case uint64:
			resulto := v.Reduce(SumUint64)
			result, rok := resulto.Value()
			if rok {
				return NewOptional(SumUint64(result, next))
			} else {
				return NewOptional(next)
			}
		case float32:
			resulto := v.Reduce(SumFloat32)
			result, rok := resulto.Value()
			if rok {
				return NewOptional(SumFloat32(result, next))
			} else {
				return NewOptional(next)
			}
		case float64:
			resulto := v.Reduce(SumFloat64)
			result, rok := resulto.Value()
			if rok {
				return NewOptional(SumFloat64(result, next))
			} else {
				return NewOptional(next)
			}
		default:
			panic("Can't sum on non int/int32/int64/uint32/uint64/float32/float64 elements, please write your own reduce function and use reduce instead")
		}
	}
}

func (v *baseStream) Close() {
	if v.closefunc != nil {
		v.closefunc()
	}
}

func (v *baseStream) Count() int {
	var count int = 0
	for _, ok := v.src.Next(); ok; _, ok = v.src.Next() {
		count++
	}
	return count
}

func (v *baseStream) Each(f interface{}) {
	mf := toMapCallFunc(f)
	for val, ok := v.src.Next(); ok; val, ok = v.src.Next() {
		mf(val)
	}
}

func (v *baseStream) Filter(f interface{}) Stream {
	pf := toPredictFunc(f)
	iter := v.src
	dest := &filterIterWrapper{iter, pf}
	return &baseStream{dest, nil}
}

func (v *baseStream) Limit(limit int) Stream {
	iter := v.src
	dest := &limitIterWrapper{iter, limit, 0}
	return &baseStream{dest, nil}
}

func (v *baseStream) OnClose(f func()) Stream {
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

// Return an optional has no value
// call to result.Value() will return nil, false
func NewEmptyOptional() Optional {
	return &baseOptional{nil, false}
}

// Return an optional has a value
// call to result.Value() will return val, true
func NewOptional(val interface{}) Optional {
	return &baseOptional{val, true}
}

func (v *baseStream) Reduce(f interface{}) Optional {
	rf := toReduceFunc(f)
	last_val, ok := v.src.Next()
	if !ok {
		return NewEmptyOptional()
	}

	for nv, ok := v.src.Next(); ok; nv, ok = v.src.Next() {
		last_val = rf(last_val, nv)
	}
	return NewOptional(last_val)
}

func (v *baseStream) Iterator() Iterator {
	return v.src
}

type concatIter struct {
	first, second                   Iterator
	firstExhausted, secondExhausted bool
}

func (v *concatIter) Next() (interface{}, bool) {
	if v.firstExhausted && v.secondExhausted {
		return nil, false
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
	return nil, false
}

func (v *baseStream) Concat(other Stream) Stream {
	iter := v.src
	dest := &concatIter{iter, other.Iterator(), false, false}
	return &baseStream{dest, nil}
}

func (v *baseStream) CollectTo(dest interface{}) (count int) {
	max := reflect.ValueOf(dest).Len()
	count = 0
	for val, ok := v.src.Next(); ok; val, ok = v.src.Next() {
		reflect.Indirect(reflect.ValueOf(dest)).Index(int(count)).Set(reflect.ValueOf(val))
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

func (v *fileLineIter) Next() (interface{}, bool) {
	ok := v.src.Scan()
	if !ok {
		return nil, false
	} else {
		return v.src.Text(), true
	}
}

// Return a stream from file lines. Caller must call close
// on return stream. If file open fails, the stream will
// be nil and error will be returned.
func FromFileLines(filepath string) (Stream, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	iter := &fileLineIter{scanner}
	result := FromIterator(iter)
	return result.OnClose(func() {
		fmt.Println("Closing file")
		file.Close()
	}), nil
}

// Create new stream from a reader, with the delimiter for tokenizer
func FromReader(r io.Reader, delimeter rune) Stream {
	scanner := bufio.NewScanner(r)
	scanner.Split(bufio.ScanLines)
	iter := &fileLineIter{scanner}
	return FromIterator(iter)
}

// Return Stream of range from low (inclusive) to high(exclusive).
// values are all int.
// Again, it is lazy, don't worry about the pre-allocation of memory.
// It is safe to do stream.Range(0, 10000000)
func Range(low, high int) Stream {
	return Iterate(low, func(i interface{}) interface{} {
		return i.(int) + 1
	}).Limit(high - low)
}

type skipIter struct {
	src     Iterator
	skipN   int
	skipped int
}

func (v *skipIter) Next() (interface{}, bool) {
	if v.skipped >= v.skipN {
		return v.src.Next()
	} else {
		var i int
		for i = 0; i < v.skipN; i++ {
			_, ok := v.src.Next()
			if !ok {
				return nil, false
			} else {
				v.skipped++
			}
		}
		return v.src.Next()
	}
}

func (v *baseStream) Skip(number int) Stream {
	return &baseStream{&skipIter{v.src, number, 0}, nil}
}

func (v *baseStream) SendTo(c interface{}) {
	v.Each(func(i interface{}) {
		reflect.ValueOf(c).Send(reflect.ValueOf(i))
	})
}

type peekIter struct {
	src Iterator
	f   MapCallFunc
}

func (v *peekIter) Next() (interface{}, bool) {
	next, ok := v.src.Next()
	if !ok {
		return next, ok
	} else {
		v.f(next)
		return next, ok
	}
}
func (v *baseStream) Peek(f interface{}) Stream {
	mcf := toMapCallFunc(f)
	return &baseStream{&peekIter{v.src, mcf}, nil}
}

// Create a stream from a set of values.
// e.g.
//   stream.Of(1,2,3,4,5).Sum().Value() => 15, true
// 15 is the result, true means the optional actually have a value
//   stream.Of().Sum().Value() => nil, false
// nil is default result, false means the stream is empty, so sum
// is non-existent
func Of(vars ...interface{}) Stream {
	return FromIterator(NewArrayIterator(vars))
}

func toLessThanCmpFunc(arg interface{}) LessThanCmpFunc {
	// arg must be a argument
	fv := reflect.ValueOf(arg)
	ft := reflect.TypeOf(arg)

	if ft.Kind() != reflect.Func {
		panic("Can't convert if arg isn't a function")
	}

	if ft.NumIn() != 2 {
		panic("Can't convert to less cmp func with numin != 2")
	}
	
	if ft.NumOut() < 1 {
		panic("Can't convert to less cmp func func with numout < 1")
	}

	if ft.Out(0).Kind() != reflect.Bool {
		panic("First return type must be boolean for less cmp func")
	}

	return func(arg1 interface{}, arg2 interface{}) bool {
		args := make([]reflect.Value, 2)
		args[0] = reflect.ValueOf(arg1)
		args[1] = reflect.ValueOf(arg2)
		var rv []reflect.Value
		if ft.IsVariadic() {
			rv = fv.CallSlice(args)
		} else {
			rv = fv.Call(args)
		}
		return rv[0].Bool()
	}
}

func toVoidFunc(args ...interface{}) VoidFunc {
	if len(args) < 1 {
		return func() {}
	}
	fv := reflect.ValueOf(args[0])
	argsV := toValues(args[1:])
	ft := reflect.TypeOf(args[0])
	if ft.Kind() != reflect.Func {
		panic("Can't convert if arg isn't a function")
	}

	return func() {
		//fmt.Println("Args0 =", args[0])
		//fmt.Println("Arguments are:", args[1:])
		if ft.IsVariadic() {

			//fmt.Println("Calling slice")
			fv.CallSlice(argsV)
		} else {
			//fmt.Println("Calling normal")
			//fmt.Println(argsV[0])
			fv.Call(argsV)
		}
	}
}

func toReduceFunc(arg interface{}) ReduceFunc {
	// arg must be a argument
	fv := reflect.ValueOf(arg)
	ft := reflect.TypeOf(arg)
	if ft.Kind() != reflect.Func {
		panic("Can't convert if arg isn't a function")
	}
	if ft.NumIn() != 2 {
		panic("Can't convert to reduce func with numin != 2")
	}
	if ft.NumOut() < 1 {
		panic("Can't conver to reduce func with numout < 1")
	}
	return func(arg1 interface{}, arg2 interface{}) interface{} {
		args := make([]reflect.Value, 2)
		args[0] = reflect.ValueOf(arg1)
		args[1] = reflect.ValueOf(arg2)
		var rv []reflect.Value
		if ft.IsVariadic() {
			rv = fv.CallSlice(args)
		} else {
			rv = fv.Call(args)
		}
		rvi := toInterfaces(rv)
		return rvi[0]
	}

}

func toPredictFunc(arg interface{}) PredictFunc {
	// arg must be a argument
	fv := reflect.ValueOf(arg)
	ft := reflect.TypeOf(arg)
	if ft.Kind() != reflect.Func {
		panic("Can't convert if arg isn't a function")
	}
	if ft.NumIn() != 1 {
		panic("Can't convert to predict func with numin != 1")
	}
	if ft.NumOut() < 1 {
		panic("Can't conver to predict func with numout < 1")
	}

	if ft.Out(0).Kind() != reflect.Bool {
		panic("First return type must be boolean for predict func")
	}

	return func(arg interface{}) bool {
		args := make([]reflect.Value, 1)
		args[0] = reflect.ValueOf(arg)
		var rv []reflect.Value
		if ft.IsVariadic() {
			rv = fv.CallSlice(args)
		} else {
			rv = fv.Call(args)
		}
		return rv[0].Bool()
	}

}
func toMapFunc(arg interface{}) MapFunc {
	// arg must be a argument
	fv := reflect.ValueOf(arg)
	ft := reflect.TypeOf(arg)
	if ft.Kind() != reflect.Func {
		panic("Can't convert if arg isn't a function")
	}
	if ft.NumIn() != 1 {
		panic("Can't convert to map func with numin != 1")
	}
	if ft.NumOut() < 1 {
		panic("Can't conver to map func with numout < 1")
	}
	return func(arg interface{}) interface{} {
		args := make([]reflect.Value, 1)
		args[0] = reflect.ValueOf(arg)
		var rv []reflect.Value
		if ft.IsVariadic() {
			rv = fv.CallSlice(args)
		} else {
			rv = fv.Call(args)
		}
		rvi := toInterfaces(rv)
		return rvi[0]
	}
}

func toMapCallFunc(arg interface{}) MapCallFunc {
	// arg must be a argument
	fv := reflect.ValueOf(arg)
	ft := reflect.TypeOf(arg)
	if ft.Kind() != reflect.Func {
		panic("Can't convert if arg isn't a function")
	}
	if ft.NumIn() != 1 {
		panic("Can't convert to map call func with numin != 1")
	}
	return func(arg interface{}) {
		args := make([]reflect.Value, 1)
		args[0] = reflect.ValueOf(arg)
		if ft.IsVariadic() {
			fv.CallSlice(args)
		} else {
			fv.Call(args)
		}
	}
}

func toGenFunc(arg interface{}) GenFunc {
	// arg must be a argument
	fv := reflect.ValueOf(arg)
	ft := reflect.TypeOf(arg)
	if ft.Kind() != reflect.Func {
		panic("Can't convert if arg isn't a function")
	}
	if ft.NumIn() != 0 {
		panic("Can't convert to generate func with numin != 0")
	}
	if ft.NumOut() < 1 {
		panic("Can't conver to generate func with numout < 1")
	}
	return func() interface{} {
		args := make([]reflect.Value, 0)
		var rv []reflect.Value
		if ft.IsVariadic() {
			rv = fv.CallSlice(args)
		} else {
			rv = fv.Call(args)
		}
		rvi := toInterfaces(rv)
		return rvi[0]
	}
}

func toFunc(args ...interface{}) func() []interface{} {
	if len(args) < 1 {
		return func() []interface{} {
			return nil
		}
	}
	fv := reflect.ValueOf(args[0])
	argsV := toValues(args[1:])
	ft := reflect.TypeOf(args[0])
	if ft.Kind() != reflect.Func {
		panic("Can't convert if arg isn't a function")
	}
	return func() []interface{} {
		//fmt.Println("Args0 =", args[0])
		//fmt.Println("Arguments are:", args[1:])
		if ft.IsVariadic() {
			//fmt.Println("Calling slice")
			rv := fv.CallSlice(argsV)
			return toInterfaces(rv)
		} else {
			//fmt.Println("Calling normal")
			//fmt.Println(argsV[0])

			rv := fv.Call(argsV)
			return toInterfaces(rv)
		}
	}
}

func toInterfaces(args []reflect.Value) []interface{} {
	result := make([]interface{}, len(args))
	for idx, v := range args {
		result[idx] = v.Interface()
	}
	return result
}

func toValues(args []interface{}) []reflect.Value {
	result := make([]reflect.Value, len(args))
	for idx, v := range args {
		result[idx] = reflect.ValueOf(v)
	}
	return result
}
