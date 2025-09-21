const dropzone = document.getElementById('dropzone');
const fileInput = document.getElementById('dropzone-file');
const fileListDisplay = document.getElementById('fileList');
const uploadButton = document.getElementById('uploadButton');
const statusDisplay = document.getElementById('status');

dropzone.addEventListener('click', function() {
    fileInput.click();
});

fileInput.addEventListener('change', handleFileSelection);
dropzone.addEventListener('dragover', function(event) {
    event.preventDefault();
    dropzone.classList.add('bg-gray-100'); // Change background on dragover
});

dropzone.addEventListener('dragleave', function(event) {
    dropzone.classList.remove('bg-gray-100'); // Revert background on drag leave
});

dropzone.addEventListener('drop', function(event) {
    event.preventDefault();
    dropzone.classList.remove('bg-gray-100'); // Revert background on drop
    handleFileSelection(event);
});

function handleFileSelection(event) {
    const files = event.target.files || event.dataTransfer.files;
    fileListDisplay.innerHTML = ''; // Clear previous file list

    for (let i = 0; i < files.length; i++) {
        const file = files[i];
        const fileElement = document.createElement('p');
        fileElement.textContent = file.name;
        fileListDisplay.appendChild(fileElement);
    }
}
uploadButton.addEventListener('click', async function() {
    const files = fileInput.files;
    if (files.length === 0) {
        statusDisplay.textContent = 'No files selected.';
        return;
    }

    const formData = new FormData();
    for (let i = 0; i < files.length; i++) {
        formData.append('uploadFile', files[i]);
    }

    // Disable the button and change text
    uploadButton.disabled = true;
    uploadButton.textContent = 'Uploading...';

    try {
        const response = await fetch('/upload', {
            method: 'POST',
            body: formData
        });

        if (response.ok) {
            statusDisplay.textContent = 'Files uploaded successfully!';
            fileInput.value = ''; // Clear the file input
            fileListDisplay.innerHTML = ''; // Clear the file list display
        } else {
            statusDisplay.textContent = 'Failed to upload files.';
        }
    } catch (error) {
        statusDisplay.textContent = 'Error: ' + error.message;
    } finally {
        // Re-enable the button and reset text
        uploadButton.disabled = false;
        uploadButton.textContent = 'Upload';
    }
});
