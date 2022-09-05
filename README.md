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

// create stream from static list
s := stream.Of("1","2","3")

// Create stream from slice
s := stream.FromArray([]string{"1","2","3"})

// Create stream from a file
s := stream.FromFileLines("/tmp/input.txt")
// Note that s.Close() must be called for file streams - it will close 
// the file handle

// Create stream from a receiver channel
c := make(chan int)
s := stream.FromChannel(c)
// Note here the other go routines must send to the channel and close the channel

// Create stream from a generator function
s := stream.Generate(func() int {
  return 5
})
// This is pretty much useless, it returns a infinite stream of integer 5.
// But you can combine it with other features like Limit, Skip, and Sum

s := stream.Generate(func() int {
  return rand.Intn(12)
})
// This will create a infinite stream of random integers betwee 0 and 12

// Create stream from iterative function
s := stream.Iterate(1, func(i int) int {
  return i + 2
})
// This will generate a stream of even integers, starting from 1.

seed := 1
fibonaci_func := func(i int) int {
    result := i + seed
    seed = i
    return result
}
stream.Iterate(1, fibonaci).Limit(30)
// This creates a stream of Fibonacci numbers, limited to first 30 items

// Create stream from a range
s := stream.Range(1, 100)
fmt.Println(s.Sum())
// You will get 4950, true <- true means there is a sum for retrieval

// Create stream from Iterator
// dummy type implements Iterator, which has a Next() function 
// Next() function returns the next value and whether the value exists.
type dummy struct {
}

func (v dummy) Next() (int, bool) {
  return 5, true
}

s := stream.FromIterator(dummy{}) // this is same as static generator function

// Create stream from Map keys and values
mp := make(map[string]string)
mp["1"] = "Jessy"
mp["2"] = "Steve"
sk := stream.FromMapKeys(mp)
sv := stream.FromMapValues(mp)
se := stream.FromMapEntries(mp)
```

#### Stream operators
##### Concat - consumes first stream first, when exhausted, consume the second
```go
s1 := stream.Range(0,5)
s2 := stream.Range(5,10)
s3 := s1.Concat(s2) 
// s3 is the same as stream.Range(0,10)
// this applies for all kinds of streams, not only ranges
```
##### Map - Map a transformation to existing stream
```go
// Due to go type system, Stream[T1].Map(func (T1) T2) => Stream[T2] won't work. We have use a static function like this unfortunately.
s1 := stream.Range(0, 5) // this stream, when consumed, will produce 0,1,2,3,4
s2 := stream.Map(s1, func(i int) int {
  return i + 5
})
// s2, when consumed, will consume s1, and produce 5,6,7,8,9
s3 := s2.Map(func(i int) string {
  return fmt.Sprintf("Student number %d", i)
}
// s3, when consumed, will consume s2 and produce 
// "Student number 5", "Student number 6",...,"Student number 9"
```

##### Reduce - reduce a stream to a single object
The result is a stream.Optional object. The optional object has a Value()
function to retrieve its value
```go
s1 := stream.Of(1,2,3,4,5)
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
fmt.Println(s2.Sum().Value()) // will produce 100, true
```

##### Count - consume all elements and return the count
```go
fmt.Println(stream.Of("1","2").Count()) // produces 2
fmt.Println(stream.Range(0,100).Count()) // produces 100
fmt.Println(stream.Generate(func() int{ return 5 })) // never returns, infinite loop
```
##### Filter - filter elements based on predict
```go
s1 := stream.FromFileLines("/tmp/test.txt")
defer s1.Close()
s2 := s1.Filter(func(line string) bool {
  return len(line) < 100
})
// s2 is stream of strings, where the lines are shorter than 100 characters
```

##### Each - do a func for each of elements, doesn't expect return result
```go
stream.Range(0, 5).Each(func(i int) {
    fmt.Println(i, "* 2 =", i*2)
})
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
})
defer s1.Close()
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
s1 := stream.Range(0, 100).Skip(50) // 50, 51, 52,...., 99
```

##### CollectTo - collect elements to slice, up to stream's availability and buffer capacity
```go
buffer := make([]int, 100)
collect_count := stream.Range(0, 100000).CollectTo(buffer)
// collect_count will be 100, and buffer will contain 0~99

collect_count1 := stream.Range(0,1).CollectTo(buffer)
// collect_count1 will be 1, and buffer will contain 0. Rest of values untouched
```

##### MaxCmp - Do reduce use the less comparator to find max value
```go
max_cmp_func := func(i, j int) bool {
  return i < j
}
fmt.Println(stream.Range(0, 100).MaxCmp(max_cmp_func)) // 99, true
```

##### MinCmp - Do reduce using less comparator to find min value
```go
max_cmp_func := func(i, j int) bool {
  return i < j
}
fmt.Println(stream.Range(0, 100).MinCmp(max_cmp_func)) // 99, true
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
fmt.Println(stream.Range(0, 5).Peek(func(i int) {
  count = count + i
}).Count())
fmt.Println(count)
// Produces 5 (the count), 10 (the sum)
```

##### WithIndex - index the number when consuming
```go
stream.WithIndex(stream.Range(6, 100)).Each(func(i stream.NumberedItem) {
  fmt.Println("Object index is:", i.Index)
  fmt.Println("Object is:", i.Item)
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

stream.Range(0, 100).SendTo(destc)
close(destc)
<- done
// This will send 0~99 (100 numbers) to the channel for another go routine to consume.
```


##### Pack2, Pack3, Pack4, Pack5 Package stream items in pairs, triples, quadruples, and quintuples
```go
stream.Pack2(stream.Of(1,2,3,4,5)) => [1,2], [3,4]
stream.Pack3(stream.Of(1,2,3,4,5,6)) => [1,2,3], [4,5,6]
stream.Pack4(stream.Of(1,2,3,4,5,6)) => [1,2,3,4]
```


##### Flatten, expand elements in place
```go
	s13 := Of("1,2", "3,4,5", "", "6,7,8,9", "10")
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
	})
        //s14-> [1,2,3,4,5,6,7,8,9,10]
```
