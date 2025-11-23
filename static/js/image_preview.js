        // Preview functionality
        function showPreview(img) {
            const preview = document.getElementById('preview');
            const previewImg = document.getElementById('previewImg');
            previewImg.src = img.src;
            preview.style.display = 'flex';
        }

        function hidePreview() {
            document.getElementById('preview').style.display = 'none';
        }

        // // Preview event listeners
        // document.getElementById('preview').addEventListener('click', function (e) {
        //     if (e.target === this) hidePreview();
        // });

        // Close preview on any click
        document.getElementById('preview').addEventListener('click', hidePreview);

        document.addEventListener('keydown', function (e) {
            if (e.key === 'Escape') hidePreview();
        });