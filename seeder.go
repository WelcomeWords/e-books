package main

import (
	"log"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// Listas de archivos de imágenes y PDFs
var bookImageFilenames = []string{
	"1984.jpg", "alicia_en_el_pais_de_las_maravillas.jpg", "aura.jpg", "carrie.jpg", "cementerio_de_animales.jpg",
	"cien_anos_de_soledad.jpg", "cien_anos_de_soledad_2.jpg", "cien_anos_de_soledad_3.jpg", "cien_anos_de_soledad_4.jpg",
	"cronica_de_una_muerte_anunciada.jpg", "dona_barbara.jpg", "dracula.jpg", "el_amor_en_los_tiempos_del_colera.jpg",
	"el_codigo_da_vinci.jpg", "el_coleccionista.jpg", "el_coronel_no_tiene_quien_le_escriba.jpg", "el_cuento_de_la_criada.jpg",
	"el_exorcista.jpg", "el_gran_gatsby.jpg", "el_hombre_ilustrado.jpg", "el_hobbit.jpg", "el_juego_de_gerald.jpg",
	"el_lazarillo_de_tormes.jpg", "el_principito.jpg", "el_problema_de_los_tres_cuerpos.jpg", "el_resplandor.jpg",
	"el_senor_de_los_anillos.jpg", "el_viejo_y_el_mar.jpg", "etica_para_amador.jpg", "fahrenheit_451.jpg",
	"ficciones.jpg", "frankestein.jpg", "hamlet.jpg", "historia_de_dos_ciudades.jpg", "it.jpg", "la_bruja_de_portobello.jpg",
	"la_casa_de_los_espiritus.jpg", "la_casa_infernal.jpg", "la_ciudad_y_los_perros.jpg", "la_divina_comedia.jpg",
	"la_insoportable_levedad_del_ser.jpg", "la_metamorfosis.jpg", "la_maldicion_de_hill_house.jpg", "la_sombra_del_viento.jpg",
	"la_sombra_del_viento_2.jpg", "la_sombra_del_viento_3.jpg", "la_semilla_del_diablo.jpg", "los_crimenes_de_oxford.jpg",
	"los_juegos_del_hambre.jpg", "los_miserables.jpg", "los_mitos_de_cthulhu.jpg", "los_pilares_de_la_tierra.jpg",
	"los_viajes_de_gulliver.jpg", "maus.jpg", "moby_dick.jpg", "morgana.jpg", "orgullo_y_prejuicio.jpg", "psicosis.jpg",
	"rebelion_en_la_granja.jpg", "ready_player_one.jpg", "salems_lot.jpg", "sobre_la_libertad.jpg",
	"un_mundo_feliz.jpg", "ulises.jpg", "viaje_al_centro_de_la_tierra.jpg", "el_nombre_del_viento.jpg",
	"donde_viven_los_monstruos.jpg", "cien_anos_de_soledad_5.jpg", "el_amor_en_los_tiempos_del_colera_2.jpg",
	"cien_anos_de_soledad_6.jpg", "cien_anos_de_soledad_7.jpg",
}

var bookPdfFilenames = []string{
	"1984.pdf", "alicia_en_el_pais_de_las_maravillas.pdf", "aura.pdf", "carrie.pdf", "cementerio_de_animales.pdf",
	"cien_anos_de_soledad.pdf", "cien_anos_de_soledad_2.pdf", "cien_anos_de_soledad_3.pdf", "cien_anos_de_soledad_4.pdf",
	"cronica_de_una_muerte_anunciada.pdf", "dona_barbara.pdf", "dracula.pdf", "el_amor_en_los_tiempos_del_colera.pdf",
	"el_codigo_da_vinci.pdf", "el_coleccionista.pdf", "el_coronel_no_tiene_quien_le_escriaba.pdf", "el_cuento_de_la_criada.pdf",
	"el_exorcista.pdf", "el_gran_gatsby.pdf", "el_hombre_ilustrado.pdf", "el_hobbit.pdf", "el_juego_de_gerald.pdf",
	"el_lazarillo_de_tormes.pdf", "el_principito.pdf", "el_problema_de_los_tres_cuerpos.pdf", "el_resplandor.pdf",
	"el_senor_de_los_anillos.pdf", "el_viejo_y_el_mar.pdf", "etica_para_amador.pdf", "fahrenheit_451.pdf",
	"ficciones.pdf", "frankestein.pdf", "hamlet.pdf", "historia_de_dos_ciudades.pdf", "it.pdf", "la_bruja_de_portobello.pdf",
	"la_casa_de_los_espiritus.pdf", "la_casa_infernal.pdf", "la_ciudad_y_los_perros.pdf", "la_divina_comedia.pdf",
	"la_insoportable_levedad_del_ser.pdf", "la_metamorfosis.pdf", "la_maldicion_de_hill_house.pdf", "la_sombra_del_viento.pdf",
	"la_sombra_del_viento_2.pdf", "la_sombra_del_viento_3.pdf", "la_semilla_del_diablo.pdf", "los_crimenes_de_oxford.pdf",
	"los_juegos_del_hambre.pdf", "los_miserables.pdf", "los_mitos_de_cthulhu.pdf", "los_pilares_de_la_tierra.pdf",
	"los_viajes_de_gulliver.pdf", "maus.pdf", "moby_dick.pdf", "morgana.pdf", "orgullo_y_prejuicio.pdf", "psicosis.pdf",
	"rebelion_en_la_granja.pdf", "ready_player_one.pdf", "salems_lot.pdf", "sobre_la_libertad.pdf",
	"un_mundo_feliz.pdf", "ulises.pdf", "viaje_al_centro_de_la_tierra.pdf", "el_nombre_del_viento.pdf",
	"donde_viven_los_monstruos.pdf", "cien_anos_de_soledad_5.pdf", "el_amor_en_los_tiempos_del_colera_2.pdf",
	"cien_anos_de_soledad_6.pdf", "cien_anos_de_soledad_7.pdf",
}

var bookAuthors = map[string]string{
	"1984": "George Orwell", "alicia_en_el_pais_de_las_maravillas": "Lewis Carroll", "aura": "Carlos Fuentes",
	"carrie": "Stephen King", "cementerio_de_animales": "Stephen King", "cien_anos_de_soledad": "Gabriel García Márquez",
	"cien_anos_de_soledad_2": "Gabriel García Márquez", "cien_anos_de_soledad_3": "Gabriel García Márquez",
	"cien_anos_de_soledad_4": "Gabriel García Márquez", "cronica_de_una_muerte_anunciada": "Gabriel García Márquez",
	"dona_barbara": "Rómulo Gallegos", "dracula": "Bram Stoker", "el_amor_en_los_tiempos_del_colera": "Gabriel García Márquez",
	"el_codigo_da_vinci": "Dan Brown", "el_coleccionista": "John Fowles", "el_coronel_no_tiene_quien_le_escriba": "Gabriel García Márquez",
	"el_cuento_de_la_criada": "Margaret Atwood", "el_exorcista": "William Peter Blatty", "el_gran_gatsby": "F. Scott Fitzgerald",
	"el_hombre_ilustrado": "Ray Bradbury", "el_hobbit": "J.R.R. Tolkien", "el_juego_de_gerald": "Stephen King",
	"el_lazarillo_de_tormes": "Anónimo", "el_principito": "Antoine de Saint-Exupéry", "el_problema_de_los_tres_cuerpos": "Liu Cixin",
	"el_resplandor": "Stephen King", "el_senor_de_los_anillos": "J.R.R. Tolkien", "el_viejo_y_el_mar": "Ernest Hemingway",
	"etica_para_amador": "Fernando Savater", "fahrenheit_451": "Ray Bradbury", "ficciones": "Jorge Luis Borges",
	"frankestein": "Mary Shelley", "hamlet": "William Shakespeare", "historia_de_dos_ciudades": "Charles Dickens",
	"it": "Stephen King", "la_bruja_de_portobello": "Paulo Coelho", "la_casa_de_los_espiritus": "Isabel Allende",
	"la_casa_infernal": "Richard Matheson", "la_ciudad_y_los_perros": "Mario Vargas Llosa", "la_divina_comedia": "Dante Alighieri",
	"la_insoportable_levedad_del_ser": "Milan Kundera", "la_metamorfosis": "Franz Kafka", "la_maldicion_de_hill_house": "Shirley Jackson",
	"la_sombra_del_viento": "Carlos Ruiz Zafón", "la_sombra_del_viento_2": "Carlos Ruiz Zafón", "la_sombra_del_viento_3": "Carlos Ruiz Zafón",
	"la_semilla_del_diablo": "Ira Levin", "los_crimenes_de_oxford": "Guillermo Martínez", "los_juegos_del_hambre": "Suzanne Collins",
	"los_miserables": "Victor Hugo", "los_mitos_de_cthulhu": "H.P. Lovecraft", "los_pilares_de_la_tierra": "Ken Follett",
	"los_viajes_de_gulliver": "Jonathan Swift", "maus": "Art Spiegelman", "moby_dick": "Herman Melville",
	"morgana": "Federico Moccia", "orgullo_y_prejuicio": "Jane Austen", "psicosis": "Robert Bloch",
	"rebelion_en_la_granja": "George Orwell", "ready_player_one": "Ernest Cline", "salems_lot": "Stephen King",
	"sobre_la_libertad": "John Stuart Mill", "un_mundo_feliz": "Aldous Huxley", "ulises": "James Joyce",
	"viaje_al_centro_de_la_tierra": "Julio Verne", "el_nombre_del_viento": "Patrick Rothfuss",
	"donde_viven_los_monstruos": "Maurice Sendak",
	"cien_anos_de_soledad_5":    "Gabriel García Márquez", "el_amor_en_los_tiempos_del_colera_2": "Gabriel García Márquez",
	"cien_anos_de_soledad_6": "Gabriel García Márquez", "cien_anos_de_soledad_7": "Gabriel García Márquez",
}

var bookGenres = map[string]string{
	"carrie": "Terror", "cementerio_de_animales": "Terror", "dracula": "Terror", "el_exorcista": "Terror",
	"el_juego_de_gerald": "Terror", "el_resplandor": "Terror", "it": "Terror", "la_casa_infernal": "Terror",
	"la_maldicion_de_hill_house": "Terror", "la_semilla_del_diablo": "Terror", "psicosis": "Terror", "salems_lot": "Terror",
	"dune": "Ciencia Ficción", "el_problema_de_los_tres_cuerpos": "Ciencia Ficción", "fahrenheit_451": "Ciencia Ficción",
	"ready_player_one": "Ciencia Ficción", "un_mundo_feliz": "Ciencia Ficción", "el_nombre_del_viento": "Fantasía",
	"el_hobbit": "Fantasía", "el_senor_de_los_anillos": "Fantasía", "alicia_en_el_pais_de_las_maravillas": "Fantasía",
	"donde_viven_los_monstruos": "Infantil", "el_principito": "Infantil", "maus": "Novela Gráfica",
	"el_codigo_da_vinci": "Misterio", "el_coleccionista": "Suspense", "1984": "Distopía",
	"cien_anos_de_soledad": "Realismo Mágico", "cien_anos_de_soledad_2": "Realismo Mágico",
	"cien_anos_de_soledad_3": "Realismo Mágico", "cien_anos_de_soledad_4": "Realismo Mágico",
	"cien_anos_de_soledad_5": "Realismo Mágico", "cien_anos_de_soledad_6": "Realismo Mágico",
	"cien_anos_de_soledad_7":          "Realismo Mágico",
	"cronica_de_una_muerte_anunciada": "Novela", "dona_barbara": "Novela",
	"el_amor_en_los_tiempos_del_colera": "Novela Romántica", "el_amor_en_los_tiempos_del_colera_2": "Novela Romántica",
	"el_coronel_no_tiene_quien_le_escriba": "Novela", "el_cuento_de_la_criada": "Distopía",
	"el_gran_gatsby": "Clásico", "el_hombre_ilustrado": "Ciencia Ficción", "el_lazarillo_de_tormes": "Clásico",
	"el_viejo_y_el_mar": "Clásico", "etica_para_amador": "Filosofía", "ficciones": "Ficción Corta",
	"frankestein": "Gótico", "hamlet": "Tragedia", "historia_de_dos_ciudades": "Histórica",
	"la_bruja_de_portobello": "Novela", "la_casa_de_los_espiritus": "Realismo Mágico",
	"la_ciudad_y_los_perros": "Novela", "la_divina_comedia": "Épica",
	"la_insoportable_levedad_del_ser": "Filosofía", "la_metamorfosis": "Novela Corta",
	"la_sombra_del_viento": "Misterio", "la_sombra_del_viento_2": "Misterio", "la_sombra_del_viento_3": "Misterio",
	"los_crimenes_de_oxford": "Misterio", "los_juegos_del_hambre": "Distopía", "los_miserables": "Clásico",
	"los_pilares_de_la_tierra": "Histórica", "los_viajes_de_gulliver": "Sátira", "moby_dick": "Aventura",
	"morgana": "Romance", "orgullo_y_prejuicio": "Clásico", "rebelion_en_la_granja": "Distopía",
	"sobre_la_libertad": "Filosofía", "ulises": "Clásico", "viaje_al_centro_de_la_tierra": "Aventura",
}

// seedDatabase orquesta el poblamiento de todas las tablas
func (app *App) seedDatabase() {
	log.Println("Iniciando poblamiento de la base de datos...")
	app.seedUsers()
	app.seedBooks()
	app.seedLoans() // Añadida la función para poblar préstamos
	log.Println("Poblamiento de la base de datos completado.")
}

// seedUsers inserta usuarios de prueba si la tabla está vacía.
func (app *App) seedUsers() {
	var count int
	app.DB.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if count > 0 {
		log.Println("La tabla 'users' ya está poblada. No se realizarán cambios.")
		return
	}

	log.Println("Poblando la base de datos con usuarios de prueba...")
	// Contraseñas hasheadas para los usuarios de ejemplo (usando bcrypt)
	adminPass, _ := hashPassword("admin123")
	userPass, _ := hashPassword("user123")

	_, err := app.DB.Exec(`
		INSERT INTO users (username, name, email, password, role) VALUES
		('admin', 'Administrador', 'admin@example.com', ?, 'admin'),
		('usuario1', 'Usuario Prueba Uno', 'user1@example.com', ?, 'user'),
		('usuario2', 'Usuario Prueba Dos', 'user2@example.com', ?, 'user')
	`, adminPass, userPass, userPass)

	if err != nil {
		log.Fatalf("FATAL: No se pudo insertar usuarios de prueba: %v", err)
	}
	log.Println("¡Poblado de usuarios completado!")
}

func (app *App) seedBooks() {
	var count int
	app.DB.QueryRow("SELECT COUNT(*) FROM books").Scan(&count)
	if count >= len(bookImageFilenames) {
		log.Println("La tabla 'books' ya está poblada. No se realizarán cambios.")
		return
	}

	log.Printf("Poblando la base de datos con %d libros...", len(bookImageFilenames))

	stmt, err := app.DB.Prepare("INSERT INTO books (title, author, genre, stock, description, cover_image_path, pdf_file_path, release_date) VALUES (?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		log.Fatalf("FATAL: No se pudo preparar la sentencia para insertar libros: %v", err)
	}
	defer stmt.Close()

	// Libros que tendrán una fecha de lanzamiento futura
	upcomingBooks := map[string]bool{
		"aura": true, "dracula": true, "frankestein": true, "hamlet": true, "it": true,
		"la_casa_infernal": true, "los_mitos_de_cthulhu": true, "psicosis": true,
		"rebelion_en_la_granja": true, "ulises": true,
	}

	for i, imgFilename := range bookImageFilenames {
		baseName := strings.TrimSuffix(imgFilename, ".jpg")
		title := strings.ReplaceAll(baseName, "_", " ")
		caser := cases.Title(language.Spanish)
		title = caser.String(title)

		author, ok := bookAuthors[baseName]
		if !ok {
			author = "Autor Desconocido"
		}
		genre, ok := bookGenres[baseName]
		if !ok {
			genre = "Clásico"
		}
		pdfFilename := bookPdfFilenames[i]

		var releaseDate time.Time
		if upcomingBooks[baseName] {
			releaseDate = time.Now().AddDate(0, 1, 0)
		} else {
			releaseDate = time.Now().AddDate(0, -1, 0)
		}

		// Insertamos el libro con stock inicial de 20
		_, err := stmt.Exec(title, author, genre, 20, "Descripción de "+title, imgFilename, pdfFilename, releaseDate)
		if err != nil {
			log.Printf("ADVERTENCIA: No se pudo insertar el libro '%s': %v", title, err)
		}
	}
	log.Println("¡Poblado de libros completado!")
}

func (app *App) seedLoans() {
	var count int
	app.DB.QueryRow("SELECT COUNT(*) FROM loans").Scan(&count)
	if count > 0 {
		log.Println("La tabla 'loans' ya está poblada. No se realizarán cambios.")
		return
	}

	log.Println("Poblando la base de datos con préstamos de prueba para el usuario1...")

	var userID int
	err := app.DB.QueryRow("SELECT id FROM users WHERE username = 'usuario1'").Scan(&userID)
	if err != nil {
		log.Printf("ADVERTENCIA: No se pudo encontrar 'usuario1' para poblar préstamos. Saltando préstamos. Error: %v", err)
		return
	}

	var bookID1, bookID2, bookID3 int
	app.DB.QueryRow("SELECT id FROM books WHERE title = '1984'").Scan(&bookID1)
	app.DB.QueryRow("SELECT id FROM books WHERE title = 'El Principito'").Scan(&bookID2)
	app.DB.QueryRow("SELECT id FROM books WHERE title = 'Maus'").Scan(&bookID3)

	// Solo insertar si los IDs de los libros se encontraron
	if bookID1 != 0 {
		// Préstamo activo
		_, err := app.DB.Exec("INSERT INTO loans (user_id, book_id, loan_date, status) VALUES (?, ?, ?, 'active')",
			userID, bookID1, time.Now().AddDate(0, 0, -7)) // Prestado hace 7 días
		if err != nil {
			log.Printf("ADVERTENCIA: No se pudo insertar préstamo para libro ID %d: %v", bookID1, err)
		} else {
			// Decrementar stock del libro prestado
			app.DB.Exec("UPDATE books SET stock = stock - 1 WHERE id = ?", bookID1)
		}
	}

	if bookID2 != 0 {
		// Préstamo devuelto
		_, err := app.DB.Exec("INSERT INTO loans (user_id, book_id, loan_date, return_date, status) VALUES (?, ?, ?, ?, 'returned')",
			userID, bookID2, time.Now().AddDate(0, 0, -30), time.Now().AddDate(0, 0, -15)) // Prestado hace 30, devuelto hace 15
		if err != nil {
			log.Printf("ADVERTENCIA: No se pudo insertar préstamo devuelto para libro ID %d: %v", bookID2, err)
		}
		// No decrementar stock ya que ya fue devuelto
	}

	if bookID3 != 0 {
		// Otro préstamo activo
		_, err := app.DB.Exec("INSERT INTO loans (user_id, book_id, loan_date, status) VALUES (?, ?, ?, 'active')",
			userID, bookID3, time.Now().AddDate(0, 0, -2))
		if err != nil {
			log.Printf("ADVERTENCIA: No se pudo insertar segundo préstamo activo para libro ID %d: %v", bookID3, err)
		} else {
			// Decrementar stock del libro prestado
			app.DB.Exec("UPDATE books SET stock = stock - 1 WHERE id = ?", bookID3)
		}
	}

	log.Println("¡Poblado de préstamos completado!")
}

func hashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}
