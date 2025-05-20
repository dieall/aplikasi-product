package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"golang.org/x/crypto/bcrypt"
)

var db *gorm.DB
var err error

// User adalah model untuk tabel users
type User struct {
	gorm.Model
	Username string `json:"username"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

func initDB() {
	// Gunakan environment variables dalam produksi
	dbHost := "localhost"
	dbUser := "root"
	dbPass := ""
	dbName := "golang_db"
	dbPort := "3306"

	// Koneksi ke database
	connectionString := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local",
		dbUser, dbPass, dbHost, dbPort, dbName)

	db, err = gorm.Open("mysql", connectionString)
	if err != nil {
		log.Fatal(err)
	}

	// Migrasi database
	db.AutoMigrate(&User{})
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		username := r.FormValue("username")
		password := r.FormValue("password")

		var user User
		if db.Where("username = ?", username).First(&user).RecordNotFound() {
			http.Error(w, "Username tidak ditemukan", http.StatusUnauthorized)
			return
		}

		err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
		if err != nil {
			http.Error(w, "Password salah", http.StatusUnauthorized)
			return
		}

		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}

	// Tampilkan halaman login
	http.ServeFile(w, r, "templates/login.html")
}

func dashboardHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "templates/dashboard.html")
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		// Di sini bisa tambahkan penghapusan session jika sudah ada
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func main() {
	initDB()
	defer db.Close()

	r := mux.NewRouter()

	// Serve static files
	fs := http.FileServer(http.Dir("static"))
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))

	// Routes
	r.HandleFunc("/", loginHandler)
	r.HandleFunc("/login", loginHandler)
	r.HandleFunc("/dashboard", dashboardHandler)
	r.HandleFunc("/logout", logoutHandler)

	port := "8080"
	fmt.Printf("Server berjalan di port %s...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
