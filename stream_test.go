package stream

import (
	"fmt"
	"strconv"
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

	var i = 0
	new_stream := Generate(func() int {
		i = i + 1
		return i
	}).Limit(120)
	Pack2(new_stream).Each(
		func(i Pair[int, int]) {
			first, second := i.Value()
			fmt.Printf("Pair[%d %d]\n", first, second)
		})
	i = 0
	new_stream = Generate(func() int {
		i = i + 1
		return i
	}).Limit(115)
	Pack3(new_stream).Each(
		func(i Triple[int, int, int]) {
			first, second, third := i.Value()
			fmt.Printf("Triple[%d %d %d]\n", first, second, third)
		})
	i = 4
	new_stream = Generate(func() int {
		i = i + 1
		return i
	}).Limit(115)
	Pack4(new_stream).Each(
		func(i Quadruple[int, int, int, int]) {
			first, second, third, forth := i.Value()
			fmt.Printf("Quadruple[%d %d %d %d]\n", first, second, third, forth)
		})
	i = 0
	new_stream = Generate(func() int {
		i = i + 1
		return i
	}).Limit(115)
	Pack5(new_stream).Each(
		func(i Quintuple[int, int, int, int, int]) {
			first, second, third, forth, fifth := i.Value()
			fmt.Printf("Qunintuple[%d %d %d %d %d]\n", first, second, third, forth, fifth)
		})

	s6 := Of(1, 2, 3, 4, 5)
	s6.OnClose(func() {
		fmt.Println("s6 is closed")
	})
	s7 := s6.Peek(func(i int) {})
	s8 := s7.Limit(3)
	s9 := Map(s8, func(i int) int {
		return i + 18
	})
	s10 := s9.Limit(3)
	s10.Each(func(i int) {
		fmt.Println(i)
	})
	s10.Close()

	s11 := Of([]int{1, 2, 3}, []int{4, 5}, []int{}, []int{6, 7, 8, 9, 10})
	s12 := Flatten(s11, func(i []int) []int { return i })
	s12.Each(func(i int) {
		fmt.Println("Hello", i)
	})

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
	WithIndex(s14).Each(func(i NumberedItem[int]) {
		fmt.Println("Hello", i.Index, "->", i.Item)
	})
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
