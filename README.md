#Sistema de Gestion de E-Books#
##Descripcion General del Proyecto##

Este es un sistema web para la gestion y prestamo de libros electronicos, desarrollado en Go. La aplicacion permite a los usuarios explorar un catalogo de libros, ver próximos lanzamientos, tomar libros prestado y devolverlos. Incluye un modulo de administracion para la gestion completa de usuarios y del inventario de libros.

##Caracteristicas Principales##

Autenticacion de usuarios:Registro e inicio de sesion para usuarios y administradores.
Catalogo de libros: Visualización de todos los libros disponibles.
Próximos Lanzamientos: Sección dedicada a los libros que serán lanzados en el futuro.
##Gestión de Prestamos##
    * Prestamo de libros con actualizacion automatica de stock.
    * Devolución de libros con incremento de stock.
    * Sección "Mis Prestamos" para que cada usuario vea su historial y gestione sus prestamos activos.
##Panel de Administracion##
    * Gestion completa (CRUD) de usuarios y libros.
    * Subida de portadas de libros y archivos PDF.

##Tecnologías Utilizadas##

Lenguaje de Programación: Go
Base de Datos:MySQL / MariaDB Hostname/IP:127.0.0.1  Port:3306
Seguridad:`golang.org/x/crypto/bcrypt`

##Proceso de instalacion y ejecucion##

Pasos para ejecutar el codigo

##Prerrequisitos##
 Go: (https://golang.org/doc/install) 
 Laragon (MySQL/MariaDB): [Descargar e instalar Laragon](https://laragon.org/download/)

##Clonar el repositorio##

Abre tu terminal y clona el repositorio del proyecto. Despues de clonarlo, navega a la carpeta del proyecto:

##Cambia al [URL_DE_TU_PC]
##Navega a la carpeta del proyecto
# Ejemplo si el repositorio crea una carpeta 'e-books':
cd e-books
