package main

import (
	"fmt"
	"hjh/tool"
	"time"
)

func main() {
	var rdb tool.RedisCli
	rdb.Init("34.134.141.170:6379", "cmVkaXNhZG1pbnBhc3N3b3JkCg==", 0)
	rdb.SetPexpire(time.Hour)
	rdb.SavePage("abcd", 1, "hello1")
	rdb.SavePage("abcd", 4, "hello4")
	rdb.SavePage("abcd", 8, "hello8")
	rdb.SavePage("abcd", 6, "hello6")
	count, _ := rdb.GetMaxPage("abcd")
	fmt.Println("count: ", count)
	val, _ := rdb.GetPage("abcd", 8)
	fmt.Println(val)
}
