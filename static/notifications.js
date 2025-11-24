// Notification utility function
function showNotification(message, type = 'success') {
    const notification = document.createElement('div');
    notification.className = `notification ${type}`;
    notification.textContent = message;
    document.body.appendChild(notification);
    setTimeout(() => notification.remove(), 5000);
}

// Load notification methods
window.loadMethods = async function() {
    try {
        const profileId = localStorage.getItem('profile_id');
        if (!profileId) {
            console.log('No profile ID found, skipping method loading');
            return;
        }

        const response = await fetch('/api/notifications/methods', {
            headers: {
                'X-Profile-ID': profileId
            }
        });

        if (!response.ok) {
            throw new Error('Failed to load notification methods');
        }

        const methods = await response.json();
        console.log(`Loaded ${methods.length} notification methods`);

        // Optional: Update total methods count if element exists
        const totalMethodsEl = document.getElementById('totalMethods');
        if (totalMethodsEl) {
            totalMethodsEl.innerHTML = `<i class="fas fa-bell" style="color: var(--accent-color); margin-right: 0.75rem; font-size: 2rem;"></i> ${methods.length}`;
        }

        // Optional: Populate methods list if element exists
        const methodsListEl = document.getElementById('methodsList');
        if (methodsListEl) {
            if (methods.length === 0) {
                methodsListEl.innerHTML = `
                    <tr>
                        <td colspan="3" style="padding: 2rem; text-align: center; color: var(--text-secondary);">
                            <i class="fas fa-bell-slash" style="font-size: 2rem; margin-bottom: 1rem; display: block;"></i>
                            <span style="font-size: 1.2rem; font-weight: 500;">No notification methods found</span>
                            <p style="margin-top: 0.5rem; color: #888;">Click "Add Method" to create your first notification method.</p>
                        </td>
                    </tr>
                `;
                return;
            }
            
            methodsListEl.innerHTML = methods.map(method => {
                const config = typeof method.config === 'string' ? JSON.parse(method.config) : method.config;
                const status = method.status || 'active';
                
                // Create status badge with appropriate styling
                let statusIcon = '';
                let statusClass = '';
                let statusBadge = '';
                
                if (status === 'unauthorized') {
                    statusIcon = '<i class="fas fa-lock" style="font-size: 1.1rem; margin-right: 0.5rem;"></i>';
                    statusClass = 'status-unauthorized';
                    statusBadge = `<span style="display: inline-block; padding: 0.25rem 0.5rem; border-radius: 0.25rem; background: linear-gradient(135deg, rgba(255, 140, 0, 0.1) 0%, rgba(255, 140, 0, 0.2) 100%); border: 1px solid var(--orange-accent); color: var(--orange-accent); font-weight: bold; text-transform: uppercase; letter-spacing: 1px; font-size: 0.75rem;">${statusIcon} UNAUTHORIZED</span>`;
                    
                    // Send notification for unauthorized status
                    sendUnauthorizedNotification(method.name || method.type, config.webhook_url);
                } else {
                    statusIcon = '<i class="fas fa-check-circle" style="font-size: 1.1rem; margin-right: 0.5rem;"></i>';
                    statusClass = 'status-up';
                    statusBadge = `<span style="display: inline-block; padding: 0.25rem 0.5rem; border-radius: 0.25rem; background: linear-gradient(135deg, rgba(40, 167, 69, 0.1) 0%, rgba(40, 167, 69, 0.2) 100%); border: 1px solid var(--status-up); color: var(--status-up); font-weight: bold; text-transform: uppercase; letter-spacing: 1px; font-size: 0.75rem;">${statusIcon} ACTIVE</span>`;
                }
                
                // Get icon for method type
                let methodIcon = '';
                switch (method.type.toLowerCase()) {
                    case 'email':
                        methodIcon = '<i class="fas fa-envelope" style="color: var(--accent-color); margin-right: 0.75rem; font-size: 1.2rem;"></i>';
                        break;
                    case 'slack':
                        methodIcon = '<i class="fab fa-slack" style="color: var(--accent-color); margin-right: 0.75rem; font-size: 1.2rem;"></i>';
                        break;
                    case 'teams':
                        methodIcon = '<i class="fas fa-users" style="color: var(--accent-color); margin-right: 0.75rem; font-size: 1.2rem;"></i>';
                        break;
                    case 'webhook':
                        methodIcon = '<i class="fas fa-link" style="color: var(--accent-color); margin-right: 0.75rem; font-size: 1.2rem;"></i>';
                        break;
                    default:
                        methodIcon = '<i class="fas fa-bell" style="color: var(--accent-color); margin-right: 0.75rem; font-size: 1.2rem;"></i>';
                }
                
                // Create method type badge
                const methodTypeBadge = `<span style="display: inline-block; padding: 0.25rem 0.5rem; border-radius: 0.25rem; background: linear-gradient(135deg, rgba(197, 165, 114, 0.1) 0%, rgba(197, 165, 114, 0.2) 100%); border: 1px solid var(--accent-color); color: var(--accent-color); font-weight: bold; text-transform: uppercase; letter-spacing: 1px; font-size: 0.75rem;">${methodIcon}${method.type.toUpperCase()}</span>`;

                return `
                    <tr style="transition: all 0.3s ease; border-bottom: 1px solid var(--border-color);">
                        <td style="padding: 1.25rem 1rem; background: linear-gradient(135deg, rgba(42, 42, 42, 0.5) 0%, rgba(42, 42, 42, 0.8) 100%);">
                            <div style="display: flex; flex-direction: column; gap: 0.75rem;">
                                ${methodTypeBadge}
                                ${statusBadge}
                            </div>
                        </td>
                        <td style="padding: 1.25rem 1rem; background: linear-gradient(135deg, rgba(42, 42, 42, 0.5) 0%, rgba(42, 42, 42, 0.8) 100%);">
                            <div style="display: flex; flex-direction: column; gap: 0.5rem;">
                                <strong style="font-size: 1.1rem; color: var(--text-secondary); font-weight: 600;">${method.name || method.type}</strong>
                                <div style="color: var(--text-primary); font-size: 0.9rem;">
                                    ${formatMethodDetails(method)}
                                </div>
                            </div>
                        </td>
                        <td style="padding: 1.25rem 1rem; background: linear-gradient(135deg, rgba(42, 42, 42, 0.5) 0%, rgba(42, 42, 42, 0.8) 100%);">
                            <div style="display: flex; gap: 0.75rem;">
                                <button onclick="editMethod('${method.id}')" class="btn btn-small" style="display: flex; align-items: center; justify-content: center; gap: 0.5rem; padding: 0.5rem 1rem; background: linear-gradient(135deg, #c5a572 0%, #b38b5d 100%); border-radius: 0.375rem; border: none; color: white; font-weight: 600; cursor: pointer; transition: all 0.2s; text-transform: uppercase; letter-spacing: 1px; font-size: 0.8rem; box-shadow: 0 2px 4px rgba(197, 165, 114, 0.3);">
                                    <i class="fas fa-edit"></i> Edit
                                </button>
                                <button onclick="deleteMethod('${method.id}')" class="btn btn-small btn-danger" style="display: flex; align-items: center; justify-content: center; gap: 0.5rem; padding: 0.5rem 1rem; background: linear-gradient(135deg, #dc3545 0%, #c82333 100%); border-radius: 0.375rem; border: none; color: white; font-weight: 600; cursor: pointer; transition: all 0.2s; text-transform: uppercase; letter-spacing: 1px; font-size: 0.8rem; box-shadow: 0 2px 4px rgba(220, 53, 69, 0.3);">
                                    <i class="fas fa-trash"></i> Delete
                                </button>
                            </div>
                        </td>
                    </tr>
                `;
            }).join('');
        }
    } catch (error) {
        console.error('Error loading methods:', error);
        showNotification('Failed to load notification methods', 'error');
    }
};

// Format method details for display
function formatMethodDetails(method) {
    try {
        // Parse the config if it's a string
        const config = typeof method.config === 'string' ? JSON.parse(method.config) : method.config;
        
        switch (method.type) {
            case 'email':
                return `${config.smtp_email} → ${config.recipient_email}`;
            case 'slack':
                return `${config.channel}`;
            case 'teams':
                return 'Teams Webhook';
            default:
                return '';
        }
    } catch (error) {
        console.error('Error parsing config:', error);
        return 'Configuration error';
    }
}

// Delete method function
async function deleteMethod(methodId) {
    if (!confirm('Are you sure you want to delete this notification method?')) {
        return;
    }

    try {
        const profileId = localStorage.getItem('profile_id');
        const response = await fetch(`/api/notifications/methods/${methodId}`, {
            method: 'DELETE',
            headers: {
                'Content-Type': 'application/json',
                'X-Profile-ID': profileId
            }
        });

        if (!response.ok) {
            throw new Error('Failed to delete notification method');
        }

        showNotification('Notification method deleted successfully');
        loadMethods();
    } catch (error) {
        console.error('Error deleting method:', error);
        showNotification('Failed to delete notification method', 'error');
    }
}

// Placeholder for sendUnauthorizedNotification function
function sendUnauthorizedNotification(name, url) {
    console.warn(`Unauthorized notification method: ${name} (${url})`);
}

// Placeholder for editMethod function
function editMethod(methodId) {
    console.log(`Editing method: ${methodId}`);
}

// Prevent multiple event listeners
const loadMethodsHandler = function() {
    loadMethods();
};

// Remove any existing listeners first
document.removeEventListener('DOMContentLoaded', loadMethodsHandler);
document.addEventListener('DOMContentLoaded', loadMethodsHandler); 