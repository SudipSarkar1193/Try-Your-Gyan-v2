package types

import "time"

type User struct {
	Id         int64       `json:"id"`
	Username   string      `json:"username" validate:"required"`
	Email      string      `json:"email" validate:"required,email"`
	Password   interface{} `json:"password" validate:"required"`
	IsVarified bool        `json:"isVarified"`
	ProfileImg string      `json:"profileImg"`
	Bio        *string     `json:"bio"`
}

type QuizRequest struct {
	UserID       int64 `json:"user_id"`
	Topic        string `json:"topic"`
	NumQuestions int    `json:"num_questions"`
	Difficulty   string `json:"difficulty"`
}

type Quiz struct {
	ID             int       `json:"id" db:"id"`
	QuizName       string    `json:"quiz_name" validate:"required" db:"quiz_name"`
	Level          string    `json:"level" validate:"required" db:"level"`
	Score          int       `json:"score" db:"score"`
	TotalQuestions int       `json:"totalQuestions" db:"totalQuestions"`
	UserID         int       `json:"user_id" validate:"required" db:"user_id"` // Foreign key to the users table
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}

type Question struct {
	ID            int      `json:"id" db:"id"`
	SerialNumber  int      `json:"serial_number" validate:"required" db:"serial_number"`
	Question      string   `json:"question" validate:"required" db:"question"`
	Options       []string `json:"options" validate:"required,dive,required" db:"options"` // JSONB field in PostgreSQL
	CorrectAnswer string   `json:"correctAnswer" validate:"required" db:"correct_answer"`
	UserAnswer    string   `json:"user_answer"  db:"user_answer"`
	Description   string   `json:"description" db:"description"`
	QuizID        int      `json:"quiz_id" validate:"required" db:"quiz_id"` // Foreign key to the quizzes table
}

type GoogleTokenInfo struct {
	Email         string `json:"email"`
	EmailVerified string `json:"email_verified"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}

type UserUpdateRequest struct {
	Bio               string `json:"bio"`
	CurrentPassword   string `json:"currentPassword"`
	Email             string `json:"email"`
	IsPasswordChanged bool   `json:"isPasswordChanged"`
	IsUsernameChanged bool   `json:"isUsernameChanged"`
	IsBioChanged      bool   `json:"isbioChanged"`
	IsEmailChanged    bool   `json:"isemailChanged"`
	NewPassword       string `json:"newPassword"`
	Username          string `json:"username"`
}
