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

#### Stream from slice/array

```
import "github.com/wushilin/stream"

var s Stream = stream.Range(0, 10) // a stream (which is lazy), that will contain [0..10)
var sum Optional = s.Filter(func(i interface{}) bool {
	return i.(int) >= 5
}).Sum()
// Note that sum is optional since stream might be empty.
// Be sure to get sum.Value() => return the value and a bool
// indicating the value's presense
fmt.Println(sum.OrValue("There is no element in the stream"))
```

Other typical scenarios:

```
	slice := make([]string, 300)
	// Read file as stream - you need to close the stream
	st, err := stream.FromFileLines("stream.go")
	if err != nil {
		panic(err)
	}
	// Collect first 300 lines 
	// (note this will return an int on elements collected)
	st.CollectTo(slice)

	// Attach a close handler for the stream.
	st = st.OnClose(func() {
		fmt.Println("Doing other things too")
	})
	// handlers are called in the order they are attached
	st = st.OnClose(func() {
		fmt.Println("Don't forget not to create loop!")
	})
	// Close stream. Typically only required for file IO
	st.Close()

	// Make stream from an array/slice
	st = stream.FromArray(slice)

	// Count elements in stream. Note this will consume stream
	fmt.Println(st.Count())

	st1 := stream.FromArray(slice)
	// Do a func for each of the elements
	st1.Each(print)

	fmt.Println("Testing filter")
	st2 := stream.FromArray(slice)
	// Lazy filter the stream with a predict. LAZY means stream is not consumed
	st2.Filter(isLength5).Each(print)

	fmt.Println("Testing limit")
	st3 := stream.FromArray(slice)
	// Limit limits the elements in stream. Again, it is LAZY. 
	// It won't consume stream
	st3.Limit(4).Each(print)

	fmt.Println("Testing Map")
	st4 := stream.FromArray(slice)
	// Map stream elements to a transform function and return the result
	// as a stream
	st4.Map(toupper).Each(print)

	fmt.Println("Testing reduce")
	st5 := stream.FromArray(slice)
	// Reduce a stream with reduce function
	fmt.Println(st5.Reduce(concat))

	fmt.Println("Test generator")
	// Generate an infinite stream
	sum := stream.Generate(func() interface{} {
		return 5
	// You can limit on infinite stream, then apply add as reduction function
	}).Limit(2000000).Reduce(add)
	fmt.Println("Sum is", sum, "which should be 10000000")

	fmt.Println("Testing fibonacci generator again")
	seed := 1
	fibonaci := func(i interface{}) interface{} {
		ii := i.(int)
		result := ii + seed
		seed = ii
		return result
	}
	// You can use Iterate to generate a stream of fibonacci numbers.
    // Again, it is LAZY!
	stream.Iterate(1, fibonaci).Limit(30).Each(print)

	fmt.Println("Testing sequence generator again")
	add1 := func(i interface{}) interface{} {
		ii := i.(int)
		return ii + 1
	}
	// You can use init value + iterative function to generate stream too
	stream.Iterate(1, add1).Limit(30).Each(print)

	fmt.Println("Testing range")
	// You can use Range to create stream too! Note high is not returned
	fmt.Println("0++30 = ", stream.Range(0, 30).Reduce(stream.SumInt))

	fmt.Println("Testing skip")
	// Count elements! You can skip elements too!
	fmt.Println(stream.Range(0, 30).Skip(50).Count())

	fmt.Println("Testing Of and filter")
	// You can create stream using the arguments
	stream.Of("a", "b", "c").Filter(func(a interface{}) bool {
		return a.(string) == "a"
	}).Each(print)

	fmt.Println("Testing peek")
	// You can add peek function, return the same stream, but on consumption
    // the peek function is called
	// You can also use Max() to get max value if elements are numbers/strings
	// Note this returns an optional
	fmt.Println(stream.Range(1, 10).Peek(print).Max()) 

	fmt.Println("Testing sum")
	// You can use Sum() too! Note that the sum return an Optional
	fmt.Println(stream.Range(1, 100).Sum().Value())

	// You can collect to slice from string. The max
    // to collect is actually the target's cap
	target := make([]int, 30)
	n := stream.Range(0, 20).CollectTo(target)
	fmt.Println(n)

	// You can create stream from Channel too!
    // You can send all elements from stream to channel too!
	fmt.Println("Testing channel")
	sourcec := make(chan int)
	done := make(chan bool)
	go func() {
		news := stream.FromChannel(sourcec)
		print(news.Skip(5).Sum().OrValue("No value returned"))
		done <- true
	}()
	stream.Range(0, 30).SendTo(sourcec)
	// Be sure you close this, otherwise stream will believe there 
	// are more elements so any attempt to read more than provided 
	// value, will haang there.
	close(sourcec)
	<-done
``` 


