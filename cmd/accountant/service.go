package main

import (
	"fmt"
	"net"
	"net/rpc"
	"os"
)

//Accountant int
type Accountant int

//GetBalance func(account int, accountsBalance, balance *float32) error
func (t *Accountant) GetBalance(account int, balance *float32) error {
	//var Network bytes.Buffer

	fmt.Println("GetBalance", account)
	*balance = 123.45
	fmt.Println("reply: ", *balance)
	return nil
}

func startService() {
	accountant := new(Accountant)
	rpc.Register(accountant)
	service := "127.0.0.1:1200"
	tcpAddr, err := net.ResolveTCPAddr("tcp", service)
	checkError(err)

	listener, err := net.ListenTCP("tcp", tcpAddr)
	checkError(err)

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go rpc.ServeConn(conn)
	}

}

func checkError(err error) {
	if err != nil {
		fmt.Println("Fatal error ", err.Error())
		os.Exit(1)
	}
}
