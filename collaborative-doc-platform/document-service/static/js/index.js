document.addEventListener("DOMContentLoaded", () => {
    fetch("/documents")
        .then(response => {
            if (!response.ok) throw new Error("Error fetching documents");
            return response.json();
        })
        .then(documents => {
            const documentList = document.getElementById("document-list");
            documents.forEach(doc => {
                const listItem = document.createElement("li");
                listItem.textContent = doc.title;
                listItem.onclick = () => window.location.href = `/editor.html?doc=${doc.id}`;
                documentList.appendChild(listItem);
            });
        })
        .catch(error => console.error("Error fetching documents:", error));
});

function createNewDocument() {
    fetch("/documents", {
        method: "POST",
        headers: {
            "Content-Type": "application/json"
        },
        body: JSON.stringify({ title: "Untitled", content: "" }) // Use an empty string for initial content
    }).then(response => {
        if (!response.ok) {
            return response.text().then(text => { throw new Error(text) });
        }
        return response.json();
    }).then(doc => {
        window.location.href = `/editor.html?doc=${doc.id}`;
    }).catch(error => {
        console.error("Error creating document:", error);
        alert(`Error creating document: ${error.message}`);
    });
}
