# stream

## Install
```
go get github.com/wushilin/stream
```

## Documentation
```
$ godoc -http=":16666"

Browse http://localhost:16666
```

## Usage

### Stream API
What is a stream? Stream is a concept of collection of items, from various sources.
#### Creating Stream
There are couple of ways of creating streams. Please read throught code below.
```go
import "github.com/wushilin/stream"

// The most powerful stream construct is to build from iterator
// Iterator[T] implement a single method
//     Next() T, bool
// It returns [next value], true if a valud is available, or [anything], false if no more values available
s: = stream.FromIterator(iter)

// create stream from static list
s := stream.Of("1","2","3") // Stream[string]
s := stream.Of(1, 2, 3, 4, 5, 6) // Stream[int]

// Stream is generic enabled, you can infer types automatically

// Create stream from slice
s := stream.FromArray([]string{"1","2","3"}) // Stream[string]

// Create stream from a file
s := stream.FromFileLines("/tmp/input.txt") // Stream[string]
// Note that s.Close() must be called for file streams - it will close 
// the file handle

// Create stream from a receiver channel
c := make(chan int) 
s := stream.FromChannel(c)  // Stream[int]
// Note here the other go routines must send to the channel and close the channel

// Create stream from a generator function
s := stream.Generate(func() int {
  return 5
}) // Stream[int]
// This is pretty much useless, it returns a infinite stream of integer 5.
// But you can combine it with other features like Limit, Skip, and Sum

s := stream.Generate(func() int {
  return rand.Intn(12)
}) // Stream[int]
// This will create a infinite stream of random integers betwee 0 and 12

// Create stream from iterative function
s := stream.Iterate(1, func(i int) int {
  return i + 2
}) // Stream[int]
// This will generate a stream of even integers, starting from 1.

seed := 1
fibonaci_func := func(i int) int {
    result := i + seed
    seed = i
    return result
}
stream.Iterate(1, fibonaci).Limit(30) //Stream[int]
// This creates a stream of Fibonacci numbers, limited to first 30 items

// Create stream from a range
s := stream.Range(1, 100)
result := s.Sum() // Optional[int]
fmt.Println(s.Sum().Value()) // Prints 4950, true
// You will get 4950, true <- true means there is a sum for retrieval

// Create stream from Iterator
// dummy type implements Iterator, which has a Next() function 
// Next() function returns the next value and whether the value exists.
type dummy struct {
}

func (v dummy) Next() (int, bool) {
  return 5, true
}

s := stream.FromIterator(dummy{}) // Stream[int],  this is same as static generator function

// Create stream from Map keys and values
mp := make(map[string]string)
mp["1"] = "Jessy"
mp["2"] = "Steve"
sk := stream.FromMapKeys(mp) // Stream[string]
sv := stream.FromMapValues(mp) // Stream[string]
se := stream.FromMapEntries(mp) // Stream[MapEntry[string, string]]
```

#### Stream operators
##### Concat - consumes first stream first, when exhausted, consume the second
```go
s1 := stream.Range(0,5) // Stream[int]
s2 := stream.Range(5,10) // Stream[int]
s3 := s1.Concat(s2) // Stream[int]
// s3 is the same as stream.Range(0,10)
// this applies for all kinds of streams, not only ranges
```
##### Map - Map a transformation to existing stream
```go
// Due to go type system, Stream[T1].Map(func (T1) T2) => Stream[T2] won't work. We have use a static function like this unfortunately.
s1 := stream.Range(0, 5) // this stream, when consumed, will produce 0,1,2,3,4
s2 := stream.Map(s1, func(i int) int {
  return i + 5
}) // Stream[int]
// s2, when consumed, will consume s1, and produce 5,6,7,8,9
s3 := s2.Map(func(i int) string {
  return fmt.Sprintf("Student number %d", i)
} // Stream[string]
// s3, when consumed, will consume s2 and produce 
// "Student number 5", "Student number 6",...,"Student number 9"
```

##### Reduce - reduce a stream to a single object
The result is a stream.Optional object. The optional object has a Value()
function to retrieve its value
```go
s1 := stream.Of(1,2,3,4,5) // Stream[int]
sum := 0
reduce_func := func(i,j int) int { return i + j }
fmt.Println(s1.Reduce(reduce_func).Value())
// This will produce 15, true (15 is the result, true means there is a result)
// stream.Range(5, -1).Reduce(reduce_func).Value() will produce nil, false
// since stream.Range(5, -1) contains no elements.
```
##### Limit - limit the number of elements to emit. 
The remaining items of upper stream is not consumed
```go
s1 := stream.Generate(func() int { return 5 }) // s1 is infinite number of 5s
s2 := s1.Limit(20)
result := s2.Sum() // Optional[int]
fmt.Println(result.Value()) // will produce 100, true
```

##### Count - consume all elements and return the count
```go
fmt.Println(stream.Of("1","2").Count()) // produces 2
fmt.Println(stream.Range(0,100).Count()) // produces 100
fmt.Println(stream.Generate(func() int{ return 5 })) // never returns, infinite loop
```
##### Filter - filter elements based on predict
```go
s1 := stream.FromFileLines("/tmp/test.txt") // Stream[string], with OnClose handler to close the file automatically
s2 := s1.Filter(func(line string) bool {
  return len(line) < 100
}) // Stream[string]

defer s2.Close() // Eventually closes S2, which will close s1 as well!
// s2 is stream of strings, where the lines are shorter than 100 characters
```

##### Each - do a func for each of elements, doesn't expect return result
```go
stream.Range(0, 5).Each(func(i int) {
    fmt.Println(i, "* 2 =", i*2)
}) // Consume a stream
// Produces
// 0 * 2 = 0
// 1 * 2 = 2
// ...
// 4 * 2 = 8
```
##### Close - close the stream, and trigger OnClose function
Streams by default, has no close handler. so Close() does nothing. 
The only one has a default close handler is the stream.FromFileLines("filename")
When it is closed, the file gets closed. 
You can attach a close handler, so the handler is called when stream is closed.

```go
s1 := stream.Range(0, 100).OnClose(func() {
  fmt.Println("I am closed, good job man, I will call some other functions")
}).OnClose(func() {
  fmt.Println("I am closing some database connection")
}) // multiple OnClose handler will be called in sequence of the original attachment
defer s1.Close() // Close() by default is noop call
// Then you will see the two OnClose functions are called in order they are attached. (earlier OnClose func got called first).

```
##### Iterator - return iterator of the stream
streams are implemented using iterators. So it simply returns the underlying iterator
```go
iter := stream.Range(0,100).Iterator()
for val, has_next := iter.Next(); val, has_next = iter.Next(); has_next {
  // do something
}
```

##### Skip - skip N elements
```go
s1 := stream.Range(0, 100).Skip(50) // Stream[int] 50, 51, 52,...., 99
```
##### CollectAll - collect all elements to slice
```go
arr := stream.Range(0, 5).CollectAll() // []{0, 1, 2, 3, 4}
```
##### CollectTo - collect elements to slice, up to stream's availability and buffer capacity
```go
buffer := make([]int, 100)
collect_count := stream.Range(0, 100000).CollectTo(buffer)
// collect_count will be 100, and buffer will contain 0~99. Rest of values untouched

collect_count1 := stream.Range(0,1).CollectTo(buffer)
// collect_count1 will be 1, and buffer will contain 0. Rest of values untouched
```

##### MaxCmp - Do reduce use the less comparator to find max value
```go
max_cmp_func := func(i, j int) bool {
  return i < j
}
fmt.Println(stream.Range(0, 100).MaxCmp(max_cmp_func).Value()) // 99, true
```

##### MinCmp - Do reduce using less comparator to find min value
```go
max_cmp_func := func(i, j int) bool {
  return i < j
}
fmt.Println(stream.Range(0, 100).MinCmp(max_cmp_func).Value()) // 99, true
```

##### Max - do natural comparison max value. Builtin support for all numbers and strings
```go
fmt.Println(stream.Range(0, 100).MaxCmp(func(i, j int) int {
	if i < j {
		return -1
	}
  if i == j {
  	return 0
  }
  return 1
}).Value()) // 99, true
```

##### Min - do natural comparison min value. Builtin support for all numbers and strings
```go
fmt.Println(stream.Of(1.1, 1.2, 1.3).MinCmp(func(j, j int) int {
  if i < j {
    return -1
  }
  if i == j {
    return 0
  }
  return 1
}).Value()) // 1.1, true
```

##### Peek - return a stream that contains same element, but attach a trigger on consumption
```go
count := 0
s1 := stream.Range(0, 5) // Stream[int]
s2 := s1.Peek(func (i int) {
  count = count + 1
}) // Stream[int], same as s1, but consuming s2 now has a side effect of adding counter

fmt.Println(s2.Count()) // 5, and count will be 10 after consumption
```

##### WithIndex - index the number when consuming
```go
// NumberedItem[T] contains => Index: int64 and Item: T
s1 := stream.WithIndex(stream.Range(6, 100)) // Stream[NumberedItem[int]]

s1.Each(func(i stream.NumberedItem) {
  fmt.Println("Object index is:", i.Index) //int64
  fmt.Println("Object is:", i.Item) //int
})
```

##### SendTo - send stream to a channel
```go
destc := make(chan int)
done := make(chan bool)
go func() {
    for next_int := range(destc) {
      fmt.Println("Received:", next_int)
    }
    done <- true
}()

stream.Range(0, 100).SendTo(destc) // Consumes all elements and send to channel
close(destc)
<- done
// This will send 0~99 (100 numbers) to the channel for another go routine to consume.
```


##### Pack2, Pack3, Pack4, Pack5 Package stream items in pairs, triples, quadruples, and quintuples
```go
stream.Pack2(stream.Of(1,2,3,4,5)) => Pair[1,2], Pair[3,4]
stream.Pack3(stream.Of(1,2,3,4,5,6)) => Triple[1,2,3], Triple[4,5,6]
stream.Pack4(stream.Of(1,2,3,4,5,6)) => Quadruple[1,2,3,4]
...

// If remaining items are not enough for the packaging, they are ignored, but not silently ignored.
// The lasat value will be returned if you use Iterator to retrieve. The value is returned as partial Pair, and the value present is still false.
```


##### Flatten, expand elements in place
```go
	s13 := Of("1,2", "3,4,5", "", "6,7,8,9", "10") //Stream[string]
	s14 := Flatten(s13, func(i string) []int {
		tokens := strings.Split(i, ",")
		result := []int{}
		for _, str := range tokens {
			if str == "" {
				continue
			}
			num, _ := strconv.Atoi(str)
			result = append(result, num)
		}
		return result
	}) // Stream[int]
  //s14-> [1,2,3,4,5,6,7,8,9,10]
```

##### Parallel map and Future value support
```go
  // Warning - this may create many go routines. Only call it if you know the stream is bounded!
	src = Range(0, 1000) //Stream[int]
	// Spawn as many goroutines as possible to map each elements in the stream
  // using the mapper function.
	futs = ParallelMapUbounded(src, func(i int) int {
		<-time.After(5 * time.Second)
		return i + 100
	}) // Stream[future.Future[int]]
  // future.Future[T] supports:
  //   GetWait() T => Wait for the value to become available
  // . GetNow() (T,bool) => Try to get value immeidately, and do not block
  // . GetTimeout(duration) (T,bool) => Try to get value with timeout
  // . Then(func (T)) => Execute func (T) when value is available
  // Print the results
	futs.Each(func(i future.Future[int]) {
		fmt.Println(i.GetWait())
	})
  // The above code would take almost exactly 5 seconds
```

##### Parallel do with an executor
```go
	src := Range(0, 1000)
  // Create a new executor with max of 300 parallel goroutines
  // And max pending task is 1 million. See github.com/wushilin/threads
	threadPool := threads.NewPool(300, 1000000)
  // Eventually cleanup the pool
	defer func() {
		threadPool.Shutdown()
		threadPool.Wait()
		fmt.Println("Thread pool stopped")
	}()
  // Start the executor
	tp.Start()
  // Parall Map the object to future using the pool so max concurrency is enforced
	futs := ParallelMap(src, func(i int) int {
		<-time.After(5 * time.Second)
		return i + 100
	}, threadPool)


	futs.Each(func(i future.Future[int]) {
		fmt.Println(i.GetWait())
	})
  // The above code would take almost exact 20 seconds (4 batches literally)
```