package stream

import (
	"fmt"
	"strings"
	"testing"
)

func TestStream(*testing.T) {
	slice := make([]string, 300)
	st0, err := FromFileLines("stream.go")
	if err != nil {
		panic(err)
	}
	st0.CollectTo(slice)
	st0 = st0.OnClose(func() {
		fmt.Println("Doing other things too")
	})
	st0 = st0.OnClose(func() {
		fmt.Println("Don't forget not to create loop!")
	})
	defer st0.Close()
	st0, err = FromFileLines("stream.go")
	st := Map(st0, func(i string) string { return "->>>" + i })
	defer st.Close()

	st = st.Peek(func(i string) {
		fmt.Println(i)
	})
	fmt.Println("Consuming ST now")
	fmt.Println(st.Count())

	st1 := FromArray(slice)
	st1.Each(printString)

	fmt.Println("Testing filter")
	st2 := FromArray(slice)
	st2.Filter(isLength5).Each(printString)

	fmt.Println("Testing limit")
	st3 := FromArray(slice)
	st3.Limit(4).Each(printString)

	fmt.Println("Testing Map")
	st4 := FromArray(slice)
	Map(st4, toupper).Each(printString)

	fmt.Println("Testing reduce")
	st5 := FromArray(slice)
	fmt.Println(st5.Reduce(concat))

	fmt.Println("Testing generator")
	sum := Generate(func() int {
		return 5
	}).Limit(2000000).Reduce(add)
	fmt.Println("Sum is", sum, "which should be 10000000")

	fmt.Println("Testing fibonacci generator again")
	seed := 1
	fibonaci := func(i int) int {
		ii := i
		result := ii + seed
		seed = ii
		return result
	}
	Iterate(1, fibonaci).Limit(30).Each(printInt)

	fmt.Println("Testing sequence generator again")

	add1 := func(i int) int {
		ii := i
		return ii + 1
	}
	Iterate(1, add1).Limit(30).Each(printInt)

	fmt.Println("Testing range")
	fmt.Println("0++30 = ", Sum(Range(0, 30)))

	fmt.Println("Testing skip")
	fmt.Println(Range(0, 30).Skip(50).Count())

	fmt.Println("Testing Of and filter")
	Of("a", "b", "c").Filter(func(a string) bool {
		return a == "a"
	}).Each(printString)

	fmt.Println("Testing peek & Max, should be 9")
	fmt.Println(Range(1, 10).Peek(printInt).MaxCmp(func(i, j int) int {
		if i < j {
			return -1
		}
		if i == j {
			return 0
		}
		return 1
	}))

	fmt.Println("Testing sum")
	fmt.Println(Sum(Range(1, 100)))

	target := make([]int, 30)
	n := Range(0, 20).CollectTo(target)
	fmt.Println(n)

	fmt.Println("Testing channel")
	sourcec := make(chan int)
	done := make(chan bool)
	go func() {
		news := FromChannel(sourcec)
		print(Sum(news.Skip(5)).OrValue(0))
		done <- true
	}()
	Range(0, 30).SendTo(sourcec)
	close(sourcec)
	<-done
}

func printString(i string) {
	fmt.Println(i)
}
func printInt(i int) {
	fmt.Println(i)
}

func add(i, j int) int {
	return i + j
}

func concat(i, j string) string {
	return i + "\\n" + j
}

func toupper(i string) string {
	return strings.ToUpper(i)
}

func isLength5(i string) bool {
	return len([]rune(i)) == 5
}
