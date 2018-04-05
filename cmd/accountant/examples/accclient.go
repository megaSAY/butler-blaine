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

	/*
		fmt.Println(accountant.UpdateDB)
		err = client.Call(accountant.UpdateDB, to, &reply)
		if err != nil {
			log.Fatal("Account error:", err)
		}
		fmt.Println("reply: ", reply)
	*/

	fmt.Println(accountant.GetBalance)
	err = client.Call(accountant.GetBalance, 8336, &reply)
	if err != nil {
		log.Fatal("Account error:", err)
	}
	fmt.Println("reply: ", reply)

	iReply := 0
	fmt.Println(accountant.GetAccountsQuantity)
	err = client.Call(accountant.GetAccountsQuantity, 0, &iReply)
	if err != nil {
		log.Fatal("Account error:", err)
	}
	fmt.Println("reply: ", iReply)

	balances := []accountant.Balance{}
	fmt.Println(accountant.GetFullBalance)
	err = client.Call(accountant.GetFullBalance, iReply, &balances)
	if err != nil {
		log.Fatal("Account error:", err)
	}
	fmt.Println("reply: ", balances)

}
