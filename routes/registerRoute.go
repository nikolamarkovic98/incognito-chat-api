package routes

import (
	"encoding/json"
	"incognito-chat-api/types"
	"incognito-chat-api/utils"
	"net/http"
)

type RegisterInput struct {
	ChatId   string `json:"chatId"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type RegisterOutput struct {
	IsUsernameTaken bool   `json:"isUsernameTaken"`
	Token 			string `json:"token"`
}

func Register(w http.ResponseWriter, r *http.Request, chats map[string]types.Chat) {
	var input RegisterInput
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		utils.SendErrorMessage(w, "Invalid input", http.StatusBadRequest)
		return
	}

	if input.ChatId == "" || input.Username == "" {
		utils.SendErrorMessage(w, "Inalid input", http.StatusBadRequest)
		return
	}

	
	var isUsernameTaken = false
	if chat, exists := chats[input.ChatId]; exists {
		if input.Password != chat.Password {
			utils.SendErrorMessage(w, "Invalid input", http.StatusBadRequest)
			return
		}

		// loop through connections to find match
		for _, chat := range chat.Connections {
			if chat.Username == input.Username {
				isUsernameTaken = true
				break
			}
		}
	} else {
		// chat doesnt exist
		utils.SendErrorMessage(w, "Inalid input", http.StatusNotFound)
		return
	}

	token := utils.GenJWTToken(input.Username, input.ChatId)

	output := RegisterOutput{
		IsUsernameTaken: isUsernameTaken,
		Token: token,
	}

	utils.SendSuccessMessage(w, output)
}
