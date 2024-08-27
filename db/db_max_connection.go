package db

import "net/http"

func LoadTestHandler(w http.ResponseWriter, r *http.Request) {
	// Simulate chat message storage
	err := StoreMessage("room1", "Test message")
	if err != nil {
		http.Error(w, "Failed to store message", http.StatusInternalServerError)
		return
	}

	// Simulate chat message retrieval
	_, err = GetChatHistory("room1")
	if err != nil {
		http.Error(w, "Failed to retrieve chat history", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Load test successful"))
}


//command 
//hey -n 10 -c 10 http://localhost:8080/load-test