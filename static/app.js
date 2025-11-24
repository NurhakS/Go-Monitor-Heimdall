// Global variables
let currentProfileId = null;

function getCurrentProfileId() {
    console.group('🔍 Profile ID Determination');
    
    // First, try to get from localStorage
    const storedProfileId = localStorage.getItem('profile_id');
    if (storedProfileId) {
        console.log('✅ Profile ID from localStorage:', storedProfileId);
        console.groupEnd();
        return storedProfileId;
    }

    // If not in localStorage, try to find from the profile select element
const profileSelect = document.getElementById('profileSelect');
    if (profileSelect) {
        console.log('Profile Select Element:', profileSelect);
        console.log('Profile Select Value:', profileSelect.value);
        
        // Check for a selected option
        const selectedOption = profileSelect.options[profileSelect.selectedIndex];
        if (selectedOption && selectedOption.value) {
            console.log('✅ Selected Profile Option:', selectedOption.value);
            console.groupEnd();
            return selectedOption.value;
        }
    }

    // Log more diagnostic information
    console.error('❌ Profile selection diagnostics:');
    console.error('localStorage profile_id:', localStorage.getItem('profile_id'));
    console.error('profileSelect element:', profileSelect);
    console.error('profileSelect value:', profileSelect ? profileSelect.value : 'No select element');

    // If no profile is found, show a more informative notification
    showNotification('No active profile. Please create or select a profile.', 'error');
    console.groupEnd();
    return null;
}

// Function to show notifications
function showNotification(message, type = 'info') {
    console.log(`Notification (${type}): ${message}`);
    
    // Create notification element
    const notification = document.createElement('div');
    notification.className = `notification ${type}`;
    notification.textContent = message;
    
    // Add to document
    document.body.appendChild(notification);
    
    // Remove after 5 seconds
    setTimeout(() => {
        notification.style.opacity = '0';
        setTimeout(() => {
            document.body.removeChild(notification);
        }, 300);
    }, 5000);
}

async function loadCredentials() {
    console.group('🔍 Loading Credentials');
    try {
        const profileId = getCurrentProfileId();
        if (!profileId) {
            console.warn('No active profile found. Skipping credential load.');
            console.groupEnd();
            return;
        }

        const response = await fetch(`/api/credentials?profileId=${encodeURIComponent(profileId)}`, {
            headers: {
                'X-Profile-ID': profileId,
                'Accept': 'application/json'
            }
        });

        if (!response.ok) {
            const errorText = await response.text();
            throw new Error(`Failed to load credentials: ${errorText}`);
        }

        const credentials = await response.json();
        console.log('Loaded Credentials:', credentials);

        // Update credential select in the UI
        const credentialSelect = document.getElementById('credentialSelect');
        if (credentialSelect) {
            if (credentials.length === 0) {
                credentialSelect.innerHTML = `
                    <option value="">No Credentials Available</option>
                `;
            } else {
                credentialSelect.innerHTML = `
                    <option value="">No Credential</option>
                    ${credentials.map(cred => {
                        let icon = '';
                        switch(cred.type.toLowerCase()) {
                            case 'bearer':
                                icon = 'key';
                                break;
                            case 'basic':
                                icon = 'user-lock';
                                break;
                            case 'oauth2':
                                icon = 'shield-alt';
                                break;
                            case 'api':
                                icon = 'code';
                                break;
                            default:
                                icon = 'lock';
                        }
                        return `
                            <option value="${cred.id}">
                                <i class="fas fa-${icon}"></i> ${cred.name} (${cred.type})
                            </option>
                        `;
                    }).join('')}
                `;
            }
            console.log('Credential select updated with options:', credentialSelect.innerHTML);
        }

        console.groupEnd();
    } catch (error) {
        console.error('Error loading credentials:', error);
        console.groupEnd();
    }
}

function clearCredentialSelect() {
    const credentialSelect = document.getElementById('credentialSelect');
    if (credentialSelect) {
        credentialSelect.innerHTML = `
            <option value="">Loading credentials...</option>
        `;
    }
}

function clearMonitorList() {
    const monitorList = document.getElementById('monitor-list');
    if (monitorList) {
        monitorList.innerHTML = `
            <tr>
                <td colspan="6" style="text-align: center;">Loading monitors...</td>
            </tr>
        `;
    }
}

// Updated loadMonitors function with enhanced styling and icons
async function loadMonitors() {
    console.group('🔍 Loading Monitors');
    try {
        const profileId = getCurrentProfileId();
        if (!profileId) {
            console.warn('No active profile found. Skipping monitor load.');
            console.groupEnd();
            return;
        }

        const response = await fetch(`/api/monitors?profileId=${encodeURIComponent(profileId)}`, {
            headers: {
                'X-Profile-ID': profileId,
                'Accept': 'application/json'
            }
        });

        if (!response.ok) {
            const errorText = await response.text();
            throw new Error(`Failed to load monitors: ${errorText}`);
        }

        const monitors = await response.json();
        console.log('Loaded Monitors:', monitors);

        // Update monitor list in the UI with enhanced styling
        const monitorList = document.getElementById('monitor-list');
        if (monitorList) {
            monitorList.innerHTML = monitors.map(monitor => {
                // Determine status icon and class with enhanced styling
                let statusIcon = '';
                let statusClass = '';
                let statusBadge = '';
                
                switch(monitor.status.toLowerCase()) {
                    case 'up':
                        statusIcon = '<i class="fas fa-check-circle" style="font-size: 1.2rem; margin-right: 0.5rem;"></i>';
                        statusClass = 'status-up';
                        statusBadge = `<span style="display: inline-block; padding: 0.25rem 0.5rem; border-radius: 0.25rem; background: linear-gradient(135deg, rgba(40, 167, 69, 0.1) 0%, rgba(40, 167, 69, 0.2) 100%); border: 1px solid var(--status-up); color: var(--status-up); font-weight: bold; text-transform: uppercase; letter-spacing: 1px; font-size: 0.75rem;">${statusIcon} ONLINE</span>`;
                        break;
                    case 'down':
                        statusIcon = '<i class="fas fa-times-circle" style="font-size: 1.2rem; margin-right: 0.5rem;"></i>';
                        statusClass = 'status-down';
                        statusBadge = `<span style="display: inline-block; padding: 0.25rem 0.5rem; border-radius: 0.25rem; background: linear-gradient(135deg, rgba(220, 53, 69, 0.1) 0%, rgba(220, 53, 69, 0.2) 100%); border: 1px solid var(--status-down); color: var(--status-down); font-weight: bold; text-transform: uppercase; letter-spacing: 1px; font-size: 0.75rem;">${statusIcon} OFFLINE</span>`;
                        break;
                    case 'pending':
                        statusIcon = '<i class="fas fa-clock" style="font-size: 1.2rem; margin-right: 0.5rem;"></i>';
                        statusClass = 'status-pending';
                        statusBadge = `<span style="display: inline-block; padding: 0.25rem 0.5rem; border-radius: 0.25rem; background: linear-gradient(135deg, rgba(255, 193, 7, 0.1) 0%, rgba(255, 193, 7, 0.2) 100%); border: 1px solid var(--status-pending); color: var(--status-pending); font-weight: bold; text-transform: uppercase; letter-spacing: 1px; font-size: 0.75rem;">${statusIcon} PENDING</span>`;
                        break;
                    case 'unauthorized':
                        statusIcon = '<i class="fas fa-lock" style="font-size: 1.2rem; margin-right: 0.5rem;"></i>';
                        statusClass = 'status-unauthorized';
                        statusBadge = `<span style="display: inline-block; padding: 0.25rem 0.5rem; border-radius: 0.25rem; background: linear-gradient(135deg, rgba(255, 140, 0, 0.1) 0%, rgba(255, 140, 0, 0.2) 100%); border: 1px solid var(--orange-accent); color: var(--orange-accent); font-weight: bold; text-transform: uppercase; letter-spacing: 1px; font-size: 0.75rem;">${statusIcon} UNAUTHORIZED</span>`;
                        break;
                    default:
                        statusIcon = '<i class="fas fa-question-circle" style="font-size: 1.2rem; margin-right: 0.5rem;"></i>';
                        statusClass = 'status-pending';
                        statusBadge = `<span style="display: inline-block; padding: 0.25rem 0.5rem; border-radius: 0.25rem; background: linear-gradient(135deg, rgba(255, 193, 7, 0.1) 0%, rgba(255, 193, 7, 0.2) 100%); border: 1px solid var(--status-pending); color: var(--status-pending); font-weight: bold; text-transform: uppercase; letter-spacing: 1px; font-size: 0.75rem;">${statusIcon} UNKNOWN</span>`;
                }
                
                // Format check time and failure count with enhanced styling
                const lastChecked = monitor.last_checked ? new Date(monitor.last_checked).toLocaleString() : 'Never';
                const failureCount = monitor.failure_count || 0;
                const failureThreshold = monitor.failure_threshold || 1;
                const checkTime = monitor.check_interval || 60;
                
                // Create method badge with appropriate color based on HTTP method
                let methodColor = '#007bff'; // Default blue
                switch(monitor.method) {
                    case 'GET':
                        methodColor = '#28a745'; // Green
                        break;
                    case 'POST':
                        methodColor = '#ff8c00'; // Orange
                        break;
                    case 'PUT':
                        methodColor = '#6f42c1'; // Purple
                        break;
                    case 'DELETE':
                        methodColor = '#dc3545'; // Red
                        break;
                    case 'PATCH':
                        methodColor = '#17a2b8'; // Teal
                        break;
                }
                
                const methodBadge = `<span style="display: inline-block; padding: 0.15rem 0.4rem; border-radius: 0.25rem; background: linear-gradient(135deg, rgba(${hexToRgb(methodColor)}, 0.1) 0%, rgba(${hexToRgb(methodColor)}, 0.2) 100%); border: 1px solid ${methodColor}; color: ${methodColor}; font-weight: bold; text-transform: uppercase; letter-spacing: 1px; font-size: 0.7rem; margin-right: 0.5rem;">${monitor.method}</span>`;
                
                const requestTypeBadge = `<span style="display: inline-block; padding: 0.15rem 0.4rem; border-radius: 0.25rem; background: linear-gradient(135deg, rgba(0, 123, 255, 0.1) 0%, rgba(0, 123, 255, 0.2) 100%); border: 1px solid #007bff; color: #007bff; font-weight: bold; text-transform: uppercase; letter-spacing: 1px; font-size: 0.7rem;">${monitor.request_type || 'HTTP'}</span>`;
                
                // Create a gradient card-like style for each monitor row
                return `
                    <tr style="transition: all 0.3s ease; border-bottom: 1px solid var(--border-color);">
                        <td style="padding: 1.25rem 1rem; background: linear-gradient(135deg, rgba(42, 42, 42, 0.5) 0%, rgba(42, 42, 42, 0.8) 100%);">
                            <div style="display: flex; align-items: center; margin-bottom: 0.5rem;">
                                <strong style="font-size: 1.2rem; color: var(--text-secondary); background: var(--metallic-gold); -webkit-background-clip: text; -webkit-text-fill-color: transparent; font-weight: 700;">${monitor.name}</strong>
                            </div>
                            
                            <div style="display: flex; flex-wrap: wrap; gap: 0.5rem; margin-bottom: 0.75rem;">
                                ${methodBadge}
                                ${requestTypeBadge}
                            </div>
                            
                            <div style="margin-bottom: 0.75rem; background: rgba(26, 26, 26, 0.5); padding: 0.5rem; border-radius: 0.375rem; border: 1px solid var(--border-color);">
                                <small style="display: flex; align-items: center; color: #aaa; word-break: break-all;">
                                    <i class="fas fa-link" style="color: var(--accent-color); margin-right: 0.5rem; min-width: 16px;"></i>
                                    <span style="color: var(--text-primary); font-weight: 600;">${monitor.url}</span>
                                </small>
                            </div>
                            
                            <div style="display: flex; flex-wrap: wrap; gap: 0.75rem; margin-top: 0.75rem; background: rgba(26, 26, 26, 0.3); padding: 0.5rem; border-radius: 0.375rem;">
                                <small style="display: flex; align-items: center; color: var(--text-secondary); font-weight: 600;">
                                    <i class="fas fa-clock" style="color: var(--accent-color); margin-right: 0.5rem; min-width: 16px;"></i> Check: ${checkTime}s
                                </small>
                                <small style="display: flex; align-items: center; color: ${failureCount > 0 ? 'var(--status-down)' : 'var(--text-secondary)'}; font-weight: 600;">
                                    <i class="fas fa-exclamation-triangle" style="color: ${failureCount > 0 ? 'var(--status-down)' : 'var(--warning-color)'}; margin-right: 0.5rem; min-width: 16px;"></i> Failures: ${failureCount}/${failureThreshold}
                                </small>
                            </div>
                        </td>
                        <td style="padding: 1.25rem 1rem; text-align: center; vertical-align: middle; background: linear-gradient(135deg, rgba(42, 42, 42, 0.5) 0%, rgba(42, 42, 42, 0.8) 100%);">
                            ${statusBadge}
                        </td>
                        <td style="padding: 1.25rem 1rem; text-align: center; vertical-align: middle; background: linear-gradient(135deg, rgba(42, 42, 42, 0.5) 0%, rgba(42, 42, 42, 0.8) 100%);">
                            <span style="font-size: 1.1rem; font-weight: bold; color: var(--text-secondary);">${monitor.response_time || 'N/A'}</span>
                            <small style="display: block; color: #aaa; margin-top: 0.25rem;">milliseconds</small>
                        </td>
                        <td style="padding: 1.25rem 1rem; vertical-align: middle; background: linear-gradient(135deg, rgba(42, 42, 42, 0.5) 0%, rgba(42, 42, 42, 0.8) 100%);">
                            <span style="color: var(--text-secondary);">${lastChecked}</span>
                        </td>
                        <td style="padding: 1.25rem 1rem; text-align: center; vertical-align: middle; background: linear-gradient(135deg, rgba(42, 42, 42, 0.5) 0%, rgba(42, 42, 42, 0.8) 100%);">
                            <span style="display: inline-block; padding: 0.25rem 0.5rem; border-radius: 0.25rem; background: ${monitor.response_code >= 200 && monitor.response_code < 300 ? 'linear-gradient(135deg, rgba(40, 167, 69, 0.1) 0%, rgba(40, 167, 69, 0.2) 100%)' : monitor.response_code >= 400 ? 'linear-gradient(135deg, rgba(220, 53, 69, 0.1) 0%, rgba(220, 53, 69, 0.2) 100%)' : 'linear-gradient(135deg, rgba(255, 193, 7, 0.1) 0%, rgba(255, 193, 7, 0.2) 100%)'}; 
                                  border: 1px solid ${monitor.response_code >= 200 && monitor.response_code < 300 ? 'var(--status-up)' : monitor.response_code >= 400 ? 'var(--status-down)' : 'var(--status-pending)'}; 
                                  color: ${monitor.response_code >= 200 && monitor.response_code < 300 ? 'var(--status-up)' : monitor.response_code >= 400 ? 'var(--status-down)' : 'var(--status-pending)'};
                                  font-weight: bold; font-size: 0.9rem;">
                                ${monitor.response_code || 'N/A'}
                            </span>
                        </td>
                        <td style="padding: 1.25rem 1rem; vertical-align: middle; background: linear-gradient(135deg, rgba(42, 42, 42, 0.5) 0%, rgba(42, 42, 42, 0.8) 100%);">
                            <div style="display: flex; flex-direction: column; gap: 0.5rem;">
                                <button onclick="deleteMonitor('${monitor.id}')" class="btn btn-small btn-danger" style="display: flex; align-items: center; justify-content: center; gap: 0.5rem; padding: 0.5rem 1rem; background: linear-gradient(135deg, #dc3545 0%, #c82333 100%); border-radius: 0.375rem; border: none; color: white; font-weight: 600; cursor: pointer; transition: all 0.2s; text-transform: uppercase; letter-spacing: 1px; font-size: 0.8rem; box-shadow: 0 2px 4px rgba(220, 53, 69, 0.3);">
                                    <i class="fas fa-trash"></i> Delete
                                </button>
                                ${failureCount > 0 ? `
                                <button onclick="resetFailures('${monitor.id}')" class="btn btn-small" style="display: flex; align-items: center; justify-content: center; gap: 0.5rem; padding: 0.5rem 1rem; background: linear-gradient(135deg, #c5a572 0%, #b38b5d 100%); border-radius: 0.375rem; border: none; color: white; font-weight: 600; cursor: pointer; transition: all 0.2s; text-transform: uppercase; letter-spacing: 1px; font-size: 0.8rem; box-shadow: 0 2px 4px rgba(197, 165, 114, 0.3);">
                                    <i class="fas fa-redo"></i> Reset Failures
                                </button>
                                ` : ''}
                            </div>
                        </td>
                    </tr>
                `;
            }).join('');
        }

        // Update monitor stats with enhanced styling
        const totalMonitors = document.getElementById('totalMonitors');
        const upMonitors = document.getElementById('upMonitors');
        const downMonitors = document.getElementById('downMonitors');

        // Count monitors by status
        const upCount = monitors.filter(m => m.status.toLowerCase() === 'up').length;
        const downCount = monitors.filter(m => m.status.toLowerCase() === 'down').length;
        const totalCount = monitors.length;

        // Update the UI elements with counts and icons
        if (totalMonitors) {
            totalMonitors.innerHTML = `<i class="fas fa-chart-line" style="color: var(--accent-color); margin-right: 0.5rem;"></i> ${totalCount}`;
            console.log('Updated total monitors count:', totalCount);
        }
        
        if (upMonitors) {
            upMonitors.innerHTML = `<i class="fas fa-check-circle" style="color: var(--status-up); margin-right: 0.5rem;"></i> ${upCount}`;
            console.log('Updated up monitors count:', upCount);
        }
        
        if (downMonitors) {
            downMonitors.innerHTML = `<i class="fas fa-times-circle" style="color: var(--status-down); margin-right: 0.5rem;"></i> ${downCount}`;
            console.log('Updated down monitors count:', downCount);
        }

        console.groupEnd();
    } catch (error) {
        console.error('Error loading monitors:', error);
        console.groupEnd();
    }
}

// Helper function to convert hex to rgb
function hexToRgb(hex) {
    // Remove the # if present
    hex = hex.replace('#', '');
    
    // Parse the hex values
    const r = parseInt(hex.substring(0, 2), 16);
    const g = parseInt(hex.substring(2, 4), 16);
    const b = parseInt(hex.substring(4, 6), 16);
    
    // Return the RGB values as a string
    return `${r}, ${g}, ${b}`;
}

async function loadProfiles() {
    try {
        const response = await fetch('/api/profiles');
        if (!response.ok) throw new Error('Failed to load profiles');
        
        const profiles = await response.json();
        console.log('Loaded Profiles:', profiles);

        const profileSelect = document.getElementById('profileSelect');
        profileSelect.innerHTML = `
            <option value="">Select Profile</option>
            ${profiles.map(p => `
                <option value="${p.id}" ${p.is_active ? 'selected' : ''}>
                    ${p.name}
                </option>
            `).join('')}
        `;

        const activeProfile = profiles.find(p => p.is_active);
        if (activeProfile) {
            currentProfileId = activeProfile.id;
            localStorage.setItem('profile_id', currentProfileId);
            console.log('Active Profile Set:', currentProfileId);
            
            // Ensure the correct option is selected
            const activeOption = Array.from(profileSelect.options).find(
                option => option.value === activeProfile.id
            );
            if (activeOption) {
                activeOption.selected = true;
            }

            // Explicitly load monitors and credentials
            try {
                await loadMonitors(); 
                await loadCredentials();
            } catch (loadError) {
                console.error('Error loading monitors or credentials:', loadError);
                showNotification('Failed to load monitors or credentials', 'error');
            }
        } else {
            console.warn('No active profile found');
            showNotification('Please create or activate a profile', 'warning');
            
            // Clear credentials and monitors if no active profile
            clearCredentialSelect();
            clearMonitorList();
        }
    } catch (error) {
        console.error('Error loading profiles:', error);
        showNotification('Failed to load profiles', 'error');
        
        // Clear credentials and monitors on error
        clearCredentialSelect();
        clearMonitorList();
    }
}

async function handleMonitorSubmit(event) {
    event.preventDefault();

    // Existing validation
    const name = document.getElementById('name').value.trim();
    const url = document.getElementById('url').value.trim();
    const method = document.getElementById('method').value;
    const checkInterval = parseInt(document.getElementById('check_interval').value);
    const failureThreshold = parseInt(document.getElementById('failureThreshold').value);
    const timeout = parseInt(document.getElementById('timeout').value);
    const credentialId = document.getElementById('credentialSelect').value;

    console.log('Monitor Submission Details:');
    console.log('  Name:', name);
    console.log('  URL:', url);
    console.log('  Method:', method);
    console.log('  Check Interval:', checkInterval);
    console.log('  Failure Threshold:', failureThreshold);
    console.log('  Timeout:', timeout);
    console.log('  Credential ID:', credentialId);

    // Validate inputs
    if (!name || !url) {
        showNotification('Please fill in all required fields', 'error');
        return;
    }

    try {
        const monitorData = {
            name,
            url,
            method,
            check_interval: checkInterval,
            failure_threshold: failureThreshold,
            timeout,
            is_active: true,
            credential_id: credentialId || null  // Allow optional credential
        };

        const response = await fetch('/api/monitors', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-Profile-ID': getCurrentProfileId()
            },
            body: JSON.stringify(monitorData)
        });

        console.log('Monitor Creation Response:', response);

        if (!response.ok) {
            const errorText = await response.text();
            console.error('Monitor Creation Error:', errorText);
            throw new Error(errorText || 'Failed to create monitor');
        }

        const newMonitor = await response.json();
        console.log('Monitor created:', newMonitor);
        showNotification('Monitor created successfully', 'success');

        // Reset form and reload monitors
        event.target.reset();
        loadMonitors();
    } catch (error) {
        console.error('Error creating monitor:', error);
        showNotification(error.message, 'error');
    }
}
