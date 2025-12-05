function showStatusModal(title, message, type = "success") {
  const modal = document.getElementById("statusModal");
  const backdrop = document.getElementById("modalBackdrop");
  const panel = document.getElementById("modalPanel");

  // Elements for content
  const modalIcon = document.getElementById("modalIcon");
  const modalTitle = document.getElementById("modalTitle");
  const modalMessage = document.getElementById("modalMessage");

  // 1. Set Content
  modalTitle.innerText = title;
  modalMessage.innerText = message;

  // 2. Set Icon Styles
  modalIcon.className = "fa fa-5x mb-3 animated zoomIn";
  if (type === "success") {
    modalIcon.classList.add("fa-check-circle", "text-green-500");
  } else if (type === "error") {
    modalIcon.classList.add("fa-times-circle", "text-red-500");
  } else {
    modalIcon.classList.add("fa-info-circle", "text-blue-500");
  }

  // 3. Show the modal wrapper (display: block)
  modal.classList.remove("hidden");

  // 4. Trigger Animations (Small delay ensures transition runs)
  setTimeout(() => {
    // Fade in backdrop
    backdrop.classList.remove("opacity-0");
    backdrop.classList.add("opacity-100");

    // Fly in / Scale up panel
    panel.classList.remove(
      "opacity-0",
      "translate-y-4",
      "sm:translate-y-0",
      "sm:scale-95"
    );
    panel.classList.add("opacity-100", "translate-y-0", "sm:scale-100");
  }, 10);
}

function closeStatusModal() {
  const modal = document.getElementById("statusModal");
  const backdrop = document.getElementById("modalBackdrop");
  const panel = document.getElementById("modalPanel");

  // 1. Trigger Fade Out Animations
  backdrop.classList.remove("opacity-100");
  backdrop.classList.add("opacity-0");

  panel.classList.remove("opacity-100", "translate-y-0", "sm:scale-100");
  panel.classList.add(
    "opacity-0",
    "translate-y-4",
    "sm:translate-y-0",
    "sm:scale-95"
  );

  // 2. Wait for animation to finish (300ms matches duration-300) then hide
  setTimeout(() => {
    modal.classList.add("hidden");
  }, 300);
}

// Close on backdrop click
window.onclick = function (event) {
  const modal = document.getElementById("statusModal");
  // We check if the target is the modal container (which covers the screen)
  // Note: In the HTML structure above, the click might register on the
  // container div wrapping the panel.
  if (
    event.target.id === "statusModal" ||
    event.target.closest("#modalBackdrop")
  ) {
    // Logic depends slightly on z-index layering, simpler approach:
    // If they click the backdrop div specifically:
  }
};
// Improved Backdrop Click Handler for this structure
document
  .getElementById("statusModal")
  .addEventListener("click", function (event) {
    // If the click is strictly on the background wrapper (not the white panel)
    const panel = document.getElementById("modalPanel");
    if (!panel.contains(event.target)) {
      closeStatusModal();
    }
  });
