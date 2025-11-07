// /static/js/hover_preview.js
// Floating full-size preview on the LEFT; no square zoom/transform.

document.addEventListener('DOMContentLoaded', function () {
  // 1) Kill typical zoom/lens UI purely via CSS (do NOT block events)
  const killerStyle = document.createElement('style');
  killerStyle.textContent = `
    img.thumb, .photo-container img, td img { transform: none !important; }
    img.thumb:hover, .photo-container img:hover, td img:hover { transform: none !important; }
    .zoom-box, .zoomLens, .zoomPup, .zoomWindow, .zoomContainer, .elevatezoom,
    .cloud-zoom-lens, .cloud-zoom-big, .MagicZoom, .mz-zoom-window,
    .xzoom-lens, .xzoom-preview, #hover-box, #thumb-preview,
    .image-hover-preview, #image-hover-preview {
      display: none !important; visibility: hidden !important; opacity: 0 !important;
    }
  `;
  document.head.appendChild(killerStyle);

  // 2) Build the LEFT-anchored floating preview
  let previewDiv = document.getElementById('hover-preview');
  if (!previewDiv) {
    previewDiv = document.createElement('div');
    previewDiv.id = 'hover-preview';
    previewDiv.style.cssText = `
      position: fixed;
      left: 20px;
      top: 50%;
      transform: translateY(-50%);
      width: 50vw;
      max-width: 800px;
      max-height: 500px;
      border: 3px solid #007bff;
      border-radius: 8px;
      box-shadow: 0 10px 40px rgba(0,0,0,0.3);
      opacity: 0;
      visibility: hidden;
      transition: opacity 0.3s ease, visibility 0.3s ease;
      z-index: 9999;
      background-color: white;
      pointer-events: none;
      padding: 10px;
      display: flex;
      align-items: center;
      justify-content: center;
    `;
    const img = document.createElement('img');
    img.id = 'hover-preview-img';
    img.style.cssText = `
      max-width: 100%;
      max-height: 100%;
      object-fit: contain;
      transform: none !important;
      transition: none !important;
    `;
    previewDiv.appendChild(img);
    document.body.appendChild(previewDiv);
  }
  const previewImg = document.getElementById('hover-preview-img');

  const show = (src) => {
    previewImg.src = src;
    previewDiv.style.opacity = '1';
    previewDiv.style.visibility = 'visible';
  };
  const hide = () => {
    previewDiv.style.opacity = '0';
    previewDiv.style.visibility = 'hidden';
  };

  // 3) Event delegation (no capture blocking)
  const isThumb = (el) =>
    !!el && (
      el.matches?.('img.thumb') ||
      (el.closest?.('.photo-container') && el.closest('.photo-container').querySelector('img'))
    );

  document.addEventListener('mouseover', (e) => {
    const t = e.target;
    if (t && t.tagName === 'IMG' && isThumb(t)) {
      t.style.transform = 'none';
      t.style.transition = 'none';
      show(t.src);
    }
  });

  document.addEventListener('mouseout', (e) => {
    const t = e.target;
    if (t && t.tagName === 'IMG' && isThumb(t)) {
      hide();
    }
  });

  // Touch support
  document.addEventListener('touchstart', (e) => {
    const t = e.target;
    if (t && t.tagName === 'IMG' && isThumb(t)) show(t.src);
  }, { passive: true });
  document.addEventListener('touchend', (e) => {
    const t = e.target;
    if (t && t.tagName === 'IMG' && isThumb(t)) hide();
  });
});
