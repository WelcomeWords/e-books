package main

import (
	"database/sql"
	"time"
)

type User struct {
	ID        int
	Username  string
	Name      string
	Email     string
	Password  string
	Role      string
	CreatedAt time.Time
}

type Book struct {
	ID             int
	Title          string
	Author         string
	Genre          string
	Stock          int
	Description    string
	CoverImagePath string
	PdfFilePath    string
	ReleaseDate    string
	IsAvailable    bool
}

type Loan struct {
	ID                  int
	UserID              int
	BookID              int
	Book                Book
	LoanDate            time.Time
	ReturnDate          sql.NullTime
	Status              string
	LoanDateFormatted   string
	ReturnDateFormatted string
}
