# incognito-chat-api

Server code for Incognito Chat application written in Go with gorilla/mux and gorilla/websocket packages.

On start two goroutines are created, one for http and one for ws connections.

Once ws recives a handshake, a new goroutine is created for that socket/connection which listens for incoming messages.
