        // function showPreview(img) {
        //     const preview = document.getElementById('preview');
        //     const previewImg = document.getElementById('previewImg');
        //     previewImg.src = img.src;
        //     preview.style.display = 'flex';
        // }

        // function hidePreview() {
        //     document.getElementById('preview').style.display = 'none';
        // }

        // // Close preview when clicking outside
        // document.getElementById('preview').addEventListener('click', function(e) {
        //     if (e.target === this) {
        //         hidePreview();
        //     }
        // });

        // // Close preview with ESC key
        // document.addEventListener('keydown', function(e) {
        //     if (e.key === 'Escape') {
        //         hidePreview();
        //     }
        // });

        document.addEventListener('DOMContentLoaded', function() {
            // Get current sort parameters from URL
            const urlParams = new URLSearchParams(window.location.search);
            let currentSort = urlParams.get('sort') || 'updated_at';
            let currentOrder = urlParams.get('order') || 'DESC';

            // Add click handlers to all sort links
            document.querySelectorAll('.sort-link').forEach(link => {
                link.addEventListener('click', function(e) {
                    e.preventDefault();
                    const column = this.dataset.column;

                    // Toggle order if clicking the same column
                    if (column === currentSort) {
                        currentOrder = currentOrder === 'ASC' ? 'DESC' : 'ASC';
                    } else {
                        currentSort = column;
                        currentOrder = 'ASC';
                    }

                    // Build new URL with sort parameters
                    const newUrl = new URL(window.location.href);
                    newUrl.searchParams.set('sort', currentSort);
                    newUrl.searchParams.set('order', currentOrder);

                    // Preserve search query if exists
                    const searchQuery = urlParams.get('q');
                    if (searchQuery) {
                        newUrl.searchParams.set('q', searchQuery);
                    }

                    // Navigate to new URL
                    window.location.href = newUrl.toString();
                });
            });
        });