package main

import (
	"database/sql"
	"log"
	"net"
	"net/rpc"
	"strconv"

	"github.com/megasay/butlerblaine/accountant"
)

//Accountant int
type Accountant struct {
	db *sql.DB
}

//UpdateDB func(send int, reply *int) error
func (a *Accountant) UpdateDB(send int, reply *int) error {
	log.Println("INFO:\tUpdateDB called.")
	err := processOperations(a.db)
	return err
}

//GetBalance func(account int, accountsBalance, balance *float64) error
func (a *Accountant) GetBalance(account int, balance *float64) error {
	type Item struct {
		Dostupno string
		Date     string
	}
	var item Item
	log.Println("INFO:\tGetBalance called.")
	log.Println("INFO:\tRequest for account: ", account)
	query := "SELECT dostupno, MAX(date) FROM operations WHERE account = '" + strconv.Itoa(account) + "' group by account;"

	if a.db == nil {
		log.Fatalf("ERROR:\tDB pointer is nil.")
	}
	rows, err := a.db.Query(query)
	if err != nil {
		log.Fatalf("ERROR:\tUnable to query DB: %v", err)
	}
	defer rows.Close()

	for rows.Next() {

		err = rows.Scan(&item.Dostupno, &item.Date)
		if err != nil {
			log.Fatalf("ERROR:\tUnable to scan rows DB: %v", err)
		}
	}

	*balance, _ = strconv.ParseFloat(item.Dostupno, 64)
	log.Println("INFO:\tBalance: ", *balance)
	return nil
}

//GetAccountsQuantity func(send int, quantity *int) error
func (a *Accountant) GetAccountsQuantity(send int, quantity *int) error {
	log.Println("INFO:\tGetAccountsQuantity called.")
	query := "SELECT COUNT(*) AS count FROM (SELECT account FROM operations group by account);"

	if a.db == nil {
		log.Fatalf("ERROR:\tDB pointer is nil.")
	}
	rows, err := a.db.Query(query)
	if err != nil {
		log.Fatalf("ERROR:\tUnable to query DB: %v", err)
	}
	defer rows.Close()

	for rows.Next() {

		err = rows.Scan(quantity)
		if err != nil {
			log.Fatalf("ERROR:\tUnable to scan rows DB: %v", err)
		}
	}
	log.Println("INFO:\tAccounts quantity: ", *quantity)
	return nil
}

//GetFullBalance func(account int, accountsBalance, balance *float64) error
func (a *Accountant) GetFullBalance(accountsQuantity int, balances *[]accountant.Balance) error {
	type Item struct {
		Account  int
		Dostupno string
		Date     string
	}
	var item Item
	log.Println("INFO:\tGetFullBalance called.")
	query := "SELECT account, dostupno, MAX(date) FROM operations group by account;"

	if a.db == nil {
		log.Fatalf("ERROR:\tDB pointer is nil.")
	}
	rows, err := a.db.Query(query)
	if err != nil {
		log.Fatalf("ERROR:\tUnable to query DB: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var balance accountant.Balance
		err = rows.Scan(&item.Account, &item.Dostupno, &item.Date)
		if err != nil {
			log.Fatalf("ERROR:\tUnable to scan rows DB: %v", err)
		}
		balance.Account = item.Account
		balance.Dostupno = item.Dostupno
		*balances = append(*balances, balance)
	}

	return nil
}

//GetLastUndefinedOperation func(send int, Description *string) error
func (a *Accountant) GetLastUndefinedOperation(send int, description *string) error {
	log.Println("INFO:\tGetLastUndefinedOperation called.")
	query := "SELECT operations.description FROM operations WHERE description NOT IN (SELECT description FROM defined_operations) LIMIT 1;"

	if a.db == nil {
		log.Fatalf("ERROR:\tDB pointer is nil.")
	}
	rows, err := a.db.Query(query)
	if err != nil {
		log.Fatalf("ERROR:\tUnable to query DB: %v", err)
	}
	defer rows.Close()

	for rows.Next() {

		err = rows.Scan(description)
		if err != nil {
			log.Fatalf("ERROR:\tUnable to scan rows DB: %v", err)
		}
	}
	log.Println("INFO:\tLast undefined operation description: ", *description)
	return nil
}

//DefineOperation func(do *accountant.DefinedOperation, reply *int) error
func (a *Accountant) DefineOperation(do *accountant.DefinedOperation, reply *int) error {

	log.Println("INFO:\tDefineOperation called.")
	if a.db == nil {
		log.Fatalf("ERROR:\tDB pointer is nil.")
	}
	var gid int64
	gid = -1
	rows, err := a.db.Query("SELECT id FROM groups WHERE name = '" + do.Group + "'")
	if err != nil {
		log.Fatalf("ERROR:\tUnable to query DB: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&gid)
		if err != nil {
			log.Fatalf("ERROR:\tUnable to scan rows DB: %v", err)
		}
	}

	if gid != -1 {
		log.Println("INFO:\tgid ", gid)
	} else {
		log.Println("INFO:\tGroup not found. Creating new group.")
		stmt, _ := a.db.Prepare("INSERT INTO groups(id, name, description) values(?,?,?);")
		result, err := stmt.Exec(nil, do.Group, "reserved")
		if err != nil {
			log.Fatalf("ERROR:\tInsert new group %v", err)
		} else {
			gid, _ = result.LastInsertId()
		}
		log.Println("INFO:\tNew group created successfuly.")
	}

	stmt, _ := a.db.Prepare("INSERT INTO defined_operations(id, gid, description) values(?,?,?);")
	_, err = stmt.Exec(nil, gid, do.Operation)
	if err != nil {
		log.Fatalf("ERROR:\t%v", err)
	}
	log.Println("INFO:\tOperation defined successfuly.")

	return nil
}

func startService(db *sql.DB) {

	accountant := new(Accountant)
	accountant.db = db
	rpc.Register(accountant)
	service := "127.0.0.1:1200"
	tcpAddr, err := net.ResolveTCPAddr("tcp", service)
	if err != nil {
		log.Fatalf("ERROR:\tResolveTCPAddr: %v", err)
	}

	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		log.Fatalf("ERROR:\tListenTCP: %v", err)
	}
	log.Printf("INFO:\tService started at %s.\n", service)
	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go rpc.ServeConn(conn)
	}

}
