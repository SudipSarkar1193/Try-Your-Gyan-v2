async function generateQuiz(token, quizData) {
    try {
      const response = await fetch('http://127.0.0.1:5000/api/quiz/new', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`, // Attach the JWT token
        },
        body: JSON.stringify(quizData), // The quiz data payload
      });
  
      // Check if the response was successful
      if (!response.ok) {
        const errorText = await response.text();
        throw new Error(`Error: ${errorText}`);
      }
  
      // Parse the response data
      const jsonRes = await response.json();
      console.log('Quiz generated successfully:');
      let data = null;
        if (jsonRes) data = JSON.parse(jsonRes.data.Candidates[0].Content.Parts);

        if (data) console.log("quizData", data);

      return jsonRes;
  
    } catch (error) {
      console.error('Failed to generate quiz:', error);
    }
  }
  




  async function login() {
    try {
      const response = await fetch('http://127.0.0.1:5000/api/users/login', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',

        },
        body: JSON.stringify({
            identifier:"Sudip",
            password:"Sudip"
        }), // The quiz data payload
      });
  
      // Check if the response was successful
      if (!response.ok) {
        const errorText = await response.text();
        throw new Error(`Error: ${errorText}`);
      }
  
      // Parse the response data
      const result = await response.json();
      console.log('Logged in successfully:', result);
      return result;
  
    } catch (error) {
      console.error('Failed to Login:', error);
    }
  }
  

async function m() {
    const log = await login()


    console.log(log.data.access_token)

  generateQuiz(log.data.access_token,{
    topic :"History",
    num_questions:5,
    difficulty:"Easy"
  })
} 

m()

 