package stream

import (
	stream "."
	"fmt"
	"strings"
	"testing"
)

func TestStream(*testing.T) {
	slice := make([]string, 300)
	st, err := stream.FromFileLines("stream.go")
	if err != nil {
		panic(err)
	}
	st.CollectTo(slice)
	st = st.OnClose(func() {
		fmt.Println("Doing other things too")
	})
	st = st.OnClose(func() {
		fmt.Println("Don't forget not to create loop!")
	})
	st.Close()

	st = stream.FromArray(slice)
	fmt.Println(st.Count())

	st1 := stream.FromArray(slice)
	st1.Each(print)

	fmt.Println("Testing filter")
	st2 := stream.FromArray(slice)
	st2.Filter(isLength5).Each(print)

	fmt.Println("Testing limit")
	st3 := stream.FromArray(slice)
	st3.Limit(4).Each(print)

	fmt.Println("Testing Map")
	st4 := stream.FromArray(slice)
	st4.Map(toupper).Each(print)

	fmt.Println("Testing reduce")
	st5 := stream.FromArray(slice)
	fmt.Println(st5.Reduce(concat))

	fmt.Println("Test generator")
	sum := stream.Generate(func() interface{} {
		return 5
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
	stream.Iterate(1, fibonaci).Limit(30).Each(print)

	fmt.Println("Testing sequence generator again")

	add1 := func(i interface{}) interface{} {
		ii := i.(int)
		return ii + 1
	}
	stream.Iterate(1, add1).Limit(30).Each(print)

	fmt.Println("Testing range")
	fmt.Println("0++30 = ", stream.Range(0, 30).Reduce(stream.SumInt))

	fmt.Println("Testing skip")
	fmt.Println(stream.Range(0, 30).Skip(50).Count())

	fmt.Println("Testing Of and filter")
	stream.Of("a", "b", "c").Filter(func(a interface{}) bool {
		return a.(string) == "a"
	}).Each(print)

	fmt.Println("Testing peek")
	fmt.Println(stream.Range(1, 10).Peek(print).Max())

	fmt.Println("Testing sum")
	fmt.Println(stream.Range(1, 100).Sum().Value())

	target := make([]int, 30)
	n := stream.Range(0, 20).CollectTo(target)
	fmt.Println(n)

	fmt.Println("Testing channel")
	sourcec := make(chan int)
	done := make(chan bool)
	go func() {
		news := stream.FromChannel(sourcec)
		print(news.Skip(5).Sum().OrValue("No value returned"))
		done <- true
	}()
	stream.Range(0, 30).SendTo(sourcec)
	close(sourcec)
	<-done
}

func print(i interface{}) {
	fmt.Println(i)
}

func add(i, j interface{}) interface{} {
	return i.(int) + j.(int)
}

func concat(i, j interface{}) interface{} {
	return i.(string) + j.(string)
}

func toupper(i interface{}) interface{} {
	return strings.ToUpper(i.(string))
}

func isLength5(i interface{}) bool {
	return len(i.(string)) == 5
}
