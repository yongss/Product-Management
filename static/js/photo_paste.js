document.addEventListener('DOMContentLoaded', function () {
                const pasteArea = document.getElementById('pasteArea');
                const photoInput = document.getElementById('photoInput');
                const previewContainer = document.getElementById('previewContainer');
                const imageFiles = new DataTransfer();

                // Handle paste events
                document.addEventListener('paste', function (e) {
                    if (document.activeElement.tagName === 'INPUT' ||
                        document.activeElement.tagName === 'TEXTAREA') {
                        return;
                    }

                    const items = e.clipboardData.items;
                    for (let item of items) {
                        if (item.type.indexOf('image') !== -1) {
                            const file = item.getAsFile();
                            addImageFile(file);
                        }
                    }
                });

                // Handle drag and drop
                pasteArea.addEventListener('dragover', function (e) {
                    e.preventDefault();
                    this.classList.add('dragover');
                });

                pasteArea.addEventListener('dragleave', function (e) {
                    e.preventDefault();
                    this.classList.remove('dragover');
                });

                pasteArea.addEventListener('drop', function (e) {
                    e.preventDefault();
                    this.classList.remove('dragover');

                    for (let file of e.dataTransfer.files) {
                        if (file.type.startsWith('image/')) {
                            addImageFile(file);
                        }
                    }
                });

                // Handle click to paste
                pasteArea.addEventListener('click', function () {
                    navigator.clipboard.read().then(items => {
                        items.forEach(item => {
                            if (item.types.includes('image/png') ||
                                item.types.includes('image/jpeg')) {
                                item.getType('image/png').then(blob => {
                                    addImageFile(new File([blob], 'pasted-image.png'));
                                });
                            }
                        });
                    }).catch(err => {
                        console.log('Click to paste not supported:', err);
                    });
                });

                function addImageFile(file) {
                    // Add to DataTransfer object
                    imageFiles.items.add(file);
                    photoInput.files = imageFiles.files;

                    // Create preview
                    const reader = new FileReader();
                    reader.onload = function (e) {
                        const img = document.createElement('img');
                        img.src = e.target.result;
                        img.classList.add('preview-image');
                        previewContainer.appendChild(img);
                    };
                    reader.readAsDataURL(file);
                }
            });