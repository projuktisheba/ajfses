// --- 1. Reusable Modal Helper Function ---
    function showStatusModal(title, message, type = 'success') {
        const modalElement = document.getElementById('statusModal');
        const modalIcon = document.getElementById('modalIcon');
        const modalTitle = document.getElementById('modalTitle');
        const modalMessage = document.getElementById('modalMessage');

        // Set Content
        modalTitle.innerText = title;
        modalMessage.innerText = message;

        // Reset Classes
        modalIcon.className = 'fa fa-5x mb-3 animated zoomIn'; // utilizing your template's animate.css

        // Style based on type
        if (type === 'success') {
            modalIcon.classList.add('fa-check-circle', 'text-success');
        } else if (type === 'error') {
            modalIcon.classList.add('fa-times-circle', 'text-danger');
        } else {
            modalIcon.classList.add('fa-info-circle', 'text-primary');
        }

        // Initialize and Show Bootstrap Modal
        const myModal = new bootstrap.Modal(modalElement);
        myModal.show();
    }