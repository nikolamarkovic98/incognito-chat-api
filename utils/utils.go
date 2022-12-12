package utils

import (
	"encoding/json"
	"fmt"
	"incognito-chat-api/types"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"

	uuid "github.com/nu7hatch/gouuid"
)

func GetPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	return ":" + port
}

func Genuuid() string {
	id, _ := uuid.NewV4()
	return id.String()
}

func SaveFile(chatId string, file multipart.File, fileHeader multipart.FileHeader) (string, error) {
	// Create the uploads folder if it doesn't already exist
	filesPath := fmt.Sprintf("./uploads/%s", chatId)
	err := os.MkdirAll(filesPath, os.ModePerm)
	if err != nil {
		return "Error creating chat directory", err
	}

	// Construct new filename and create file
	filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), filepath.Ext(fileHeader.Filename))
	createdFilePath := fmt.Sprintf("%s/%s", filesPath, filename)
	createdFile, err := os.Create(createdFilePath)
	if err != nil {
		return "Error creating file", err
	}

	// write to newly created file
	if _, err := io.Copy(createdFile, file); err != nil {
		return "Error writing to a file", err
	}

	// close file
	createdFile.Close()

	return filename, nil
}

func SendErrorMessage(w http.ResponseWriter, message any, status int) {
	response, _ := json.Marshal(message)
	w.WriteHeader(status)
	w.Write(response)
}

func SendSuccessMessage(w http.ResponseWriter, message any) {
	response, _ := json.Marshal(message)
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

func SocketResponse(chat types.Chat, message types.WS_Signal) {
	response, _ := json.Marshal(message)
	for _, connection := range chat.Connections {
		if err := connection.Conn.WriteMessage(1, response); err != nil {
			continue
		}
	}
}

func IsValidDate(date string) bool {
	_, err := time.Parse("Mon Jan 02 2006 15:04:05 GMT-0700 (Central European Standard Time)", date)
	return err != nil
}
