// Add this to your static/js folder
// duplicate_file_warning.js

// Track original file names for each file type
const existingFiles = {
    photos: new Set(),
    drawings: new Set(),
    cad: new Set(),
    cnc: new Set(),
    invoice: new Set()
};

// Initialize existing file names on page load
document.addEventListener('DOMContentLoaded', function() {
    // Populate existing files for each type
    ['photos', 'drawings', 'cad', 'cnc', 'invoice'].forEach(type => {
        const container = document.getElementById(`current${capitalize(type)}`);
        if (container) {
            const fileItems = container.querySelectorAll('.file-item');
            fileItems.forEach(item => {
                const filename = item.getAttribute('data-filename');
                if (filename) {
                    existingFiles[type].add(filename.toLowerCase());
                }
            });
        }
    });

    // Add change listeners to all file inputs
    addFileInputListeners();
});

function capitalize(str) {
    if (str === 'cad') return 'Cad';
    if (str === 'cnc') return 'Cnc';
    return str.charAt(0).toUpperCase() + str.slice(1);
}

function addFileInputListeners() {
    // Photos
    const photoInput = document.querySelector('input[name="photos"]');
    if (photoInput) {
        photoInput.addEventListener('change', function(e) {
            handleFileInputChange(e, 'photos');
        });
    }

    // Drawings
    const drawingInput = document.querySelector('input[name="drawings"]');
    if (drawingInput) {
        drawingInput.addEventListener('change', function(e) {
            handleFileInputChange(e, 'drawings');
        });
    }

    // CAD
    const cadInput = document.querySelector('input[name="cad"]');
    if (cadInput) {
        cadInput.addEventListener('change', function(e) {
            handleFileInputChange(e, 'cad');
        });
    }

    // CNC
    const cncInput = document.querySelector('input[name="cnc"]');
    if (cncInput) {
        cncInput.addEventListener('change', function(e) {
            handleFileInputChange(e, 'cnc');
        });
    }

    // Invoice
    const invoiceInput = document.querySelector('input[name="invoice"]');
    if (invoiceInput) {
        invoiceInput.addEventListener('change', function(e) {
            handleFileInputChange(e, 'invoice');
        });
    }
}

function handleFileInputChange(event, fileType) {
    const input = event.target;
    const files = Array.from(input.files);
    
    if (files.length === 0) return;

    const duplicates = [];
    
    // Check for duplicates
    files.forEach(file => {
        const filename = file.name.toLowerCase();
        if (existingFiles[fileType].has(filename)) {
            duplicates.push(file.name);
        }
    });

    // If duplicates found, show warning
    if (duplicates.length > 0) {
        const message = duplicates.length === 1
            ? `A file named "${duplicates[0]}" already exists.\n\nDo you want to replace it?`
            : `The following files already exist:\n\n${duplicates.join('\n')}\n\nDo you want to replace them?`;
        
        if (!confirm(message)) {
            // User cancelled - clear the input
            input.value = '';
            return;
        }
    }
    
    // If user confirmed or no duplicates, allow upload
    // Files will be processed normally with automatic rename handling on server
}

// Update existing files list when a file is removed
function updateExistingFilesList(filename, fileType) {
    const lowerFilename = filename.toLowerCase();
    existingFiles[fileType].delete(lowerFilename);
}

// Override the removeFile function to update our tracking
const originalRemoveFile = window.removeFile;
window.removeFile = function(button, filename, fileType) {
    if (originalRemoveFile) {
        originalRemoveFile(button, filename, fileType);
    }
    
    // Update our tracking after successful removal
    setTimeout(() => {
        updateExistingFilesList(filename, fileType);
    }, 500);
};