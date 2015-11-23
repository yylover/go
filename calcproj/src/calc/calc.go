package main

import (
	"os"
	"fmt"
	"simplemath"
	"strconv"
	)

var Usage = func() {
	fmt.Println("USAGE: calc command [arguments] ...")
	fmt.Println("\nThe commands are:\n\tadd\tAddition of two values.\n\tsqrt\tSquare root of a non-negative value;")
}

func main() {
	args := os.Args
	fmt.Println("args:", args)
	if args == nil || len(args) < 2 {
		fmt.Println("参数少")
		Usage()
		return;
	}

	switch args[1] {
	case "add":
		v1, _ := strconv.Atoi(args[2])
		v2, _ := strconv.Atoi(args[3])

		ret := simplemath.Add(v1, v2)
		fmt.Println("Result: ", ret)
	case "sqrt":
		v, err := strconv.Atoi(args[2])
		if err != nil {
			fmt.Println("sqrl error", err)
		}
		ret := simplemath.Sqrt(v)
		fmt.Println("Result", ret);
	default:
		fmt.Println("方法不对:", args[0])
		Usage()
	}
}

