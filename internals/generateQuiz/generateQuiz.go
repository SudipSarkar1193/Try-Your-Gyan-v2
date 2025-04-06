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
	return fmt.Sprintf(`Generate a quiz with the following details:

	- **Topic**: "%s"
	- **Number of Questions**: %d
	- **Difficulty Level**: "%s"
	
	### Instructions:
	1. Create an array of quiz questions in the following format:
	[ 
	  { "ok": true }, 
	  [
		{
		  "serial_number": "1",
		  "question": "What is the capital of France?",
		  "options": ["Berlin", "Madrid", "Paris", "Rome"],
		  "correctAnswer": "Paris",
		  "description": "Paris is the capital and most populous city of France, known for its rich history in art, fashion, and culture."
		},
		{
		  "serial_number": "2",
		  "question": "Which element has the atomic number 1?",
		  "options": ["Helium", "Oxygen", "Hydrogen", "Carbon"],
		  "correctAnswer": "Hydrogen",
		  "description": "Hydrogen is the lightest and most abundant element in the universe."
		}
	  ]
	]
	
	### Guidelines:
	- Each question must include a serial number, question, four options, a correct answer, and a description explaining the answer.
	- The content should be accurate, clear, and related to the specified topic and difficulty level.
	- If you cannot generate the requested number of questions, provide as many as possible with the correct format.
	
	### Fallback Response:
	- If the topic is inappropriate or you cannot generate questions, return:
	[ 
	  { "ok": false }, 
	  ["The requested topic is inappropriate or cannot be used to generate quiz questions."]
	]

	Always generate the response as an array containing two elements. The first element should be an object with the key "ok", and the second element should be an array (either of questions in case of success or a single error message in case of failure/fallback). Always adhere to this structure, regardless of whether the generation was successful.

	Now generate the quiz by strictly following the structure.
	`, topic, number, difficulty)
}

func GenerateQuiz(quizRequest *types.QuizRequest) (any, error) {

	fmt.Println()
	fmt.Println("Point 1")

	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("API_KEY")))
	if err != nil {
		return nil, err
	}
	defer client.Close()
	fmt.Println("Point 2")
	model := client.GenerativeModel("gemini-1.5-flash")
	model.GenerationConfig = genai.GenerationConfig{
		ResponseMIMEType: "application/json",
	}
	fmt.Println("Point 3")
	prompt := generatePrompt(quizRequest.Topic, quizRequest.NumQuestions, quizRequest.Difficulty)
	fmt.Println("Point 4")
	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return nil, err
	}
	fmt.Println("Point 5")

	return resp, nil

}
