package main

import (
	"fmt"
	"log"
	"net/http"
)

func homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	http.ServeFile(w, r, "index.html")
}

func dashboardHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/dashboard" {
		http.NotFound(w, r)
		return
	}

	http.ServeFile(w, r, "dashboard.html")
}

func main() {
	err := setupDatabase()

	if err != nil {
		log.Fatal("Database error: ", err)
	}

	defer db.Close()

	http.Handle(
		"/static/",
		http.StripPrefix(
			"/static/",
			http.FileServer(http.Dir("static")),
		),
	)

	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/dashboard", dashboardHandler)

	fmt.Println("Student Dashboard:")
	fmt.Println("http://localhost:8080/dashboard")

	err = http.ListenAndServe(":8080", nil)

	if err != nil {
		log.Println(err)
	}
}