package generateQuiz

import (
	"context"
	"fmt"
	
	"os"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"

	//"github.com/SudipSarkar1193/Try-Your-Gyan-v2.git/internals/config"
	"github.com/SudipSarkar1193/Try-Your-Gyan-v2.git/internals/types"
)

func generatePrompt(topic string, number int, difficulty string) string {
	return fmt.Sprintf(`Generate an array of quiz questions on the topic "%s" with %d questions. Each question should have four multiple-choice options, one correct answer, and a description explaining the correct answer. The difficulty level should be "%s".

Example format: 
[
  { "ok": true }, 
  [
    { 
      "serial_number": "1",
      "question": "What is the capital of France?",
      "options": ["Berlin", "Madrid", "Paris", "Rome"],
      "correctAnswer": "Paris",
      "description": "Paris is the capital and most populous city of France. It has been a major center of finance, diplomacy, commerce, fashion, science, and the arts for centuries."
    },
    { 
      "serial_number": "2",
      "question": "Which element has the atomic number 1?",
      "options": ["Helium", "Oxygen", "Hydrogen", "Carbon"],
      "correctAnswer": "Hydrogen",
      "description": "Hydrogen is the lightest and most abundant element in the universe, making up roughly 75 percent of all normal matter."
    }
  ]
]

(ALWAYS FOLLOW THIS FORMAT if you can generate the questions)

Guidelines:
- **Clarity**: If you do not understand the topic or cannot generate questions, return a response in the format: 
[
  { "ok": false }, 
  ["The requested topic is inappropriate or cannot be used to generate quiz questions."]
]. 

(ALWAYS FOLLOW THIS FORMAT if you can not generate the questions for any reasons)




- **Completeness**: Ensure that each question includes all required fields: serial number, question, options, correct answer, and description.
- **Description**: Provide detailed and relevant historical or contextual information regarding the correct answer.
- **Accuracy**: Validate the correctness of each answer. Provide the best available information if there are uncertainties.
- **Fallback Handling**: If you cannot generate the exact number of questions, provide as many high-quality questions as possible and clearly indicate any limitations.
- **Format Compliance**: Adhere strictly to the provided format to ensure consistency and clarity.
- **Inappropriate Content**: If the input topic is unethical, inappropriate (e.g., related to sex, porn, rape, etc.), or if you fail to generate the questions for any reason, return a response in the format: [{ "ok": false }, ["The requested topic is inappropriate or cannot be used to generate quiz questions."]].


Now generate the quiz questions.`, topic, number, difficulty)
}

func GenerateQuiz(quizRequest *types.QuizRequest) (any, error) {

	// if err := config.LoadEnvFile(".env"); err != nil {
	// 	fmt.Println("Error loading Env file",err)
	// }

	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("API_KEY")))
	if err != nil {
		return nil, err
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-1.5-flash")
	model.GenerationConfig = genai.GenerationConfig{
		ResponseMIMEType: "application/json",
	}

	prompt := generatePrompt(quizRequest.Topic, quizRequest.NumQuestions, quizRequest.Difficulty)

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return nil, err
	}

	return resp, nil

}
