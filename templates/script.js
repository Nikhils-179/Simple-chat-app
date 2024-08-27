$(function() {
let socket = null;
const msgBox = $("#messageInput");
const messages = $(".messages");

function showWelcomePopup() {
    let username = prompt("Hello! Welcome to ChatWave. Please enter your username to join the chat:");
    if (username) {
        startWebSocket(username);
    } else {
        alert("Username is required to join the chat.");
        showWelcomePopup(); 
    }
}

function startWebSocket(username) {
    const url = `ws://localhost:8080/room?username=${encodeURIComponent(username)}`;
    socket = new WebSocket(url);

    socket.onopen = function() {
        console.log("WebSocket connection established");
    };

    socket.onclose = function() {
        console.log("WebSocket connection closed");
    };

    socket.onmessage = function(e) {
        console.log(e)
        const data = e.data;
        console.log(data)
        if (data.includes("Chat History:")) {
                // Process chat history separately
                const historyMessages = data.split('\n').slice(1); // Remove "Chat History:" line
                const formattedHistory = historyMessages.map(message => message.trim()).filter(Boolean);
                const historyArrayString = `[${formattedHistory.join(' , ')}]`;
                // Create a new message element for the history
                const newMessage = $("<div>").addClass("message history").text(historyArrayString);
                messages.append(newMessage);
        } else {
            const messageClass = data.startsWith(username) ? "sent" : "received";
            const newMessage = $("<div>").addClass("message " + messageClass).text(data);
            messages.append(newMessage);
        }
    };

    socket.onerror = function(e) {
        console.log("WebSocket error: ", e);
    };
}

$("#sendButton").on("click", function() {
    if (!msgBox.val()) return;
    if (!socket || socket.readyState !== WebSocket.OPEN) {
        alert("Error: There is no WebSocket connection");
        return;
    }
    const message = msgBox.val(); 
    socket.send(message);
    msgBox.val("");
});

$("#getHistoryButton").on("click", function() {
    if (!socket || socket.readyState !== WebSocket.OPEN) {
        alert("Error: There is no WebSocket connection");
        return;
    }
    socket.send(JSON.stringify({action: "getHistory"}));
});
    showWelcomePopup();
});
