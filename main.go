package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	switch os.Args[1] {
	case "init":
		Init()
	default:
		log.Fatal("unknown command")
	}
}

func Init() {
	fmt.Println("hello world")
}
