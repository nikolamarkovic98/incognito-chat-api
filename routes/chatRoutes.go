package routes

import (
	"encoding/json"
	"fmt"
	"incognito-chat-api/types"
	"incognito-chat-api/utils"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
)

type CreateChatRouteInput struct {
	CreatedAt string `json:"createdAt"`
	Name      string `json:"name"`
	Password  string `json:"password"`
	Duration  int    `json:"duration"`
}

type CreateChatRouteOutput struct {
	ID          string          `json:"id"`
	CreatedAt   string          `json:"createdAt"`
	Name        string          `json:"name"`
	Duration    int             `json:"duration"`
	Messages    []types.Message `json:"messages"`
	UsersTyping []string 	    `json:"usersTyping"`
}

func CreateChatRoute(w http.ResponseWriter, r *http.Request, chats map[string]types.Chat) {	
	var input CreateChatRouteInput
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		utils.SendErrorMessage(w, "Invalid input", http.StatusBadRequest)
		return
	}

	nameLen := len(input.Name)
	if 10 > input.Duration || 60 < input.Duration || 
	   nameLen == 0 || nameLen > 20 {
		utils.SendErrorMessage(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// if utils.IsValidDate(input.CreatedAt) {
	// 	utils.SendErrorMessage(w, "Invalid input", http.StatusBadRequest)
	// 	return
	// }

	chatId := utils.Genuuid()
	messages := []types.Message{}
	usersTyping := []string{}

	chats[chatId] = types.Chat{
		CreatedAt:   input.CreatedAt,
		Duration:    input.Duration,
		Password: 	 input.Password,
		Name:        input.Name,
		Messages:    messages,
		UsersTyping: usersTyping,
	}

	output := CreateChatRouteOutput{
		ID:          chatId,
		CreatedAt:   input.CreatedAt,
		Duration:    input.Duration,
		Name:        input.Name,
		Messages:    messages,
		UsersTyping: usersTyping,
	}

	// start chat timer
	go removeChat(chats, chatId, input.Duration)

	utils.SendSuccessMessage(w, output)
}

// removes chat data after chat time expires
func removeChat(chats map[string]types.Chat, chatId string, duration int) {
	timer := time.NewTimer(time.Duration(duration) * time.Minute)
	<-timer.C

	for _, conn := range chats[chatId].Connections {
		conn.Conn.Close()
	}

	os.RemoveAll(fmt.Sprintf("uploads/%s", chatId))
	delete(chats, chatId)
}

type GetChatOutput struct {
	ID          string          `json:"id"`
	CreatedAt   string          `json:"createdAt"`
	Name        string          `json:"name"`
	Duration    int             `json:"duration"`
	UsersTyping []string        `json:"usersTyping"`
	Messages    []types.Message `json:"messages"`
}

func GetChat(w http.ResponseWriter, r *http.Request, chats map[string]types.Chat) {
	chatId := mux.Vars(r)["chatId"]

	if chat, exists := chats[chatId]; exists {
		jwtToken := utils.GetJWTFromHeader(r, chatId)
		if jwtToken == "" {
			utils.SendErrorMessage(w, "Invalid input", http.StatusBadRequest)
			return
		}

		output := GetChatOutput{
			ID:          chatId,
			CreatedAt:   chat.CreatedAt,
			Name:        chat.Name,
			Duration:    chat.Duration,
			UsersTyping: chat.UsersTyping,
			Messages:    chat.Messages,
		}
		
		utils.SendSuccessMessage(w, output)
	} else {
		utils.SendErrorMessage(w, "Invalid input", http.StatusNotFound)
	}
}
