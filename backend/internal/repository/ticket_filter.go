package repository

type TicketFilter struct {
	Q        string
	Status   string
	Priority string
	Category string
	Assignee string
	Limit    int
	Offset   int
	Sort     string // created_at, updated_at, priority
	Order    string // asc|desc
}
