package email

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

// Function to send the OTP email using Sendinblue
func SendOTPEmail(to, otp string) error {

	apiKey := os.Getenv("BREVO_API_KEY")

	url := "https://api.brevo.com/v3/smtp/email"

	// Beautiful HTML email content with OTP
	htmlContent := fmt.Sprintf(`
	<!DOCTYPE html>
	<html>
	<head>
	<link rel="preconnect" href="https://fonts.googleapis.com">
<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
<link href="https://fonts.googleapis.com/css2?family=Montserrat:ital,wght@0,100..900;1,100..900&display=swap" rel="stylesheet">
		<style>
			body {
				font-family:'Gill Sans', 'Gill Sans MT', Calibri, 'Trebuchet MS', sans-serif;
    			font-style:italic;
    			background-color: #f4f4f4;
    			color: #333333;
    			margin: 0;
    			padding: 0;
			}
			.container {
				width: 100%%;
				max-width: 600px;
				margin: 20px auto;
				background-color: #ffffff;
				border-radius: 8px;
				box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
				overflow: hidden;
				
			}
			.header {
				font-family: "Montserrat", sans-serif;
 				font-optical-sizing: auto;
  				font-style: normal;
				background-color: #392e4d;
				color: #ffffff;
				padding: 20px;
				text-align: center;
				font-size: 28px;
				font-weight: 500;
			}
			.body {
				padding: 20px;
				text-align: center;
			}
			.otp {
				font-size: 32px;
				font-weight: bold;
				color: #ebe8f0;
				margin: 20px 0;
			}
			.footer {
				margin-top: 20px;
				font-size: 12px;
				color: #888888;
			}
		</style>
	</head>
	<body>
		<div class="container">
			<div class="header">
				Your OTP Code
			</div>
			<div class="body">
				<p>Hi,</p>
				<p>Use the following OTP to verify your email address:</p>
				<p class="otp">%s</p>
				<p class="footer">This OTP is valid for 3 minutes. Please do not share it with anyone.</p>
			</div>
		</div>
	</body>
	</html>
	`, otp)

	// Email payload
	payload := map[string]interface{}{
		"sender":      map[string]string{"name": "Try-your-জ্ঞান", "email": "netajibosethesudip@gmail.com"},
		"to":          []map[string]string{{"email": to}},
		"subject":     "Your OTP Code",
		"htmlContent": htmlContent,
	}

	payloadBytes, _ := json.Marshal(payload)

	// Create request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return err
	}

	// Add headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", apiKey)

	// Execute request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to send email: %s", resp.Status)
	}
	return nil
}
