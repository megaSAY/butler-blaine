package main

import (
	"bufio"
	"container/list"
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/Syfaro/telegram-bot-api"
	_ "github.com/mattn/go-sqlite3"
)

const homeDir = "/opt/butlerblaine/"
const logFullName = "/var/log/butlerblain/tbot.log"

const operationsParser = "oppars"
const operationsParserDataBaseName = "ops.db"
const allowedListName = "allowed.list"
const botTokenName = "tbottoken"

type Operation struct {
	name     string
	dostupno string
	date     string
}

func checkerr(e error) {
	if e != nil {
		log.Fatalf("ERROR:\t%v", e)
	}
}
func userAllowed(allowedIDs []int, fromID int) bool {
	for _, userID := range allowedIDs {
		if userID == fromID {
			return true
		}
	}
	return false
}
func answerReplace(msg string) string {
	db, err := sql.Open("sqlite3", homeDir+operationsParserDataBaseName)
	if err != nil {
		log.Fatalf("ERROR:\tUnable to open DB: %v", err)
	}
	defer db.Close()
	rows, err := db.Query("SELECT search, replace FROM replace;")
	if err != nil {
		log.Fatalf("ERROR:\t%v", err)
	}

	for rows.Next() {
		var search, replace string
		err = rows.Scan(&search, &replace)
		log.Printf(search + replace)
		if err != nil {
			log.Fatalf("ERROR:\t%v", err)
		}
		msg = strings.Replace(msg, search, replace, -1)
	}
	return msg
}
func responseReplace(search string, replace string) string {
	var response string = "Хорошо, в своих следующих соощениях буду заменять '" + search + "' на '" + replace + "'.\n"

	db, err := sql.Open("sqlite3", homeDir+operationsParserDataBaseName)
	if err != nil {
		log.Fatalf("ERROR:\tUnable to open DB: %v", err)
	}
	defer db.Close()
	stmt, err := db.Prepare("CREATE TABLE IF NOT EXISTS replace (id INTEGER PRIMARY KEY AUTOINCREMENT, search VARCHAR(32), replace VARCHAR(32), UNIQUE(search));")
	if err != nil {
		log.Fatalf("ERROR:\tUnable to Prepare SQL Querry: %v", err)
	}
	stmt.Exec()
	stmt, _ = db.Prepare("INSERT INTO replace(id, search, replace) values(?,?,?);")
	_, err = stmt.Exec(nil, search, replace)
	if err != nil {
		log.Println("WARNING:\t" + err.Error())
		return "Ой, что-то пошло не так.\nЯ не смог выполнить вашу команду. :(\nПожалуйста, перешлите это сообщение Админинистратору.\nОписане проблемы: WARNING:\t" + err.Error()
	}
	log.Println("INFO:\tNew operation added successfully")
	return response
}
func responseBalance() string {
	var response string = "Доступные средства:\n"
	cmd := exec.Command(homeDir + operationsParser)
	log.Printf("Running command and waiting for it to finish...")
	err := cmd.Run()
	if err != nil {
		log.Printf("Command finished with error: %v", err)
		//ToDo: вывести в сообщении лог oppars
	}
	db, err := sql.Open("sqlite3", homeDir+operationsParserDataBaseName)
	if err != nil {
		log.Fatalf("ERROR:\tUnable to open DB: %v", err)
	}
	defer db.Close()
	rows, err := db.Query("SELECT account, dostupno, MAX(date) FROM operations group by account;")
	if err != nil {
		log.Fatalf("ERROR:\t%v", err)
	}
	ops := make([]*Operation, 0)
	for rows.Next() {
		op := new(Operation)
		err = rows.Scan(&op.name, &op.dostupno, &op.date)
		if err != nil {
			log.Fatalf("ERROR:\t%v", err)
		}
		ops = append(ops, op)
	}
	for _, op := range ops {
		response += "Карта " + op.name + " доступно " + op.dostupno + " руб.\n"
	}
	return response
}
func main() {
	//ToDo: ввести параметр лог в файл или на экран. По умолчанию на экран.
	//ToDo: сделать man
	/*
		f, err := os.OpenFile(logFullName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		checkerr(err)
		defer f.Close()
		log.SetOutput(f)
	*/
	listMsgs := list.New()

	fileAllowedList, err := os.OpenFile(homeDir+allowedListName, os.O_RDWR, 0666)
	checkerr(err)
	defer fileAllowedList.Close()

	scanner := bufio.NewScanner(fileAllowedList)
	var allowedIDs []int
	for scanner.Scan() {
		fmt.Println(scanner.Text())
		id, _ := strconv.Atoi(scanner.Text())
		allowedIDs = append(allowedIDs, id)
	}
	fmt.Printf("INFO:\tAllowed users ID %d\n", allowedIDs)

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	tbottoken, err := ioutil.ReadFile(homeDir + botTokenName)
	checkerr(err)
	bot, err := tgbotapi.NewBotAPI(string(tbottoken))
	checkerr(err)

	bot.Debug = false

	log.Printf("INFO:\tAuthorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		if userAllowed(allowedIDs, update.Message.From.ID) {
			var msg string

			log.Println("ALLOWED")

			//ToDo: Заменить на регулярные выражения
			splitMsg := strings.Split(update.Message.Text, " ")
			//ToDo: Список для каждого пользователя с временем актуальности
			listMsgs.PushBack(splitMsg[0])
			switch strings.TrimSpace(splitMsg[0]) {
			case "баланс", "Баланс":
				msg = responseBalance()
			case "Заменяй", "заменяй", "Замени", "замени":
				log.Println(splitMsg[1] + splitMsg[3])
				msg = responseReplace(splitMsg[1], splitMsg[3])
			default:
				msg = "Эту комманду я не знаю, но могу выполнить следующие:\n\n"
				msg += "Баланс - доступные средства на ваших банковских картах;\n"
				msg += "Заменяй слово1 на слово2 - после этой комманды в своих ответах я буду заменять слово1 на слово2;"
			}
			msg = answerReplace(msg)
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, msg))
		} else {
			log.Println("DEINED")
		}
	}
}
