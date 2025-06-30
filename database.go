package main

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql" // El driver de MySQL
)

// InitDB inicializa y devuelve una conexión a la base de datos
func InitDB() (*sql.DB, error) {
	// Data Source Name (DSN) para la conexión a la base de datos.
	// Formato: username:password@tcp(host:port)/dbname
	// Laragon usualmente usa 'root' como usuario y sin contraseña.
	// ¡Asegúrate de que 'ebooks_db' sea el nombre correcto de tu BD!
	dsn := "root:@tcp(127.0.0.1:3306)/ebooks_db?parseTime=true"

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("error al abrir la base de datos: %w", err)
	}

	// Configurar el pool de conexiones
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	// Hacer un ping para verificar que la conexion es exitosa
	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("error al conectar con la base de datos: %w", err)
	}

	fmt.Println("¡Conexión a la base de datos exitosa!")
	return db, nil
}
