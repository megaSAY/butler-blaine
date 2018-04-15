package main

import (
	"database/sql"
	"log"
	"net"
	"net/rpc"
	"strconv"

	"github.com/megasay/butler-blaine/accountant"
)

//Accountant int
type Accountant struct {
	db *sql.DB
}

//UpdateDB func(send int, reply *int) error
func (a *Accountant) UpdateDB(send int, reply *int) error {
	log.Println("INFO:\t*UpdateDB called.")
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
	log.Println("INFO:\t*GetBalance called.")
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
	log.Println("INFO:\t*GetAccountsQuantity called.")
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
	log.Println("INFO:\t*GetFullBalance called.")
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
	log.Println("INFO:\t*GetLastUndefinedOperation called.")
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

	log.Println("INFO:\t*DefineOperation called.")
	if a.db == nil {
		log.Fatalf("ERROR:\tDB pointer is nil.")
	}
	var gid int64
	gid = -1
	rows, err := a.db.Query("SELECT id FROM groups WHERE name = $1", do.Group)
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

//CancelOperationDefinition func(do *accountant.DefinedOperation, reply *int) error
func (a *Accountant) CancelOperationDefinition(do *accountant.DefinedOperation, reply *int) error {

	log.Println("INFO:\t*CancelOperationDefinition called.")
	if a.db == nil {
		log.Fatalf("ERROR:\tDB pointer is nil.")
	}
	var gid int64
	row := a.db.QueryRow("SELECT id FROM groups WHERE name = $1", do.Group)
	err := row.Scan(&gid)
	if err == sql.ErrNoRows {
		log.Printf("INFO:\tGroup '%s' not found.\n", do.Group)
	} else if err != nil {
		log.Fatalf("ERROR:\tUnable to scan id: %v", err)
	}

	log.Printf("INFO:\tCancel defenition. Gid %d. Operation %s.\n", gid, do.Operation)
	_, err = a.db.Exec("DELETE FROM defined_operations WHERE description = $1 AND gid = $2", do.Operation, gid)
	if err != nil {
		log.Fatalf("ERROR:\tUnable cancel definition: %v", err)
	}

	log.Println("INFO:\tOperation definition cancel successfuly.")

	return nil
}

//GetOperationsGroupsNames func(0, groupsNames *[]string) error
func (a *Accountant) GetOperationsGroupsNames(send int, groupsNames *[]string) error {
	log.Println("INFO:\t*GetOperationsGroupsNames.")
	query := "SELECT name FROM groups;"

	if a.db == nil {
		log.Fatalf("ERROR:\tDB pointer is nil.")
	}
	rows, err := a.db.Query(query)
	if err != nil {
		log.Fatalf("ERROR:\tUnable to query DB: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		err = rows.Scan(&name)
		if err != nil {
			log.Fatalf("ERROR:\tUnable to scan rows DB: %v", err)
		}
		*groupsNames = append(*groupsNames, name)
	}
	log.Println("INFO:\tGroups names: ", *groupsNames)
	return nil
}

//DeleteOperationsGroup func(name string, reply 0) error
func (a *Accountant) DeleteOperationsGroups(name string, reply int) error {
	log.Println("INFO:\t*DeleteOperationsGroup.")

	if a.db == nil {
		log.Fatalf("ERROR:\tDB pointer is nil.")
	}

	var gid int64
	row := a.db.QueryRow("SELECT id FROM groups WHERE name = $1", name)
	err := row.Scan(&gid)
	if err == sql.ErrNoRows {
		log.Printf("INFO:\tGroup '%s' not found.\n", name)
	} else if err != nil {
		log.Fatalf("ERROR:\tUnable to scan id: %v", err)
	}

	log.Printf("INFO:\tDelete group. Gid %d. name %s.\n", gid, name)
	_, err = a.db.Exec("DELETE FROM groups WHERE name = $1", name)
	if err != nil {
		log.Fatalf("ERROR:\tUnable delete group: %v", err)
	}

	_, err = a.db.Exec("DELETE FROM defined_operations WHERE gid = $1", gid)
	if err != nil {
		log.Fatalf("ERROR:\tUnable definitions for this group: %v", err)
	}

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
