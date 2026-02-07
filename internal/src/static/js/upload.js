const dropzone = document.getElementById('dropzone');
const fileInput = document.getElementById('dropzone-file');
const fileListDisplay = document.getElementById('fileList');
const uploadButton = document.getElementById('uploadButton');
const statusDisplay = document.getElementById('status');
const fileToggle = document.getElementById('fileToggle');
const folderToggle = document.getElementById('folderToggle');
const fileUploadSection = document.getElementById('fileUploadSection');
const folderUploadSection = document.getElementById('folderUploadSection');
const folderInput = document.getElementById('folderInput');
const folderDropzone = document.getElementById('folderDropzone');

let uploadMode = 'file'; // 'file' or 'folder'

// Toggle between file and folder upload modes
fileToggle.addEventListener('click', function(e) {
    e.preventDefault();
    uploadMode = 'file';
    fileUploadSection.style.display = 'block';
    folderUploadSection.style.display = 'none';
    fileToggle.classList.remove('bg-gray-400');
    fileToggle.classList.add('bg-orange-500');
    folderToggle.classList.remove('bg-orange-500');
    folderToggle.classList.add('bg-gray-400');
    fileListDisplay.innerHTML = '';
    fileInput.value = '';
    folderInput.value = '';
});

folderToggle.addEventListener('click', function(e) {
    e.preventDefault();
    uploadMode = 'folder';
    fileUploadSection.style.display = 'none';
    folderUploadSection.style.display = 'block';
    folderToggle.classList.remove('bg-gray-400');
    folderToggle.classList.add('bg-orange-500');
    fileToggle.classList.remove('bg-orange-500');
    fileToggle.classList.add('bg-gray-400');
    fileListDisplay.innerHTML = '';
    fileInput.value = '';
    folderInput.value = '';
});

// File upload events
dropzone.addEventListener('click', function() {
    fileInput.click();
});

fileInput.addEventListener('change', handleFileSelection);
dropzone.addEventListener('dragover', function(event) {
    event.preventDefault();
    dropzone.classList.add('bg-gray-100');
});

dropzone.addEventListener('dragleave', function(event) {
    dropzone.classList.remove('bg-gray-100');
});

dropzone.addEventListener('drop', function(event) {
    event.preventDefault();
    dropzone.classList.remove('bg-gray-100');
    handleFileSelection(event);
});

// Folder upload events
folderDropzone.addEventListener('click', function() {
    folderInput.click();
});

folderInput.addEventListener('change', handleFolderSelection);
folderDropzone.addEventListener('dragover', function(event) {
    event.preventDefault();
    folderDropzone.classList.add('bg-gray-100');
});

folderDropzone.addEventListener('dragleave', function(event) {
    folderDropzone.classList.remove('bg-gray-100');
});

folderDropzone.addEventListener('drop', function(event) {
    event.preventDefault();
    folderDropzone.classList.remove('bg-gray-100');
    handleFolderSelection(event);
});

function handleFileSelection(event) {
    const files = event.target.files || event.dataTransfer.files;
    fileListDisplay.innerHTML = '';

    for (let i = 0; i < files.length; i++) {
        const file = files[i];
        const fileElement = document.createElement('p');
        fileElement.textContent = file.name;
        fileListDisplay.appendChild(fileElement);
    }
}

function handleFolderSelection(event) {
    const files = event.target.files || event.dataTransfer.files;
    fileListDisplay.innerHTML = '';

    if (files.length === 0) {
        fileListDisplay.innerHTML = '<p class="text-gray-500">No files selected</p>';
        return;
    }

    // Show summary of folder contents
    const pathSet = new Set();
    for (let i = 0; i < files.length; i++) {
        if (files[i].webkitRelativePath) {
            const folderName = files[i].webkitRelativePath.split('/')[0];
            pathSet.add(folderName);
        }
    }

    const summary = document.createElement('p');
    summary.textContent = `Folder: ${Array.from(pathSet).join(', ')} (${files.length} files)`;
    summary.className = 'font-semibold text-blue-600';
    fileListDisplay.appendChild(summary);

    // Show first few files
    const displayLimit = 5;
    for (let i = 0; i < Math.min(displayLimit, files.length); i++) {
        const file = files[i];
        const fileElement = document.createElement('p');
        fileElement.className = 'text-sm text-gray-600';
        fileElement.textContent = file.webkitRelativePath || file.name;
        fileListDisplay.appendChild(fileElement);
    }

    if (files.length > displayLimit) {
        const more = document.createElement('p');
        more.className = 'text-sm text-gray-500 italic';
        more.textContent = `... and ${files.length - displayLimit} more files`;
        fileListDisplay.appendChild(more);
    }
}
uploadButton.addEventListener('click', async function() {
    let files;
    
    if (uploadMode === 'file') {
        files = fileInput.files;
    } else {
        files = folderInput.files;
    }
    
    if (files.length === 0) {
        statusDisplay.textContent = 'No files selected.';
        return;
    }

    const formData = new FormData();
    
    if (uploadMode === 'file') {
        for (let i = 0; i < files.length; i++) {
            formData.append('uploadFile', files[i]);
        }
    } else {
        // For folder upload, include file paths to preserve structure
        for (let i = 0; i < files.length; i++) {
            formData.append('uploadFile', files[i], files[i].webkitRelativePath || files[i].name);
        }
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
            statusDisplay.className = 'text-green-600 font-semibold';
            fileInput.value = '';
            folderInput.value = '';
            fileListDisplay.innerHTML = '';
        } else {
            statusDisplay.textContent = 'Failed to upload files.';
            statusDisplay.className = 'text-red-600 font-semibold';
        }
    } catch (error) {
        statusDisplay.textContent = 'Error: ' + error.message;
        statusDisplay.className = 'text-red-600 font-semibold';
    } finally {
        // Re-enable the button and reset text
        uploadButton.disabled = false;
        uploadButton.textContent = 'Upload';
    }
});
