package main
import (
  "fmt"
)
func main() {
  var s1 []int
  if s1 == nil {
    fmt.Println("s1==nil")
  } else {
    fmt.Println("s1!=nil")
  }
  var arr = [5]int{}
  s1 = arr[:]
  s1 = make([]int, 0)
  if s1 == nil {
    fmt.Println("s1==nil")
  } else {
    fmt.Println("s1!=nil")
  }
  fmt.Println(s1)
}