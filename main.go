package main

import (
	"fmt"
	"net/http"

	"incognito-chat-api/middlewares"
	"incognito-chat-api/routes"
	"incognito-chat-api/types"
	"incognito-chat-api/utils"

	"github.com/gorilla/mux"
)

var chats = make(map[string]types.Chat)

func setupRoutes(router *mux.Router) {
	// register route
	router.HandleFunc("/api/register", func(w http.ResponseWriter, r *http.Request) {
		routes.Register(w, r, chats)
	}).Methods("POST", "OPTIONS")

	// get single chat
	router.HandleFunc("/api/chat/{chatId}", func(w http.ResponseWriter, r *http.Request) {
		routes.GetChat(w, r, chats)
	}).Methods("GET")

	// create chat
	router.HandleFunc("/api/chat", func(w http.ResponseWriter, r *http.Request) {
		routes.CreateChatRoute(w, r, chats)
	}).Methods("POST", "OPTIONS")

	// ws handler
	router.HandleFunc("/ws/{chatId}/{username}", func(w http.ResponseWriter, r *http.Request) {
		routes.WebSocketEndpoint(w, r, chats)
	}).Methods("GET")

	// upload file
	router.HandleFunc("/api/upload/{chatId}", func(w http.ResponseWriter, r *http.Request) {
		routes.UploadFileEndpoint(w, r, chats)
	}).Methods("POST", "OPTIONS")

	// get uploaded file
	router.HandleFunc("/api/file/{chatId}/{filename}", routes.GetFileEndpoint).Methods("GET")
}

func main() {
	port := utils.GetPort()
	router := mux.NewRouter()

	router.Use(middlewares.Cors)
	setupRoutes(router)

	fmt.Println("Starting started on port " + port)
	http.ListenAndServe(port, router)
}
