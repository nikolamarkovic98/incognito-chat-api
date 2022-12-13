package routes

import (
	"fmt"
	"incognito-chat-api/types"
	"incognito-chat-api/utils"
	"net/http"

	"github.com/gorilla/mux"
)

func GetFileEndpoint(w http.ResponseWriter, r *http.Request) {
	// get file name
	fileName := mux.Vars(r)["filename"]
	chatId := mux.Vars(r)["chatId"]

	// create path to file
	filePath := fmt.Sprintf("./uploads/%s/%s", chatId, fileName)

	// send/serve file
	http.ServeFile(w, r, filePath)
}

func UploadFileEndpoint(w http.ResponseWriter, r *http.Request, chats map[string]types.Chat) {
	chatId := mux.Vars(r)["chatId"]
	chat := chats[chatId]

	// parse form data
	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		utils.SendErrorMessage(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// get file from form data
	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// get other data from form data
	sentBy := r.FormValue("sentBy")
	sentAt := r.FormValue("sentAt")

	// save file
	filename, err := utils.SaveFile(chatId, file, *fileHeader)
	if err != nil {
		fmt.Println(err)
		return
	}

	var output types.WS_Signal
	fileURL := fmt.Sprintf("api/file/%s/%s", chatId, filename)

	// prepare message
	message := types.Message{
		ID:     utils.Genuuid(),
		Type:   "img",
		Text:   "",
		File:   fileURL,
		SentBy: sentBy,
		SentAt: sentAt,
		Likes:  []string{},
	}

	// prepare output
	output.EventType = types.CREATE
	output.Message = message

	// update chat data
	chat.Messages = append(chat.Messages, message)
	chats[chatId] = chat

	// sending everyone uploaded file...
	utils.SocketResponse(chat.Connections, output)
}
