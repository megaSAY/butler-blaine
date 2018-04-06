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
	iReply := 0
	var sReply string

	fmt.Println(accountant.UpdateDB)
	err = client.Call(accountant.UpdateDB, 0, &iReply)
	if err != nil {
		log.Fatal("Account error:", err)
	}
	fmt.Println("reply: ", reply)

	fmt.Println(accountant.GetBalance)
	err = client.Call(accountant.GetBalance, 8336, &reply)
	if err != nil {
		log.Fatal("Account error:", err)
	}
	fmt.Println("reply: ", reply)

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

	fmt.Println(accountant.GetLastUndefinedOperation)
	err = client.Call(accountant.GetLastUndefinedOperation, 0, &sReply)
	if err != nil {
		log.Fatal("Account error:", err)
	}
	fmt.Println("reply: ", sReply)

	fmt.Println(accountant.DefineOperation)
	df := accountant.DefinedOperation{}
	df.Group = "SomeGroup3"
	df.Operation = sReply
	err = client.Call(accountant.DefineOperation, df, &iReply)
	if err != nil {
		log.Fatal("Account error:", err)
	}
	fmt.Println("reply: ", sReply)

	fmt.Println(accountant.GetLastUndefinedOperation)
	err = client.Call(accountant.GetLastUndefinedOperation, 0, &sReply)
	if err != nil {
		log.Fatal("Account error:", err)
	}
	fmt.Println("reply: ", sReply)

	fmt.Println(accountant.CancelOperationDefinition)
	/*df := accountant.DefinedOperation{}
	df.Group = "SomeGroup3"
	df.Operation = sReply*/
	err = client.Call(accountant.CancelOperationDefinition, df, &iReply)
	if err != nil {
		log.Fatal("Account error:", err)
	}
	fmt.Println("reply: ", sReply)

	fmt.Println(accountant.GetLastUndefinedOperation)
	err = client.Call(accountant.GetLastUndefinedOperation, 0, &sReply)
	if err != nil {
		log.Fatal("Account error:", err)
	}
	fmt.Println("reply: ", sReply)

}
