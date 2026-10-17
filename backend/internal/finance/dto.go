package finance

// ExpenseInput adalah tipe input pengeluaran (dipindah verbatim dari
// service/expense.go, BE-F6).
type ExpenseInput struct {
	Date     string
	Category string
	Note     string
	Amount   int64
	Method   string
	Source   string
}
