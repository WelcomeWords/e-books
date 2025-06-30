package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/alexedwards/scs/v2"
)

type App struct {
	DB             *sql.DB
	SessionManager *scs.SessionManager
}

func main() {
	sessionManager := scs.New()
	sessionManager.Lifetime = 24 * time.Hour
	sessionManager.Cookie.Persist = true
	sessionManager.Cookie.SameSite = http.SameSiteLaxMode
	sessionManager.Cookie.Secure = false

	db, err := InitDB()
	if err != nil {
		log.Fatalf("No se pudo conectar a la base de datos: %v", err)
	}
	defer db.Close()

	app := &App{
		DB:             db,
		SessionManager: sessionManager,
	}

	// app.seedDatabase()

	mux := http.NewServeMux()
	fileServer := http.FileServer(http.Dir("./static/"))
	mux.Handle("/static/", http.StripPrefix("/static/", fileServer))

	// --- Rutas Públicas ---
	mux.HandleFunc("/login", app.loginHandler)
	mux.HandleFunc("/", app.homeRedirectHandler)
	mux.HandleFunc("/logout", app.logoutHandler)

	// --- Rutas Protegidas ---
	mux.Handle("/catalog", app.requireAuthentication(http.HandlerFunc(app.catalogHandler)))
	mux.Handle("/upcoming", app.requireAuthentication(http.HandlerFunc(app.upcomingReleasesHandler)))
	mux.Handle("/book", app.requireAuthentication(http.HandlerFunc(app.bookDetailHandler)))
	mux.Handle("/loan/create", app.requireAuthentication(http.HandlerFunc(app.createLoanHandler)))
	mux.Handle("/loan/return", app.requireAuthentication(http.HandlerFunc(app.returnLoanHandler)))
	mux.Handle("/my-loans", app.requireAuthentication(http.HandlerFunc(app.myLoansHandler)))

	// --- Rutas de Admin ---
	adminRouter := http.NewServeMux()
	adminRouter.HandleFunc("/admin/dashboard", app.adminDashboardHandler)
	adminRouter.HandleFunc("/admin/books/new", app.adminBookFormHandler)
	adminRouter.HandleFunc("/admin/books/save", app.adminBookSaveHandler)
	adminRouter.HandleFunc("/admin/books/delete", app.adminBookDeleteHandler)
	adminRouter.HandleFunc("/admin/users/new", app.adminUserFormHandler)
	adminRouter.HandleFunc("/admin/users/edit", app.adminUserFormHandler)
	adminRouter.HandleFunc("/admin/users/save", app.adminUserSaveHandler)
	adminRouter.HandleFunc("/admin/users/delete", app.adminUserDeleteHandler)
	mux.Handle("/admin/", app.requireAuthentication(app.requireAdmin(adminRouter)))

	port := ":8080"
	fmt.Printf("Servidor escuchando en http://localhost%s\n", port)
	err = http.ListenAndServe(port, app.SessionManager.LoadAndSave(mux))
	if err != nil {
		log.Fatalf("No se pudo iniciar el servidor: %v", err)
	}
}
