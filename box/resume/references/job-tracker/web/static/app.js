let applications = [];
let currentFilter = 'all';
let isEditMode = false;
let pendingFiles = [];

async function loadApplications() {
    try {
        console.log('Loading applications...');
        const response = await fetch('/api/applications');
        applications = await response.json();
        console.log('Loaded applications:', applications.length, applications);
        renderApplications();
        updateStatusBar();
        console.log('Applications rendered and status updated');
    } catch (error) {
        console.error('Failed to load applications:', error);
    }
}

async function updateStatusBar() {
    try {
        const response = await fetch('/api/status');
        const status = await response.json();
        
        document.getElementById('gpg-recipient-compact').textContent = status.gpg_recipient || 'Unknown';
        document.getElementById('data-status-compact').textContent = status.encrypted ? 'Encrypted' : 'Plain';
        document.getElementById('file-size-compact').textContent = status.file_size || '-';
        document.getElementById('last-updated-compact').textContent = status.last_updated ? 
            new Date(status.last_updated).toLocaleString('ko-KR', {
                month: '2-digit', 
                day: '2-digit', 
                hour: '2-digit', 
                minute: '2-digit'
            }) : '-';
    } catch (error) {
        document.getElementById('gpg-recipient').textContent = 'Error';
        console.error('Failed to load status:', error);
    }
}

function renderApplications() {
    const tbody = document.getElementById('applications-body');
    const emptyState = document.getElementById('empty-state');
    
    let filtered = applications;
    if (currentFilter !== 'all') {
        filtered = applications.filter(app => app.status === currentFilter);
    }
    
    if (filtered.length === 0) {
        tbody.innerHTML = '';
        emptyState.style.display = 'block';
        return;
    }
    
    emptyState.style.display = 'none';
    tbody.innerHTML = filtered.map(app => `
        <tr ${app.final_result === 'rejected' ? 'class="rejected-row"' : ''}>
            <td>
                <strong>${app.company}</strong>
                ${app.notes ? `<span class="notes-indicator" onmouseenter="showTooltip(event, \`${escapeHtml(app.notes)}\`)" onmouseleave="hideTooltip()">📝</span>` : ''}
            </td>
            <td>${app.position}</td>
            <td>
                <span class="status ${app.status}">${getStatusLabel(app.status)}</span>
            </td>
            <td class="result-cell">
                ${getFinalResult(app.final_result)}
            </td>
            <td>${formatDateWithDay(app.applied_date)}</td>
            <td>${app.platform || '-'}</td>
            <td>
                ${app.url ? `<a href="${app.url}" target="_blank" class="url-link">🔗</a>` : '-'}
            </td>
            <td class="files-cell">
                ${getFilesDisplay(app)}
            </td>
            <td class="actions-cell" style="display: ${isEditMode ? 'table-cell' : 'none'};">
                ${isEditMode ? `
                    <div class="actions">
                        <button class="action-btn" onclick="editApplication('${app.id}')">Edit</button>
                        <button class="action-btn" onclick="deleteApplication('${app.id}')">Delete</button>
                    </div>
                ` : ''}
            </td>
        </tr>
    `).join('');
}

function filterStatus(status) {
    currentFilter = status;
    renderApplications();
}

function openModal(appId = null) {
    const modal = document.getElementById('application-modal');
    const form = document.getElementById('application-form');
    const title = document.getElementById('modal-title');
    
    if (appId) {
        const app = applications.find(a => a.id === appId);
        if (app) {
            title.textContent = 'Edit Application';
            document.getElementById('app-id').value = app.id;
            document.getElementById('company').value = app.company;
            document.getElementById('position').value = app.position;
            document.getElementById('status').value = app.status;
            document.getElementById('final-result').value = app.final_result || '';
            document.getElementById('applied-date').value = app.applied_date;
            document.getElementById('platform').value = app.platform || '';
            document.getElementById('url').value = app.url || '';
            document.getElementById('notes').value = app.notes || '';
            displayFiles(app.files || []);
        }
    } else {
        title.textContent = 'New Application';
        form.reset();
        document.getElementById('app-id').value = '';
        displayFiles([]);
        pendingFiles = [];
    }
    
    modal.classList.add('active');
}

function closeModal() {
    document.getElementById('application-modal').classList.remove('active');
}

function editApplication(id) {
    openModal(id);
}

async function deleteApplication(id) {
    if (!confirm('Are you sure you want to delete this application?')) {
        return;
    }
    
    try {
        await fetch(`/api/application?id=${id}`, {
            method: 'DELETE'
        });
        await loadApplications();
    } catch (error) {
        console.error('Failed to delete application:', error);
    }
}

async function updateStatus(appId, newStatus) {
    const app = applications.find(a => a.id === appId);
    if (!app) return;
    
    const originalStatus = app.status;
    app.status = newStatus; // Optimistic update
    
    try {
        const response = await fetch('/api/application', {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(app)
        });
        
        if (!response.ok) {
            throw new Error('Failed to update status');
        }
        
        // Show success feedback
        showToast(`Status updated: ${app.company} → ${newStatus.toUpperCase()}`);
        
        // Update status bar
        updateStatusBar();
        
    } catch (error) {
        // Revert on error
        app.status = originalStatus;
        renderApplications();
        showToast('Failed to update status', 'error');
        console.error('Failed to update status:', error);
    }
}

function showToast(message, type = 'success') {
    // Create toast container if it doesn't exist
    let toastContainer = document.getElementById('toast-container');
    if (!toastContainer) {
        toastContainer = document.createElement('div');
        toastContainer.id = 'toast-container';
        toastContainer.className = 'toast-container';
        document.body.appendChild(toastContainer);
    }
    
    const toast = document.createElement('div');
    toast.className = `toast ${type}`;
    toast.textContent = message;
    toastContainer.appendChild(toast);
    
    setTimeout(() => {
        toast.classList.add('show');
    }, 100);
    
    setTimeout(() => {
        toast.classList.remove('show');
        setTimeout(() => {
            if (toast.parentNode) {
                toast.parentNode.removeChild(toast);
            }
        }, 300);
    }, 3000);
}

async function syncWithGit() {
    const statusDiv = document.getElementById('sync-status');
    statusDiv.textContent = 'Syncing with Git...';
    statusDiv.classList.add('show');
    
    try {
        const response = await fetch('/api/sync', { method: 'POST' });
        const result = await response.json();
        statusDiv.textContent = result.message || 'Sync completed';
        setTimeout(() => {
            statusDiv.classList.remove('show');
        }, 3000);
    } catch (error) {
        statusDiv.textContent = 'Sync failed: ' + error.message;
        setTimeout(() => {
            statusDiv.classList.remove('show');
        }, 5000);
    }
}

document.getElementById('application-form').addEventListener('submit', async (e) => {
    e.preventDefault();
    
    const id = document.getElementById('app-id').value;
    const application = {
        id: id || Date.now().toString(),
        company: document.getElementById('company').value,
        position: document.getElementById('position').value,
        status: document.getElementById('status').value,
        final_result: document.getElementById('final-result').value,
        applied_date: document.getElementById('applied-date').value,
        platform: document.getElementById('platform').value,
        url: document.getElementById('url').value,
        notes: document.getElementById('notes').value
    };
    
    try {
        const response = await fetch('/api/application', {
            method: id ? 'PUT' : 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(application)
        });
        
        if (!response.ok) {
            const errorText = await response.text();
            throw new Error(`HTTP ${response.status}: ${errorText}`);
        }
        
        const savedApp = await response.json();
        console.log('Application saved:', savedApp);
        showToast('Application saved successfully', 'success');
        
        // Upload pending files if any
        if (pendingFiles.length > 0 && savedApp.id) {
            showToast(`Uploading ${pendingFiles.length} pending file(s)...`, 'success');
            for (const fileObj of pendingFiles) {
                try {
                    await uploadFile(savedApp.id, fileObj.file, fileObj.description);
                } catch (error) {
                    console.error('Failed to upload pending file:', error);
                    showToast(`Failed to upload: ${fileObj.file.name}`, 'error');
                }
            }
            pendingFiles = [];
            showToast('All files uploaded successfully', 'success');
        }
        
        // Force reload applications
        console.log('Reloading applications...');
        await loadApplications();
        console.log('Applications reloaded, closing modal...');
        closeModal();
    } catch (error) {
        console.error('Failed to save application:', error);
        showToast(`Save failed: ${error.message}`, 'error');
    }
});

function toggleEditMode() {
    const toggle = document.getElementById('edit-mode-toggle');
    isEditMode = toggle.checked;
    
    const newAppBtn = document.getElementById('new-app-btn');
    const actionsHeader = document.getElementById('actions-header');
    
    if (isEditMode) {
        newAppBtn.disabled = false;
        actionsHeader.style.display = 'table-cell';
        showToast('Edit mode enabled', 'success');
    } else {
        newAppBtn.disabled = true;
        actionsHeader.style.display = 'none';
        showToast('View mode enabled', 'success');
        
        // Auto-save when exiting edit mode
        saveAllChanges();
    }
    
    renderApplications();
}

async function saveAllChanges() {
    try {
        const response = await fetch('/api/save', { method: 'POST' });
        if (response.ok) {
            updateStatusBar();
        }
    } catch (error) {
        console.error('Failed to save changes:', error);
    }
}

function getStatusLabel(status) {
    const statusLabels = {
        'applied': 'APPLIED',
        'screening': 'SCREENING',
        'interview1': 'INTERVIEW I',
        'interview2': 'INTERVIEW II',
        'interview3': 'INTERVIEW III',
        'final': 'FINAL',
        'offer': 'OFFER',
        'rejected': 'REJECTED',
        'withdrawn': 'WITHDRAWN'
    };
    return statusLabels[status] || status.toUpperCase();
}

function getFinalResult(finalResult) {
    if (finalResult === 'accepted') return '<span class="result-text result-accepted">ACCEPTED</span>';
    if (finalResult === 'rejected') return '<span class="result-text result-rejected">REJECTED</span>';
    if (finalResult === 'withdrawn') return '<span class="result-text result-withdrawn">WITHDRAWN</span>';
    return '<span class="result-text result-pending">PENDING</span>';
}

function formatDateWithDay(dateString) {
    if (!dateString) return '-';
    
    const date = new Date(dateString);
    if (isNaN(date.getTime())) return dateString;
    
    const dayNames = ['일', '월', '화', '수', '목', '금', '토'];
    const dayOfWeek = dayNames[date.getDay()];
    
    const year = date.getFullYear();
    const month = String(date.getMonth() + 1).padStart(2, '0');
    const day = String(date.getDate()).padStart(2, '0');
    
    return `${year}-${month}-${day} (${dayOfWeek})`;
}

function escapeHtml(text) {
    if (!text) return '';
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

function escapeHtmlPreserveNewlines(text) {
    if (!text) return '';
    // First escape HTML, then replace newlines with <br>
    const escaped = escapeHtml(text);
    return escaped.replace(/\n/g, '<br>');
}

let tooltipElement = null;

function showTooltip(event, text) {
    hideTooltip(); // Remove any existing tooltip
    
    tooltipElement = document.createElement('div');
    tooltipElement.className = 'tooltip';
    // Support multiline by preserving line breaks and escaping HTML
    const safeText = escapeHtml(text).replace(/\n/g, '<br>');
    tooltipElement.innerHTML = safeText;
    document.body.appendChild(tooltipElement);
    
    const rect = event.target.getBoundingClientRect();
    const tooltipRect = tooltipElement.getBoundingClientRect();
    
    let top = rect.top - tooltipRect.height - 10;
    let left = rect.left + rect.width / 2 - tooltipRect.width / 2;
    
    // Adjust if tooltip goes off screen
    if (top < 10) {
        top = rect.bottom + 10;
    }
    if (left < 10) {
        left = 10;
    }
    if (left + tooltipRect.width > window.innerWidth - 10) {
        left = window.innerWidth - tooltipRect.width - 10;
    }
    
    tooltipElement.style.top = top + 'px';
    tooltipElement.style.left = left + 'px';
    tooltipElement.style.opacity = '1';
}

function hideTooltip() {
    if (tooltipElement) {
        tooltipElement.remove();
        tooltipElement = null;
    }
}

function getFilesDisplay(app) {
    const files = app.files || [];
    if (files.length === 0) {
        return '<span class="file-count file-count-zero">0</span>';
    }
    
    const fileLinks = files.map(file => {
        const filename = typeof file === 'string' ? file : file.filename;
        const description = typeof file === 'string' ? '' : (file.description || '');
        const size = typeof file === 'string' ? 0 : (file.size || 0);
        const sizeText = size > 0 ? ` (${formatFileSize(size)})` : '';
        const titleText = description ? `${filename} - ${description}` : filename;
        const descriptionText = description ? ` - ${description}` : '';
        return `<a href="/api/download?app_id=${app.id}&filename=${filename}" class="file-link" title="${titleText}">${filename}</a><span class="file-description">${descriptionText}</span><span class="file-size">${sizeText}</span>`;
    }).join('<br>');
    
    return `<div class="file-display">
        <span class="file-count">${files.length}</span>
        <div class="file-dropdown">${fileLinks}</div>
    </div>`;
}

function displayFiles(files) {
    const fileList = document.getElementById('file-list');
    fileList.innerHTML = '';
    
    files.forEach(file => {
        const filename = typeof file === 'string' ? file : file.filename;
        const description = typeof file === 'string' ? '' : (file.description || '');
        
        const fileItem = document.createElement('div');
        fileItem.className = 'file-item';
        fileItem.innerHTML = `
            <span class="file-name">${escapeHtml(filename)}</span>
            <input type="text" class="file-desc-input" placeholder="Add description..." 
                   value="${escapeHtml(description)}" 
                   data-filename="${escapeHtml(filename)}">
            <button type="button" class="file-delete" data-filename="${escapeHtml(filename)}">&times;</button>
        `;
        
        // Add event listeners programmatically to avoid quote issues
        const input = fileItem.querySelector('.file-desc-input');
        const deleteBtn = fileItem.querySelector('.file-delete');
        
        input.addEventListener('blur', function() {
            updateFileDescription(filename, this.value);
        });
        
        input.addEventListener('keypress', function(event) {
            if (event.key === 'Enter') {
                this.blur();
            }
        });
        
        deleteBtn.addEventListener('click', function() {
            deleteFile(filename);
        });
        fileList.appendChild(fileItem);
    });
}

function displayPendingFiles(pendingFileList) {
    const fileList = document.getElementById('file-list');
    fileList.innerHTML = '';
    
    pendingFileList.forEach((fileObj, index) => {
        const filename = fileObj.file ? fileObj.file.name : fileObj.name;
        
        const fileItem = document.createElement('div');
        fileItem.className = 'file-item pending';
        fileItem.innerHTML = `
            <span class="file-name">${escapeHtml(filename)} (pending)</span>
            <span style="flex: 1;"></span>
            <button type="button" class="file-delete" data-index="${index}">&times;</button>
        `;
        
        // Add event listener programmatically
        const deleteBtn = fileItem.querySelector('.file-delete');
        deleteBtn.addEventListener('click', function() {
            removePendingFile(index);
        });
        
        fileList.appendChild(fileItem);
    });
}

async function uploadFile(appId, file, description = '') {
    const formData = new FormData();
    formData.append('file', file);
    formData.append('app_id', appId);
    formData.append('description', description);
    
    try {
        const response = await fetch('/api/upload', {
            method: 'POST',
            body: formData
        });
        
        if (!response.ok) {
            throw new Error('Upload failed');
        }
        
        const result = await response.json();
        showToast(`File uploaded: ${result.filename}`, 'success');
        return result.filename;
    } catch (error) {
        showToast('File upload failed', 'error');
        throw error;
    }
}

async function deleteFile(filename) {
    const appId = document.getElementById('app-id').value;
    if (!appId || !confirm(`Delete file: ${filename}?`)) {
        return;
    }
    
    try {
        await fetch(`/api/file/delete?app_id=${appId}&filename=${filename}`, {
            method: 'DELETE'
        });
        
        // Remove from UI
        const app = applications.find(a => a.id === appId);
        if (app) {
            app.files = app.files.filter(f => {
                const fname = typeof f === 'string' ? f : f.filename;
                return fname !== filename;
            });
            displayFiles(app.files);
        }
        
        showToast('File deleted successfully', 'success');
    } catch (error) {
        showToast('Failed to delete file', 'error');
    }
}

function removePendingFile(index) {
    pendingFiles.splice(index, 1);
    displayPendingFiles(pendingFiles);
}

async function updateFileDescription(filename, description) {
    const appId = document.getElementById('app-id').value;
    console.log(`Updating description for file: ${filename}, appId: ${appId}, description: "${description}"`);
    
    if (!appId) {
        showToast('Cannot update description: application not saved', 'error');
        return;
    }
    
    try {
        const url = `/api/file/description?app_id=${appId}&filename=${encodeURIComponent(filename)}`;
        console.log(`Sending PUT request to: ${url}`);
        
        const response = await fetch(url, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ description: description })
        });
        
        console.log(`Response status: ${response.status}`);
        
        if (!response.ok) {
            const errorText = await response.text();
            console.error(`Error response: ${errorText}`);
            throw new Error(`Failed to update description: ${response.status}`);
        }
        
        const result = await response.json();
        console.log('Update successful:', result);
        
        // Update local data
        const app = applications.find(a => a.id === appId);
        if (app && app.files) {
            const fileObj = app.files.find(f => 
                (typeof f === 'string' ? f : f.filename) === filename
            );
            if (fileObj && typeof fileObj === 'object') {
                fileObj.description = description;
                console.log('Local data updated');
            }
        }
        
        showToast('Description updated', 'success');
        
        // Reload applications to reflect changes
        await loadApplications();
    } catch (error) {
        console.error('Failed to update file description:', error);
        showToast(`Failed to update description: ${error.message}`, 'error');
    }
}

// Handle file upload input
document.getElementById('file-upload').addEventListener('change', async (e) => {
    const appId = document.getElementById('app-id').value;
    const files = Array.from(e.target.files);
    
    if (!appId) {
        // For new applications, store files temporarily
        pendingFiles = pendingFiles || [];
        pendingFiles.push(...files.map(f => ({ file: f })));
        
        displayPendingFiles(pendingFiles);
        showToast(`${files.length} file(s) selected. Save application to upload.`, 'success');
        return;
    }
    
    // For existing applications, upload immediately
    for (const file of files) {
        try {
            await uploadFile(appId, file);
        } catch (error) {
            console.error('Upload error:', error);
        }
    }
    
    // Reload applications to show new files
    await loadApplications();
    
    // Clear the input
    e.target.value = '';
});

// Handle ESC key to close modal
document.addEventListener('keydown', (e) => {
    if (e.key === 'Escape') {
        const modal = document.getElementById('application-modal');
        if (modal.classList.contains('active')) {
            closeModal();
        }
    }
});

// Format file size helper function
function formatFileSize(bytes) {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
}

// Load applications on page load
loadApplications();