package routes

import (
	"encoding/json"
	"incognito-chat-api/types"
	"incognito-chat-api/utils"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

func checkOrigin(r *http.Request) bool {
	// checks origin of request - determine whether or not an icoming request from a different domain is allowed to connect
	// if they are not allowed to connect they will be hit with CORS error
	return true
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     checkOrigin,
}

func WebSocketEndpoint(w http.ResponseWriter, r *http.Request, chats map[string]types.Chat) {
	// establishes websocket connection and return pointer to socket or an error
	chatId := mux.Vars(r)["chatId"]
	username := mux.Vars(r)["username"]

	if chat, exists := chats[chatId]; exists {
		socket, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			utils.SendErrorMessage(w, "Error creating socket", http.StatusServiceUnavailable)
			return
		}

		connId := utils.Genuuid()
		conn := types.Connection{
			ID:       connId,
			Username: username,
			Conn:     *socket,
		}

		chat.Connections = append(chat.Connections, conn)

		chats[chatId] = chat

		// creating new goroutine for each ws connection
		go handleSocket(conn, chatId, chats)
	}
}

func handleSocket(connection types.Connection, chatId string, chats map[string]types.Chat) {
	socket := connection.Conn
	isSocketActive := true

	for isSocketActive {
		// listen for incoming messages/events
		messageType, rawMessage, err := socket.ReadMessage()

		// chat has ended
		if _, exists := chats[chatId]; !exists {
			return
		}

		chat := chats[chatId]
		var input types.WS_Signal

		if messageType == -1 {
			// user left chat
			typingIndex := utils.GetIndex(chat.UsersTyping, connection.Username)
			if typingIndex != -1 {
				chat.UsersTyping = utils.RemoveIndexFromSlice(chat.UsersTyping, typingIndex)
			}

			// remove connection and destroy socket
			connectionIndex := utils.GetIndexById(chat.Connections, connection.ID)
			destroyConnection(chats, chatId, connectionIndex)

			// prepare input
			input.EventType = types.TYPING
			input.Message = types.Message{
				SentBy: connection.Username,
			}

			isSocketActive = false
		} else {
			// all okay

			// parse meessage
			err = json.Unmarshal([]byte(rawMessage), &input)
			if err != nil {
				log.Println(err)
				return
			}
	
			var eventType = input.EventType
			var message = input.Message
	
			// process signal based on eventType
			if eventType == types.CREATE {
				message.ID = utils.Genuuid()
				chat.Messages = append(chat.Messages, message)
				input.Message = message
			} else if eventType == types.LIKE {
				for i, loopMessage := range chat.Messages {
					if loopMessage.ID == message.ID {
						chat.Messages[i].Likes = message.Likes
					}
				}
			} else if eventType == types.DELETE {
				messageIndex := utils.GetIndexById(chat.Messages, message.ID)
				chat.Messages = utils.RemoveIndexFromSlice(chat.Messages, messageIndex)
			} else if eventType == types.TYPING {
				if message.Text == "" {
					index := utils.GetIndex(chat.UsersTyping, message.SentBy)
					chat.UsersTyping = utils.RemoveIndexFromSlice(chat.UsersTyping, index)
				} else {
					chat.UsersTyping = append(chat.UsersTyping, message.SentBy)
				}
			}

		}


		// update current chat with new data
		chats[chatId] = chat

		// send message to everyone else in the chat
		utils.SocketResponse(chat.Connections, input)
	}
}

// closes connection and updates chat data
func destroyConnection(chats map[string]types.Chat, chatId string, connIndex int) {
	chat := chats[chatId]
	chat.Connections[connIndex].Conn.Close()
	chat.Connections = utils.RemoveIndexFromSlice(chat.Connections, connIndex)
	chats[chatId] = chat
}
