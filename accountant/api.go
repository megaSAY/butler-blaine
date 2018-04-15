package accountant

//UpdateDB Send: 0. Reply: 0
const UpdateDB = "Accountant.UpdateDB"

//GetBalance Send: int Account number. Reply: *float64 Balance
const GetBalance = "Accountant.GetBalance"

//GetAccountsQuantity Send: 0. Reply: *int Accounts quantity
const GetAccountsQuantity = "Accountant.GetAccountsQuantity"

//GetFullBalance Send int Accounts quantity. Reply: *[]Balance Accounts balance slice
const GetFullBalance = "Accountant.GetFullBalance"

//Balance Account int, Dostupno string
type Balance struct {
	Account  int
	Dostupno string
}

//GetLastUndefinedOperation Send 0. Reply *string
const GetLastUndefinedOperation = "Accountant.GetLastUndefinedOperation"

//DefineOperation Send DefinedOperation. Reply :0
const DefineOperation = "Accountant.DefineOperation"

//CancelOperationDefinition Send DefinedOperation. Reply :0
const CancelOperationDefinition = "Accountant.CancelOperationDefinition"

//DefinedOperation Send.
type DefinedOperation struct {
	Group     string
	Operation string
}

//GetOperationsGroupsNames Send 0. Reply :0
const GetOperationsGroupsNames = "Accountant.GetOperationsGroupsNames"

//DeleteOperationsGroups Send group name. Reply :0
const DeleteOperationsGroup = "Accountant.DeleteOperationsGroup"
