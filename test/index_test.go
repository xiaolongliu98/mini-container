package test

import (
	"fmt"
	"testing"
)

func TestAny(t *testing.T) {
	deferFunc := func() {
		fmt.Println("deferFunc")
	}

	fmt.Println("TestAny A")
	defer deferFunc()
	fmt.Println("TestAny B")
}
