package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"net"
	"net/rpc"
	"os"
)

type Accountant int

type Account struct {
	Name   string
	Amount float32
}

//GetBallance func(account int, accountsBalance *bytes.Buffer) error
func (t *Accountant) GetBallance(account int, accountsBalance *bytes.Buffer) error {
	var network bytes.Buffer
	// if account is 0 then all accunts ballance
	if account == 0 {

		s := make([]Account, 3)

		enc := gob.NewEncoder(&network)

		s[0].Name = "w"
		s[0].Amount = 212.3
		enc.Encode(&s)
		fmt.Println(network.Len())

		s[1].Name = "w"
		s[1].Amount = 212.3
		enc.Encode(&s)
		fmt.Println(network.Len())
	}
	accountsBalance = &network
	return nil
}

func startService() {
	service := "127.0.0.1:1200"
	tcpAddr, err := net.ResolveTCPAddr("tcp", service)
	checkError(err)

	listener, err := net.ListenTCP("tcp", tcpAddr)
	checkError(err)

	for {
		conn, _ := listener.Accept()
		go rpc.ServeConn(conn)
		/*
			if err != nil {
				continue
			}

			encoder := gob.NewEncoder(conn)
			decoder := gob.NewDecoder(conn)

			for n := 0; n < 10; n++ {
				var person Person
				decoder.Decode(&person)
				fmt.Println(person.String())
				encoder.Encode(person)
			}
			conn.Close() // we're finished
		*/
	}
}

func checkError(err error) {
	if err != nil {
		fmt.Println("Fatal error ", err.Error())
		os.Exit(1)
	}
}

/*
func handler() {
	fmt.Printf("DEBUG:\thandler.\n")
}
func startService() {
	var _hendler = &handler()
	l,err := net.Listen("")
}
*/
