package main

import (
	"fmt"
	"log"
	"net/rpc"
	"os"

	"github.com/megasay/butlerblaine/accountant"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: ", os.Args[0], "server:port")
		os.Exit(1)
	}
	service := os.Args[1]

	client, err := rpc.Dial("tcp", service)
	if err != nil {
		log.Fatal("dialing:", err)
	}

	var reply float32
	err = client.Call(accountant.GetBalance, 8336, &reply)
	if err != nil {
		log.Fatal("Account error:", err)
	}

	fmt.Println("reply: ", reply)
}
