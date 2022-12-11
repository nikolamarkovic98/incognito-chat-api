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
			utils.SendErrorMessage(w, "Error creating socket")
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

	for {
		// listen for incoming messages/events
		messageType, rawMessage, err := socket.ReadMessage()

		// chat has ended
		if _, exists := chats[chatId]; !exists {
			return
		}

		// user left chat
		if messageType == -1 {
			index := getConnectionIndexByID(chats[chatId], connection.ID)
			destroyConnection(chats, chatId, index)
			return
		}

		chat := chats[chatId]
		var input types.WS_Signal

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
					chat.Messages[i] = message
				}
			}
		}

		// update current chat with new data
		chats[chatId] = chat

		// send message to everyone else in the chat
		utils.SocketResponse(chat, input)
	}
}

// returns connection index, if not found -1
func getConnectionIndexByID(chat types.Chat, socketId string) int {
	for index, chatConnection := range chat.Connections {
		if chatConnection.ID == socketId {
			return index
		}
	}

	return -1
}

// closes connection and updates chat data
func destroyConnection(chats map[string]types.Chat, chatId string, connIndex int) {
	chat := chats[chatId]

	chat.Connections[connIndex].Conn.Close()
	connsLen := len(chat.Connections)
	chat.Connections[connIndex] = chat.Connections[connsLen-1]
	chat.Connections = chat.Connections[:connsLen-1]

	chats[chatId] = chat
}
