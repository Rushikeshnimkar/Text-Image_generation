package controller

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.lazarusAI/models"
)

func GetChatGptResponse(prompt string) (string, error) {
	var (
		ChatGptURL = "https://api.openai.com/v1/chat/completions"
	)

	openAIKey := os.Getenv("OPENAI_KEY")
	if openAIKey == "" {
		return "", errors.New("OPENAI_KEY not found in environment variables")
	}

	// Create the request body struct
	requestData := models.ChatGPTRequest{
		Model: "gpt-4-turbo-preview",
		Messages: []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		}{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	// Marshal request body to JSON
	requestBody, err := json.Marshal(requestData)
	if err != nil {
		return "", err
	}

	// Create HTTP POST request
	req, err := http.NewRequest("POST", ChatGptURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return "", err
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+openAIKey)

	// Create HTTP client
	client := &http.Client{}

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Check response status code
	if resp.StatusCode != http.StatusOK {
		return "", errors.New("non-OK status code received: " + resp.Status)
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Return response body as string
	return string(body), nil
}
func GenerateImage(prompt string) (string, error) {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		return "", err
	}

	// Get OpenAI API key from environment variable
	apiKey := os.Getenv("OPENAI_KEY")
	if apiKey == "" {
		return "", errors.New("OPENAI_KEY not found in environment variables")
	}

	// Create the request data
	requestData := models.ChatGptInput{
		// Model:  "dall-e-3",
		Prompt: prompt,
		N:      1,
		Size:   "1024x1024",
	}

	// Marshal the request data to JSON
	requestBody, err := json.Marshal(requestData)
	if err != nil {
		return "", err
	}

	// Create HTTP POST request to ChatGPT API for image generation
	url := "https://api.openai.com/v1/images/generations"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return "", err
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	// Create HTTP client
	client := &http.Client{}

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	// bd, _ := req.GetBody()
	// fmt.Println(bd.)
	// Check response status code
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("non-OK status code received: %s", resp.Status, resp.StatusCode)
	}

	// Extract the image URL from the response
	var responseData models.AutoGenerated
	// var responseData = make(map[string]interface{})

	err = json.Unmarshal(body, &responseData)
	if err != nil {
		return "", err
	}

	fmt.Printf("%+v\n", &responseData)

	// imageURL, _ := responseData["url"].(string)
	if responseData.Data[0].URL == "" {
		return "", errors.New("image URL not found in response")
	} else {
		return responseData.Data[0].URL, nil
	}

	// return imageURL, nil
}
func TextRouter(c *gin.Context) {
	// Parse the request body
	var requestBody map[string]string
	if err := c.BindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad request"})
		return
	}

	// Get the prompt from the request body
	prompt, ok := requestBody["prompt"]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Prompt not found in request body"})
		return
	}

	// Call GetChatGptResponse function from the utils package
	response, err := GetChatGptResponse(prompt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting ChatGPT response"})
		return
	}

	// Write the response back to the client
	c.JSON(http.StatusOK, gin.H{"response": response})
}

func ImageRouter(c *gin.Context) {
	var requestBody map[string]string
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad request"})
		return
	}

	// Get the prompt from the request body
	prompt, ok := requestBody["prompt"]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Prompt not found in request body"})
		return
	}

	// Call GenerateImage function from the utils package
	imageURL, err := GenerateImage(prompt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating image: " + err.Error()})
		return
	}

	// Write the image URL back to the client
	c.JSON(http.StatusOK, gin.H{"image_url": imageURL})
}
