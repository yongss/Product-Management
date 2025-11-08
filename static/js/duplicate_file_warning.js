// Enhanced duplicate_file_warning.js - Simplified and more reliable
// This version won't block form submission unnecessarily

// Track original file names for each file type
const existingFiles = {
    photos: new Set(),
    drawings: new Set(),
    cad: new Set(),
    cnc: new Set(),
    invoice: new Set()
};

// Store pending file operations
let pendingFileOperation = null;
let formSubmissionBlocked = false;

// Initialize on page load
document.addEventListener('DOMContentLoaded', function() {
    console.log('Enhanced duplicate warning system loaded');
    
    // Create modal HTML
    createDuplicateModal();
    
    // Populate existing files for each type
    ['photos', 'drawings', 'cad', 'cnc', 'invoice'].forEach(type => {
        const container = document.getElementById(`current${capitalize(type)}`);
        if (container) {
            const fileItems = container.querySelectorAll('.file-item');
            fileItems.forEach(item => {
                const filename = item.getAttribute('data-filename');
                if (filename) {
                    existingFiles[type].add(filename.toLowerCase());
                    console.log(`Added existing ${type} file: ${filename}`);
                }
            });
        }
    });

    // Add file input listeners
    addFileInputListeners();
    
    // Add form submission handler
    setupFormSubmission();
});

function capitalize(str) {
    if (str === 'cad') return 'Cad';
    if (str === 'cnc') return 'Cnc';
    return str.charAt(0).toUpperCase() + str.slice(1);
}

function createDuplicateModal() {
    // Remove existing modal if present
    const existingModal = document.getElementById('duplicateModal');
    if (existingModal) {
        existingModal.remove();
    }
    
    const modalHTML = `
        <div id="duplicateModal" class="duplicate-modal" style="display: none;">
            <div class="duplicate-modal-content">
                <div class="duplicate-modal-header">
                    <h2>⚠️ Duplicate Files Detected</h2>
                </div>
                <div class="duplicate-modal-body">
                    <p id="duplicateMessage"></p>
                    <div id="duplicateFileList" class="duplicate-file-list"></div>
                    <p class="duplicate-hint">What would you like to do?</p>
                </div>
                <div class="duplicate-modal-actions">
                    <button class="btn-modal btn-replace" onclick="handleDuplicateAction('replace')">
                        🔄 Replace Existing
                    </button>
                    <button class="btn-modal btn-keep-both" onclick="handleDuplicateAction('keepBoth')">
                        📋 Keep Both (will rename)
                    </button>
                    <button class="btn-modal btn-cancel-action" onclick="handleDuplicateAction('cancel')">
                        ❌ Cancel
                    </button>
                </div>
            </div>
        </div>
        
        <style>
            .duplicate-modal {
                position: fixed;
                top: 0;
                left: 0;
                width: 100%;
                height: 100%;
                background-color: rgba(0, 0, 0, 0.6);
                display: flex;
                justify-content: center;
                align-items: center;
                z-index: 10000;
                animation: fadeIn 0.2s ease-in-out;
            }
            
            @keyframes fadeIn {
                from { opacity: 0; }
                to { opacity: 1; }
            }
            
            .duplicate-modal-content {
                background: white;
                border-radius: 12px;
                box-shadow: 0 8px 32px rgba(0, 0, 0, 0.3);
                max-width: 500px;
                width: 90%;
                max-height: 80vh;
                overflow-y: auto;
                animation: slideIn 0.3s ease-out;
            }
            
            @keyframes slideIn {
                from { 
                    transform: translateY(-50px);
                    opacity: 0;
                }
                to { 
                    transform: translateY(0);
                    opacity: 1;
                }
            }
            
            .duplicate-modal-header {
                background: linear-gradient(135deg, #ff6b6b 0%, #ee5a6f 100%);
                color: white;
                padding: 20px;
                border-radius: 12px 12px 0 0;
            }
            
            .duplicate-modal-header h2 {
                margin: 0;
                font-size: 20px;
                font-weight: 600;
            }
            
            .duplicate-modal-body {
                padding: 24px;
            }
            
            .duplicate-modal-body p {
                margin: 0 0 16px 0;
                color: #333;
                line-height: 1.5;
            }
            
            .duplicate-hint {
                font-weight: 600;
                color: #555;
                margin-top: 20px !important;
            }
            
            .duplicate-file-list {
                background: #f8f9fa;
                border: 1px solid #dee2e6;
                border-radius: 6px;
                padding: 12px;
                margin: 16px 0;
                max-height: 200px;
                overflow-y: auto;
            }
            
            .duplicate-file-list div {
                padding: 6px 0;
                color: #495057;
                font-family: 'Courier New', monospace;
                font-size: 13px;
                border-bottom: 1px solid #e9ecef;
            }
            
            .duplicate-file-list div:last-child {
                border-bottom: none;
            }
            
            .duplicate-file-list div:before {
                content: "📄 ";
                margin-right: 6px;
            }
            
            .duplicate-modal-actions {
                display: flex;
                gap: 12px;
                padding: 20px 24px;
                background: #f8f9fa;
                border-radius: 0 0 12px 12px;
                justify-content: space-between;
            }
            
            .btn-modal {
                flex: 1;
                padding: 12px 16px;
                border: none;
                border-radius: 6px;
                font-size: 14px;
                font-weight: 600;
                cursor: pointer;
                transition: all 0.2s ease;
                display: flex;
                align-items: center;
                justify-content: center;
                gap: 6px;
            }
            
            .btn-replace {
                background: #28a745;
                color: white;
            }
            
            .btn-replace:hover {
                background: #218838;
                transform: translateY(-2px);
                box-shadow: 0 4px 12px rgba(40, 167, 69, 0.3);
            }
            
            .btn-keep-both {
                background: #007bff;
                color: white;
            }
            
            .btn-keep-both:hover {
                background: #0056b3;
                transform: translateY(-2px);
                box-shadow: 0 4px 12px rgba(0, 123, 255, 0.3);
            }
            
            .btn-cancel-action {
                background: #6c757d;
                color: white;
            }
            
            .btn-cancel-action:hover {
                background: #5a6268;
                transform: translateY(-2px);
                box-shadow: 0 4px 12px rgba(108, 117, 125, 0.3);
            }
            
            @media (max-width: 600px) {
                .duplicate-modal-actions {
                    flex-direction: column;
                }
                
                .btn-modal {
                    width: 100%;
                }
            }
        </style>
    `;
    
    document.body.insertAdjacentHTML('beforeend', modalHTML);
}

function showDuplicateModal(duplicates, fileType) {
    const modal = document.getElementById('duplicateModal');
    const message = document.getElementById('duplicateMessage');
    const fileList = document.getElementById('duplicateFileList');
    
    if (duplicates.length === 1) {
        message.textContent = `A file with the same name already exists:`;
    } else {
        message.textContent = `${duplicates.length} files with the same names already exist:`;
    }
    
    fileList.innerHTML = duplicates.map(name => `<div>${name}</div>`).join('');
    
    modal.style.display = 'flex';
    formSubmissionBlocked = true;
}

function hideDuplicateModal() {
    const modal = document.getElementById('duplicateModal');
    if (modal) {
        modal.style.display = 'none';
    }
}

// function handleDuplicateAction(action) {
//     console.log('Action selected:', action);
    
//     if (!pendingFileOperation) {
//         hideDuplicateModal();
//         formSubmissionBlocked = false;
//         return;
//     }
    
//     const { input, fileType } = pendingFileOperation;
    
//     switch(action) {
//         case 'replace':
//             console.log('User chose to replace files - files will be uploaded and replaced');
//             input.dataset.duplicateHandled = 'true';
//             input.dataset.duplicateAction = 'replace';
//             hideDuplicateModal();
//             formSubmissionBlocked = false;
//             break;
            
//         case 'keepBoth':
//             console.log('User chose to keep both files - server will add (1), (2) etc');
//             input.dataset.duplicateHandled = 'true';
//             input.dataset.duplicateAction = 'keepBoth';
//             hideDuplicateModal();
//             formSubmissionBlocked = false;
//             break;
            
//         case 'cancel':
//             console.log('User cancelled upload - clearing file input');
//             input.value = '';
//             input.dataset.duplicateHandled = 'true';
//             input.dataset.duplicateAction = 'cancel';
//             hideDuplicateModal();
//             formSubmissionBlocked = false;
//             break;
//     }
    
//     pendingFileOperation = null;
// }

function handleDuplicateAction(action) {
  // ...existing code...
  const { input, fileType } = pendingFileOperation;

  switch(action) {
    case 'replace':
      input.dataset.duplicateHandled = 'true';
      input.dataset.duplicateAction = 'replace';
      document.getElementById(fileType + 'Action').value = 'replace'; // NEW
      hideDuplicateModal();
      formSubmissionBlocked = false;
      break;

    case 'keepBoth':
      input.dataset.duplicateHandled = 'true';
      input.dataset.duplicateAction = 'keepBoth';
      document.getElementById(fileType + 'Action').value = 'keepBoth'; // NEW
      hideDuplicateModal();
      formSubmissionBlocked = false;
      break;

    case 'cancel':
      input.value = '';
      input.dataset.duplicateHandled = 'true';
      input.dataset.duplicateAction = 'cancel';
      hideDuplicateModal();
      formSubmissionBlocked = false;
      break;
  }

  pendingFileOperation = null;
}


function addFileInputListeners() {
    const fileInputs = [
        { name: 'photos', type: 'photos' },
        { name: 'drawings', type: 'drawings' },
        { name: 'cad', type: 'cad' },
        { name: 'cnc', type: 'cnc' },
        { name: 'invoice', type: 'invoice' }
    ];

    fileInputs.forEach(({ name, type }) => {
        const input = document.querySelector(`input[name="${name}"]`);
        if (input) {
            // Remove any existing event listeners by cloning
            const newInput = input.cloneNode(true);
            input.parentNode.replaceChild(newInput, input);
            
            // Mark as handled by default (no files selected)
            newInput.dataset.duplicateHandled = 'true';
            
            // Add new event listener
            newInput.addEventListener('change', function(e) {
                handleFileInputChange(e, type);
            }, false);
            
            console.log(`Added listener for ${type}`);
        }
    });
}

function handleFileInputChange(event, fileType) {
    console.log('File input changed for type:', fileType);
    
    const input = event.target;
    const files = Array.from(input.files);
    
    // Reset the flag
    input.dataset.duplicateHandled = 'false';
    input.dataset.duplicateAction = '';
    
    if (files.length === 0) {
        console.log('No files selected - marking as handled');
        input.dataset.duplicateHandled = 'true';
        return;
    }

    console.log(`Checking ${files.length} files for duplicates`);

    const duplicates = [];
    
    // Check for duplicates
    files.forEach(file => {
        const filename = file.name.toLowerCase();
        if (existingFiles[fileType].has(filename)) {
            duplicates.push(file.name);
            console.log(`Duplicate found: ${file.name}`);
        }
    });

    // If duplicates found, show modal
    if (duplicates.length > 0) {
        console.log(`Found ${duplicates.length} duplicate(s), showing modal`);
        pendingFileOperation = { input, fileType };
        showDuplicateModal(duplicates, fileType);
        return;
    }
    
    // No duplicates, mark as handled and approved
    console.log('No duplicates found, approved for upload');
    input.dataset.duplicateHandled = 'true';
    input.dataset.duplicateAction = 'approved';
}

function setupFormSubmission() {
    const form = document.querySelector('form[action="/update"], form[action="/save"]');
    if (!form) {
        console.log('No form found');
        return;
    }
    
    console.log('Setting up form submission handler');
    
    form.addEventListener('submit', function(e) {
        console.log('Form submit event triggered');
        console.log('formSubmissionBlocked:', formSubmissionBlocked);
        console.log('pendingFileOperation:', pendingFileOperation);
        
        // If modal is currently open, block submission
        if (formSubmissionBlocked || pendingFileOperation) {
            e.preventDefault();
            e.stopPropagation();
            console.log('BLOCKED: Modal is open, please choose an action');
            alert('Please handle the duplicate file warning before submitting.');
            return false;
        }
        
        // Check all file inputs to see if any need attention
        const fileInputs = form.querySelectorAll('input[type="file"]');
        let needsAttention = false;
        
        fileInputs.forEach(input => {
            const filesSelected = input.files.length > 0;
            const handled = input.dataset.duplicateHandled === 'true';
            
            console.log(`Input ${input.name}: files=${filesSelected}, handled=${handled}`);
            
            if (filesSelected && !handled) {
                console.log(`  -> NEEDS ATTENTION: ${input.name}`);
                needsAttention = true;
            }
        });
        
        if (needsAttention) {
            e.preventDefault();
            e.stopPropagation();
            console.log('BLOCKED: Some files not processed yet');
            alert('Please wait a moment for file validation to complete, then try again.');
            return false;
        }
        
        console.log('✅ Form submission ALLOWED - all checks passed');
        return true;
    });
}

// Update existing files list when a file is removed
function updateExistingFilesList(filename, fileType) {
    const lowerFilename = filename.toLowerCase();
    existingFiles[fileType].delete(lowerFilename);
    console.log(`Removed ${filename} from ${fileType} tracking`);
}

// Override the removeFile function to update our tracking
if (typeof window.removeFile !== 'undefined') {
    const originalRemoveFile = window.removeFile;
    window.removeFile = function(button, filename, fileType) {
        originalRemoveFile(button, filename, fileType);
        
        // Update our tracking after successful removal
        setTimeout(() => {
            updateExistingFilesList(filename, fileType);
        }, 500);
    };
}

// Make handleDuplicateAction globally accessible
window.handleDuplicateAction = handleDuplicateAction;

console.log('Duplicate file warning system initialized');
console.log('Existing files:', existingFiles);