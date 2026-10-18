package main

import (
	"fmt"
	"time"
)

type demoStudent struct {
	name       string
	email      string
	university string
	department string
	year       string
	semester   string
	skills     string
	interests  string
	bio        string
}

func seedDemoStudents(db *Database) error {
	students := []demoStudent{
		{
			name:       "Ayesha Rahman",
			email:      "demo.student01@example.invalid",
			university: "Shahjalal University of Science and Technology",
			department: "Computer Science and Engineering",
			year:       "2nd Year",
			semester:   "2nd Semester",
			skills:     "Go, C++, SQL",
			interests:  "Web Development, Database Systems",
			bio:        "Interested in building useful web applications.",
		},
		{
			name:       "Nusrat Jahan",
			email:      "demo.student02@example.invalid",
			university: "University of Dhaka",
			department: "Computer Science and Engineering",
			year:       "3rd Year",
			semester:   "1st Semester",
			skills:     "Python, Machine Learning, Pandas",
			interests:  "Artificial Intelligence, Data Science",
			bio:        "Interested in machine learning and data analysis.",
		},
		{
			name:       "Tanvir Ahmed",
			email:      "demo.student03@example.invalid",
			university: "Bangladesh University of Engineering and Technology",
			department: "Computer Science and Engineering",
			year:       "2nd Year",
			semester:   "1st Semester",
			skills:     "C, C++, Algorithms",
			interests:  "Competitive Programming, Algorithms",
			bio:        "Enjoys solving programming problems.",
		},
		{
			name:       "Sadia Islam",
			email:      "demo.student04@example.invalid",
			university: "University of Rajshahi",
			department: "Information and Communication Engineering",
			year:       "3rd Year",
			semester:   "2nd Semester",
			skills:     "Java, SQL, JavaScript",
			interests:  "Software Engineering, Web Applications",
			bio:        "Interested in software development projects.",
		},
		{
			name:       "Fahim Hossain",
			email:      "demo.student05@example.invalid",
			university: "University of Chittagong",
			department: "Computer Science and Engineering",
			year:       "2nd Year",
			semester:   "2nd Semester",
			skills:     "Python, Flask, HTML, CSS",
			interests:  "Web Development, Cybersecurity",
			bio:        "Interested in secure and accessible websites.",
		},
		{
			name:       "Maliha Akter",
			email:      "demo.student06@example.invalid",
			university: "Chittagong University of Engineering and Technology",
			department: "Computer Science and Engineering",
			year:       "3rd Year",
			semester:   "1st Semester",
			skills:     "C++, Data Structures, Git",
			interests:  "Algorithms, Research Collaboration",
			bio:        "Enjoys working with a collaborative research team.",
		},
		{
			name:       "Sakib Hasan",
			email:      "demo.student07@example.invalid",
			university: "Khulna University of Engineering and Technology",
			department: "Computer Science and Engineering",
			year:       "2nd Year",
			semester:   "1st Semester",
			skills:     "Java, Spring Boot, MySQL",
			interests:  "Backend Development, Distributed Systems",
			bio:        "Interested in backend systems and APIs.",
		},
		{
			name:       "Sumaiya Sultana",
			email:      "demo.student08@example.invalid",
			university: "Rajshahi University of Engineering and Technology",
			department: "Computer Science and Engineering",
			year:       "3rd Year",
			semester:   "2nd Semester",
			skills:     "Python, TensorFlow, NumPy",
			interests:  "Deep Learning, Computer Vision",
			bio:        "Interested in AI-based research projects.",
		},
		{
			name:       "Rafiul Karim",
			email:      "demo.student09@example.invalid",
			university: "Jahangirnagar University",
			department: "Information Technology",
			year:       "2nd Year",
			semester:   "2nd Semester",
			skills:     "JavaScript, React, Node.js",
			interests:  "Full-Stack Development, Human-Computer Interaction",
			bio:        "Interested in building user-friendly applications.",
		},
		{
			name:       "Tasnim Chowdhury",
			email:      "demo.student10@example.invalid",
			university: "Khulna University",
			department: "Computer Science and Engineering",
			year:       "3rd Year",
			semester:   "1st Semester",
			skills:     "R, Python, Data Visualization",
			interests:  "Data Mining, Health Informatics",
			bio:        "Interested in using data to solve real-world problems.",
		},
		{
			name:       "Imran Hossain",
			email:      "demo.student11@example.invalid",
			university: "Jagannath University",
			department: "Computer Science and Engineering",
			year:       "2nd Year",
			semester:   "1st Semester",
			skills:     "C++, Linux, Networking",
			interests:  "Computer Networks, Information Security",
			bio:        "Interested in networking and system security.",
		},
		{
			name:       "Nabila Karim",
			email:      "demo.student12@example.invalid",
			university: "Islamic University, Bangladesh",
			department: "Information and Communication Technology",
			year:       "3rd Year",
			semester:   "2nd Semester",
			skills:     "Python, SQL, Data Analysis",
			interests:  "Natural Language Processing, Data Science",
			bio:        "Interested in language technology and data analysis.",
		},
	}

	for _, s := range students {
		// Create the student if no conflicting record exists.
		_, err := db.Exec(`
			INSERT INTO users
				(name, email, password_hash, role, verified, created_at,
				 institution, department, academic_year, semester)
			VALUES (?, ?, ?, 'student', 1, ?, ?, ?, ?, ?)
			ON CONFLICT DO NOTHING
		`,
			s.name,
			s.email,
			"demo-account-no-login",
			time.Now().Unix(),
			s.university,
			s.department,
			s.year,
			s.semester,
		)
		if err != nil {
			return fmt.Errorf("insert demo student %s: %w", s.email, err)
		}

		// Update existing demo students.
		_, err = db.Exec(`
			UPDATE users
			SET name = ?,
			    institution = ?,
			    department = ?,
			    academic_year = ?,
			    semester = ?
			WHERE email = ? AND role = 'student'
		`,
			s.name,
			s.university,
			s.department,
			s.year,
			s.semester,
			s.email,
		)
		if err != nil {
			return fmt.Errorf("update demo student %s: %w", s.email, err)
		}

		var userID int64

		err = db.QueryRow(`
			SELECT id FROM users
			WHERE email = ?
		`, s.email).Scan(&userID)
		if err != nil {
			return fmt.Errorf("find demo student %s: %w", s.email, err)
		}

		// Add or update the student's research profile.
		_, err = db.Exec(`
			INSERT INTO researcher_profiles
				(user_id, bio, interests, expertise)
			VALUES (?, ?, ?, ?)
			ON CONFLICT(user_id) DO UPDATE SET
				bio = excluded.bio,
				interests = excluded.interests,
				expertise = excluded.expertise
		`,
			userID,
			s.bio,
			s.interests,
			s.skills,
		)
		if err != nil {
			return fmt.Errorf("save research profile %s: %w", s.email, err)
		}
	}

	return nil
}