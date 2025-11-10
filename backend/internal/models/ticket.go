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
	CreatedBy   string    `json:"createdBy"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	Comments    []Comment `json:"comments,omitempty"`

	// --- Optional display fields ---
	// Populated automatically when joining with users table.
	AssigneeName  string `json:"assignee_name,omitempty"`
	AssigneeEmail string `json:"assignee_email,omitempty"`
}

type Comment struct {
	ID        string    `json:"id"`
	TicketID  string    `json:"ticketId"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"createdAt"`
}
