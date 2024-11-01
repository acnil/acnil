---
title: "Colaboradores"
---

Puedes introducir el teléfono de una persona para saber si pertenecen a Acnil

<style>
    #fetchButton {
        background-color: #007BFF; /* Blue background */
        font-size: 18px; /* Larger text size */
        padding: 8px 20px; /* Padding for a bigger button */
        border: none; /* Remove default border */
        border-radius: 5px; /* Rounded corners */
        cursor: pointer; /* Pointer cursor on hover */
        box-shadow: 2px 2px 5px rgba(0, 0, 0, 0.3); /* Subtle shadow */
        transition: background-color 0.3s ease; /* Smooth color transition */
        color: black;
        width:  195px
    }

    #fetchButton:hover {
        background-color: #0056b3; /* Darker blue on hover */
    }

    #fetchButton:active {
        background-color: #004080; /* Even darker blue when clicked */
    }

    #phoneInput {
        padding: 8px;
        font-size: 16px;
        margin-bottom: 10px;
        border: 1px solid #ccc;
        border-radius: 5px;
        width: calc(100% - 200px);
        color: black;
    }

    #result {
        color: black;
        margin-top: 20px;
        font-family: monospace; /* Monospace font for better readability */
        white-space: pre-wrap; /* Preserve formatting */
        background-color: #f9f9f9;
        padding: 10px;
        border-radius: 5px;
        border: 1px solid #ddd;
    }
</style>

<input type="text" id="phoneInput" placeholder="Teléfono">
<button id="fetchButton">Comprobar</button>
<div id="result"></div>

<script>
    document.getElementById('fetchButton').addEventListener('click', function() {
        const dni = document.getElementById('phoneInput').value;
        const url = `https://2ona3c7j6easp4k4636aibgdxm0cukhy.lambda-url.eu-west-1.on.aws/?phone=${encodeURIComponent(dni)}`;

        document.getElementById('result').textContent = ""
        console.log(url)
        fetch(url)
            .then(response => {
                if (!response.ok) {
                    throw new Error('Network response was not ok ' + response.statusText);
                }
                return response.json(); // or response.text() if you expect plain text
            })
            .then(data => {
                console.log('Success:', data);
                document.getElementById('result').textContent = data.result
            })
            .catch(error => {
                console.error('Error:', error);
                document.getElementById('result').textContent = 'Error: ' + error.message;
            });
    });
</script>