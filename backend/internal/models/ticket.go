package models

import "time"

type Ticket struct {
	ID          string    `json:"id"`
	Alias       string    `json:"alias"`
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
	AssigneeName  string `json:"assigneeName,omitempty"`
	AssigneeEmail string `json:"assigneeEmail,omitempty"`
}

type Comment struct {
	ID        string    `json:"id"`
	TicketID  string    `json:"ticketId"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"createdAt"`
}
