const panels = document.querySelectorAll('.form-panel');
let activePanel = document.querySelector('.form-panel.active');

panels.forEach(panel => {
	panel.addEventListener('click', () => {

		if (activePanel) {
			activePanel.classList.remove('active');
			activePanel.querySelector('.panel-content').classList.add('inactive');
		}

		panel.classList.add('active');
		panel.querySelector('.panel-content').classList.remove('inactive');
		document.getElementById('active-panel').value = panel.dataset.target;
		activePanel = panel;
	});
});
	
document.addEventListener('DOMContentLoaded', () => {
    const dropzone = document.getElementById('dropzone');
    const fileInput = document.getElementById('fileInput');
    const dropzoneText = dropzone.querySelector('.dropzone-content p');

    function isValidFileType(file) {
        return file && (file.type === 'application/pdf' || file.name.toLowerCase().endsWith('.pdf'));
    }

    dropzone.addEventListener('dragover', (event) => {
        event.preventDefault();

        let isValid = true;
        if (event.dataTransfer.items) {
            for (let i = 0; i < event.dataTransfer.items.length; i++) {
                const item = event.dataTransfer.items[i];
                if (item.kind === 'file' && item.type) {
                    const file = item.getAsFile();
                    if (!isValidFileType(file)) {
                        isValid = false;
                        break;
                    }
                } else {
                    isValid = false;
                }
            }
        }

        dropzone.style.borderColor = isValid ? '#4f46e5' : '#e53e3e';
    });

    dropzone.addEventListener('dragleave', () => {
        dropzone.style.borderColor = '#6366f1';
    });

    dropzone.addEventListener('drop', (event) => {
        event.preventDefault();
        dropzone.style.borderColor = '#6366f1';

        const files = event.dataTransfer.files;
        if (files.length === 0) return;

        const isValid = Array.from(files).every(file => isValidFileType(file));
        if (!isValid) {
            dropzone.style.borderColor = '#e53e3e';
            dropzoneText.textContent = 'Ошибка: Принимаются только PDF-файлы';
            setTimeout(() => {
                resetDropzoneText();
                dropzone.style.borderColor = '#6366f1';
            }, 2000);
            return;
        }

        fileInput.files = files;
        updateDropzoneText(files[0].name);
    });

    dropzone.addEventListener('click', () => {
        fileInput.click();
    });

    fileInput.addEventListener('change', () => {
        const file = fileInput.files[0];
        if (file && isValidFileType(file)) {
            updateDropzoneText(file.name);
        } else {
            dropzone.style.borderColor = '#e53e3e';
            dropzoneText.textContent = 'Ошибка: Неверный формат файла';
            fileInput.value = '';
            setTimeout(() => {
                resetDropzoneText();
                dropzone.style.borderColor = '#6366f1';
            }, 2000);
        }
    });

    function updateDropzoneText(filename) {
        dropzoneText.textContent = `Файл "${filename}" выбран`;
    }

    function resetDropzoneText() {
        dropzoneText.textContent = 'Перетащите файл сюда, или нажмите для выбора';
    }
});