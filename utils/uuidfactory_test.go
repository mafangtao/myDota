package utils

import (
	"testing"
	"fmt"
)

func TestUUIDGenerator(t *testing.T) {
	//新建UUIDGennerator
	UUIDFactory := NewUUIDGenerator("idtest")

	//获取UUID
	for i:= 0; i < 50; i++{
		fmt.Println(UUIDFactory.Get())
	}

	//获取uint32 形式的UUID
	for i := 0; i < 50; i++{
		fmt.Println(UUIDFactory.GetUint32())
	}
}
