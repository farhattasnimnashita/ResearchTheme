
package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"
)

type Student struct {
	Name, Department, Skills, Interests string
	Year, Semester                      int
}

var students = []Student{
	{"Ayesha Rahman", "CSE", "Python, Machine Learning, SQL", "Artificial Intelligence", 2, 4},
	{"Tanvir Ahmed", "CSE", "Go, JavaScript, HTML, CSS", "Web Development", 2, 4},
	{"Nusrat Jahan", "CSE", "Python, Pandas, Data Analysis", "Data Science", 3, 1},
	{"Rafiul Islam", "EEE", "C, Arduino, Digital Logic", "IoT and Robotics", 2, 4},
	{"Sadia Akter", "CSE", "Java, SQL, Spring Boot", "Database Systems", 2, 3},
	{"Mehedi Hasan", "CSE", "C++, Algorithms, Git", "Competitive Programming", 3, 2},
	{"Tania Sultana", "CSE", "Python, NLP, TensorFlow", "Natural Language Processing", 3, 1},
	{"Fahim Chowdhury", "SWE", "React, Node.js, MongoDB", "Software Engineering", 2, 4},
	{"Maliha Karim", "CSE", "Python, OpenCV, PyTorch", "Computer Vision", 2, 4},
	{"Sakib Hossain", "EEE", "MATLAB, C, Microcontrollers", "Embedded Systems", 3, 2},
	{"Sumaiya Islam", "CSE", "Linux, Python, Cybersecurity", "Network Security", 2, 3},
	{"Adnan Kabir", "CSE", "Go, SQLite, REST API", "Backend Development", 3, 1},
}

func main() {
	if os.Getenv("APP_ENV") == "production" {
		log.Fatal("Demo accounts must not be created in production.")
	}

	password := os.Getenv("DEMO_PASSWORD")
	if len(password) < 12 || len(password) > 72 {
		log.Fatal("Set DEMO_PASSWORD to 12-72 characters.")
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "researchbridge.db"
	}

	if _, err := os.Stat(dbPath); err != nil {
		log.Fatal("Database not found. Start the app once first: ", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	db.Exec("PRAGMA foreign_keys=ON; PRAGMA busy_timeout=5000")

	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	defer tx.Rollback()

	added := 0

	for i, s := range students {
		email := fmt.Sprintf(
			"student%02d@demo.researchbridge.test", i+1,
		)

		hash, err := bcrypt.GenerateFromPassword(
			[]byte(password),
			bcrypt.DefaultCost,
		)
		if err != nil {
			log.Fatal(err)
		}

		result, err := tx.Exec(`
			INSERT INTO users
			(name, email, password_hash, role, verified,
			 created_at, institution, department,
			 academic_year, semester)
			VALUES (?, ?, ?, 'student', 1, ?, ?, ?, ?, ?)
			ON CONFLICT(email) DO NOTHING
		`,
			s.Name,
			email,
			string(hash),
			time.Now().Unix(),
			"Demo University",
			s.Department,
			s.Year,
			s.Semester,
		)

		if err != nil {
			log.Fatal(err)
		}

		n, err := result.RowsAffected()
		if err != nil {
			log.Fatal(err)
		}
		added += int(n)

		var userID int64
		err = tx.QueryRow(`
			SELECT id FROM users
			WHERE email = ? AND role = 'student'
		`, email).Scan(&userID)

		if err != nil {
			log.Fatal(err)
		}

		_, err = tx.Exec(`
			INSERT INTO researcher_profiles
			(user_id, bio, interests, expertise)
			VALUES (?, ?, ?, ?)
			ON CONFLICT(user_id) DO NOTHING
		`,
			userID,
			"Fictional student profile for development and testing.",
			s.Interests,
			s.Skills,
		)
		if err != nil {
			log.Fatal(err)
		}
	}

	if err := tx.Commit(); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Demo profiles added:", added)
	fmt.Println("Total demo profiles:", len(students))
	fmt.Println("Demo email: student01@demo.researchbridge.test")
	fmt.Println("Up to: student12@demo.researchbridge.test")
	fmt.Println("Password: the DEMO_PASSWORD you selected")
	fmt.Println("Existing demo accounts keep their old passwords.")
}
