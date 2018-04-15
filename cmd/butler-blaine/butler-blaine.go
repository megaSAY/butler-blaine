package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net/rpc"
	"os"
	"strconv"
	"strings"

	"github.com/megasay/butler-blaine/accountant"
	"github.com/megasay/checkerr"
)

// Default values
const defaultStdoutVerbose = false
const defaultTelegramToken = "./telegram.token"
const defaultSlackToken = "./slack.token"
const defaultAllowed = "./allowed.list"
const defaultStartTelegramBot = false
const defaultStartSlackBot = false
const defaultStartSignalBot = false
const defaultStartVKBot = false

// Command line argunments parsing
var flagStdoutVerbose = flag.Bool("stdout", defaultStdoutVerbose, "Print to stdout.")
var flagTelegramToken = flag.String("telegram-token", defaultTelegramToken, "Path to telegram token file.")
var flagSlackToken = flag.String("slack-token", defaultSlackToken, "Path to slack token file.")
var flagAllowed = flag.String("allowed", defaultAllowed, "Path to allowed list file.")
var flagStartTelegramBot = flag.Bool("telegram", defaultStartTelegramBot, "Start Telegram bot.")
var flagStartSlackBot = flag.Bool("slack", defaultStartSlackBot, "Start Slack bot.")
var flagStartSignalBot = flag.Bool("signal", defaultStartSignalBot, "Start Signal bot.")
var flagStartVKBot = flag.Bool("vk", defaultStartVKBot, "Start VK bot.")

// ToDo: Записать в базу данных
var cmdBalance = [...]string{"Баланс", "баланс"}
var cmdDefineOperation = [...]string{"Определить операцию", "определить операцию"}
var respDefault = "Эту комманду я не знаю, но могу выполнить следующие:\n\nБаланс - доступные средства на ваших банковских картах;\n"

//Служащие
var staffAccountant *rpc.Client

var allowedIDs []int

func scanAllowed() {
	fileAllowedList, err := os.OpenFile(*flagAllowed, os.O_RDWR, 0666)
	checkerr.Check(err, checkerr.ERROR, "Unable to open allowed list file")
	defer fileAllowedList.Close()

	scanner := bufio.NewScanner(fileAllowedList)

	for scanner.Scan() {
		log.Println(scanner.Text())
		id, _ := strconv.Atoi(scanner.Text())
		allowedIDs = append(allowedIDs, id)
	}
	log.Printf("INFO:\tAllowed users ID %d\n", allowedIDs)

	err = scanner.Err()
	checkerr.Check(err, checkerr.ERROR, "Unable scan allowed list file")
}
func falgsProcessing() {
	flag.Parse()

	if *flagStdoutVerbose != defaultStdoutVerbose {
		log.Printf("INFO:\tDefault print output changed to stdout.")
	}
	if !*flagStdoutVerbose {
		f, err := os.OpenFile("/var/log/butler-blaine/butler-blaine.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		checkerr.Check(err, checkerr.ERROR, "Unable to create log file.")
		defer f.Close()
		log.SetOutput(f)
	}
	if *flagTelegramToken != defaultTelegramToken {
		log.Printf("INFO:\tDefault telegram token file changed to %s.", *flagTelegramToken)
	}
	if *flagSlackToken != defaultSlackToken {
		log.Printf("INFO:\tDefault slack token file changed to %s.", *flagSlackToken)
	}
	if *flagAllowed != defaultAllowed {
		log.Printf("INFO:\tDefault alowed list file changed to %s.", *flagAllowed)
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

func getResponse(request string) string {
	var response string
	splitRequest := strings.Split(request, " ")
	wordsQuantity := len(splitRequest)
	switch strings.TrimSpace(splitRequest[0]) {
	case cmdBalance[0], cmdBalance[1]:
		switch wordsQuantity {
		case 1:
			response = responseBalance(8336)

		case 2:
			if len(splitRequest[1]) != 4 {
				return "Не понимаю команду.\nПосле команды 'Баланс' должен следовать четырех значный номер счета.\nПример: Баланс 3886"
			} else {
				account, err := strconv.ParseInt(splitRequest[1], 10, 16)
				if err != nil {
					return "Не могу разобрать номер счёта.\nНомер должен быть четырехзначным числом.\nПример: Баланс 3886"
				}
				response = responseBalance(int(account))
			}

		default:
			return "Не понимаю команду.\nПосле команды 'Баланс' должен следовать четырех значный номер счета.\nПример: Баланс 3886"
		}

	case "Определить":
		response = responseGetLastUndefinedOperation()

	case "Группы":
		response = responseGetOperationsGroupsNames()

	case "Удалить":
		if wordsQuantity == 3 {
			switch splitRequest[1] {
			case "группу":
				response = "Пока еще не умею. :("
				response = responseDeleteOperationsGroup(splitRequest[2])
			case "определение":
				response = "Пока еще не умею. :("
			default:
				response = "Не понимаю команду.\nПосле команды 'Удалить' надо уточнить, что вы хотите удалить.\nПример: Удалить группу ИМЯ_ГРУППЫ."
			}
		} else {
			response = "Не понимаю команду.\nПосле команды 'Удалить' надо уточнить, что вы хотите удалить.\nПример: Удалить группу ИМЯ_ГРУППЫ."
		}
	case "Команды":
		response = "Я знаю следующие команды:\n\n"
		response += "Баланс, Определить, Группы;"

	default:
		response = "Эту команду я не знаю, но могу выполнить следующие:\n\n"
		response += "Баланс - доступные средства на ваших банковских картах;\n"
		response += "Команды - подробное описание всех команд;\n"
	}
	return response
}

//Запрос баланса
func responseBalance(account int) string {
	var iReply int
	var f32Reply float32

	staffAccountant, err := rpc.Dial("tcp", "127.0.0.1:1200")
	checkerr.Check(err, checkerr.WARNING, "Unable connect to Accountant service.")
	if err != nil {
		return "К сожалению бухгалер не доступен в данный момент, попробуйте повторить запрос позже.\nИли сообщите об этом Александру, он всё поправит. ;)"
	}

	err = staffAccountant.Call(accountant.UpdateDB, 0, &iReply)
	checkerr.Check(err, checkerr.ERROR, "Unable to update accountant DB.")
	err = staffAccountant.Call(accountant.GetBalance, &account, &f32Reply)
	checkerr.Check(err, checkerr.ERROR, "Unable to get balance from accountant.")
	staffAccountant.Close()

	return fmt.Sprintf("Доступные средства счета *%v: %8.2f руб.", account, f32Reply)
}

//Информация о последней не определенной операции
func responseGetLastUndefinedOperation() string {
	var sReply string

	staffAccountant, err := rpc.Dial("tcp", "127.0.0.1:1200")
	checkerr.Check(err, checkerr.WARNING, "Unable connect to Accountant service.")
	if err != nil {
		return "К сожалению бухгалер не доступен в данный момент, попробуйте повторить запрос позже.\nИли сообщите об этом Александру, он всё поправит. ;)"
	}

	err = staffAccountant.Call(accountant.GetLastUndefinedOperation, 0, &sReply)
	checkerr.Check(err, checkerr.ERROR, "Unable to get last undefined operation from accountant.")
	staffAccountant.Close()

	return fmt.Sprintf("Не определенная операция %v.", sReply)
}

//Определение операции
func responseDefineOperation(operation string, group string) string {
	var iReply int

	staffAccountant, err := rpc.Dial("tcp", "127.0.0.1:1200")
	checkerr.Check(err, checkerr.WARNING, "Unable connect to Accountant service.")
	if err != nil {
		return "К сожалению бухгалер не доступен в данный момент, попробуйте повторить запрос позже.\nИли сообщите об этом Александру, он всё поправит. ;)"
	}

	df := accountant.DefinedOperation{}
	df.Group = group
	df.Operation = operation
	err = staffAccountant.Call(accountant.DefineOperation, df, &iReply)
	checkerr.Check(err, checkerr.ERROR, "Unable to define operation accountant.")
	staffAccountant.Close()

	return "_" //fmt.Sprintf("Не определенная операция %v.", sReply)
}

//Список групп
func responseGetOperationsGroupsNames() string {
	var sReply []string

	staffAccountant, err := rpc.Dial("tcp", "127.0.0.1:1200")
	checkerr.Check(err, checkerr.WARNING, "Unable connect to Accountant service.")
	if err != nil {
		return "К сожалению бухгалер не доступен в данный момент, попробуйте повторить запрос позже.\nИли сообщите об этом Александру, он всё поправит. ;)"
	}

	err = staffAccountant.Call(accountant.GetOperationsGroupsNames, 0, &sReply)
	checkerr.Check(err, checkerr.ERROR, "Unable to obtein groups list accountant.")
	staffAccountant.Close()

	response := "Список групп операций:\n"
	for _, groupName := range sReply {
		response += groupName + "\n"
	}

	return response
}

func responseDeleteOperationsGroup(group string) string {
	var iReply int

	staffAccountant, err := rpc.Dial("tcp", "127.0.0.1:1200")
	checkerr.Check(err, checkerr.WARNING, "Unable connect to Accountant service.")
	if err != nil {
		return "К сожалению бухгалер не доступен в данный момент, попробуйте повторить запрос позже.\nИли сообщите об этом Александру, он всё поправит. ;)"
	}

	err = staffAccountant.Call(accountant.DeleteOperationsGroup, group, &iReply)
	checkerr.Check(err, checkerr.ERROR, "Unable to delete group.")

	staffAccountant.Close()

	return "Группа и связаные с ней операции успешно удалены."
}

func main() {
	falgsProcessing()
	scanAllowed()

	//question := make(chan string)
	//answer := make(chan string)

	if *flagStartTelegramBot {
		//	go startTelegramBot(question, answer)
		go startTelegramBot()
	}
	if *flagStartSlackBot {
		go startSlackBot()
	}
	if *flagStartSignalBot {
		go startSignalBot()
	}
	if *flagStartVKBot {
		go startVKBot()
	}

	var input string
	fmt.Scanln(&input)
}
