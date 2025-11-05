package models

import "time"

type Ticket struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Category    string    `json:"category"`
	Priority    string    `json:"priority"`
	Status      string    `json:"status"`
	Assignee    string    `json:"assignee"`
	Department  string    `json:"department"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	Comments    []Comment `json:"comments,omitempty"`
}

type Comment struct {
	ID        string    `json:"id"`
	TicketID  string    `json:"ticketId"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"createdAt"`
}
