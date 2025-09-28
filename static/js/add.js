// Add product functionality
const uploadedFiles = {
    photos: [],
    drawings: [],
    cad: []
};

function setupFileUpload(inputId, listId, category) {
    const input = document.getElementById(inputId);
    const list = document.getElementById(listId);
    
    if (!input || !list) return;
    
    input.addEventListener('change', (e) => {
        handleFiles(e.target.files, category, list);
    });

    // Add drag and drop functionality
    const uploadArea = input.parentElement;
    
    uploadArea.addEventListener('dragover', (e) => {
        e.preventDefault();
        uploadArea.style.borderColor = '#007bff';
        uploadArea.style.backgroundColor = '#f0f8ff';
    });
    
    uploadArea.addEventListener('dragleave', (e) => {
        e.preventDefault();
        uploadArea.style.borderColor = '#ddd';
        uploadArea.style.backgroundColor = '';
    });
    
    uploadArea.addEventListener('drop', (e) => {
        e.preventDefault();
        uploadArea.style.borderColor = '#ddd';
        uploadArea.style.backgroundColor = '';
        
        const files = e.dataTransfer.files;
        handleFiles(files, category, list);
    });
}

function handleFiles(files, category, listContainer) {
    for (let file of files) {
        // Validate file type
        const ext = '.' + file.name.split('.').pop().toLowerCase();
        if (!isValidFileType(category, ext)) {
            alert(`Invalid file type "${ext}" for ${category}. Please check allowed file types.`);
            continue;
        }
        
        uploadFile(file, category, listContainer);
    }
}

function isValidFileType(category, ext) {
    const validTypes = {
        photos: ['.jpg', '.jpeg', '.png'],
        drawings: ['.pdf', '.dxf', '.dwg'],
        cad: ['.stp', '.step', '.fcstd', '.igs', '.iges']
    };
    
    return validTypes[category] && validTypes[category].includes(ext);
}

function uploadFile(file, category, listContainer) {
    const formData = new FormData();
    formData.append('file', file);
    formData.append('category', category);
    
    // Show upload progress
    const progressItem = document.createElement('div');
    progressItem.className = 'file-item';
    progressItem.innerHTML = `
        <span>${escapeHtml(file.name)} - Uploading...</span>
        <div style="width: 100px; height: 4px; background: #eee; border-radius: 2px;">
            <div style="width: 0%; height: 100%; background: #007bff; border-radius: 2px; transition: width 0.3s;"></div>
        </div>
    `;
    listContainer.appendChild(progressItem);
    
    fetch('/upload', {
        method: 'POST',
        body: formData
    })
    .then(response => response.json())
    .then(data => {
        listContainer.removeChild(progressItem);
        
        if (data.success) {
            uploadedFiles[category].push(data.path);
            updateFileList(listContainer, uploadedFiles[category], category);
        } else {
            alert('Upload failed: ' + (data.error || 'Unknown error'));
        }
    })
    .catch(error => {
        listContainer.removeChild(progressItem);
        console.error('Upload error:', error);
        alert('Upload failed: ' + error.message);
    });
}

function updateFileList(container, files, category) {
    container.innerHTML = files.map((file, index) => `
        <div class="file-item">
            <span>${escapeHtml(file.split('/').pop())}</span>
            <button type="button" onclick="removeFile('${category}', ${index})">Remove</button>
        </div>
    `).join('');
}

function removeFile(category, index) {
    uploadedFiles[category].splice(index, 1);
    const listId = category === 'photos' ? 'photosList' : 
                  category === 'drawings' ? 'drawingsList' : 'cadList';
    const container = document.getElementById(listId);
    if (container) {
        updateFileList(container, uploadedFiles[category], category);
    }
}

function validateForm() {
    const partNo = document.getElementById('partNo').value.trim();
    const partName = document.getElementById('partName').value.trim();
    const description = document.getElementById('description').value.trim();
    const cost = document.getElementById('cost').value.trim();
    const qty = document.getElementById('qty').value;
    const material = document.getElementById('material').value.trim();
    
    if (!partNo) {
        alert('PartNo number is required');
        return false;
    }
    
    if (!partName) {
        alert('PartName is required');
        return false;
    }
    if (!description) {
        alert('Description is required');
        return false;
    }
    if (!cost) {
        alert('Cost is required');
        return false;
    }
    
    if (!qty || qty < 0) {
        alert('Please enter a valid quantity');
        return false;
    }
    
    if (!material) {
        alert('Material is required');
        return false;
    }
    
    return true;
}

function escapeHtml(text) {
    const map = {
        '&': '&amp;',
        '<': '&lt;',
        '>': '&gt;',
        '"': '&quot;',
        "'": '&#039;'
    };
    return text.replace(/[&<>"']/g, function(m) { return map[m]; });
}