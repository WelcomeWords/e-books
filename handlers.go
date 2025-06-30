package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type MyLoansPageData struct {
	UserName       string
	IsAdmin        bool
	Loans          []Loan
	SuccessMessage string
	ErrorMessage   string
}

// AdminDashboardData se utiliza para pasar datos específicos a la plantilla admin_dashboard.html
type AdminDashboardData struct {
	UserName       string
	IsAdmin        bool
	UserCount      int
	BookCount      int
	LoanCount      int
	Users          []User // Usa la struct User de models.go
	Books          []Book // Usa la struct Book de models.go
	SuccessMessage string
	SearchQuery    string
	ErrorMessage   string
}

// FormPageData se utiliza para pasar datos específicos a los formularios de admin
type FormPageData struct {
	UserName   string
	IsAdmin    bool
	Book       Book // Usa la struct Book de models.go
	User       User // Usa la struct User de models.go
	IsUpcoming bool
}

// BookDetailPageData se utiliza para pasar datos específicos a la plantilla book_detail.html
type BookDetailPageData struct {
	UserName    string
	IsAdmin     bool
	Book        Book // Usa la struct Book de models.go
	UserHasLoan bool
}

// --- Handlers de Autenticacion y Rutas Publicas ---

// homeRedirectHandler redirige la ruta raiz a la pagina de login.
func (app *App) homeRedirectHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// loginHandler maneja el proceso de inicio de sesion de usuarios.
func (app *App) loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		fp := filepath.Join("templates", "login.html")
		tmpl, err := template.ParseFiles(fp)
		if err != nil {
			log.Printf("Error al parsear plantilla de login (GET): %v", err)
			http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
			return
		}
		if err := tmpl.Execute(w, nil); err != nil {
			log.Printf("Error al ejecutar plantilla de login (GET): %v", err)
			http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		}
		return
	}
	if r.Method == http.MethodPost {
		r.ParseForm()
		username := r.FormValue("username")
		password := r.FormValue("password")

		if len(password) != 4 {
			log.Printf("Intento de login fallido para %s: longitud de password incorrecta", username)
			http.Redirect(w, r, "/login?error=true", http.StatusSeeOther)
			return
		}

		var userID int
		var userRole, userName string
		query := "SELECT id, name, role FROM users WHERE username = ?"
		err := app.DB.QueryRow(query, username).Scan(&userID, &userName, &userRole)
		if err != nil {
			if err == sql.ErrNoRows {
				log.Printf("Intento de login fallido para %s: usuario no encontrado", username)
				http.Redirect(w, r, "/login?error=true", http.StatusSeeOther)
			} else {
				log.Printf("Error de DB durante el login para %s: %v", username, err)
				http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
			}
			return
		}

		app.SessionManager.RenewToken(r.Context())
		app.SessionManager.Put(r.Context(), "authenticatedUserID", userID)
		app.SessionManager.Put(r.Context(), "userName", userName)
		app.SessionManager.Put(r.Context(), "userRole", userRole)
		log.Printf("Inicio de sesión exitoso (modo simple) para %s (%s)", userName, userRole)
		http.Redirect(w, r, "/catalog", http.StatusSeeOther)
	}
}

// logoutHandler destruye la sesion y redirige al login.
func (app *App) logoutHandler(w http.ResponseWriter, r *http.Request) {
	app.SessionManager.Destroy(r.Context())
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// --- Middlewares de Seguridad ---

// requireAuthentication es un middleware que asegura que el usuario esté autenticado.
func (app *App) requireAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !app.SessionManager.Exists(r.Context(), "authenticatedUserID") {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// requireAdmin es un middleware que asegura que el usuario tenga rol de administrador.
func (app *App) requireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if app.SessionManager.GetString(r.Context(), "userRole") != "admin" {
			log.Printf("Acceso denegado: Usuario no administrador intentó acceder a %s", r.URL.Path)
			http.Redirect(w, r, "/catalog", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// --- Handlers de Paginas Protegidas ---

// catalogHandler muestra el catálogo de libros disponibles.
func (app *App) catalogHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := app.DB.Query("SELECT id, title, author, cover_image_path, DATE_FORMAT(release_date, '%Y') as release_year FROM books WHERE release_date <= NOW() ORDER BY title")
	if err != nil {
		log.Println(err)
		http.Error(w, "Error de servidor al cargar el catálogo", 500)
		return
	}
	defer rows.Close()

	var books []Book
	for rows.Next() {
		var book Book
		var releaseYear string
		if err := rows.Scan(&book.ID, &book.Title, &book.Author, &book.CoverImagePath, &releaseYear); err != nil {
			log.Println(err)
			http.Error(w, "Error de servidor al escanear libro del catálogo", 500)
			return
		}
		book.ReleaseDate = releaseYear
		book.IsAvailable = true
		books = append(books, book)
	}

	data := struct {
		UserName string
		IsAdmin  bool
		Books    []Book
	}{
		UserName: app.SessionManager.GetString(r.Context(), "userName"),
		IsAdmin:  app.SessionManager.GetString(r.Context(), "userRole") == "admin",
		Books:    books,
	}

	files := []string{"templates/catalog.html", "templates/partials/navbar.html"}
	ts, err := template.ParseFiles(files...)
	if err != nil {
		log.Println(err)
		http.Error(w, "Error de servidor al parsear plantillas de catálogo", 500)
		return
	}
	ts.ExecuteTemplate(w, "catalog.html", data)
}

// upcomingReleasesHandler muestra los libros con lanzamientos futuros.
func (app *App) upcomingReleasesHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := app.DB.Query("SELECT id, title, author, cover_image_path, DATE_FORMAT(release_date, 'Enero de %Y') as release_date FROM books WHERE release_date > NOW() ORDER BY release_date")
	if err != nil {
		log.Println(err)
		http.Error(w, "Error de servidor al cargar próximos lanzamientos", 500)
		return
	}
	defer rows.Close()

	var books []Book // Usa la struct Book de models.go
	for rows.Next() {
		var book Book
		var releaseDateStr string
		if err := rows.Scan(&book.ID, &book.Title, &book.Author, &book.CoverImagePath, &releaseDateStr); err != nil {
			log.Println(err)
			http.Error(w, "Error de servidor al escanear próximo lanzamiento", 500)
			return
		}
		book.ReleaseDate = releaseDateStr
		books = append(books, book)
	}

	data := struct {
		UserName string
		IsAdmin  bool
		Books    []Book
	}{
		UserName: app.SessionManager.GetString(r.Context(), "userName"),
		IsAdmin:  app.SessionManager.GetString(r.Context(), "userRole") == "admin",
		Books:    books,
	}

	files := []string{"templates/upcoming.html", "templates/partials/navbar.html"}
	ts, err := template.ParseFiles(files...)
	if err != nil {
		log.Println(err)
		http.Error(w, "Error de servidor al parsear plantillas de próximos lanzamientos", 500)
		return
	}
	ts.ExecuteTemplate(w, "upcoming.html", data)
}

// bookDetailHandler muestra los detalles de un libro específico.
func (app *App) bookDetailHandler(w http.ResponseWriter, r *http.Request) {
	bookID, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil || bookID < 1 {
		http.NotFound(w, r)
		return
	}

	userID := app.SessionManager.GetInt(r.Context(), "authenticatedUserID")
	var book Book
	var releaseDateTime time.Time

	row := app.DB.QueryRow("SELECT id, title, author, genre, stock, description, cover_image_path, pdf_file_path, release_date FROM books WHERE id = ?", bookID)
	err = row.Scan(&book.ID, &book.Title, &book.Author, &book.Genre, &book.Stock, &book.Description, &book.CoverImagePath, &book.PdfFilePath, &releaseDateTime)
	if err != nil {
		if err == sql.ErrNoRows {
			http.NotFound(w, r)
			return
		}
		log.Println(err)
		http.Error(w, "Error de servidor al cargar detalle del libro", 500)
		return
	}
	book.ReleaseDate = releaseDateTime.Format("2006")

	var loanCount int
	// Verifica si el usuario tiene un prestamo activo para este libro
	app.DB.QueryRow("SELECT COUNT(*) FROM loans WHERE user_id = ? AND book_id = ? AND status = 'active'", userID, bookID).Scan(&loanCount)

	data := BookDetailPageData{
		UserName:    app.SessionManager.GetString(r.Context(), "userName"),
		IsAdmin:     app.SessionManager.GetString(r.Context(), "userRole") == "admin",
		Book:        book,
		UserHasLoan: loanCount > 0,
	}

	files := []string{"templates/book_detail.html", "templates/partials/navbar.html"}
	ts, err := template.ParseFiles(files...)
	if err != nil {
		log.Println(err)
		http.Error(w, "Error de servidor al parsear plantillas de detalle de libro", 500)
		return
	}
	ts.ExecuteTemplate(w, "book_detail.html", data)
}

// createLoanHandler maneja la creación de un nuevo préstamo de libro.
func (app *App) createLoanHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}
	r.ParseForm()
	bookID, err := strconv.Atoi(r.FormValue("book_id"))
	if err != nil {
		http.Error(w, "ID de libro inválido", http.StatusBadRequest)
		return
	}
	userID := app.SessionManager.GetInt(r.Context(), "authenticatedUserID")
	if userID == 0 { // Seguridad extra, aunque el middleware ya debería cubrirlo
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	tx, err := app.DB.Begin()
	if err != nil {
		log.Println(err)
		http.Error(w, "Error de servidor al iniciar transacción", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback() // Asegurarse de hacer rollback si algo falla

	// --- LÓGICA: Verificar si ya existe un préstamo ACTIVO para este usuario y libro ---
	var activeLoanCount int
	err = tx.QueryRow("SELECT COUNT(*) FROM loans WHERE user_id = ? AND book_id = ? AND status = 'active'", userID, bookID).Scan(&activeLoanCount)
	if err != nil {
		log.Println(err)
		app.SessionManager.Put(r.Context(), "flashError", "Error de servidor al verificar préstamos existentes.")
		http.Error(w, "Error de servidor al verificar préstamos existentes", http.StatusInternalServerError)
		return
	}
	if activeLoanCount > 0 {
		// Ya existe un préstamo activo para este libro y usuario. Prevenir duplicados.
		log.Printf("Intento de crear préstamo: Usuario %d ya tiene el libro %d activo.", userID, bookID)
		app.SessionManager.Put(r.Context(), "flashError", "Ya tienes este libro prestado.") // Mensaje flash
		http.Redirect(w, r, fmt.Sprintf("/book?id=%d", bookID), http.StatusSeeOther)
		return
	}

	res, err := tx.Exec("UPDATE books SET stock = stock - 1 WHERE id = ? AND stock > 0", bookID)
	if err != nil {
		log.Println(err)
		app.SessionManager.Put(r.Context(), "flashError", "Error al actualizar stock del libro.")
		http.Error(w, "Error de servidor al actualizar stock del libro", http.StatusInternalServerError)
		return
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		log.Println(err)
		app.SessionManager.Put(r.Context(), "flashError", "Error al verificar stock afectado.")
		http.Error(w, "Error al verificar filas afectadas", http.StatusInternalServerError)
		return
	}
	if rowsAffected == 0 {
		app.SessionManager.Put(r.Context(), "flashError", "No hay stock disponible para este libro.")
		http.Redirect(w, r, fmt.Sprintf("/book?id=%d", bookID), http.StatusSeeOther) // Redirigir con error
		return
	}

	_, err = tx.Exec("INSERT INTO loans (user_id, book_id, status, loan_date) VALUES (?, ?, 'active', NOW())", userID, bookID)
	if err != nil {
		log.Println(err)
		app.SessionManager.Put(r.Context(), "flashError", "Error al registrar el préstamo.")
		http.Error(w, "Error de servidor al registrar préstamo", http.StatusInternalServerError)
		return
	}

	err = tx.Commit()
	if err != nil {
		log.Println(err)
		app.SessionManager.Put(r.Context(), "flashError", "Error al finalizar el préstamo.")
		http.Error(w, "Error de servidor al finalizar transacción de préstamo", http.StatusInternalServerError)
		return
	}
	app.SessionManager.Put(r.Context(), "flashSuccess", "¡Libro prestado con éxito!")
	http.Redirect(w, r, fmt.Sprintf("/book?id=%d", bookID), http.StatusSeeOther)
}

func (app *App) returnLoanHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}
	r.ParseForm()
	bookID, err := strconv.Atoi(r.FormValue("book_id"))
	if err != nil {
		http.Error(w, "ID de libro inválido", http.StatusBadRequest)
		return
	}
	userID := app.SessionManager.GetInt(r.Context(), "authenticatedUserID")
	if userID == 0 { // Seguridad extra
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	tx, err := app.DB.Begin()
	if err != nil {
		log.Println(err)
		http.Error(w, "Error de servidor al iniciar transacción", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	var loanID int
	// Selecciona el préstamo más reciente activo para ese user_id y book_id
	err = tx.QueryRow("SELECT id FROM loans WHERE user_id = ? AND book_id = ? AND status = 'active' ORDER BY loan_date DESC LIMIT 1", userID, bookID).Scan(&loanID)
	if err != nil {
		if err == sql.ErrNoRows {
			// Si no se encuentra un préstamo activo, redirigimos con un mensaje de error.
			log.Printf("Intento de devolver libro (ID: %d) para usuario (ID: %d): No se encontró un préstamo activo para devolver.", bookID, userID)
			app.SessionManager.Put(r.Context(), "flashError", "No se encontró un préstamo activo para este libro.")
			http.Redirect(w, r, "/my-loans", http.StatusSeeOther)
			return
		}
		log.Println(err)
		app.SessionManager.Put(r.Context(), "flashError", "Error de servidor al buscar el préstamo.")
		http.Error(w, "Error de servidor al buscar préstamo activo", http.StatusInternalServerError)
		return
	}

	res, err := tx.Exec("UPDATE loans SET status = 'returned', return_date = NOW() WHERE id = ?", loanID)
	if err != nil {

		log.Println("Error ejecutando UPDATE loans:", err)
		app.SessionManager.Put(r.Context(), "flashError", "Error al actualizar el estado del préstamo.")
		http.Error(w, "Error de servidor al actualizar estado del préstamo", http.StatusInternalServerError)
		return
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		log.Println("Error obteniendo RowsAffected de UPDATE loans:", err)
		app.SessionManager.Put(r.Context(), "flashError", "Error al verificar filas afectadas por retorno.")
		http.Error(w, "Error al verificar filas afectadas por retorno", http.StatusInternalServerError)
		return
	}

	log.Printf("DEBUG: UPDATE loans afectó %d filas para loanID %d.", rowsAffected, loanID)

	if rowsAffected > 0 {
		_, err = tx.Exec("UPDATE books SET stock = stock + 1 WHERE id = ?", bookID)
		if err != nil {
			log.Println("Error ejecutando UPDATE books (incrementar stock):", err)
			app.SessionManager.Put(r.Context(), "flashError", "Error al actualizar stock del libro.")
			http.Error(w, "Error de servidor al actualizar stock del libro", http.StatusInternalServerError)
			return
		}
	} else {

		log.Printf("Advertencia: El préstamo (ID: %d) no fue afectado por la actualización de estado (rowsAffected fue 0).", loanID)
		app.SessionManager.Put(r.Context(), "flashError", "El préstamo ya fue devuelto o no es válido.")
		http.Redirect(w, r, "/my-loans", http.StatusSeeOther)
		return
	}

	err = tx.Commit()
	if err != nil {
		log.Println(err)
		app.SessionManager.Put(r.Context(), "flashError", "Error al finalizar transacción de retorno.")
		http.Error(w, "Error de servidor al finalizar transacción de retorno", http.StatusInternalServerError)
		return
	}

	app.SessionManager.Put(r.Context(), "flashSuccess", "¡Libro devuelto con éxito!")
	http.Redirect(w, r, "/my-loans", http.StatusSeeOther)
}

func (app *App) myLoansHandler(w http.ResponseWriter, r *http.Request) {
	userID := app.SessionManager.GetInt(r.Context(), "authenticatedUserID")
	if userID == 0 {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	query := `
        SELECT
            l.id,
            b.id, b.title, b.author, b.cover_image_path, b.pdf_file_path, -- Campos del libro
            l.loan_date,
            l.return_date,
            l.status
        FROM
            loans l
        JOIN
            books b ON l.book_id = b.id
        WHERE
            l.user_id = ?
        ORDER BY
            l.loan_date DESC
    `
	rows, err := app.DB.Query(query, userID)
	if err != nil {
		log.Printf("Error al consultar préstamos del usuario %d: %v", userID, err)
		http.Error(w, "Error de servidor al cargar mis préstamos", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var userLoans []Loan
	for rows.Next() {
		var loan Loan
		var book Book
		var loanDate time.Time
		var returnDate sql.NullTime
		var status string

		err := rows.Scan(
			&loan.ID,
			&book.ID, &book.Title, &book.Author, &book.CoverImagePath, &book.PdfFilePath,
			&loanDate,
			&returnDate,
			&status,
		)
		if err != nil {
			log.Printf("Error al escanear fila de préstamo: %v", err)
			http.Error(w, "Error de servidor al procesar préstamos", http.StatusInternalServerError)
			return
		}

		loan.Book = book
		loan.LoanDate = loanDate
		loan.Status = status

		// Formatea las fechas para la presentación en la plantilla
		loan.LoanDateFormatted = loanDate.Format("02/01/2006") // Formato DD/MM/YYYY
		if returnDate.Valid {
			loan.ReturnDate = returnDate // Asigna el sql.NullTime original
			loan.ReturnDateFormatted = returnDate.Time.Format("02/01/2006")
		} else {
			// Si ReturnDate no es válida (es NULL en DB), se indica como "Pendiente"
			loan.ReturnDateFormatted = "Pendiente"
		}

		userLoans = append(userLoans, loan)
	}

	successMsg := app.SessionManager.PopString(r.Context(), "flashSuccess")
	errorMsg := app.SessionManager.PopString(r.Context(), "flashError")

	// Prepara los datos para la plantilla MyLoansPageData
	data := MyLoansPageData{
		UserName:       app.SessionManager.GetString(r.Context(), "userName"),
		IsAdmin:        app.SessionManager.GetString(r.Context(), "userRole") == "admin",
		Loans:          userLoans,
		SuccessMessage: successMsg,
		ErrorMessage:   errorMsg,
	}

	// Renderiza la plantilla "my_loans.html"
	files := []string{"templates/my_loans.html", "templates/partials/navbar.html"}
	ts, err := template.ParseFiles(files...)
	if err != nil {
		log.Printf("Error al parsear plantillas para my_loans: %v", err)
		http.Error(w, "Error interno del servidor al cargar la página", http.StatusInternalServerError)
		return
	}

	err = ts.ExecuteTemplate(w, "my_loans.html", data)
	if err != nil {
		log.Printf("Error al ejecutar plantilla my_loans: %v", err)
		http.Error(w, "Error interno del servidor al renderizar la página", http.StatusInternalServerError)
	}
}

// adminDashboardHandler muestra el panel de administración con resumen y listas de usuarios/libros.
func (app *App) adminDashboardHandler(w http.ResponseWriter, r *http.Request) {
	searchQuery := r.URL.Query().Get("q")
	data := AdminDashboardData{
		UserName: app.SessionManager.GetString(r.Context(), "userName"), IsAdmin: true, SuccessMessage: r.URL.Query().Get("success"), SearchQuery: searchQuery, ErrorMessage: r.URL.Query().Get("error"),
	}

	// Obtener conteos para el dashboard
	app.DB.QueryRow("SELECT COUNT(*) FROM users").Scan(&data.UserCount)
	app.DB.QueryRow("SELECT COUNT(*) FROM books").Scan(&data.BookCount)
	app.DB.QueryRow("SELECT COUNT(*) FROM loans").Scan(&data.LoanCount)

	// Obtener libros
	var books []Book // Usa la struct Book de models.go
	var bookRows *sql.Rows
	var err error
	baseBookQuery := "SELECT id, title, author, cover_image_path FROM books"
	args := []interface{}{}
	if searchQuery != "" {
		baseBookQuery += " WHERE title LIKE ?"
		args = append(args, "%"+searchQuery+"%")
	}
	baseBookQuery += " ORDER BY id DESC"
	bookRows, err = app.DB.Query(baseBookQuery, args...)
	if err != nil {
		log.Println(err)
		http.Error(w, "Error de servidor al cargar libros en admin dashboard", 500)
		return
	}
	defer bookRows.Close()
	for bookRows.Next() {
		var book Book
		if err := bookRows.Scan(&book.ID, &book.Title, &book.Author, &book.CoverImagePath); err != nil {
			log.Println(err)
			http.Error(w, "Error de servidor al escanear libro en admin dashboard", 500)
			return
		}
		books = append(books, book)
	}
	data.Books = books

	// Obtener usuarios
	userRows, err := app.DB.Query("SELECT id, username, name, email, role FROM users ORDER BY id DESC")
	if err != nil {
		log.Println(err)
		http.Error(w, "Error de servidor al cargar usuarios en admin dashboard", 500)
		return
	}
	defer userRows.Close()
	var users []User
	for userRows.Next() {
		var user User
		if err := userRows.Scan(&user.ID, &user.Username, &user.Name, &user.Email, &user.Role); err != nil {
			log.Println(err)
			http.Error(w, "Error de servidor al escanear usuario en admin dashboard", 500)
			return
		}
		users = append(users, user)
	}
	data.Users = users

	files := []string{"templates/admin_dashboard.html", "templates/partials/navbar.html"}
	ts, err := template.ParseFiles(files...)
	if err != nil {
		log.Println(err)
		http.Error(w, "Error de servidor al parsear plantillas de admin dashboard", 500)
		return
	}
	ts.ExecuteTemplate(w, "admin_dashboard.html", data)
}

// adminBookFormHandler muestra el formulario para crear/editar libros.
func (app *App) adminBookFormHandler(w http.ResponseWriter, r *http.Request) {
	bookID := r.URL.Query().Get("id")
	pageData := FormPageData{
		UserName: app.SessionManager.GetString(r.Context(), "userName"), IsAdmin: true,
	}
	if bookID != "" {
		id, _ := strconv.Atoi(bookID)
		row := app.DB.QueryRow("SELECT id, title, author, description, release_date FROM books WHERE id = ?", id)
		var releaseDateTime time.Time
		err := row.Scan(&pageData.Book.ID, &pageData.Book.Title, &pageData.Book.Author, &pageData.Book.Description, &releaseDateTime)
		if err != nil {
			http.Error(w, "Libro no encontrado", http.StatusNotFound)
			return
		}
		pageData.IsUpcoming = releaseDateTime.After(time.Now())
	}
	files := []string{"templates/admin_book_form.html", "templates/partials/navbar.html"}
	ts, err := template.ParseFiles(files...)
	if err != nil {
		log.Println(err)
		http.Error(w, "Error de servidor al parsear plantillas de formulario de libro", 500)
		return
	}
	ts.ExecuteTemplate(w, "admin_book_form.html", pageData)
}

func (app *App) adminBookSaveHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(32 << 20)
	bookID := r.FormValue("book_id")
	title := r.FormValue("title")
	author := r.FormValue("author")
	description := r.FormValue("description")
	isUpcoming := r.FormValue("is_upcoming")
	var releaseDate time.Time
	if isUpcoming == "on" {
		releaseDate = time.Now().AddDate(0, 1, 0) // Ejemplo: un mes a partir de ahora
	} else {
		releaseDate = time.Now()
	}

	coverPath, err := app.uploadFile(r, "cover_image", "./static/book_covers/")
	if err != nil {
		log.Printf("Error al subir imagen de portada: %v", err)
		http.Error(w, "Error al subir imagen de portada", http.StatusInternalServerError)
		return
	}
	pdfPath, err := app.uploadFile(r, "pdf_file", "./static/book_pdfs/")
	if err != nil {
		log.Printf("Error al subir archivo PDF: %v", err)
		http.Error(w, "Error al subir archivo PDF", http.StatusInternalServerError)
		return
	}

	if bookID == "" || bookID == "0" {
		_, err := app.DB.Exec("INSERT INTO books (title, author, description, release_date, cover_image_path, pdf_file_path) VALUES (?, ?, ?, ?, ?, ?)", title, author, description, releaseDate, coverPath, pdfPath)
		if err != nil {
			log.Printf("Error al insertar libro: %v", err)
			http.Error(w, "Error de servidor al guardar libro", 500)
			return
		}
	} else { // Actualización de libro existente
		tx, err := app.DB.Begin()
		if err != nil {
			log.Printf("Error al iniciar transacción para actualizar libro: %v", err)
			http.Error(w, "Error de servidor", http.StatusInternalServerError)
			return
		}
		defer tx.Rollback()

		// Actualizar rutas de archivos solo si se cargaron nuevos
		if coverPath != "" {
			_, err = tx.Exec("UPDATE books SET cover_image_path = ? WHERE id = ?", coverPath, bookID)
			if err != nil {
				log.Printf("Error al actualizar cover_image_path: %v", err)
				http.Error(w, "Error de servidor", http.StatusInternalServerError)
				return
			}
		}
		if pdfPath != "" {
			_, err = tx.Exec("UPDATE books SET pdf_file_path = ? WHERE id = ?", pdfPath, bookID)
			if err != nil {
				log.Printf("Error al actualizar pdf_file_path: %v", err)
				http.Error(w, "Error de servidor", http.StatusInternalServerError)
				return
			}
		}
		// Actualizar campos de texto
		_, err = tx.Exec("UPDATE books SET title = ?, author = ?, description = ?, release_date = ? WHERE id = ?", title, author, description, releaseDate, bookID)
		if err != nil {
			log.Printf("Error al actualizar campos de texto del libro: %v", err)
			http.Error(w, "Error de servidor", http.StatusInternalServerError)
			return
		}
		tx.Commit()
	}
	http.Redirect(w, r, "/admin/dashboard?success=book_saved", http.StatusSeeOther)
}

// adminBookDeleteHandler elimina un libro.
func (app *App) adminBookDeleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}
	bookID := r.URL.Query().Get("id")
	if bookID == "" {
		http.Error(w, "ID de libro no proporcionado", http.StatusBadRequest)
		return
	}
	var coverPath, pdfPath string
	// Obtener rutas de archivos antes de eliminar el registro de la DB
	err := app.DB.QueryRow("SELECT cover_image_path, pdf_file_path FROM books WHERE id = ?", bookID).Scan(&coverPath, &pdfPath)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("Error al obtener rutas de archivos para eliminar libro: %v", err)
		http.Error(w, "Error de servidor", http.StatusInternalServerError)
		return
	}

	_, err = app.DB.Exec("DELETE FROM books WHERE id = ?", bookID)
	if err != nil {
		log.Printf("Error al eliminar libro de la base de datos: %v", err)
		http.Error(w, "Error al eliminar libro", http.StatusInternalServerError)
		return
	}

	// Eliminar archivos físicos (si existen)
	if coverPath != "" {
		fullPath := filepath.Join("./static/book_covers/", coverPath)
		if _, err := os.Stat(fullPath); err == nil { //
			if err := os.Remove(fullPath); err != nil {
				log.Printf("Advertencia: No se pudo eliminar archivo de portada %s: %v", fullPath, err)
			}
		}
	}
	if pdfPath != "" {
		fullPath := filepath.Join("./static/book_pdfs/", pdfPath)
		if _, err := os.Stat(fullPath); err == nil { //
			if err := os.Remove(fullPath); err != nil {
				log.Printf("Advertencia: No se pudo eliminar archivo PDF %s: %v", fullPath, err)
			}
		}
	}
	http.Redirect(w, r, "/admin/dashboard?success=book_deleted", http.StatusSeeOther)
}

// adminUserFormHandler muestra el formulario para crear/editar usuarios.
func (app *App) adminUserFormHandler(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("id")
	pageData := FormPageData{
		UserName: app.SessionManager.GetString(r.Context(), "userName"), IsAdmin: true,
	}
	if userIDStr != "" {
		id, _ := strconv.Atoi(userIDStr)
		row := app.DB.QueryRow("SELECT id, username, name, email, role FROM users WHERE id = ?", id)
		err := row.Scan(&pageData.User.ID, &pageData.User.Username, &pageData.User.Name, &pageData.User.Email, &pageData.User.Role)
		if err != nil {
			http.Error(w, "Usuario no encontrado", http.StatusNotFound)
			return
		}
	}
	files := []string{"templates/admin_user_form.html", "templates/partials/navbar.html"}
	ts, err := template.ParseFiles(files...)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "Internal Server Error", 500)
		return
	}
	ts.ExecuteTemplate(w, "admin_user_form.html", pageData)
}

// adminUserSaveHandler maneja el guardado (creación o actualización) de un usuario.
func (app *App) adminUserSaveHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}
	r.ParseForm()
	userID := r.FormValue("user_id")
	username := r.FormValue("username")
	name := r.FormValue("name")
	email := r.FormValue("email")
	password := r.FormValue("password")
	role := r.FormValue("role")

	// Validaciones básicas de entrada
	if username == "" || name == "" || email == "" || role == "" {
		http.Redirect(w, r, "/admin/users/new?error=campos_requeridos", http.StatusSeeOther)
		return
	}

	if userID == "" || userID == "0" {
		if password == "" {
			http.Redirect(w, r, "/admin/users/new?error=password_requerida", http.StatusSeeOther)
			return
		}
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("Error al hashear contraseña: %v", err)
			http.Error(w, "Error interno al procesar contraseña", http.StatusInternalServerError)
			return
		}
		_, err = app.DB.Exec("INSERT INTO users (username, name, email, password, role) VALUES (?, ?, ?, ?, ?)", username, name, email, string(hashedPassword), role)
		if err != nil {
			log.Printf("Error al insertar usuario: %v", err)
			http.Error(w, "Error al crear usuario", http.StatusInternalServerError)
			return
		}
	} else { // Es una actualización de usuario
		// Si se proporcionó una nueva contraseña, hashearla y actualizarla
		if password != "" {
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
			if err != nil {
				log.Printf("Error al hashear nueva contraseña: %v", err)
				http.Error(w, "Error interno al procesar nueva contraseña", http.StatusInternalServerError)
				return
			}
			_, err = app.DB.Exec("UPDATE users SET username=?, name=?, email=?, password=?, role=? WHERE id=?", username, name, email, string(hashedPassword), role, userID)
			if err != nil {
				log.Printf("Error al actualizar usuario con contraseña: %v", err)
				http.Error(w, "Error al actualizar usuario", http.StatusInternalServerError)
				return
			}
		} else { // Si no se proporcionó contraseña, actualizar solo los otros campos
			_, err := app.DB.Exec("UPDATE users SET username=?, name=?, email=?, role=? WHERE id=?", username, name, email, role, userID)
			if err != nil {
				log.Printf("Error al actualizar usuario sin contraseña: %v", err)
				http.Error(w, "Error al actualizar usuario", http.StatusInternalServerError)
				return
			}
		}
	}
	http.Redirect(w, r, "/admin/dashboard?success=user_saved", http.StatusSeeOther)
}

// adminUserDeleteHandler elimina un usuario.
func (app *App) adminUserDeleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}
	// Es más seguro obtener el ID del formulario POST si es un delete button
	// o usar r.URL.Query().Get("id") si el delete es via enlace (menos seguro).
	// Por simplicidad, mantendré Query().Get("id") como en tu código original.
	userIDToDelete, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		http.Error(w, "ID de usuario inválido", http.StatusBadRequest)
		return
	}
	currentUserID := app.SessionManager.GetInt(r.Context(), "authenticatedUserID")
	if userIDToDelete == currentUserID {
		http.Redirect(w, r, "/admin/dashboard?error=self_delete", http.StatusSeeOther)
		return
	}
	_, err = app.DB.Exec("DELETE FROM users WHERE id = ?", userIDToDelete)
	if err != nil {
		log.Printf("Error al eliminar usuario: %v", err)
		http.Error(w, "Error al eliminar usuario", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/admin/dashboard?success=user_deleted", http.StatusSeeOther)
}

// uploadFile es una función auxiliar para manejar la subida de archivos (portadas, PDFs).
func (app *App) uploadFile(r *http.Request, inputName, destPath string) (string, error) {
	file, handler, err := r.FormFile(inputName)
	if err != nil {
		if err == http.ErrMissingFile {
			return "", nil // Si el archivo es opcional, no es un error
		}
		return "", fmt.Errorf("error al obtener archivo '%s': %w", inputName, err)
	}
	defer file.Close()

	// Crea el directorio de destino si no existe
	if _, err := os.Stat(destPath); os.IsNotExist(err) {
		err = os.MkdirAll(destPath, 0755) // Permisos de lectura/escritura/ejecución para propietario
		if err != nil {
			return "", fmt.Errorf("error al crear directorio de destino '%s': %w", destPath, err)
		}
	}

	// Genera un nombre de archivo único para evitar conflictos
	fileName := fmt.Sprintf("%d-%s", time.Now().UnixNano(), handler.Filename)
	dst, err := os.Create(filepath.Join(destPath, fileName))
	if err != nil {
		return "", fmt.Errorf("error al crear archivo de destino: %w", err)
	}
	defer dst.Close()

	// Copia el contenido del archivo subido al archivo de destino
	_, err = io.Copy(dst, file)
	if err != nil {
		return "", fmt.Errorf("error al copiar archivo: %w", err)
	}
	return fileName, nil
}
