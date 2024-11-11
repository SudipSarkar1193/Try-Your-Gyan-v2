

const topic = "Iran before Islam"
const value = 10
const difficulty = "Hard"


async function f1(){
    try {
        console.log("-->>", topic, value, difficulty);
    
        const response = await fetch(`http://127.0.0.1:5000/api/quiz/new`, {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            topic,
            num_questions: value,
            difficulty,
          }),
          
        });
    
        if (!response.ok) {
           // Try to get the error as JSON, or fallback to plain text
           let errorMessage;
           const contentType = response.headers.get("content-type");
           console.log("------------------------")
           console.log("response.headers => ",response.headers)
           console.log("------------------------")

           console.log("contentType",contentType)
           console.log("contentType.includes(application/json)",contentType.includes("application/json"))

           if (contentType && contentType.includes("application/json")) {

               //⭐⭐ Bcz if I'm direcly sending a text response from Go backend using http.Error() function , the conten-type will be "text/plain; charset=utf-8" not application/json 

               const errorData = await response.json();
               console.log("errorData =>",errorData)
               console.log("------------------------")
               errorMessage = errorData.errorMessage || errorData.body ||  "Unknown JSON error";
           } 
           else {

               errorMessage = await response.text(); // Read as plain text
               console.log("response.status =>",response.status)
               console.log("------------------------")
               console.log("errorMessage => ",errorMessage)
               console.log("------------------------")
           }

           console.error(`Error: ${errorMessage} (status: ${response.status})`);
           return;
        }
    
        const jsonRes = await response.json();
        
        //console.log("jsonRes",jsonRes.data.Candidates[0].Content.Parts)
        
        let data = null;
        if (jsonRes) data = JSON.parse(jsonRes.data.Candidates[0].Content.Parts);

        if (data) {
       
        console.log("quizData", data);
        
        // Navigate to QuizPage and pass quizData as state
        
      }
      // Navigate to QuizPage with QuizDataProp -- fix it chatGPT
    
      } catch (error) {
        console.log(error)
      }
}

f1()
