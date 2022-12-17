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

func SocketResponse(connections []types.Connection, message types.WS_Signal) {
	response, _ := json.Marshal(message)
	for _, connection := range connections {
		if err := connection.Conn.WriteMessage(1, response); err != nil {
			continue
		}
	}
}

func IsValidDate(date string) bool {
	var testDate = "Mon Jan 02 2006 15:04:05 GMT-0700 (Central European Standard Time)"
	_, err := time.Parse(testDate, date)
	return err != nil
}

func GetIndex[sliceType string | int](slice []sliceType, value sliceType) int {
	for index, el := range slice {
		if el == value {
			return index
		}
	}

	return -1
}

type ID interface {
	GetId() string
}

// returns connection index, if not found -1
func GetIndexById[T ID](slice []T, id string) int {
	for index, el := range slice {
		if el.GetId() == id {
			return index
		}
	}

	return -1
}

// removes element from slice by provided index
func RemoveIndexFromSlice[sliceType any](slice []sliceType, index int) []sliceType {
	sliceLen := len(slice)
	slice[index] = slice[sliceLen-1]
	slice = slice[:sliceLen-1]
	return slice
}