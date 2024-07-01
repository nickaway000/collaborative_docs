const urlParams = new URLSearchParams(window.location.search);
const documentID = urlParams.get('doc');
const ws = new WebSocket(`ws://localhost:8081/ws?doc=${documentID}`);

let quill;
let localChange = false;

// Function to initialize Quill editor and set up WebSocket
function initializeEditor() {
    quill = new Quill('#editor-container', {
        theme: 'snow'
    });

    // Event listener for text changes in Quill editor
    quill.on('text-change', (delta, oldDelta, source) => {
        if (localChange) {
            localChange = false; // Reset the flag
            return;
        }
        if (source === 'user') {
            sendMessage(delta);
        }
    });
}

// Function to handle WebSocket messages
function handleWebSocketMessage(event) {
    const message = JSON.parse(event.data);
    console.log("Message received:", message);

    if (message.type === "initial") {
        document.getElementById('title').value = message.title || '';
        quill.setContents(message.content);
    } else if (message.type === "edit" && message.document_id === parseInt(documentID)) {
        localChange = true;  // Prevent broadcast of this change
        quill.updateContents(message.delta);
    }
}

// Function to send edits over WebSocket
function sendMessage(delta) {
    const message = {
        type: "edit",
        document_id: parseInt(documentID),
        delta: delta
    };
    ws.send(JSON.stringify(message));
}

// Function to save document changes via fetch API
function saveDocument() {
    const content = quill.getContents();
    const title = document.getElementById('title').value;

    fetch(`/documents/${documentID}`, {
        method: "PUT",
        headers: {
            "Content-Type": "application/json"
        },
        body: JSON.stringify({ title: title, content: content })
    }).then(response => {
        if (response.ok) {
            console.log("Document saved successfully");
        } else {
            console.error("Failed to save document");
        }
    }).catch(error => {
        console.error("Error saving document:", error);
    });
}

// Initialize editor and WebSocket connection when DOM is ready
document.addEventListener("DOMContentLoaded", () => {
    initializeEditor();

    ws.onopen = () => {
        console.log("WebSocket connection established");
        // Request initial document state
        ws.send(JSON.stringify({
            type: "load",
            document_id: parseInt(documentID)
        }));
    };

    ws.onmessage = handleWebSocketMessage;

    ws.onclose = (event) => {
        console.log("WebSocket connection closed:", event);
    };

    ws.onerror = (error) => {
        console.error("WebSocket error:", error);
    };

    // Example: Observing changes in the title input field
    const titleObserver = new MutationObserver(mutationsList => {
        for (let mutation of mutationsList) {
            if (mutation.type === 'attributes' && mutation.attributeName === 'value') {
                console.log("Title field value changed:", mutation.target.value);
                // Perform actions based on the title field changes if needed
            }
        }
    });

    const titleInput = document.getElementById('title');
    if (titleInput) {
        titleObserver.observe(titleInput, { attributes: true });
    }
});
