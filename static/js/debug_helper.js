// Debug Helper - Add this to modify.html temporarily to diagnose the issue
// Put this at the very end of your page, after all other scripts

console.log('🔍 DEBUG HELPER LOADED');

// Wait for page to fully load
window.addEventListener('load', function() {
    console.log('🔍 Page fully loaded, starting diagnostics...');
    
    // Check if form exists
    const form = document.querySelector('form[action="/update"], form[action="/save"]');
    console.log('🔍 Form found:', form ? 'YES' : 'NO');
    if (form) {
        console.log('🔍 Form action:', form.action);
        console.log('🔍 Form method:', form.method);
        console.log('🔍 Form enctype:', form.enctype);
    }
    
    // Check all file inputs
    const fileInputs = document.querySelectorAll('input[type="file"]');
    console.log('🔍 File inputs found:', fileInputs.length);
    
    fileInputs.forEach((input, index) => {
        console.log(`🔍 Input ${index + 1}:`, {
            name: input.name,
            id: input.id,
            hasFiles: input.files.length > 0,
            fileCount: input.files.length,
            duplicateHandled: input.dataset.duplicateHandled,
            duplicateAction: input.dataset.duplicateAction
        });
        
        // Add a test listener
        input.addEventListener('change', function() {
            console.log(`🔍 FILE CHANGE DETECTED on ${input.name}:`, {
                fileCount: input.files.length,
                files: Array.from(input.files).map(f => f.name)
            });
        });
    });
    
    // Monitor form submission
    if (form) {
        // Capture in capture phase (before other handlers)
        form.addEventListener('submit', function(e) {
            console.log('🔍 FORM SUBMIT CAPTURED (capture phase)');
            console.log('🔍 Event defaultPrevented:', e.defaultPrevented);
        }, true);
        
        // Normal phase
        form.addEventListener('submit', function(e) {
            console.log('🔍 FORM SUBMIT DETECTED (bubble phase)');
            console.log('🔍 Event defaultPrevented:', e.defaultPrevented);
            
            // Check file inputs at submission time
            const fileInputs = form.querySelectorAll('input[type="file"]');
            console.log('🔍 File inputs at submit time:');
            fileInputs.forEach((input, index) => {
                console.log(`  Input ${index + 1} (${input.name}):`, {
                    hasFiles: input.files.length > 0,
                    fileCount: input.files.length,
                    fileNames: Array.from(input.files).map(f => f.name),
                    duplicateHandled: input.dataset.duplicateHandled,
                    duplicateAction: input.dataset.duplicateAction
                });
            });
            
            // Check form data
            const formData = new FormData(form);
            console.log('🔍 Form data entries:');
            for (let [key, value] of formData.entries()) {
                if (value instanceof File) {
                    console.log(`  ${key}: FILE - ${value.name} (${value.size} bytes)`);
                } else {
                    console.log(`  ${key}: ${value}`);
                }
            }
        }, false);
    }
    
    // Check for duplicate warning system
    console.log('🔍 Duplicate warning system:', {
        existingFiles: typeof existingFiles !== 'undefined' ? existingFiles : 'NOT DEFINED',
        pendingFileOperation: typeof pendingFileOperation !== 'undefined' ? pendingFileOperation : 'NOT DEFINED',
        formSubmissionBlocked: typeof formSubmissionBlocked !== 'undefined' ? formSubmissionBlocked : 'NOT DEFINED'
    });
    
    console.log('🔍 Diagnostics complete. Try uploading files and submitting the form now.');
});

// Add a manual test button
setTimeout(function() {
    const testBtn = document.createElement('button');
    testBtn.textContent = '🔍 Run Diagnostics';
    testBtn.style.cssText = 'position: fixed; top: 10px; right: 10px; z-index: 99999; padding: 10px; background: #ff6b6b; color: white; border: none; border-radius: 5px; cursor: pointer; font-weight: bold;';
    testBtn.onclick = function() {
        console.clear();
        console.log('🔍 === MANUAL DIAGNOSTICS ===');
        
        const form = document.querySelector('form[action="/update"], form[action="/save"]');
        const fileInputs = form ? form.querySelectorAll('input[type="file"]') : [];
        
        console.log('Form exists:', !!form);
        console.log('File inputs count:', fileInputs.length);
        
        fileInputs.forEach((input, i) => {
            console.log(`\nInput ${i + 1} (${input.name}):`);
            console.log('  Files selected:', input.files.length);
            console.log('  File names:', Array.from(input.files).map(f => f.name));
            console.log('  duplicateHandled:', input.dataset.duplicateHandled);
            console.log('  duplicateAction:', input.dataset.duplicateAction);
        });
        
        if (typeof formSubmissionBlocked !== 'undefined') {
            console.log('\nformSubmissionBlocked:', formSubmissionBlocked);
        }
        if (typeof pendingFileOperation !== 'undefined') {
            console.log('pendingFileOperation:', pendingFileOperation);
        }
        
        console.log('\n🔍 Try submitting the form now and watch the console.');
    };
    document.body.appendChild(testBtn);
}, 1000);