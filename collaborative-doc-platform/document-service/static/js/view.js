document.addEventListener("DOMContentLoaded", function() {
    const urlParams = new URLSearchParams(window.location.search);
    const docId = urlParams.get('id');
    const token = urlParams.get('token');
    fetchDocument(docId, token);
});

function fetchDocument(id, token) {
    fetch(`/documents/${id}`, {
        method: "GET",
        headers: {
            "Authorization": `Bearer ${token}`
        }
    })
    .then(response => response.json())
    .then(data => {
        document.getElementById("title").value = data.title;
        document.getElementById("content").value = data.content;
    });
}

function updateDocument() {
    const urlParams = new URLSearchParams(window.location.search);
    const docId = urlParams.get('id');
    const token = urlParams.get('token');
    const title = document.getElementById("title").value;
    const content = document.getElementById("content").value;

    fetch(`/documents/${docId}`, {
        method: "PUT",
        headers: {
            "Content-Type": "application/json",
            "Authorization": `Bearer ${token}`
        },
        body: JSON.stringify({ title, content })
    })
    .then(response => response.json())
    .then(data => {
        alert("Document updated!");
    });
}

function deleteDocument() {
    const urlParams = new URLSearchParams(window.location.search);
    const docId = urlParams.get('id');
    const token = urlParams.get('token');

    fetch(`/documents/${docId}`, {
        method: "DELETE",
        headers: {
            "Authorization": `Bearer ${token}`
        }
    })
    .then(response => response.json())
    .then(data => {
        alert("Document deleted!");
        window.location.href = `index.html?token=${token}`;
    });
}
