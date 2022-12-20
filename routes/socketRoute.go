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
	
	if chat, exists := chats[chatId]; exists {
		socket, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			utils.SendErrorMessage(w, "Error creating socket", http.StatusServiceUnavailable)
			return
		}
		
		username := mux.Vars(r)["username"]
		conn := types.Connection{
			ID:       utils.Genuuid(),
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

	// first thing when ws connection is established
	// we send token from client and verify it
	// if its okay, we proceed into socket loop,
	// otherwise connection is destroyed which redirects client to home
	var token string
	_, rawMessage, _ := socket.ReadMessage()
	json.Unmarshal([]byte(rawMessage), &token)
	_, err := utils.ParseJWTToken(token, chatId)
	if err != nil {
		// remove connection and destroy socket
		chat := chats[chatId]
		chat.Connections = destroyConnection(chat.Connections, connection)
		chats[chatId] = chat
		return
	}

	listening := true
	for listening {
		// listen for incoming messages/events
		messageType, rawMessage, err := socket.ReadMessage()
		
		if _, exists := chats[chatId]; !exists {
			// chat has ended
			return
		}

		var input types.WS_Signal
		chat := chats[chatId]

		if messageType == -1 {
			// user left chat

			// filter users typing
			typingIndex := utils.GetIndex(chat.UsersTyping, connection.Username)
			if typingIndex != -1 {
				chat.UsersTyping = utils.RemoveIndexFromSlice(chat.UsersTyping, typingIndex)
			}

			// remove connection and destroy socket
			chat.Connections = destroyConnection(chat.Connections, connection)
			input.EventType = types.TYPING
			input.Message.SentBy = connection.Username

			listening = false
		} else {
			// all okay

			// parse meessage
			err = json.Unmarshal([]byte(rawMessage), &input)
			if err != nil {
				log.Println(err)
				continue
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
				if messageIndex != -1 {
					chat.Messages = utils.RemoveIndexFromSlice(chat.Messages, messageIndex)
				}
			} else if eventType == types.TYPING {
				if message.Text == "" {
					// user not typing
					userTypingIndex := utils.GetIndex(chat.UsersTyping, message.SentBy)
					if userTypingIndex != -1 {
						chat.UsersTyping = utils.RemoveIndexFromSlice(chat.UsersTyping, userTypingIndex)
					}
				} else {
					// user typing
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

func destroyConnection(connections []types.Connection, connection types.Connection) []types.Connection {
	connection.Conn.Close()
	connectionIndex := utils.GetIndexById(connections, connection.ID)
	return utils.RemoveIndexFromSlice(connections, connectionIndex)
}