package staff

import (
	"bytes"
	"encoding/gob"
	"log"
	"net/rpc"
)

func Hire(staffAddress string) (*rpc.Client, error) {
	staff, err := rpc.DialHTTP("tcp", staffAddress+":1234")
	if err != nil {
		log.Fatal("dialing:", err)
	}
	return staff, err
}

func Order(staff *rpc.Client, order string, args interface{}, results interface{}) error {
	network := new(bytes.Buffer)
	enc := gob.NewEncoder(network)
	enc.Encode(args)
	err := staff.Call(order, args, &results)
	if err != nil {
		log.Fatal("arith error:", err)
	}
	return err
}
