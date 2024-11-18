package database

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"log"

	"github.com/SudipSarkar1193/Try-Your-Gyan-v2.git/internals/types"
	_ "github.com/lib/pq"
)

func ConnectToDatabase(connectionString string) *sql.DB {

	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	fmt.Println("Successfully connected!")
	return db
}

func DisplayData(db *sql.DB) {
	fmt.Println()
	fmt.Println("INSIDE func DisplayData")
	fmt.Println()
	// Replace 'your_table' with the name of your table
	rows, err := db.Query("SELECT * FROM users")
	if err != nil {
		log.Fatal("r =>", err)
	}
	defer rows.Close()

	// Get column names for formatting
	columns, err := rows.Columns()
	if err != nil {
		log.Fatal("c =>", err)
	}

	// Iterate over rows and print each row
	for rows.Next() {
		// Create a slice of `interface{}` to hold each column’s data
		values := make([]interface{}, len(columns))
		for i := range values {
			values[i] = new(interface{})
		}

		// Scan the row into the values slice
		if err := rows.Scan(values...); err != nil {
			log.Fatal(err)
		}

		// Print the row data
		for i, val := range values {
			fmt.Printf("%s: %v ", columns[i], *(val.(*interface{})))
			fmt.Println()
		}
		fmt.Println()
		fmt.Println()
	}

	// Check for errors after loop
	if err := rows.Err(); err != nil {
		log.Fatal(">p>", err)
	}
}

func CreateUserTable(db *sql.DB) {
	// Create the students table if it doesn't exist
	createTableQuery := `
    CREATE TABLE IF NOT EXISTS users (
        id SERIAL PRIMARY KEY,
        username TEXT NOT NULL UNIQUE,
        email TEXT NOT NULL UNIQUE,
		password TEXT NOT NULL
    );`

	// Execute the query to create the table
	_, err := db.Exec(createTableQuery)
	if err != nil {
		fmt.Println("Table creation Error", err)
	}

	fmt.Println("Table 'users' created or already exists.")

}

func CreateQuizzesTable(db *sql.DB) {
	// Create the quizzes table if it doesn't exist
	createTableQuery := `
    CREATE TABLE IF NOT EXISTS quizzes (
        id SERIAL PRIMARY KEY,
        user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
        quiz_name TEXT NOT NULL,
		level TEXT NOT NULL,
        score INT NOT NULL,
		totalQuestions INT NOT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );`

	// Execute the query to create the table
	_, err := db.Exec(createTableQuery)
	if err != nil {
		fmt.Println("Table creation error:", err)
		return
	}

	fmt.Println("Table 'quizzes' created or already exists.")

}

func CreateQuestionsTable(db *sql.DB) {
	// Create the questions table if it doesn't exist
	createTableQuery := `
    CREATE TABLE IF NOT EXISTS questions (
        id SERIAL PRIMARY KEY,
        quiz_id INT NOT NULL REFERENCES quizzes(id) ON DELETE CASCADE,
        serial_number INT NOT NULL,
        question TEXT NOT NULL,
        options JSONB NOT NULL,
        correct_answer TEXT NOT NULL,
        description TEXT,
        user_answer TEXT
    );`

	// Execute the query to create the table
	_, err := db.Exec(createTableQuery)
	if err != nil {
		fmt.Println("Table creation error:", err)
		return
	}

	fmt.Println("Table 'questions' created or already exists.")

}

/*--------------------------------------------------------------------------------------------------*/

func InsertNewUser(db *sql.DB, user *types.User) (int64, error) {
	query := "INSERT INTO users (username,email,password) VALUES ($1, $2, $3) RETURNING id"

	// QueryRow allows us to capture the returned id, Exec doesn't
	err := db.QueryRow(query, user.Username, user.Email, user.Password).Scan(&user.Id)
	if err != nil {
		return -1, nil
	}

	fmt.Printf("New student inserted with ID: %d\n", user.Id)

	return user.Id, nil
}

/*--------------------------------------------------------------------------------------*/
func RetrieveUser(db *sql.DB, identifier string) (*types.User, error) {
	var user types.User
	query := `SELECT id, username, email, password FROM users WHERE email = $1 OR username = $2 LIMIT 1`
	err := db.QueryRow(query, identifier, identifier).Scan(&user.Id, &user.Username, &user.Email, &user.Password)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

/*--------------------------------------------------------------------------------------*/
func GetAllStudents(db *sql.DB) ([]types.User, error) {
	query := "SELECT * FROM users"
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []types.User
	for rows.Next() {
		var user types.User
		if err := rows.Scan(&user.Id, &user.Username, &user.Email); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

/*---------------------------------------------------------------------------------------------*/

func InsertNewQuiz(db *sql.DB, quiz *types.Quiz) error {
	query := `
		INSERT INTO quizzes (quiz_name, user_id,score ,level,totalQuestions)
		VALUES ($1, $2,$3,$4,$5)
		RETURNING id;
	`

	err := db.QueryRow(query, quiz.QuizName, quiz.UserID, quiz.Score, quiz.Level, quiz.TotalQuestions).Scan(&quiz.ID)
	if err != nil {
		return fmt.Errorf("failed to insert new quiz: %w", err)
	}

	return nil
}

/*--------------------------------------------------------------------*/

func InsertNewQuestions(tx *sql.Tx, questions []types.Question) error {
	stmt, err := tx.Prepare(`
		INSERT INTO questions (
			quiz_id, serial_number, question, options, correct_answer, description, user_answer
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %v", err)
	}
	defer stmt.Close()

	for _, question := range questions {

		// Convert Options to JSON format
		optionsJSON, err := json.Marshal(question.Options)
		if err != nil {
			return fmt.Errorf("failed to marshal options to JSON: %v", err)
		}

		// Execute the insert statement with question data
		_, err = stmt.Exec(
			question.QuizID,
			question.SerialNumber,
			question.Question,
			optionsJSON, // JSON-encoded options
			question.CorrectAnswer,
			question.Description,
			question.UserAnswer,
		)
		if err != nil {
			return fmt.Errorf("failed to insert question: %v", err)
		}
	}

	return nil
}

// -------------------- ---------------------------------

func FetchQuizzesByUser(db *sql.DB, userID int) ([]types.Quiz, error) {
	query := `SELECT id, quiz_name, score, level, totalQuestions, created_at FROM quizzes WHERE user_id = $1`
	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var quizzes []types.Quiz
	for rows.Next() {
		var quiz types.Quiz
		if err := rows.Scan(&quiz.ID, &quiz.QuizName, &quiz.Score, &quiz.Level, &quiz.TotalQuestions, &quiz.CreatedAt); err != nil {
			return nil, err
		}
		quizzes = append(quizzes, quiz)
	}
	return quizzes, nil
}

// -------------------- ---------------------------------

// FetchQuestionsByQuiz retrieves all questions for a specific quiz
func FetchQuestionsByQuiz(db *sql.DB, quizID int) ([]types.Question, error) {
	query := `SELECT id, serial_number, question, options, correct_answer, user_answer, description FROM questions WHERE quiz_id = $1`
	rows, err := db.Query(query, quizID)
	if err != nil {
		return nil, fmt.Errorf("error fetching questions: %v", err)
	}
	defer rows.Close()

	var questions []types.Question
	for rows.Next() {
		var question types.Question
		var options []byte // Read options as []byte first

		if err := rows.Scan(
			&question.ID,
			&question.SerialNumber,
			&question.Question,
			&options,
			&question.CorrectAnswer,
			&question.UserAnswer,
			&question.Description,
		); err != nil {
			return nil, err
		}

		// Unmarshal the JSONB options into []string
		if err := json.Unmarshal(options, &question.Options); err != nil {
			return nil, fmt.Errorf("error unmarshalling options: %v", err)
		}

		questions = append(questions, question)
	}
	return questions, nil
}

// -------------------- ---------------------------------
// -------------------- ---------------------------------
// -------------------- ---------------------------------

// func InsertNewQuestion(tx *sql.Tx, question *types.Question) error {

// 	stmt, err := tx.Prepare(`
// 		INSERT INTO questions (quiz_id, serial_number, question, options, correct_answer, description, user_answer)
// 		VALUES ($1, $2, $3, $4, $5, $6, $7)
// 	`)
// 	if err != nil {
// 		return err
// 	}
// 	defer stmt.Close()

// 	// Execute the insert statement with question data
// 	_, err = stmt.Exec(
// 		question.QuizID,
// 		question.SerialNumber,
// 		question.Question,
// 		question.Options,
// 		question.CorrectAnswer,
// 		question.Description,
// 		question.UserAnswer,
// 	)
// 	return err
// }
