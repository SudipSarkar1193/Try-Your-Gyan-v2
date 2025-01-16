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
	rows, err := db.Query("SELECT * FROM quizzes")
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
		password TEXT NOT NULL,
		isVarified BOOL DEFAULT false,
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

/*------------------------------------------------------------------------ */

// Create the challenges table
func CreateChallengesTable(db *sql.DB) {
	createTableQuery := `
    CREATE TABLE IF NOT EXISTS challenges (
        id SERIAL PRIMARY KEY,
        quiz_id INT NOT NULL REFERENCES quizzes(id) ON DELETE CASCADE,
        challenger_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
        topic TEXT NOT NULL,
        status VARCHAR(50) DEFAULT 'pending',
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );`
	_, err := db.Exec(createTableQuery)
	if err != nil {
		fmt.Println("Table creation error for 'challenges':", err)
		return
	}
	fmt.Println("Table 'challenges' created or already exists.")
}

// Create the challenge_users table
func CreateChallengeUsersTable(db *sql.DB) {
	createTableQuery := `
    CREATE TABLE IF NOT EXISTS challenge_users (
        id SERIAL PRIMARY KEY,
        challenge_id INT NOT NULL REFERENCES challenges(id) ON DELETE CASCADE,
        user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
        status VARCHAR(50) DEFAULT 'pending',
        score INT DEFAULT NULL,
        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );`
	_, err := db.Exec(createTableQuery)
	if err != nil {
		fmt.Println("Table creation error for 'challenge_users':", err)
		return
	}
	fmt.Println("Table 'challenge_users' created or already exists.")
}

// Create the notifications table
func CreateNotificationsTable(db *sql.DB) {
	createTableQuery := `
    CREATE TABLE IF NOT EXISTS notifications (
        id SERIAL PRIMARY KEY,
        user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
        type VARCHAR(50) NOT NULL,
        payload JSONB NOT NULL,
        status VARCHAR(50) DEFAULT 'unread',
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );`
	_, err := db.Exec(createTableQuery)
	if err != nil {
		fmt.Println("Table creation error for 'notifications':", err)
		return
	}
	fmt.Println("Table 'notifications' created or already exists.")
}

// *******************************************************************************
func CreateOtpTable(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS otp_table (
		id SERIAL PRIMARY KEY,             
		user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE, 
		otp VARCHAR(10) NOT NULL,          
		expires_at TIMESTAMP DEFAULT (NOW() + INTERVAL '3 minutes'),
		created_at TIMESTAMP DEFAULT NOW() 
	);
	`
	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create otp_table: %w", err)
	}
	fmt.Println("Created or already was there")
	return nil
}

/****************************************************************************************************/

func UpdateOtpForUser(db *sql.DB, userID int, newOtp string) error {
	query := `
		UPDATE otp_table
		SET otp = $1, expires_at = NOW() + INTERVAL '3 minutes', created_at = NOW()
		WHERE user_id = $2;
	`
	result, err := db.Exec(query, newOtp, userID)
	if err != nil {
		return fmt.Errorf("failed to update OTP: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to retrieve affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no record found for user_id %d", userID)
	}

	return nil
}

/****************************************************************************************************/

func DeleteOTPbyUserId(db *sql.DB, id int) error {
	// Prepare the SQL statement
	query := "DELETE FROM otp_table WHERE user_id = $1"

	// Execute the query
	result, err := db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete OTP: %w", err)
	}

	// Check the number of rows affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no OTP found for user ID %d", id)
	}

	return nil
}

//******************************************************************************************/

func DeleteQuizById(db *sql.DB, id int) error {
	query := "DELETE FROM quizzes WHERE id = $1"
	result, err := db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete the Quiz: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	fmt.Println("rowsAffected (DeleteQuizById)", rowsAffected)

	if rowsAffected == 0 {
		return fmt.Errorf("no Quiz found for the ID %d", id)
	}

	return nil
}

/*--------------------------------------------------------------------------------------------------*/

func InsertNewUser(db *sql.DB, user *types.User) (int64, error) {
	var query string
	var err error

	// Determine if `IsVarified` is explicitly provided or should default to false
	if user.IsVarified {
		query = "INSERT INTO users (username, email, password, isVarified,profileImg) VALUES ($1, $2, $3, $4,$5) RETURNING id"
		err = db.QueryRow(query, user.Username, user.Email, user.Password, user.IsVarified, user.ProfileImg).Scan(&user.Id)
	} else {
		// Assume false if IsVarified is not explicitly set
		query = "INSERT INTO users (username, email, password, isVarified,profileImg) VALUES ($1, $2, $3, $4,$5) RETURNING id"
		err = db.QueryRow(query, user.Username, user.Email, user.Password, false, user.ProfileImg).Scan(&user.Id)
	}

	if err != nil {
		return -1, err
	}

	fmt.Printf("New user inserted with ID: %d\n", user.Id)
	return user.Id, nil
}

/*--------------------------------------------------------------------------------------*/
func RetrieveUser(db *sql.DB, identifier any) (*types.User, error) {
	var (
		user  types.User
		query string
		err   error
	)

	switch v := identifier.(type) {
	case string:
		query = `SELECT id, username, email, password,isVarified,profileImg,bio FROM users WHERE email = $1 OR username = $2 LIMIT 1`
		err = db.QueryRow(query, v, v).Scan(&user.Id, &user.Username, &user.Email, &user.Password, &user.IsVarified, &user.ProfileImg, &user.Bio)
	case int, int64:
		query = `SELECT id, username, email, password,isVarified,profileImg,bio FROM users WHERE id = $1 LIMIT 1`
		err = db.QueryRow(query, v).Scan(&user.Id, &user.Username, &user.Email, &user.Password, &user.IsVarified, &user.ProfileImg, &user.Bio)
	default:
		return nil, fmt.Errorf("unsupported identifier type: %T", identifier)
	}

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found : %v", err)
		}
		return nil, err
	}

	return &user, nil
}

/*************************************************************************************************/

/****************************************************************************************************/

func UpdateUserById(db *sql.DB, id int64, isVarified bool) error {
	// Update the isVarified field in the database
	query := `UPDATE users SET isVarified = $1 WHERE id = $2`
	_, err := db.Exec(query, isVarified, id)
	if err != nil {
		return fmt.Errorf("error updating user with id %d: %v", id, err)
	}

	return nil
}

//*******************************

func RetrieveOTP(db *sql.DB, userID int) (string, error) {
	var otp string
	query := `SELECT otp FROM otp_table WHERE user_id = $1 AND expires_at > NOW()`
	err := db.QueryRow(query, userID).Scan(&otp)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("no valid OTP found for user ID %d", userID)
		}
		return "", fmt.Errorf("database error: %w", err)
	}
	return otp, nil
}

//********************************************************************/

func UpdateUserProfilePic(db *sql.DB, userId int, url string) error {
	query := "UPDATE users SET profileImg = $1 WHERE id = $2"

	res, err := db.Exec(query, url, userId)
	if err != nil {
		return fmt.Errorf("failed to update profileImg: %w", err)
	}

	rowsAffected, err := res.RowsAffected()

	if err != nil {
		return fmt.Errorf("failed to retrieve affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no user found with id %d ", userId)
	}
	return nil
}

//************************************************************************

func UpdateBio(db *sql.DB, userID int, newBio string) error {
	// SQL query to update the bio field
	query := "UPDATE users SET bio = $1 WHERE id = $2"

	// Execute the query with placeholders for dynamic values
	result, err := db.Exec(query, newBio, userID)
	if err != nil {
		return err
	}

	// Check if any rows were updated
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("no user found with the given ID")
	}

	return nil
}

///**********************************************************************

// UpdateUsername updates the username of a user in the PostgreSQL database
func UpdateUsername(db *sql.DB, userID int, newUsername string) error {
	// SQL query to update the username field
	query := "UPDATE users SET username = $1 WHERE id = $2"

	// Execute the query with placeholders for dynamic values
	result, err := db.Exec(query, newUsername, userID)
	if err != nil {
		return err
	}

	// Check if any rows were updated
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("no user found with the given ID")
	}

	return nil
}

//*******************************************************************/

func UpdatePassword(db *sql.DB, userID int, hashedPassword string) error {
	// SQL query to update the password field
	query := "UPDATE users SET password = $1 WHERE id = $2"

	// Execute the query with placeholders for dynamic values
	result, err := db.Exec(query, hashedPassword, userID)
	if err != nil {
		return err
	}

	// Check if any rows were updated
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("no user found with the given ID")
	}

	return nil
}

// ********************************************
func UpdateUserEmail(db *sql.DB, userId int, newEmail string) error {
	query := "UPDATE users SET email = $1 WHERE id = $2"

	// Execute the query
	result, err := db.Exec(query, newEmail, userId)
	if err != nil {
		return fmt.Errorf("failed to update email: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to retrieve affected rows: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("no user found with id %d", userId)
	}

	return nil
}

//*******************************************************

func UpdateUserBio(db *sql.DB, userId int, newBio string) error {
	// SQL query to update the bio field
	query := "UPDATE users SET bio = $1 WHERE id = $2"

	// Execute the query
	result, err := db.Exec(query, newBio, userId)
	if err != nil {
		return fmt.Errorf("failed to update bio: %w", err)
	}

	// Check rows affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to retrieve affected rows: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("no user found with id %d", userId)
	}

	return nil
}

//*******************************************************/

//******************************************************************/

func InsertNewOTP(db *sql.DB, otp string, user_id int64) (int64, error) {
	query := "INSERT INTO otp_table (user_id,otp) VALUES ($1, $2) RETURNING id"

	// QueryRow allows us to capture the returned id, Exec doesn't
	var id int64
	err := db.QueryRow(query, user_id, otp).Scan(&id)
	if err != nil {
		return -1, nil
	}

	fmt.Printf("New otp inserted with ID: %d\n", id)

	return id, nil
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

// -------------------------------------------
func FetchQuizzesByQuizId(db *sql.DB, quizId int) (*types.Quiz, error) {
	query := `SELECT quiz_name, level, created_at FROM quizzes WHERE id = $1`

	// Use QueryRow for a single result
	row := db.QueryRow(query, quizId)

	// Initialize the quiz variable
	quiz := &types.Quiz{}

	// Scan the result into the quiz structure
	err := row.Scan(&quiz.QuizName, &quiz.Level, &quiz.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			// Return a nil quiz with a meaningful error if no rows are found
			return nil, fmt.Errorf("no quiz found with id %d", quizId)
		}
		return nil, err
	}

	return quiz, nil
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

func UsernameExists(db *sql.DB, username string) (bool, error) {
	var count int
	query := "SELECT COUNT(*) FROM users WHERE username = $1"
	err := db.QueryRow(query, username).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("query error: %w", err)
	}
	return count > 0, nil
}

// -------------------- ---------------------------------

// AddColumnWithDefault adds a new column to the specified table and sets a default value.
func AddColumnWithDefault(db *sql.DB, tableName, columnName, columnType, defaultValue string) error {
	// Step 1: Add the new column to the table without a default value initially
	addColumnQuery := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s;", tableName, columnName, columnType)
	_, err := db.Exec(addColumnQuery)
	if err != nil {
		return fmt.Errorf("failed to add column: %w", err)
	}
	log.Printf("Column %s added to table %s", columnName, tableName)

	// Step 2: If a default value is provided, update existing rows with it
	if defaultValue != "" {
		updateDefaultQuery := fmt.Sprintf("UPDATE %s SET %s = $1 WHERE %s IS NULL;", tableName, columnName, columnName)
		_, err = db.Exec(updateDefaultQuery, defaultValue)
		if err != nil {
			return fmt.Errorf("failed to set default value: %w", err)
		}
		log.Printf("Default value set for column %s in table %s", columnName, tableName)
	}

	return nil
}

// -------------------- ---------------------------------
func SetupProfileImgColumn(db *sql.DB, defaultValue string) error {
	// Step 1: Set default value for future inserts
	setDefaultQuery := fmt.Sprintf("ALTER TABLE users ALTER COLUMN profileImg SET DEFAULT '%s';", defaultValue)
	_, err := db.Exec(setDefaultQuery)
	if err != nil {
		return fmt.Errorf("failed to set default value: %w", err)
	}

	// Step 2: Enforce NOT NULL constraint
	setNotNullQuery := "ALTER TABLE users ALTER COLUMN profileImg SET NOT NULL;"
	_, err = db.Exec(setNotNullQuery)
	if err != nil {
		return fmt.Errorf("failed to set NOT NULL constraint: %w", err)
	}

	// Step 3: Update existing NULL values (one-time setup)
	updateNullValuesQuery := fmt.Sprintf("UPDATE users SET profileImg = '%s' WHERE profileImg IS NULL;", defaultValue)
	_, err = db.Exec(updateNullValuesQuery)
	if err != nil {
		return fmt.Errorf("failed to update NULL values: %w", err)
	}

	fmt.Println("SetupProfileImgColumn success")

	return nil
}

func UserFindByEmailAndUpdateProfileImg(db *sql.DB, email string, newProfileImg string) error {

	updateQuery := "UPDATE users SET profileImg = $1 WHERE email = $2;"
	_, err := db.Exec(updateQuery, newProfileImg, email)
	if err != nil {
		return fmt.Errorf("failed to update profileImg for user with email %s: %w", email, err)
	}

	log.Printf("Profile image updated for user with email: %s", email)
	return nil
}

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
