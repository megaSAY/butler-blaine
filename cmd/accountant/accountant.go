package main

import (
	"database/sql"
	"encoding/base64"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strconv"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
)

// Command line argunments parsing
var flagStdoutVerbose = flag.Bool("stdout", defaultStdoutVerbose, "Print to stdout.")
var flagClientSecret = flag.String("secret", defaultClientSecret, "Path to client secret file.")
var flagDaysToProcess = flag.Int("days", defaultDaysToProcess, "Process operations during last n days.")
var flagDB = flag.String("base", defaultDB, "Path to data base file.")
var flagServiceMode = flag.Bool("server", defaultServiceMode, "Run as service.")

var stmt *sql.Stmt

const (
	iOpCardLastNum int = 1 + iota
	iOpSum
	iOpDescription
	iOpDate
	iOpDostupno
)

const (
	idxOpDateDay int = 1 + iota
	idxOpDateMon
	idxOpDateYear
	idxOpDateHour
	idxOpDateMin
)

const defaultClientSecret = "./client_secret.json"
const defaultDaysToProcess = 1
const defaultStdoutVerbose = false
const defaultDB = "./ops.db"
const defaultServiceMode = false

func insertOperation(tag string, not string) bool {
	match, _ := regexp.MatchString(".+"+tag+".+", not)
	if match {

		var compileStrings [2]string
		var data []string

		compileStrings[0] = `^Karta.*\*(\d+).*` + tag + ` (.+) RUR; (.+); (.+),.+dostupno (.+) RUR[. ]`
		compileStrings[1] = `^Schet.*\*(\d+).*.` + tag + `.(.+) RUB; (\w+).(.+);.+dostupno (.+).RUB\s`
		for i := 0; i < len(compileStrings); i++ {
			regexpOperation := regexp.MustCompile(compileStrings[i])
			data = regexpOperation.FindStringSubmatch(not)
			if len(data) == 6 {
				break
			}
		}
		if len(data) != 6 {
			log.Println("ERROR:\tRegular expression failed.")
			log.Printf("INFO:\tTag: '%s', Matches: %d", tag, len(data))
			for i := 0; i < len(compileStrings); i++ {
				log.Printf("INFO:\tRegular expression: %s", compileStrings[i])
			}
			return false
		}

		regexpDate := regexp.MustCompile(`(\d+).(\d+).(\d+) (\d+):(\d+)`)
		dataDate := regexpDate.FindStringSubmatch(data[iOpDate])
		if len(dataDate) != 6 {
			log.Println("ERROR:\tRegular expression failed.")
			log.Printf("INFO:\tString: %s", data[iOpDate])
			return match
		}

		strDate := fmt.Sprintf("%s-%s-%s %s:%s:00", dataDate[idxOpDateYear], dataDate[idxOpDateMon], dataDate[idxOpDateDay], dataDate[idxOpDateHour], dataDate[idxOpDateMin])
		cardLastNum, _ := strconv.Atoi(data[iOpCardLastNum])
		sum, _ := strconv.ParseFloat(data[iOpSum], 64)
		dostupno, _ := strconv.ParseFloat(data[iOpDostupno], 64)

		_, err := stmt.Exec(nil, tag, cardLastNum, sum, data[iOpDescription], strDate, dostupno)
		if err != nil {
			uniqueMatch, _ := regexp.MatchString("^UNIQUE.+", err.Error())
			if uniqueMatch {
				log.Println("INFO:\t" + err.Error())
				return match
			}
			log.Println("WARNING:\t" + err.Error())
			return match
		}
		log.Println("INFO:\tNew operation added successfully")
	}
	return match
}
func processOperations() bool {
	ctx := context.Background()

	b, err := ioutil.ReadFile(*flagClientSecret)
	if err != nil {
		log.Fatalf("ERROR:\tUnable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved credentials
	// at ~/.credentials/gmail.json
	config, err := google.ConfigFromJSON(b, gmail.GmailReadonlyScope)
	if err != nil {
		log.Fatalf("ERROR:\tUnable to parse client secret file to config: %v", err)
	}
	client := getClient(ctx, config)
	srv, err := gmail.New(client)
	if err != nil {
		log.Fatalf("ERROR:\tUnable to retrieve gmail Client %v", err)
	}

	pageToken := ""

	user := "me"
	now := time.Now()
	timeAfter := now.AddDate(0, 0, -*flagDaysToProcess)
	afterDate := fmt.Sprintf("%d/%d/%d", timeAfter.Year(), timeAfter.Month(), timeAfter.Day())
	req := srv.Users.Messages.List(user).Q("from:(notify@vtb24.ru OR notify@vtb.ru),after:" + afterDate)

	if pageToken != "" {
		req.PageToken(pageToken)
	}
	r, err := req.Do()
	if err != nil {
		log.Fatalf("ERROR:\tUnable to retrieve messages: %v", err)
	}
	log.Printf("INFO:\tRequest operations after %s date.\n", afterDate)
	log.Printf("INFO:\tProcessing %v operations...\n", len(r.Messages))

	tagsProc := [...]string{"spisanie", "Oplata", "snyatie", "zachislenie", "postuplenie"}
	tagsIgnore := [...]string{"voshli", "uvelichen balans scheta na", "umenshen balans scheta na", "Oplata  otklonena"}
	for _, m := range r.Messages {
		msg, _ := srv.Users.Messages.Get(user, m.Id).Do()
		ud, _ := base64.URLEncoding.DecodeString(msg.Payload.Body.Data)
		ignore := false

		log.Println("PROCESSING:\t" + string(ud))
		for i := 0; i < len(tagsIgnore); i++ {
			if match, _ := regexp.MatchString(".+"+tagsIgnore[i]+".+", string(ud)); match {
				ignore = true
				log.Println("INFO:\tThis notification type is marked as ignore.")
			}
		}
		if !ignore {
			for i := 0; i < len(tagsProc); i++ {
				if insertOperation(tagsProc[i], string(ud)) {
					break
				}
				if i == len(tagsProc)-1 {
					log.Println("WARNING:\tNo match for this notificaton.")
				}
			}
		}
	}
	return true
}
func falgsProcessing() {
	flag.Parse()

	if *flagStdoutVerbose != defaultStdoutVerbose {
		log.Printf("INFO:\tDefault print output changed to stdout.")
	}
	if !*flagStdoutVerbose {
		f, err := os.OpenFile("/var/log/butlerblaine/oppars.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("ERROR:\tUnable to create log file: %v", err)
		}
		defer f.Close()
		log.SetOutput(f)
	}
	if *flagClientSecret != defaultClientSecret {
		log.Printf("INFO:\tDefault client secret file changed to %s.", *flagClientSecret)
	}
	if *flagDaysToProcess != defaultDaysToProcess {
		log.Printf("INFO:\tDefault days number to process operations changed to %d.", *flagDaysToProcess)
	}
	if *flagDB != defaultDB {
		log.Printf("INFO:\tDefault data base file changed to %s.", *flagDB)
	}
	if *flagServiceMode != defaultServiceMode {
		log.Printf("INFO:\tAccountant run as service.")
	}
}

func main() {
	falgsProcessing()

	db, err := sql.Open("sqlite3", *flagDB)
	if err != nil {
		log.Fatalf("ERROR:\tUnable to open DB: %v", err)
	}
	defer db.Close()

	stmt, _ = db.Prepare("CREATE TABLE IF NOT EXISTS operations (id INTEGER PRIMARY KEY AUTOINCREMENT, type VARCHAR(64), account VARCHAR(32), sum REAL, description VARCHAR(128), date DATETIME, dostupno REAL, UNIQUE(account, sum, date));")
	stmt.Exec()
	stmt, _ = db.Prepare("INSERT INTO operations(id, type, account, sum, description, date, dostupno ) values(?,?,?,?,?,?,?);")

	if *flagServiceMode {
		startService()
	} else { //not server mode
		processOperations()
	}
}
