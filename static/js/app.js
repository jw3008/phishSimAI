// Main App State
const app = {
    user: null,
    userRole: null,
    currentView: 'campaigns'
};

// API Helper
const api = {
    async call(endpoint, options = {}) {
        const response = await fetch(`/api${endpoint}`, {
            ...options,
            headers: {
                'Content-Type': 'application/json',
                ...options.headers
            },
            credentials: 'include'
        });

        if (response.status === 401) {
            showPage('login');
            return null;
        }

        const data = await response.json();
        return data;
    },

    get(endpoint) {
        return this.call(endpoint);
    },

    post(endpoint, body) {
        return this.call(endpoint, {
            method: 'POST',
            body: JSON.stringify(body)
        });
    },

    put(endpoint, body) {
        return this.call(endpoint, {
            method: 'PUT',
            body: JSON.stringify(body)
        });
    },

    delete(endpoint) {
        return this.call(endpoint, { method: 'DELETE' });
    }
};

// Page Navigation
function showPage(pageName) {
    document.querySelectorAll('.page').forEach(p => p.classList.remove('active'));
    document.getElementById(`${pageName}-page`).classList.add('active');
}

function showView(viewName) {
    app.currentView = viewName;
    document.querySelectorAll('.view').forEach(v => v.classList.remove('active'));
    document.getElementById(`${viewName}-view`).classList.add('active');
    document.querySelectorAll('.nav-link').forEach(l => l.classList.remove('active'));
    document.querySelector(`[data-page="${viewName}"]`).classList.add('active');
}

// Modal
function showModal(title, content) {
    const modal = document.getElementById('modal');
    const modalBody = document.getElementById('modal-body');
    modalBody.innerHTML = `<h2>${title}</h2>${content}`;
    modal.classList.add('active');
}

function closeModal() {
    document.getElementById('modal').classList.remove('active');
}

// Login
document.getElementById('login-form').addEventListener('submit', async (e) => {
    e.preventDefault();
    const username = document.getElementById('username').value;
    const password = document.getElementById('password').value;
    const errorDiv = document.getElementById('login-error');

    const result = await api.post('/login', { username, password });

    if (result && result.success) {
        app.user = result.user;
        app.userRole = result.role;
        document.getElementById('current-user').textContent = result.user;

        // Show/hide navigation based on role
        updateUIForRole(result.role);

        showPage('dashboard');

        // Load appropriate view based on role
        if (result.role === 'admin') {
            showView('campaigns');
            loadCampaigns();
        } else {
            showView('awareness');
            loadAwarenessAssessments();
        }
    } else {
        errorDiv.textContent = 'Invalid credentials';
    }
});

// Update UI based on user role
function updateUIForRole(role) {
    const adminElements = document.querySelectorAll('.admin-only');
    const userElements = document.querySelectorAll('.user-only');

    if (role === 'admin') {
        adminElements.forEach(el => el.style.display = '');
        userElements.forEach(el => el.style.display = 'none');
    } else {
        adminElements.forEach(el => el.style.display = 'none');
        userElements.forEach(el => el.style.display = '');
    }
}

// Logout
document.getElementById('logout-btn').addEventListener('click', async () => {
    await api.post('/logout');
    showPage('login');
    app.user = null;
});

// Navigation
document.querySelectorAll('.nav-link').forEach(link => {
    link.addEventListener('click', (e) => {
        e.preventDefault();
        const page = e.target.dataset.page;
        showView(page);

        // Load data for the view
        switch(page) {
            case 'campaigns': loadCampaigns(); break;
            case 'templates': loadTemplates(); break;
            case 'pages': loadPages(); break;
            case 'groups': loadGroups(); break;
            case 'smtp': loadSMTP(); break;
            case 'assessments': loadAssessments(); break;
            case 'user-management': loadUsers(); break;
            case 'settings': loadSettings(); break;
            case 'awareness': loadAwarenessAssessments(); break;
            case 'my-results': loadMyResults(); break;
            case 'knowledge-base': initKnowledgeBase(); break;
            case 'user-settings': loadUserSettings(); break;
        }
    });
});

// Campaigns
async function loadCampaigns() {
    const campaigns = await api.get('/campaigns');
    const container = document.getElementById('campaigns-list');

    if (!campaigns || campaigns.length === 0) {
        container.innerHTML = '<div class="empty-state"><h3>No campaigns yet</h3><p>Create your first phishing simulation campaign</p></div>';
        return;
    }

    container.innerHTML = campaigns.map(c => `
        <div class="card">
            <div class="card-header">
                <div>
                    <div class="card-title">${c.name}</div>
                    <span class="badge badge-${c.status}">${c.status}</span>
                </div>
                <div class="card-actions">
                    <button class="btn btn-small btn-primary" onclick="viewCampaign(${c.id})">View</button>
                    <button class="btn btn-small btn-danger" onclick="deleteCampaign(${c.id})">Delete</button>
                </div>
            </div>
            ${c.stats ? `
            <div class="stats">
                <div class="stat-item">
                    <div class="stat-value">${c.stats.total}</div>
                    <div class="stat-label">Total</div>
                </div>
                <div class="stat-item">
                    <div class="stat-value">${c.stats.sent}</div>
                    <div class="stat-label">Sent</div>
                </div>
                <div class="stat-item">
                    <div class="stat-value">${c.stats.reported}</div>
                    <div class="stat-label">Reported</div>
                </div>
                <div class="stat-item">
                    <div class="stat-value">${c.stats.clicked}</div>
                    <div class="stat-label">Clicked</div>
                </div>
                <div class="stat-item">
                    <div class="stat-value">${c.stats.submitted}</div>
                    <div class="stat-label">Submitted</div>
                </div>
            </div>
            ` : ''}
        </div>
    `).join('');
}

async function viewCampaign(id) {
    const campaign = await api.get(`/campaigns/${id}`);
    if (!campaign) return;

    const resultsTable = campaign.results && campaign.results.length > 0 ? `
        <h3>Results</h3>
        <table>
            <thead>
                <tr>
                    <th>Name</th>
                    <th>Email</th>
                    <th>Status</th>
                    <th>Sent</th>
                    <th>Clicked</th>
                    <th>Submitted</th>
                    <th>Reported</th>
                </tr>
            </thead>
            <tbody>
                ${campaign.results.map(r => `
                    <tr>
                        <td>${r.first_name} ${r.last_name}</td>
                        <td>${r.email}</td>
                        <td><span class="badge badge-${r.status}">${r.status}</span></td>
                        <td>${r.send_date ? new Date(r.send_date).toLocaleString() : '-'}</td>
                        <td>${r.click_date ? '<span style="color: green; font-size: 18px;">✓</span>' : '<span style="color: red; font-size: 18px;">✗</span>'}</td>
                        <td>${r.submit_date ? '<span style="color: green; font-size: 18px;">✓</span>' : '<span style="color: red; font-size: 18px;">✗</span>'}</td>
                        <td>${r.report_date ? '<span style="color: green; font-size: 18px;">✓</span>' : '<span style="color: red; font-size: 18px;">✗</span>'}</td>
                    </tr>
                `).join('')}
            </tbody>
        </table>
    ` : '<p>No results yet</p>';

    const actionButtons = `
        <div style="margin: 20px 0; display: flex; gap: 10px; flex-wrap: wrap;">
            ${campaign.status === 'draft' ? `
                <button onclick="launchCampaignNow(${id})" class="btn btn-success">Launch Campaign Now</button>
            ` : ''}
            ${campaign.stats && campaign.stats.submitted > 0 ? `
                <button onclick="viewCredentials(${id})" class="btn btn-primary">View Harvested Credentials</button>
            ` : ''}
            <button onclick="downloadCredentialsPDF(${id})" class="btn btn-secondary">Download Campaign Report</button>
        </div>
    `;

    showModal(`Campaign: ${campaign.name}`, `
        <div class="card-meta">
            <span>Status: <span class="badge badge-${campaign.status}">${campaign.status}</span></span>
            <span>Created: ${new Date(campaign.created_date).toLocaleString()}</span>
        </div>
        ${campaign.stats ? `
        <div class="stats">
            <div class="stat-item">
                <div class="stat-value">${campaign.stats.sent}</div>
                <div class="stat-label">Sent (${campaign.stats.total})</div>
            </div>
            <div class="stat-item">
                <div class="stat-value" style="color: ${campaign.stats.report_rate > 50 ? 'green' : 'inherit'}">${campaign.stats.report_rate || 0}%</div>
                <div class="stat-label">Report Rate</div>
            </div>
            <div class="stat-item">
                <div class="stat-value">${campaign.stats.click_rate}%</div>
                <div class="stat-label">Click Rate</div>
            </div>
            <div class="stat-item">
                <div class="stat-value">${campaign.stats.submit_rate}%</div>
                <div class="stat-label">Submit Rate</div>
            </div>
        </div>
        ` : ''}
        ${actionButtons}
        <br>
        ${resultsTable}
    `);
}

async function deleteCampaign(id) {
    if (!confirm('Are you sure you want to delete this campaign?')) return;
    await api.delete(`/campaigns/${id}`);
    loadCampaigns();
}

document.getElementById('new-campaign-btn').addEventListener('click', async () => {
    const templates = await api.get('/templates');
    const pages = await api.get('/pages');
    const groups = await api.get('/groups');
    const smtp = await api.get('/smtp');

    showModal('New Campaign', `
        <form id="campaign-form">
            <div class="form-group">
                <label>Campaign Name</label>
                <input type="text" name="name" required>
            </div>
            <div class="form-group">
                <label>Email Template</label>
                <select name="template_id" required>
                    <option value="">Select template...</option>
                    ${templates.map(t => `<option value="${t.id}">${t.name}</option>`).join('')}
                </select>
            </div>
            <div class="form-group">
                <label>Landing Page</label>
                <select name="page_id" required>
                    <option value="">Select page...</option>
                    ${pages.map(p => `<option value="${p.id}">${p.name}</option>`).join('')}
                </select>
            </div>
            <div class="form-group">
                <label>Sending Profile</label>
                <select name="smtp_id" required>
                    <option value="">Select profile...</option>
                    ${smtp.map(s => `<option value="${s.id}">${s.name}</option>`).join('')}
                </select>
            </div>
            <div class="form-group">
                <label>URL (e.g., http://localhost:3333)</label>
                <input type="url" name="url" value="http://localhost:3333" required>
            </div>
            <div class="form-group">
                <label>Target Groups</label>
                ${groups.map(g => `
                    <div class="checkbox-group">
                        <input type="checkbox" name="groups" value="${g.id}" id="group-${g.id}">
                        <label for="group-${g.id}">${g.name}</label>
                    </div>
                `).join('')}
            </div>
            <div class="form-group">
                <div class="checkbox-group">
                    <input type="checkbox" name="launch" id="launch">
                    <label for="launch">Launch campaign immediately</label>
                </div>
            </div>
            <button type="submit" class="btn btn-primary">Create Campaign</button>
        </form>
    `);

    document.getElementById('campaign-form').addEventListener('submit', async (e) => {
        e.preventDefault();
        const formData = new FormData(e.target);
        const selectedGroups = Array.from(document.querySelectorAll('input[name="groups"]:checked')).map(cb => parseInt(cb.value));

        const campaign = {
            name: formData.get('name'),
            template_id: parseInt(formData.get('template_id')),
            page_id: parseInt(formData.get('page_id')),
            smtp_id: parseInt(formData.get('smtp_id')),
            url: formData.get('url'),
            groups: selectedGroups,
            status: formData.get('launch') ? 'launched' : 'draft'
        };

        await api.post('/campaigns', campaign);
        closeModal();
        loadCampaigns();
    });
});

// Templates
async function loadTemplates() {
    const templates = await api.get('/templates');
    const container = document.getElementById('templates-list');

    if (!templates || templates.length === 0) {
        container.innerHTML = '<div class="empty-state"><h3>No templates yet</h3><p>Create your first email template</p></div>';
        return;
    }

    container.innerHTML = templates.map(t => `
        <div class="card">
            <div class="card-header">
                <div class="card-title">${t.name}</div>
                <div class="card-actions">
                    <button class="btn btn-small btn-primary" onclick="editTemplate(${t.id})">Edit</button>
                    <button class="btn btn-small btn-danger" onclick="deleteTemplate(${t.id})">Delete</button>
                </div>
            </div>
            <div class="card-body">
                <strong>Subject:</strong> ${t.subject}
            </div>
        </div>
    `).join('');
}

async function editTemplate(id) {
    const template = await api.get(`/templates/${id}`);
    showTemplateForm(template);
}

async function deleteTemplate(id) {
    if (!confirm('Delete this template?')) return;
    await api.delete(`/templates/${id}`);
    loadTemplates();
}

function showTemplateForm(template = null) {
    showModal(template ? 'Edit Template' : 'New Template', `
        <form id="template-form">
            <div class="form-group">
                <label>Template Name</label>
                <input type="text" name="name" value="${template?.name || ''}" required>
            </div>
            <div class="form-group">
                <label>Subject</label>
                <input type="text" name="subject" value="${template?.subject || ''}" required>
                <small>Variables: {{.FirstName}}, {{.LastName}}, {{.URL}}, {{.ReportURL}}</small>
            </div>
            <div class="form-group">
                <label>HTML Content</label>
                <textarea name="html" rows="10" required>${template?.html || ''}</textarea>
                <small>Variables: {{.FirstName}}, {{.LastName}}, {{.URL}}, {{.ReportURL}} (for reporting phishing)</small>
            </div>
            <div class="form-group">
                <label>Text Content</label>
                <textarea name="text" rows="10">${template?.text || ''}</textarea>
            </div>
            <button type="submit" class="btn btn-primary">${template ? 'Update' : 'Create'} Template</button>
        </form>
    `);

    document.getElementById('template-form').addEventListener('submit', async (e) => {
        e.preventDefault();
        const formData = new FormData(e.target);
        const data = {
            name: formData.get('name'),
            subject: formData.get('subject'),
            html: formData.get('html'),
            text: formData.get('text')
        };

        if (template) {
            await api.put(`/templates/${template.id}`, data);
        } else {
            await api.post('/templates', data);
        }
        closeModal();
        loadTemplates();
    });
}

document.getElementById('new-template-btn').addEventListener('click', () => showTemplateForm());
document.getElementById('generate-ai-template-btn')?.addEventListener('click', () => generateRandomTemplate());

// Pages
async function loadPages() {
    const pages = await api.get('/pages');
    const container = document.getElementById('pages-list');

    if (!pages || pages.length === 0) {
        container.innerHTML = '<div class="empty-state"><h3>No pages yet</h3><p>Create your first landing page</p></div>';
        return;
    }

    container.innerHTML = pages.map(p => `
        <div class="card">
            <div class="card-header">
                <div class="card-title">${p.name}</div>
                <div class="card-actions">
                    <button class="btn btn-small btn-primary" onclick="editPage(${p.id})">Edit</button>
                    <button class="btn btn-small btn-danger" onclick="deletePage(${p.id})">Delete</button>
                </div>
            </div>
        </div>
    `).join('');
}

async function editPage(id) {
    const page = await api.get(`/pages/${id}`);
    showPageForm(page);
}

async function deletePage(id) {
    if (!confirm('Delete this page?')) return;
    await api.delete(`/pages/${id}`);
    loadPages();
}

function showPageForm(page = null) {
    showModal(page ? 'Edit Landing Page' : 'New Landing Page', `
        <form id="page-form">
            <div class="form-group">
                <label>Page Name</label>
                <input type="text" name="name" value="${page?.name || ''}" required>
            </div>
            <div class="form-group">
                <label>HTML Content</label>
                <textarea name="html" rows="15" required>${page?.html || ''}</textarea>
            </div>
            <div class="form-group">
                <div class="checkbox-group">
                    <input type="checkbox" name="capture_credentials" id="capture_creds" ${page?.capture_credentials ? 'checked' : ''}>
                    <label for="capture_creds">Capture Credentials</label>
                </div>
            </div>
            <div class="form-group">
                <div class="checkbox-group">
                    <input type="checkbox" name="capture_passwords" id="capture_pass" ${page?.capture_passwords ? 'checked' : ''}>
                    <label for="capture_pass">Capture Passwords</label>
                </div>
            </div>
            <div class="form-group">
                <label>Redirect URL (optional)</label>
                <input type="url" name="redirect_url" value="${page?.redirect_url || ''}">
            </div>
            <button type="submit" class="btn btn-primary">${page ? 'Update' : 'Create'} Page</button>
        </form>
    `);

    document.getElementById('page-form').addEventListener('submit', async (e) => {
        e.preventDefault();
        const formData = new FormData(e.target);
        const data = {
            name: formData.get('name'),
            html: formData.get('html'),
            capture_credentials: formData.get('capture_credentials') === 'on',
            capture_passwords: formData.get('capture_passwords') === 'on',
            redirect_url: formData.get('redirect_url')
        };

        if (page) {
            await api.put(`/pages/${page.id}`, data);
        } else {
            await api.post('/pages', data);
        }
        closeModal();
        loadPages();
    });
}

document.getElementById('new-page-btn').addEventListener('click', () => showPageForm());
document.getElementById('clone-page-btn')?.addEventListener('click', () => showClonePage());

// Groups
async function loadGroups() {
    const groups = await api.get('/groups');
    const container = document.getElementById('groups-list');

    if (!groups || groups.length === 0) {
        container.innerHTML = '<div class="empty-state"><h3>No groups yet</h3><p>Create your first target group</p></div>';
        return;
    }

    container.innerHTML = groups.map(g => `
        <div class="card">
            <div class="card-header">
                <div class="card-title">${g.name}</div>
                <div class="card-actions">
                    <button class="btn btn-small btn-primary" onclick="editGroup(${g.id})">Edit</button>
                    <button class="btn btn-small btn-danger" onclick="deleteGroup(${g.id})">Delete</button>
                </div>
            </div>
        </div>
    `).join('');
}

async function editGroup(id) {
    const group = await api.get(`/groups/${id}`);
    showGroupForm(group);
}

async function deleteGroup(id) {
    if (!confirm('Delete this group?')) return;
    await api.delete(`/groups/${id}`);
    loadGroups();
}

function showGroupForm(group = null) {
    const targets = group?.targets || [];

    showModal(group ? 'Edit Group' : 'New Group', `
        <form id="group-form">
            <div class="form-group">
                <label>Group Name</label>
                <input type="text" name="name" value="${group?.name || ''}" required>
            </div>
            <h3>Targets</h3>
            <div id="targets-container">
                ${targets.map((t, i) => `
                    <div class="card" style="padding: 10px; margin-bottom: 10px; position: relative;" data-target-index="${i}">
                        <button type="button" class="btn btn-small btn-danger" onclick="removeTarget(${i})" style="position: absolute; right: 10px; top: 10px;">Remove</button>
                        <input type="text" placeholder="First Name" value="${t.first_name}" data-field="first_name" data-index="${i}" style="margin-bottom: 5px; width: calc(100% - 80px);">
                        <input type="text" placeholder="Last Name" value="${t.last_name}" data-field="last_name" data-index="${i}" style="margin-bottom: 5px; width: calc(100% - 80px);">
                        <input type="email" placeholder="Email" value="${t.email}" data-field="email" data-index="${i}" required style="margin-bottom: 5px; width: calc(100% - 80px);">
                        <input type="text" placeholder="Position" value="${t.position}" data-field="position" data-index="${i}" style="width: calc(100% - 80px);">
                    </div>
                `).join('')}
            </div>
            <button type="button" class="btn btn-secondary" onclick="addTarget()">Add Target</button>
            <button type="submit" class="btn btn-primary">${group ? 'Update' : 'Create'} Group</button>
        </form>
    `);

    document.getElementById('group-form').addEventListener('submit', async (e) => {
        e.preventDefault();
        const formData = new FormData(e.target);
        const targetInputs = document.querySelectorAll('#targets-container input');

        const targetsMap = {};
        targetInputs.forEach(input => {
            const index = input.dataset.index;
            const field = input.dataset.field;
            if (!targetsMap[index]) targetsMap[index] = {};
            targetsMap[index][field] = input.value;
        });

        const data = {
            name: formData.get('name'),
            targets: Object.values(targetsMap)
        };

        if (group) {
            await api.put(`/groups/${group.id}`, data);
        } else {
            await api.post('/groups', data);
        }
        closeModal();
        loadGroups();
    });
}

window.addTarget = function() {
    const container = document.getElementById('targets-container');
    const index = container.children.length;
    const div = document.createElement('div');
    div.className = 'card';
    div.style.padding = '10px';
    div.style.marginBottom = '10px';
    div.style.position = 'relative';
    div.setAttribute('data-target-index', index);
    div.innerHTML = `
        <button type="button" class="btn btn-small btn-danger" onclick="removeTarget(${index})" style="position: absolute; right: 10px; top: 10px;">Remove</button>
        <input type="text" placeholder="First Name" data-field="first_name" data-index="${index}" style="margin-bottom: 5px; width: calc(100% - 80px);">
        <input type="text" placeholder="Last Name" data-field="last_name" data-index="${index}" style="margin-bottom: 5px; width: calc(100% - 80px);">
        <input type="email" placeholder="Email" data-field="email" data-index="${index}" required style="margin-bottom: 5px; width: calc(100% - 80px);">
        <input type="text" placeholder="Position" data-field="position" data-index="${index}" style="width: calc(100% - 80px);">
    `;
    container.appendChild(div);
};

window.removeTarget = function(index) {
    const container = document.getElementById('targets-container');
    const target = container.querySelector(`[data-target-index="${index}"]`);
    if (target) {
        target.remove();
    }
};

document.getElementById('new-group-btn').addEventListener('click', () => showGroupForm());

// SMTP
async function loadSMTP() {
    const configs = await api.get('/smtp');
    const container = document.getElementById('smtp-list');

    if (!configs || configs.length === 0) {
        container.innerHTML = '<div class="empty-state"><h3>No sending profiles yet</h3><p>Create your first SMTP configuration</p></div>';
        return;
    }

    container.innerHTML = configs.map(s => `
        <div class="card">
            <div class="card-header">
                <div class="card-title">${s.name}</div>
                <div class="card-actions">
                    <button class="btn btn-small btn-primary" onclick="editSMTP(${s.id})">Edit</button>
                    <button class="btn btn-small btn-danger" onclick="deleteSMTP(${s.id})">Delete</button>
                </div>
            </div>
            <div class="card-body">
                <strong>Host:</strong> ${s.host}<br>
                <strong>From:</strong> ${s.from_address}
            </div>
        </div>
    `).join('');
}

async function editSMTP(id) {
    const smtp = await api.get(`/smtp/${id}`);
    showSMTPForm(smtp);
}

async function deleteSMTP(id) {
    if (!confirm('Delete this SMTP config?')) return;
    await api.delete(`/smtp/${id}`);
    loadSMTP();
}

function showSMTPForm(smtp = null) {
    showModal(smtp ? 'Edit Sending Profile' : 'New Sending Profile', `
        <form id="smtp-form">
            <div class="form-group">
                <label>Profile Name</label>
                <input type="text" name="name" value="${smtp?.name || ''}" required>
            </div>
            <div class="form-group">
                <label>Host (e.g., smtp.gmail.com:587)</label>
                <input type="text" name="host" value="${smtp?.host || ''}" required>
            </div>
            <div class="form-group">
                <label>Username</label>
                <input type="text" name="username" value="${smtp?.username || ''}">
            </div>
            <div class="form-group">
                <label>Password</label>
                <input type="password" name="password" value="${smtp?.password || ''}">
            </div>
            <div class="form-group">
                <label>From Address</label>
                <input type="email" name="from_address" value="${smtp?.from_address || ''}" required>
            </div>
            <button type="submit" class="btn btn-primary">${smtp ? 'Update' : 'Create'} Profile</button>
        </form>
    `);

    document.getElementById('smtp-form').addEventListener('submit', async (e) => {
        e.preventDefault();
        const formData = new FormData(e.target);
        const data = {
            name: formData.get('name'),
            host: formData.get('host'),
            username: formData.get('username'),
            password: formData.get('password'),
            from_address: formData.get('from_address')
        };

        if (smtp) {
            await api.put(`/smtp/${smtp.id}`, data);
        } else {
            await api.post('/smtp', data);
        }
        closeModal();
        loadSMTP();
    });
}

document.getElementById('new-smtp-btn').addEventListener('click', () => showSMTPForm());

// Modal close
document.querySelector('.close').addEventListener('click', closeModal);
window.addEventListener('click', (e) => {
    if (e.target === document.getElementById('modal')) {
        closeModal();
    }
});

// ============================================
// ASSESSMENT MANAGEMENT (ADMIN)
// ============================================

async function loadAssessments() {
    const assessments = await api.get('/assessments');
    const container = document.getElementById('assessments-list');

    if (!assessments || assessments.length === 0) {
        container.innerHTML = '<div class="empty-state"><h3>No assessments yet</h3><p>Create your first security awareness assessment</p></div>';
        return;
    }

    container.innerHTML = assessments.map(a => `
        <div class="card">
            <div class="card-header">
                <div>
                    <div class="card-title">${a.title}</div>
                    <span class="badge badge-${a.is_published ? 'success' : 'draft'}">${a.is_published ? 'Published' : 'Draft'}</span>
                    ${a.deadline ? `<span style="margin-left: 10px;">Deadline: ${new Date(a.deadline).toLocaleDateString()}</span>` : ''}
                </div>
                <div class="card-actions">
                    <button class="btn btn-small btn-primary" onclick="viewAssessmentStats(${a.id})">Stats</button>
                    <button class="btn btn-small btn-secondary" onclick="viewAssessment(${a.id})">Edit</button>
                    ${!a.is_published ? `<button class="btn btn-small btn-success" onclick="publishAssessment(${a.id})">Publish</button>` : ''}
                    <button class="btn btn-small btn-danger" onclick="deleteAssessment(${a.id})">Delete</button>
                </div>
            </div>
        </div>
    `).join('');
}

async function viewAssessment(id) {
    const assessment = await api.get(`/assessments/${id}`);
    if (!assessment) return;

    const questionsHTML = assessment.questions && assessment.questions.length > 0 ?
        assessment.questions.map((q, idx) => `
            <div class="question-preview">
                <strong>Question ${idx + 1}:</strong> ${q.question_text} (${q.points} points)<br>
                <ul>
                    ${q.answer_options.map(opt => `
                        <li ${opt.is_correct ? 'style="color: green; font-weight: bold;"' : ''}>${opt.option_text}</li>
                    `).join('')}
                </ul>
            </div>
        `).join('')
        : '<p>No questions yet</p>';

    showModal(`Assessment: ${assessment.title}`, `
        <p><strong>Description:</strong> ${assessment.description || 'N/A'}</p>
        <p><strong>Deadline:</strong> ${assessment.deadline ? new Date(assessment.deadline).toLocaleString() : 'No deadline'}</p>
        <p><strong>Status:</strong> ${assessment.is_published ? 'Published' : 'Draft'}</p>
        <hr>
        <h3>Questions</h3>
        ${questionsHTML}
        <br>
        <button class="btn btn-primary" onclick="editAssessmentForm(${id})">Edit Assessment</button>
    `);
}

function editAssessmentForm(id) {
    // For simplicity, we'll redirect to the create form with pre-filled data
    // In production, you'd want a more sophisticated edit interface
    alert('Edit functionality - Would open a detailed editor. For now, please create a new assessment or use the API.');
}

async function publishAssessment(id) {
    if (!confirm('Publish this assessment? Users will be able to see and take it.')) return;
    await api.post(`/assessments/${id}/publish`);
    loadAssessments();
}

async function deleteAssessment(id) {
    if (!confirm('Are you sure you want to delete this assessment?')) return;
    await api.delete(`/assessments/${id}`);
    loadAssessments();
}

async function viewAssessmentStats(id) {
    const [stats, results] = await Promise.all([
        api.get(`/assessments/${id}/stats`),
        api.get(`/assessments/${id}/results`)
    ]);

    const resultsTable = results && results.length > 0 ? `
        <table>
            <thead>
                <tr>
                    <th>Username</th>
                    <th>Status</th>
                    <th>Score</th>
                    <th>Percentage</th>
                    <th>Completed</th>
                </tr>
            </thead>
            <tbody>
                ${results.map(r => `
                    <tr>
                        <td>${r.username}</td>
                        <td><span class="badge badge-${r.status === 'Completed' ? 'success' : r.status === 'In Progress' ? 'warning' : 'draft'}">${r.status}</span></td>
                        <td>${r.score} / ${r.total_points}</td>
                        <td>${r.percentage.toFixed(1)}%</td>
                        <td>${r.completed_at ? new Date(r.completed_at).toLocaleString() : '-'}</td>
                    </tr>
                `).join('')}
            </tbody>
        </table>
    ` : '<p>No user results yet</p>';

    showModal('Assessment Statistics', `
        <div class="stats">
            <div class="stat-item">
                <div class="stat-value">${stats.completed_users}</div>
                <div class="stat-label">Completed</div>
            </div>
            <div class="stat-item">
                <div class="stat-value">${stats.pending_users}</div>
                <div class="stat-label">Pending</div>
            </div>
            <div class="stat-item">
                <div class="stat-value">${stats.average_score.toFixed(1)}%</div>
                <div class="stat-label">Average Score</div>
            </div>
            <div class="stat-item">
                <div class="stat-value">${stats.pass_rate.toFixed(1)}%</div>
                <div class="stat-label">Pass Rate (≥70%)</div>
            </div>
        </div>
        <br>
        <h3>User Results</h3>
        ${resultsTable}
    `);
}

document.getElementById('new-assessment-btn')?.addEventListener('click', () => showAssessmentForm());

function showAssessmentForm() {
    const questions = [{id: 1}]; // Start with one question

    function renderForm(questionsData) {
        const questionsHTML = questionsData.map((q, idx) => `
            <div class="question-block" data-qid="${q.id}" style="border: 1px solid #ddd; padding: 15px; margin: 10px 0; border-radius: 5px;">
                <h4>Question ${idx + 1}</h4>
                <div class="form-group">
                    <label>Question Text</label>
                    <textarea name="question_text_${q.id}" required style="width: 100%; min-height: 80px;"></textarea>
                </div>
                <div class="form-group">
                    <label>Points</label>
                    <input type="number" name="points_${q.id}" value="1" min="1" required>
                </div>
                <h5>Answer Options</h5>
                ${[1,2,3,4].map(optNum => `
                    <div class="form-group" style="display: flex; gap: 10px; align-items: center;">
                        <input type="radio" name="correct_${q.id}" value="${optNum}" ${optNum === 1 ? 'checked' : ''} required>
                        <input type="text" name="option_${q.id}_${optNum}" placeholder="Option ${optNum}" required style="flex: 1;">
                    </div>
                `).join('')}
                ${questionsData.length > 1 ? `<button type="button" class="btn btn-small btn-danger" onclick="removeQuestion(${q.id})">Remove Question</button>` : ''}
            </div>
        `).join('');

        return `
            <form id="assessment-form">
                <div class="form-group">
                    <label>Assessment Title</label>
                    <input type="text" name="title" required>
                </div>
                <div class="form-group">
                    <label>Description</label>
                    <textarea name="description" style="width: 100%; min-height: 80px;"></textarea>
                </div>
                <div class="form-group">
                    <label>Deadline (optional)</label>
                    <input type="datetime-local" name="deadline">
                </div>
                <hr>
                <h3>Questions</h3>
                <div id="questions-container">${questionsHTML}</div>
                <button type="button" class="btn btn-secondary" id="add-question-btn">Add Question</button>
                <br><br>
                <button type="submit" class="btn btn-primary">Create Assessment</button>
            </form>
        `;
    }

    showModal('Create Assessment', renderForm(questions));

    let questionId = 2;
    const questionsState = questions;

    document.getElementById('add-question-btn').addEventListener('click', () => {
        questionsState.push({id: questionId++});
        document.getElementById('questions-container').innerHTML = renderForm(questionsState).match(/<div class="question-block"[\s\S]*<\/div>\s*<\/div>/g).join('');
    });

    window.removeQuestion = function(qid) {
        const idx = questionsState.findIndex(q => q.id === qid);
        if (idx !== -1) {
            questionsState.splice(idx, 1);
            document.getElementById('questions-container').innerHTML = renderForm(questionsState).match(/<div class="question-block"[\s\S]*<\/div>\s*<\/div>/g).join('');
        }
    };

    document.getElementById('assessment-form').addEventListener('submit', async (e) => {
        e.preventDefault();
        const formData = new FormData(e.target);

        const questionsData = questionsState.map((q, idx) => {
            const correctOption = parseInt(formData.get(`correct_${q.id}`));
            return {
                question_text: formData.get(`question_text_${q.id}`),
                question_order: idx,
                points: parseInt(formData.get(`points_${q.id}`)),
                answer_options: [1, 2, 3, 4].map(optNum => ({
                    option_text: formData.get(`option_${q.id}_${optNum}`),
                    is_correct: optNum === correctOption,
                    option_order: optNum - 1
                }))
            };
        });

        const deadline = formData.get('deadline') ? new Date(formData.get('deadline')).toISOString() : null;

        const data = {
            title: formData.get('title'),
            description: formData.get('description'),
            deadline: deadline,
            questions: questionsData
        };

        const result = await api.post('/assessments', data);
        if (result && result.success) {
            closeModal();
            loadAssessments();
        }
    });
}

// ============================================
// AWARENESS TRAINING (USER)
// ============================================

async function loadAwarenessAssessments() {
    // Initialize email analyzer
    initEmailAnalyzerAwareness();

    // Initialize inline chatbot
    initInlineChatbot();

    const assessments = await api.get('/user/assessments');
    const container = document.getElementById('awareness-list');

    if (!assessments || assessments.length === 0) {
        container.innerHTML = '<div class="empty-state"><h3>No assessments available</h3><p>Check back later for new training assessments</p></div>';
        return;
    }

    container.innerHTML = assessments.map(a => `
        <div class="card">
            <div class="card-header">
                <div>
                    <div class="card-title">${a.title}</div>
                    <p style="margin: 5px 0; color: #666;">${a.description || ''}</p>
                    ${a.deadline ? `<p style="margin: 5px 0; color: #d32f2f;"><strong>Deadline:</strong> ${new Date(a.deadline).toLocaleString()}</p>` : ''}
                    <span class="badge badge-${a.status === 'Completed' ? 'success' : a.status === 'In Progress' ? 'warning' : 'draft'}">${a.status}</span>
                    ${a.status === 'Completed' ? `<span style="margin-left: 10px;">Score: ${a.score}/${a.total_points} (${a.percentage.toFixed(1)}%)</span>` : ''}
                </div>
                <div class="card-actions">
                    ${a.status === 'Not Started' ?
                        `<button class="btn btn-primary" onclick="startAssessment(${a.id})">Start Assessment</button>` :
                        a.status === 'In Progress' ?
                        `<button class="btn btn-warning" onclick="continueAssessment(${a.id}, ${a.attempt_id})">Continue</button>` :
                        `<button class="btn btn-secondary" onclick="viewMyResult(${a.attempt_id})">View Details</button>
                         <button class="btn btn-primary" onclick="downloadResultPDF(${a.attempt_id})" style="margin-left: 10px;">Download PDF</button>`
                    }
                </div>
            </div>
        </div>
    `).join('');
}

async function startAssessment(assessmentId) {
    const result = await api.post(`/user/assessments/${assessmentId}/start`);
    if (result && result.success) {
        takeAssessment(assessmentId, result.attempt_id);
    }
}

async function continueAssessment(assessmentId, attemptId) {
    takeAssessment(assessmentId, attemptId);
}

async function takeAssessment(assessmentId, attemptId) {
    const detail = await api.get(`/user/assessments/${assessmentId}`);
    if (!detail || !detail.assessment) return;

    const assessment = detail.assessment;
    let currentQuestionIndex = 0;
    const userAnswers = {};

    function renderQuestion() {
        const q = assessment.questions[currentQuestionIndex];
        const progress = ((currentQuestionIndex + 1) / assessment.questions.length * 100).toFixed(0);

        // Get previously selected answer for this question
        const previousAnswer = userAnswers[q.id];

        return `
            <div style="margin-bottom: 20px;">
                <div style="background: #e0e0e0; height: 10px; border-radius: 5px; margin-bottom: 20px;">
                    <div style="background: #4CAF50; height: 100%; width: ${progress}%; border-radius: 5px;"></div>
                </div>
                <p style="color: #666;">Question ${currentQuestionIndex + 1} of ${assessment.questions.length}</p>
            </div>
            <h3>${q.question_text}</h3>
            <p style="color: #666; margin-bottom: 20px;">Points: ${q.points}</p>
            <form id="question-form">
                ${q.answer_options.map(opt => `
                    <div class="form-group" style="margin: 15px 0;">
                        <label style="display: flex; align-items: center; padding: 15px; border: 2px solid #ddd; border-radius: 5px; cursor: pointer; background: ${previousAnswer === opt.id ? '#e3f2fd' : 'transparent'};">
                            <input type="radio" name="answer" value="${opt.id}" ${previousAnswer === opt.id ? 'checked' : ''} required style="margin-right: 10px;">
                            <span>${opt.option_text}</span>
                        </label>
                    </div>
                `).join('')}
                <br>
                <div style="display: flex; gap: 10px; justify-content: space-between;">
                    ${currentQuestionIndex > 0 ? '<button type="button" class="btn btn-secondary" id="prev-btn">Previous</button>' : '<div></div>'}
                    <button type="submit" class="btn btn-primary">${currentQuestionIndex < assessment.questions.length - 1 ? 'Next Question' : 'Submit Assessment'}</button>
                </div>
            </form>
        `;
    }

    function setupQuestionHandlers() {
        const form = document.getElementById('question-form');
        if (!form) return;

        // Remove any existing listeners by cloning the form
        const newForm = form.cloneNode(true);
        form.parentNode.replaceChild(newForm, form);

        // Add submit handler
        newForm.addEventListener('submit', async (e) => {
            e.preventDefault();
            const formData = new FormData(e.target);
            const selectedOption = parseInt(formData.get('answer'));

            if (!selectedOption) {
                alert('Please select an answer before continuing.');
                return;
            }

            const q = assessment.questions[currentQuestionIndex];
            userAnswers[q.id] = selectedOption;

            // Save the response to backend
            try {
                await api.post(`/user/assessments/attempt/${attemptId}/submit`, {
                    question_id: q.id,
                    selected_option_id: selectedOption
                });
            } catch (error) {
                alert('Error saving answer. Please try again.');
                return;
            }

            // Move to next question or complete
            if (currentQuestionIndex < assessment.questions.length - 1) {
                currentQuestionIndex++;
                showModal(`Taking Assessment: ${assessment.title}`, renderQuestion());
                setupQuestionHandlers();
            } else {
                // Complete the assessment
                const result = await api.post(`/user/assessments/attempt/${attemptId}/complete`);
                if (result && result.success) {
                    showModal('Assessment Completed!', `
                        <div style="text-align: center;">
                            <h2 style="color: #4CAF50;">🎉 Congratulations!</h2>
                            <p>You have completed the assessment.</p>
                            <div class="stats">
                                <div class="stat-item">
                                    <div class="stat-value">${result.score}</div>
                                    <div class="stat-label">Score</div>
                                </div>
                                <div class="stat-item">
                                    <div class="stat-value">${result.total}</div>
                                    <div class="stat-label">Total Points</div>
                                </div>
                                <div class="stat-item">
                                    <div class="stat-value">${result.percentage.toFixed(1)}%</div>
                                    <div class="stat-label">Percentage</div>
                                </div>
                            </div>
                            <br>
                            <div style="margin-top: 20px; padding: 15px; background: #e3f2fd; border-radius: 5px;">
                                <p style="margin: 10px 0; font-size: 14px;">💡 You can view your results anytime from the <strong>My Results</strong> page in the navigation menu.</p>
                            </div>
                            <br>
                            <button class="btn btn-primary" onclick="viewMyResult(${result.attempt_id})">View Detailed Results Now</button>
                            <button class="btn btn-success" onclick="downloadResultPDF(${result.attempt_id})" style="margin-left: 10px;">Download PDF</button>
                            <br><br>
                            <button class="btn btn-info" onclick="closeModal(); showView('my-results'); loadMyResults();" style="margin-right: 10px;">Go to My Results</button>
                            <button class="btn btn-secondary" onclick="closeModal(); loadAwarenessAssessments();">Back to Assessments</button>
                        </div>
                    `);
                }
            }
        });

        // Add previous button handler
        const prevBtn = document.getElementById('prev-btn');
        if (prevBtn) {
            prevBtn.addEventListener('click', () => {
                if (currentQuestionIndex > 0) {
                    currentQuestionIndex--;
                    showModal(`Taking Assessment: ${assessment.title}`, renderQuestion());
                    setupQuestionHandlers();
                }
            });
        }
    }

    // Initial render
    showModal(`Taking Assessment: ${assessment.title}`, renderQuestion());
    setupQuestionHandlers();
}

// ============================================
// MY RESULTS (USER)
// ============================================

async function loadMyResults() {
    const results = await api.get('/user/results');
    const container = document.getElementById('my-results-list');

    if (!results || results.length === 0) {
        container.innerHTML = '<div class="empty-state"><h3>No results yet</h3><p>Complete assessments to see your results here</p></div>';
        return;
    }

    container.innerHTML = results.map(r => `
        <div class="card">
            <div class="card-header">
                <div>
                    <div class="card-title">${r.title}</div>
                    <p style="margin: 5px 0;">Score: ${r.score} / ${r.total_points} (${r.percentage.toFixed(1)}%)</p>
                    <p style="margin: 5px 0; color: #666;">Completed: ${new Date(r.completed_at).toLocaleString()}</p>
                    <span class="badge badge-${r.percentage >= 70 ? 'success' : 'danger'}">${r.percentage >= 70 ? 'Passed' : 'Failed'}</span>
                </div>
                <div class="card-actions">
                    <button class="btn btn-primary" onclick="viewMyResult(${r.attempt_id})">View Details</button>
                    <button class="btn btn-secondary" onclick="downloadResultPDF(${r.attempt_id})">Download PDF</button>
                </div>
            </div>
        </div>
    `).join('');
}

async function viewMyResult(attemptId) {
    const result = await api.get(`/user/results/${attemptId}`);
    if (!result) return;

    const responsesHTML = result.responses && result.responses.length > 0 ?
        result.responses.map((r, idx) => `
            <div class="question-result" style="border: 1px solid #ddd; padding: 15px; margin: 10px 0; border-radius: 5px; background: ${r.is_correct ? '#e8f5e9' : '#ffebee'};">
                <h4>Question ${idx + 1}</h4>
                <p><strong>${r.question_text}</strong></p>
                <p>Your answer: <span style="${r.is_correct ? 'color: green;' : 'color: red;'}">${r.selected_option}</span></p>
                ${!r.is_correct ? `<p>Correct answer: <span style="color: green;">${r.correct_answer}</span></p>` : ''}
                <p>Points earned: ${r.points_earned} / ${r.question_points}</p>
            </div>
        `).join('')
        : '<p>No responses recorded</p>';

    showModal('Assessment Result Details', `
        <h3>${result.assessment_title}</h3>
        <div class="stats">
            <div class="stat-item">
                <div class="stat-value">${result.score}</div>
                <div class="stat-label">Score</div>
            </div>
            <div class="stat-item">
                <div class="stat-value">${result.total_points}</div>
                <div class="stat-label">Total Points</div>
            </div>
            <div class="stat-item">
                <div class="stat-value">${result.percentage.toFixed(1)}%</div>
                <div class="stat-label">Percentage</div>
            </div>
            <div class="stat-item">
                <div class="stat-value" style="color: ${result.percentage >= 70 ? '#4CAF50' : '#d32f2f'}">${result.percentage >= 70 ? 'PASS' : 'FAIL'}</div>
                <div class="stat-label">Result</div>
            </div>
        </div>
        <p><strong>Completed:</strong> ${new Date(result.completed_at).toLocaleString()}</p>
        <hr>
        <h3>Question Review</h3>
        ${responsesHTML}
        <div style="margin-top: 30px; text-align: center; border-top: 1px solid #ddd; padding-top: 20px;">
            <button class="btn btn-secondary" onclick="closeModal()">← Back to My Results</button>
            <button class="btn btn-primary" onclick="downloadResultPDF(${attemptId})" style="margin-left: 10px;">📄 Download PDF</button>
        </div>
    `);
}

function downloadResultPDF(attemptId) {
    // Create a temporary link and trigger download
    const link = document.createElement('a');
    link.href = `/api/user/results/${attemptId}/pdf`;
    link.download = `assessment_result_${attemptId}.pdf`;
    link.target = '_blank';

    // Append to body, click, and remove
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);

    // Show feedback
    const messageDiv = document.createElement('div');
    messageDiv.style.cssText = 'position: fixed; top: 20px; right: 20px; background: #4CAF50; color: white; padding: 15px 20px; border-radius: 5px; z-index: 10000; box-shadow: 0 2px 5px rgba(0,0,0,0.2);';
    messageDiv.textContent = '✓ PDF download started!';
    document.body.appendChild(messageDiv);

    setTimeout(() => {
        document.body.removeChild(messageDiv);
    }, 3000);
}

// ============================================
// USER MANAGEMENT (ADMIN)
// ============================================

async function loadUsers() {
    const users = await api.get('/users');
    const container = document.getElementById('users-list');

    if (!users || users.length === 0) {
        container.innerHTML = '<div class="empty-state"><h3>No users yet</h3><p>Create your first user</p></div>';
        return;
    }

    container.innerHTML = users.map(u => `
        <div class="card">
            <div class="card-header">
                <div>
                    <div class="card-title">${u.username}</div>
                    <span class="badge badge-${u.role === 'admin' ? 'success' : 'draft'}">${u.role.toUpperCase()}</span>
                    <span style="margin-left: 10px; color: #666;">Created: ${new Date(u.created_at).toLocaleDateString()}</span>
                </div>
                <div class="card-actions">
                    <button class="btn btn-small btn-secondary" onclick="showChangeRoleForm(${u.id}, '${u.username}', '${u.role}')">Change Role</button>
                    <button class="btn btn-small btn-warning" onclick="showResetPasswordForm(${u.id}, '${u.username}')">Reset Password</button>
                    <button class="btn btn-small btn-danger" onclick="deleteUser(${u.id}, '${u.username}')">Delete</button>
                </div>
            </div>
        </div>
    `).join('');
}

async function deleteUser(id, username) {
    if (!confirm(`Are you sure you want to delete user "${username}"?`)) return;

    const result = await api.delete(`/users/${id}`);
    if (result && result.success) {
        loadUsers();
    } else {
        alert(result.error || 'Failed to delete user');
    }
}

function showChangeRoleForm(userId, username, currentRole) {
    showModal(`Change Role for ${username}`, `
        <form id="change-role-form">
            <div class="form-group">
                <label>Current Role: <strong>${currentRole}</strong></label>
            </div>
            <div class="form-group">
                <label>New Role</label>
                <select name="role" required>
                    <option value="">Select role...</option>
                    <option value="admin" ${currentRole === 'admin' ? 'selected' : ''}>Admin (Full Access)</option>
                    <option value="user" ${currentRole === 'user' ? 'selected' : ''}>User (Awareness Training Only)</option>
                </select>
            </div>
            <div class="info-box" style="background: #f5f5f5; padding: 15px; margin: 15px 0; border-radius: 5px;">
                <p style="margin: 0 0 10px 0;"><strong>Admin:</strong> Full access to phishing simulations, assessments, user management, and all reports</p>
                <p style="margin: 0;"><strong>User:</strong> Access to awareness training, assessments, and own results only</p>
            </div>
            <button type="submit" class="btn btn-primary">Update Role</button>
        </form>
    `);

    document.getElementById('change-role-form').addEventListener('submit', async (e) => {
        e.preventDefault();
        const formData = new FormData(e.target);
        const role = formData.get('role');

        const result = await api.put(`/users/${userId}`, { role });
        if (result && result.success) {
            closeModal();
            loadUsers();
        } else {
            alert(result.error || 'Failed to update role');
        }
    });
}

function showResetPasswordForm(userId, username) {
    showModal(`Reset Password for ${username}`, `
        <form id="reset-password-form">
            <div class="form-group">
                <label>New Password</label>
                <input type="password" name="new_password" minlength="6" required>
                <small>Minimum 6 characters</small>
            </div>
            <div class="form-group">
                <label>Confirm Password</label>
                <input type="password" name="confirm_password" minlength="6" required>
            </div>
            <button type="submit" class="btn btn-primary">Reset Password</button>
        </form>
    `);

    document.getElementById('reset-password-form').addEventListener('submit', async (e) => {
        e.preventDefault();
        const formData = new FormData(e.target);
        const newPassword = formData.get('new_password');
        const confirmPassword = formData.get('confirm_password');

        if (newPassword !== confirmPassword) {
            alert('Passwords do not match!');
            return;
        }

        const result = await api.put(`/users/${userId}`, { new_password: newPassword });
        if (result && result.success) {
            closeModal();
            alert('Password reset successfully');
        } else {
            alert(result.error || 'Failed to reset password');
        }
    });
}

document.getElementById('new-user-btn')?.addEventListener('click', () => showUserForm());

function showUserForm() {
    showModal('Create New User', `
        <form id="user-form">
            <div class="form-group">
                <label>Username</label>
                <input type="text" name="username" required>
                <small>Must be unique</small>
            </div>
            <div class="form-group">
                <label>Password</label>
                <input type="password" name="password" minlength="6" required>
                <small>Minimum 6 characters</small>
            </div>
            <div class="form-group">
                <label>Confirm Password</label>
                <input type="password" name="confirm_password" minlength="6" required>
            </div>
            <div class="form-group">
                <label>Role</label>
                <select name="role" required>
                    <option value="">Select role...</option>
                    <option value="user">User (Awareness Training Only) - Recommended</option>
                    <option value="admin">Admin (Full Access)</option>
                </select>
            </div>
            <div class="info-box" style="background: #f5f5f5; padding: 15px; margin: 15px 0; border-radius: 5px;">
                <p style="margin: 0 0 10px 0;"><strong>Admin:</strong> Full access to phishing simulations, assessments, user management, and all reports</p>
                <p style="margin: 0;"><strong>User:</strong> Access to awareness training, assessments, and own results only</p>
            </div>
            <button type="submit" class="btn btn-primary">Create User</button>
        </form>
    `);

    document.getElementById('user-form').addEventListener('submit', async (e) => {
        e.preventDefault();
        const formData = new FormData(e.target);
        const username = formData.get('username');
        const password = formData.get('password');
        const confirmPassword = formData.get('confirm_password');
        const role = formData.get('role');

        if (password !== confirmPassword) {
            alert('Passwords do not match!');
            return;
        }

        const result = await api.post('/users', { username, password, role });
        if (result && result.success) {
            closeModal();
            loadUsers();
        } else {
            alert(result.error || 'Failed to create user');
        }
    });
}

// Landing Page Cloner
function showClonePage() {
    showModal('Clone Landing Page from URL', `
        <form id="clone-page-form">
            <div class="form-group">
                <label>Website URL to Clone</label>
                <input type="url" name="url" placeholder="https://example.com" required>
                <small>Enter the URL of the website you want to clone (e.g., Facebook login, Google login)</small>
            </div>
            <div class="form-group">
                <label>Page Name (Optional)</label>
                <input type="text" name="name" placeholder="Leave empty to auto-generate">
            </div>
            <div class="info-box" style="background: #e3f2fd; padding: 15px; margin: 15px 0; border-radius: 5px;">
                <h4 style="margin-top: 0;">ℹ️ How it works</h4>
                <p>This feature will:</p>
                <ul>
                    <li>Fetch the HTML content from the URL</li>
                    <li>Convert relative URLs to absolute</li>
                    <li>Add tracking pixel automatically</li>
                    <li>Save as a new landing page</li>
                </ul>
            </div>
            <button type="submit" class="btn btn-primary">Clone Page</button>
        </form>
    `);

    document.getElementById('clone-page-form').addEventListener('submit', async (e) => {
        e.preventDefault();
        const formData = new FormData(e.target);
        const url = formData.get('url');
        const name = formData.get('name');

        // Show loading indicator
        const submitBtn = e.target.querySelector('button[type="submit"]');
        submitBtn.textContent = 'Cloning...';
        submitBtn.disabled = true;

        const result = await api.post('/pages/clone', { url, name });

        if (result && result.id) {
            alert('Landing page cloned successfully!');
            closeModal();
            loadPages();
        } else {
            alert(result?.error || 'Failed to clone page');
            submitBtn.textContent = 'Clone Page';
            submitBtn.disabled = false;
        }
    });
}

// Knowledge Base Chatbot
let chatHistory = [];

function initKnowledgeBase() {
    const chatMessages = document.getElementById('chat-messages');
    chatMessages.innerHTML = `
        <div class="chat-message bot-message">
            <strong>Security Assistant:</strong>
            <p>Hello! I'm your security awareness assistant. Ask me anything about phishing, cybersecurity, or online safety!</p>
        </div>
    `;
}

function addChatMessage(role, message) {
    const chatMessages = document.getElementById('chat-messages');
    const messageDiv = document.createElement('div');
    messageDiv.className = `chat-message ${role}-message`;
    messageDiv.innerHTML = `
        <strong>${role === 'user' ? 'You' : 'Security Assistant'}:</strong>
        <p>${message}</p>
    `;
    chatMessages.appendChild(messageDiv);
    chatMessages.scrollTop = chatMessages.scrollHeight;
}

async function sendChatMessage() {
    const input = document.getElementById('chat-input');
    const question = input.value.trim();

    if (!question) return;

    // Add user message to chat
    addChatMessage('user', question);
    input.value = '';

    // Add loading message
    const chatMessages = document.getElementById('chat-messages');
    const loadingDiv = document.createElement('div');
    loadingDiv.className = 'chat-message bot-message loading';
    loadingDiv.innerHTML = '<p>Thinking...</p>';
    chatMessages.appendChild(loadingDiv);
    chatMessages.scrollTop = chatMessages.scrollHeight;

    // Call API (API key is now stored in settings)
    const result = await api.call('/knowledge-base/chat', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({ question })
    });

    // Remove loading message
    chatMessages.removeChild(loadingDiv);

    if (result && result.answer) {
        addChatMessage('bot', result.answer);
    } else {
        const errorMsg = result?.error || 'Sorry, I encountered an error.';
        if (errorMsg.includes('API key') || errorMsg.includes('not configured')) {
            addChatMessage('bot', 'Gemini API key not configured. Please ask your administrator to configure the API key in Settings.\n\nGet a free API key from: https://aistudio.google.com/app/apikey');
        } else {
            addChatMessage('bot', 'Sorry, I encountered an error. Please try again.');
        }
    }
}

document.getElementById('chat-send-btn')?.addEventListener('click', sendChatMessage);
document.getElementById('chat-input')?.addEventListener('keypress', (e) => {
    if (e.key === 'Enter') {
        sendChatMessage();
    }
});

// AI Template Generation
async function generateRandomTemplate() {
    // Show loading modal
    showModal('Generating Template...', `
        <div style="text-align: center; padding: 40px;">
            <div class="loading-spinner" style="margin: 20px auto; border: 4px solid #f3f3f3; border-top: 4px solid #667eea; border-radius: 50%; width: 50px; height: 50px; animation: spin 1s linear infinite;"></div>
            <p style="margin-top: 20px; font-size: 16px;">AI is generating your phishing template...</p>
            <p style="color: #666; margin-top: 10px;">This may take 10-20 seconds</p>
        </div>
        <style>
            @keyframes spin {
                0% { transform: rotate(0deg); }
                100% { transform: rotate(360deg); }
            }
        </style>
    `);

    const result = await api.get(`/templates/generate-random`);

    if (result && result.success) {
        // Parse the generated template
        let templateData;
        try {
            templateData = JSON.parse(result.generated_text);
        } catch (e) {
            // If JSON parsing fails, try to extract from the response
            console.log('Raw response:', result.generated_text);

            // Try to extract JSON from markdown code blocks
            const jsonMatch = result.generated_text.match(/```(?:json)?\s*(\{[\s\S]*?\})\s*```/);
            if (jsonMatch) {
                try {
                    templateData = JSON.parse(jsonMatch[1]);
                } catch (e2) {
                    closeModal();
                    alert('Generated template, but could not parse the format. Please try again or check your API key.');
                    console.error('Parse error:', e2);
                    return;
                }
            } else {
                closeModal();
                alert('Generated template, but format was unexpected. Please try again.');
                console.log('Generated:', result.generated_text);
                return;
            }
        }

        // Escape HTML for safe display in form
        const escapeHtml = (text) => {
            const div = document.createElement('div');
            div.textContent = text;
            return div.innerHTML;
        };

        // Show template form pre-filled with generated data
        showModal('AI Generated Template - Review & Save', `
            <form id="ai-template-form">
                <div class="info-box" style="background: #e8f5e9; padding: 15px; margin-bottom: 15px; border-radius: 5px;">
                    <p><strong>✨ Scenario:</strong> ${escapeHtml(result.scenario)}</p>
                    <p><em>Review the generated template below and make any adjustments before saving.</em></p>
                </div>
                <div class="form-group">
                    <label>Template Name</label>
                    <input type="text" name="name" value="${escapeHtml(result.scenario)}" required>
                </div>
                <div class="form-group">
                    <label>Subject Line</label>
                    <input type="text" name="subject" value="${escapeHtml(templateData.subject || '')}" required>
                </div>
                <div class="form-group">
                    <label>HTML Content</label>
                    <textarea name="html" rows="10" required>${escapeHtml(templateData.html || '')}</textarea>
                </div>
                <div class="form-group">
                    <label>Plain Text Content</label>
                    <textarea name="text" rows="6">${escapeHtml(templateData.text || '')}</textarea>
                </div>
                <button type="submit" class="btn btn-primary">Save Template</button>
                <button type="button" class="btn btn-secondary" onclick="closeModal()">Cancel</button>
            </form>
        `);

        document.getElementById('ai-template-form').addEventListener('submit', async (e) => {
            e.preventDefault();
            const formData = new FormData(e.target);

            const data = {
                name: formData.get('name'),
                subject: formData.get('subject'),
                html: formData.get('html'),
                text: formData.get('text')
            };

            const saveResult = await api.post('/templates', data);
            if (saveResult) {
                closeModal();
                alert('Template saved successfully!');
                loadTemplates();
            }
        });
    } else {
        closeModal();
        const errorMsg = result?.error || 'Failed to generate template. Please check your API key and try again.';

        // If API key error, show helpful message
        if (errorMsg.includes('API key') || errorMsg.includes('not configured')) {
            alert('Gemini API key not configured.\n\nPlease go to Settings to configure your Google AI Studio API key.\n\nGet your free API key from: https://aistudio.google.com/app/apikey');
        } else {
            alert(errorMsg);
        }
    }
}

// PDF Download Functions
async function downloadCampaignPDF(campaignId) {
    window.open(`/api/campaigns/${campaignId}/pdf`, '_blank');
}

async function downloadAssessmentPDF(assessmentId) {
    window.open(`/api/assessments/${assessmentId}/pdf`, '_blank');
}

async function downloadResultPDF(attemptId) {
    window.open(`/api/user/results/${attemptId}/pdf`, '_blank');
}

// Credentials and Campaign Management Functions
async function viewCredentials(campaignId) {
    const credentials = await api.get(`/campaigns/${campaignId}/credentials`);

    if (!credentials || credentials.length === 0) {
        showModal('Harvested Credentials', '<p>No credentials have been submitted yet.</p>');
        return;
    }

    const credentialsTable = `
        <table class="table">
            <thead>
                <tr>
                    <th>Time</th>
                    <th>Name</th>
                    <th>Target Email</th>
                    <th>Submitted Email</th>
                    <th>Submitted Password</th>
                </tr>
            </thead>
            <tbody>
                ${credentials.map(cred => {
                    // Extract email and password from submitted data
                    const submittedEmail = cred.credentials?.email || cred.credentials?.username || cred.credentials?.user || '-';
                    const submittedPassword = cred.credentials?.password || cred.credentials?.pass || cred.credentials?.pwd || '-';

                    return `
                        <tr>
                            <td>${new Date(cred.time).toLocaleString()}</td>
                            <td>${cred.first_name} ${cred.last_name}</td>
                            <td>${cred.email}</td>
                            <td>${submittedEmail}</td>
                            <td>${submittedPassword}</td>
                        </tr>
                    `;
                }).join('')}
            </tbody>
        </table>
    `;

    showModal('Harvested Credentials', credentialsTable, 'large');
}

async function downloadCredentialsPDF(campaignId) {
    window.open(`/api/campaigns/${campaignId}/credentials-pdf`, '_blank');
}

async function endCampaign(campaignId) {
    if (!confirm('Are you sure you want to end this campaign early? This action cannot be undone.')) {
        return;
    }

    const result = await api.post(`/campaigns/${campaignId}/end`, {});
    if (result && result.success) {
        alert('Campaign ended successfully');
        loadCampaigns();
        closeModal();
    } else {
        alert('Failed to end campaign: ' + (result?.error || 'Unknown error'));
    }
}

async function launchCampaignNow(campaignId) {
    if (!confirm('Are you sure you want to launch this campaign now? Emails will be sent immediately.')) {
        return;
    }

    const result = await api.post(`/campaigns/${campaignId}/launch`, {});
    if (result && result.success) {
        alert('Campaign launched successfully! Emails are being sent.');
        loadCampaigns();
        closeModal();
    } else {
        alert('Failed to launch campaign: ' + (result?.error || 'Unknown error'));
    }
}

// Initialize
(async function() {
    const user = await api.get('/user');
    if (user && user.id) {
        app.user = user.username;
        app.userRole = user.role;
        document.getElementById('current-user').textContent = user.username;

        // Update UI for role
        updateUIForRole(user.role);

        showPage('dashboard');

        // Load appropriate view based on role
        if (user.role === 'admin') {
            showView('campaigns');
            loadCampaigns();
        } else {
            showView('awareness');
            loadAwarenessAssessments();
        }
    } else {
        showPage('login');
    }
})();

// Settings Management
async function loadSettings() {
    const settings = await api.get('/settings');
    if (settings && settings.gemini_api_key) {
        document.getElementById('gemini-api-key-input').value = settings.gemini_api_key;
    }
}

document.getElementById('save-gemini-api-key-btn')?.addEventListener('click', async () => {
    const apiKey = document.getElementById('gemini-api-key-input').value;
    const messageDiv = document.getElementById('settings-message');

    if (!apiKey) {
        messageDiv.innerHTML = '<p style="color: #e74c3c;">Please enter an API key</p>';
        return;
    }

    const result = await api.post('/settings', { key: 'gemini_api_key', value: apiKey });

    if (result && result.success) {
        messageDiv.innerHTML = '<p style="color: #27ae60;">✓ API key saved successfully!</p>';
        // Clear localStorage cache if exists
        localStorage.removeItem('gemini_api_key');
    } else {
        messageDiv.innerHTML = '<p style="color: #e74c3c;">✗ Failed to save API key</p>';
    }
});

document.getElementById('test-gemini-api-key-btn')?.addEventListener('click', async () => {
    const apiKey = document.getElementById('gemini-api-key-input').value;
    const messageDiv = document.getElementById('settings-message');

    if (!apiKey) {
        messageDiv.innerHTML = '<p style="color: #e74c3c;">Please enter an API key to test</p>';
        return;
    }

    messageDiv.innerHTML = '<p style="color: #3498db;">Testing connection...</p>';

    // Test by making a simple API call to Gemini
    try {
        const response = await fetch(`https://generativelanguage.googleapis.com/v1/models/gemini-2.5-flash:generateContent?key=${apiKey}`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                contents: [{
                    parts: [{ text: 'Hello' }]
                }]
            })
        });

        if (response.ok) {
            messageDiv.innerHTML = '<p style="color: #27ae60;">✓ API key is valid! Connection successful.</p>';
        } else {
            const error = await response.text();
            messageDiv.innerHTML = `<p style="color: #e74c3c;">✗ API key test failed: ${response.status} - Please check your API key</p>`;
        }
    } catch (error) {
        messageDiv.innerHTML = `<p style="color: #e74c3c;">✗ Connection failed: ${error.message}</p>`;
    }
});

// ============================================
// EMAIL PHISHING ANALYZER
// ============================================

function initEmailAnalyzer() {
    // Clear previous results
    document.getElementById('email-text-input').value = '';
    document.getElementById('analysis-result').style.display = 'none';
    document.getElementById('analysis-loading').style.display = 'none';
}

// Analyze Email Button
document.getElementById('analyze-email-btn')?.addEventListener('click', async () => {
    const emailText = document.getElementById('email-text-input').value.trim();
    const loadingDiv = document.getElementById('analysis-loading');
    const resultDiv = document.getElementById('analysis-result');

    if (!emailText) {
        alert('Please paste an email to analyze');
        return;
    }

    // Show loading, hide previous results
    loadingDiv.style.display = 'block';
    resultDiv.style.display = 'none';
    loadingDiv.innerHTML = '<p>🤖 Analyzing email with AI... This may take a few moments.</p>';

    try {
        const result = await api.post('/analyze-email', { email_text: emailText });

        loadingDiv.style.display = 'none';

        if (result && result.success) {
            displayAnalysisResult(result.result);
            resultDiv.style.display = 'block';
        } else if (result && result.error) {
            // Check for specific API errors
            const errorMsg = result.error;
            if (errorMsg.includes('503') || errorMsg.includes('overloaded') || errorMsg.includes('UNAVAILABLE')) {
                showApiOverloadMessage();
            } else if (errorMsg.includes('API key not configured')) {
                alert('⚠️ Gemini API Key Not Configured\n\nPlease ask your administrator to:\n1. Go to Settings\n2. Enter a Gemini API key\n3. Save and test the connection\n\nGet a free API key at: https://aistudio.google.com/app/apikey');
            } else {
                alert('Failed to analyze email: ' + errorMsg);
            }
        } else {
            alert('Failed to analyze email. Please check if Gemini API is configured in Settings.');
        }
    } catch (error) {
        loadingDiv.style.display = 'none';
        alert('Error analyzing email: ' + error.message);
    }
});

function showApiOverloadMessage() {
    const resultDiv = document.getElementById('analysis-result');
    resultDiv.innerHTML = `
        <div class="card" style="background: #fff3e0; border-left: 4px solid #ff9800;">
            <h3>⚠️ Google AI Service Temporarily Unavailable</h3>
            <p><strong>The Gemini API is experiencing high traffic (Error 503 - Service Overloaded)</strong></p>

            <h4 style="margin-top: 20px;">What This Means:</h4>
            <ul style="line-height: 1.8;">
                <li>✅ Your PhishSimAI application is working correctly</li>
                <li>⚠️ Google's AI servers are temporarily overloaded</li>
                <li>🔄 This is usually temporary and resolves within minutes</li>
            </ul>

            <h4 style="margin-top: 20px;">What To Do:</h4>
            <ol style="line-height: 1.8;">
                <li><strong>Wait 30-60 seconds</strong> and try again</li>
                <li>Try during off-peak hours (early morning/late evening)</li>
                <li>Check your API quota at: <a href="https://aistudio.google.com" target="_blank">Google AI Studio</a></li>
                <li>Consider using a paid API tier for better reliability</li>
            </ol>

            <div style="margin-top: 20px; padding: 15px; background: #e3f2fd; border-radius: 5px;">
                <h4>💡 Try These Features Instead (No API Required):</h4>
                <ul>
                    <li>📝 Complete Security Awareness Assessments</li>
                    <li>📄 Download PDF reports of your results</li>
                    <li>📊 View your assessment history</li>
                </ul>
            </div>

            <button class="btn btn-primary" onclick="document.getElementById('analyze-email-btn').click()" style="margin-top: 20px;">
                🔄 Try Again Now
            </button>
        </div>
    `;
    resultDiv.style.display = 'block';
}

// Clear Email Button
document.getElementById('clear-email-btn')?.addEventListener('click', () => {
    document.getElementById('email-text-input').value = '';
    document.getElementById('analysis-result').style.display = 'none';
});

function displayAnalysisResult(result) {
    const summaryDiv = document.getElementById('result-summary');
    const indicatorsDiv = document.getElementById('result-indicators');
    const explanationDiv = document.getElementById('result-explanation');
    const recommendationsDiv = document.getElementById('result-recommendations');

    // Determine the verdict and color
    const isPhishing = result.is_phishing;
    const riskLevel = result.risk_level || 'unknown';
    const confidenceScore = result.confidence_score || 0;

    let verdictColor = '#4CAF50'; // green for legitimate
    let verdictText = '✓ LEGITIMATE';
    let verdictBg = '#e8f5e9';

    if (isPhishing) {
        if (riskLevel === 'critical' || confidenceScore > 90) {
            verdictColor = '#d32f2f';
            verdictText = '⚠️ PHISHING - HIGH RISK';
            verdictBg = '#ffebee';
        } else if (riskLevel === 'high' || confidenceScore > 70) {
            verdictColor = '#f57c00';
            verdictText = '⚠️ LIKELY PHISHING';
            verdictBg = '#fff3e0';
        } else {
            verdictColor = '#ff9800';
            verdictText = '⚠️ SUSPICIOUS';
            verdictBg = '#fff8e1';
        }
    }

    // Helper function to escape HTML and preserve formatting
    const formatText = (text) => {
        if (!text) return '';
        return text
            .replace(/&/g, '&amp;')
            .replace(/</g, '&lt;')
            .replace(/>/g, '&gt;')
            .replace(/\n/g, '<br>');
    };

    // Summary Card
    summaryDiv.innerHTML = `
        <div style="background: ${verdictBg}; padding: 20px; border-radius: 8px; border-left: 5px solid ${verdictColor}; word-wrap: break-word; overflow-wrap: break-word;">
            <h2 style="margin: 0; color: ${verdictColor}; word-wrap: break-word;">${verdictText}</h2>
            <div style="margin-top: 15px; display: flex; gap: 30px; flex-wrap: wrap;">
                <div>
                    <p style="margin: 5px 0; color: #666; font-size: 14px;">Confidence Score</p>
                    <p style="margin: 0; font-size: 24px; font-weight: bold; color: ${verdictColor};">${confidenceScore}%</p>
                </div>
                <div>
                    <p style="margin: 5px 0; color: #666; font-size: 14px;">Risk Level</p>
                    <p style="margin: 0; font-size: 24px; font-weight: bold; color: ${verdictColor};">${riskLevel.toUpperCase()}</p>
                </div>
            </div>
        </div>
    `;

    // Phishing Indicators
    if (result.indicators && result.indicators.length > 0) {
        indicatorsDiv.innerHTML = `
            <div class="card" style="word-wrap: break-word; overflow-wrap: break-word;">
                <h3>🔍 Phishing Indicators Found</h3>
                <ul style="list-style-type: none; padding: 0; margin: 0;">
                    ${result.indicators.map(indicator => `
                        <li style="padding: 10px; margin: 5px 0; background: #fff3e0; border-left: 3px solid #ff9800; border-radius: 4px; word-wrap: break-word; overflow-wrap: break-word; max-width: 100%;">
                            <span style="display: inline-block; word-break: break-word;">⚠️ ${formatText(indicator)}</span>
                        </li>
                    `).join('')}
                </ul>
            </div>
        `;
    } else {
        indicatorsDiv.innerHTML = '';
    }

    // Explanation
    if (result.explanation) {
        explanationDiv.innerHTML = `
            <div class="card" style="word-wrap: break-word; overflow-wrap: break-word;">
                <h3>📋 Detailed Analysis</h3>
                <div style="line-height: 1.8; color: #333; word-wrap: break-word; overflow-wrap: break-word; white-space: pre-wrap; max-width: 100%;">
                    ${formatText(result.explanation)}
                </div>
            </div>
        `;
    } else {
        explanationDiv.innerHTML = '';
    }

    // Recommendations
    if (result.recommendations && result.recommendations.length > 0) {
        recommendationsDiv.innerHTML = `
            <div class="card" style="background: #e3f2fd; border-left: 4px solid #2196F3; word-wrap: break-word; overflow-wrap: break-word;">
                <h3>💡 Recommendations</h3>
                <ul style="line-height: 1.8; color: #333; padding-left: 20px;">
                    ${result.recommendations.map(rec => `
                        <li style="margin: 8px 0; word-wrap: break-word; overflow-wrap: break-word;">
                            ${formatText(rec)}
                        </li>
                    `).join('')}
                </ul>
            </div>
        `;
    } else {
        recommendationsDiv.innerHTML = '';
    }
}

// ============================================
// EMAIL ANALYZER - AWARENESS PAGE INTEGRATION
// ============================================

function initEmailAnalyzerAwareness() {
    // Clear previous results
    const emailInput = document.getElementById('email-text-awareness');
    const resultDiv = document.getElementById('analysis-result-awareness');
    const loadingDiv = document.getElementById('analysis-loading-awareness');

    if (emailInput) emailInput.value = '';
    if (resultDiv) resultDiv.style.display = 'none';
    if (loadingDiv) loadingDiv.style.display = 'none';

    // Setup event listeners
    setupEmailAnalyzerAwarenessHandlers();
}

function setupEmailAnalyzerAwarenessHandlers() {
    // Analyze Email Button (Awareness Page)
    const analyzeBtn = document.getElementById('analyze-email-awareness-btn');
    const clearBtn = document.getElementById('clear-email-awareness-btn');

    if (analyzeBtn) {
        // Remove old listeners
        const newAnalyzeBtn = analyzeBtn.cloneNode(true);
        analyzeBtn.parentNode.replaceChild(newAnalyzeBtn, analyzeBtn);

        newAnalyzeBtn.addEventListener('click', async () => {
            const emailText = document.getElementById('email-text-awareness').value.trim();
            const loadingDiv = document.getElementById('analysis-loading-awareness');
            const resultDiv = document.getElementById('analysis-result-awareness');

            if (!emailText) {
                alert('Please paste an email to analyze');
                return;
            }

            // Show loading, hide previous results
            loadingDiv.style.display = 'block';
            resultDiv.style.display = 'none';
            loadingDiv.innerHTML = '<p>🤖 Analyzing email with AI... This may take a few moments.</p>';

            try {
                const result = await api.post('/analyze-email', { email_text: emailText });

                loadingDiv.style.display = 'none';

                if (result && result.success) {
                    displayAnalysisResultAwareness(result.result);
                    resultDiv.style.display = 'block';
                    // Scroll to results
                    resultDiv.scrollIntoView({ behavior: 'smooth', block: 'nearest' });
                } else if (result && result.error) {
                    // Check for specific API errors
                    const errorMsg = result.error;
                    if (errorMsg.includes('503') || errorMsg.includes('overloaded') || errorMsg.includes('UNAVAILABLE')) {
                        showApiOverloadMessageAwareness();
                    } else if (errorMsg.includes('API key not configured')) {
                        alert('⚠️ Gemini API Key Not Configured\n\nPlease configure your own API key:\n1. Go to "My Settings" in the menu\n2. Paste your Gemini API key\n3. Click "Save My API Key"\n4. Click "Test Connection" to verify\n\nGet your FREE API key at: https://aistudio.google.com/app/apikey');
                    } else {
                        alert('Failed to analyze email: ' + errorMsg);
                    }
                } else {
                    alert('Failed to analyze email. Please configure your API key in My Settings.');
                }
            } catch (error) {
                loadingDiv.style.display = 'none';
                alert('Error analyzing email: ' + error.message);
            }
        });
    }

    if (clearBtn) {
        // Remove old listeners
        const newClearBtn = clearBtn.cloneNode(true);
        clearBtn.parentNode.replaceChild(newClearBtn, clearBtn);

        newClearBtn.addEventListener('click', () => {
            document.getElementById('email-text-awareness').value = '';
            document.getElementById('analysis-result-awareness').style.display = 'none';
        });
    }
}

function displayAnalysisResultAwareness(result) {
    const summaryDiv = document.getElementById('result-summary-awareness');
    const indicatorsDiv = document.getElementById('result-indicators-awareness');
    const explanationDiv = document.getElementById('result-explanation-awareness');
    const recommendationsDiv = document.getElementById('result-recommendations-awareness');

    // Determine the verdict and color
    const isPhishing = result.is_phishing;
    const riskLevel = result.risk_level || 'unknown';
    const confidenceScore = result.confidence_score || 0;

    let verdictColor = '#4CAF50'; // green for legitimate
    let verdictText = '✓ LEGITIMATE';
    let verdictBg = '#e8f5e9';

    if (isPhishing) {
        if (riskLevel === 'critical' || confidenceScore > 90) {
            verdictColor = '#d32f2f';
            verdictText = '⚠️ PHISHING - HIGH RISK';
            verdictBg = '#ffebee';
        } else if (riskLevel === 'high' || confidenceScore > 70) {
            verdictColor = '#f57c00';
            verdictText = '⚠️ LIKELY PHISHING';
            verdictBg = '#fff3e0';
        } else {
            verdictColor = '#ff9800';
            verdictText = '⚠️ SUSPICIOUS';
            verdictBg = '#fff8e1';
        }
    }

    // Helper function to escape HTML and preserve formatting
    const formatText = (text) => {
        if (!text) return '';
        return text
            .replace(/&/g, '&amp;')
            .replace(/</g, '&lt;')
            .replace(/>/g, '&gt;')
            .replace(/\n/g, '<br>');
    };

    // Summary Card
    summaryDiv.innerHTML = `
        <div style="background: ${verdictBg}; padding: 20px; border-radius: 8px; border-left: 5px solid ${verdictColor}; word-wrap: break-word; overflow-wrap: break-word;">
            <h3 style="margin: 0; color: ${verdictColor}; word-wrap: break-word;">${verdictText}</h3>
            <div style="margin-top: 15px; display: flex; gap: 30px; flex-wrap: wrap;">
                <div>
                    <p style="margin: 5px 0; color: #666; font-size: 14px;">Confidence Score</p>
                    <p style="margin: 0; font-size: 24px; font-weight: bold; color: ${verdictColor};">${confidenceScore}%</p>
                </div>
                <div>
                    <p style="margin: 5px 0; color: #666; font-size: 14px;">Risk Level</p>
                    <p style="margin: 0; font-size: 24px; font-weight: bold; color: ${verdictColor};">${riskLevel.toUpperCase()}</p>
                </div>
            </div>
        </div>
    `;

    // Phishing Indicators
    if (result.indicators && result.indicators.length > 0) {
        indicatorsDiv.innerHTML = `
            <div style="background: white; padding: 15px; border-radius: 8px; border: 1px solid #ddd; word-wrap: break-word; overflow-wrap: break-word;">
                <h4 style="margin-top: 0; color: #333;">🔍 Phishing Indicators Found</h4>
                <ul style="list-style-type: none; padding: 0; margin: 0;">
                    ${result.indicators.map(indicator => `
                        <li style="padding: 10px; margin: 5px 0; background: #fff3e0; border-left: 3px solid #ff9800; border-radius: 4px; word-wrap: break-word; overflow-wrap: break-word; max-width: 100%;">
                            <span style="display: inline-block; word-break: break-word; color: #333;">⚠️ ${formatText(indicator)}</span>
                        </li>
                    `).join('')}
                </ul>
            </div>
        `;
    } else {
        indicatorsDiv.innerHTML = '';
    }

    // Explanation
    if (result.explanation) {
        explanationDiv.innerHTML = `
            <div style="background: white; padding: 15px; border-radius: 8px; border: 1px solid #ddd; word-wrap: break-word; overflow-wrap: break-word;">
                <h4 style="margin-top: 0; color: #333;">📋 Detailed Analysis</h4>
                <div style="line-height: 1.8; color: #333; word-wrap: break-word; overflow-wrap: break-word; white-space: pre-wrap; max-width: 100%;">
                    ${formatText(result.explanation)}
                </div>
            </div>
        `;
    } else {
        explanationDiv.innerHTML = '';
    }

    // Recommendations
    if (result.recommendations && result.recommendations.length > 0) {
        recommendationsDiv.innerHTML = `
            <div style="background: #e3f2fd; padding: 15px; border-radius: 8px; border-left: 4px solid #2196F3; word-wrap: break-word; overflow-wrap: break-word;">
                <h4 style="margin-top: 0; color: #333;">💡 Recommendations</h4>
                <ul style="line-height: 1.8; color: #333; padding-left: 20px;">
                    ${result.recommendations.map(rec => `
                        <li style="margin: 8px 0; word-wrap: break-word; overflow-wrap: break-word;">
                            ${formatText(rec)}
                        </li>
                    `).join('')}
                </ul>
            </div>
        `;
    } else {
        recommendationsDiv.innerHTML = '';
    }
}

function showApiOverloadMessageAwareness() {
    const resultDiv = document.getElementById('analysis-result-awareness');
    resultDiv.innerHTML = `
        <div style="background: #fff3e0; padding: 20px; border-radius: 8px; border-left: 4px solid #ff9800;">
            <h4 style="color: #333; margin-top: 0;">⚠️ Google AI Service Temporarily Unavailable</h4>
            <p style="color: #333;"><strong>The Gemini API is experiencing high traffic (Error 503 - Service Overloaded)</strong></p>

            <h5 style="margin-top: 20px; color: #333;">What This Means:</h5>
            <ul style="line-height: 1.8; color: #333;">
                <li>✅ Your PhishSimAI application is working correctly</li>
                <li>⚠️ Google's AI servers are temporarily overloaded</li>
                <li>🔄 This is usually temporary and resolves within minutes</li>
            </ul>

            <h5 style="margin-top: 20px; color: #333;">What To Do:</h5>
            <ol style="line-height: 1.8; color: #333;">
                <li><strong>Wait 30-60 seconds</strong> and try again</li>
                <li>Try during off-peak hours (early morning/late evening)</li>
                <li>Check your API quota at: <a href="https://aistudio.google.com" target="_blank">Google AI Studio</a></li>
            </ol>

            <button class="btn btn-primary" onclick="document.getElementById('analyze-email-awareness-btn').click()" style="margin-top: 20px; background: #667eea; border: none;">
                🔄 Try Again Now
            </button>
        </div>
    `;
    resultDiv.style.display = 'block';
}

// Inline Chatbot Functions (on Awareness page)
function initInlineChatbot() {
    const chatMessages = document.getElementById('inline-chat-messages');
    if (!chatMessages) return;

    chatMessages.innerHTML = `
        <div class="chat-message bot-message" style="background: #e3f2fd; padding: 12px; border-radius: 8px; margin-bottom: 10px;">
            <strong style="color: #1976d2;">Security Assistant:</strong>
            <p style="margin: 5px 0 0 0; color: #333;">Hello! I'm your security awareness assistant. Ask me anything about phishing, cybersecurity, or online safety!</p>
        </div>
    `;

    setupInlineChatHandlers();
}

function setupInlineChatHandlers() {
    const sendBtn = document.getElementById('inline-chat-send-btn');
    const chatInput = document.getElementById('inline-chat-input');

    if (!sendBtn || !chatInput) return;

    // Remove old listeners
    const newSendBtn = sendBtn.cloneNode(true);
    sendBtn.parentNode.replaceChild(newSendBtn, sendBtn);

    const newChatInput = chatInput.cloneNode(true);
    chatInput.parentNode.replaceChild(newChatInput, chatInput);

    // Add click handler
    newSendBtn.addEventListener('click', sendInlineChatMessage);

    // Add enter key handler
    newChatInput.addEventListener('keypress', (e) => {
        if (e.key === 'Enter') {
            sendInlineChatMessage();
        }
    });
}

async function sendInlineChatMessage() {
    const input = document.getElementById('inline-chat-input');
    const question = input.value.trim();

    if (!question) return;

    const chatMessages = document.getElementById('inline-chat-messages');

    // Add user message
    addInlineChatMessage('user', question);
    input.value = '';

    // Add loading message
    const loadingDiv = document.createElement('div');
    loadingDiv.className = 'chat-message bot-message loading';
    loadingDiv.style.cssText = 'background: #f5f5f5; padding: 12px; border-radius: 8px; margin-bottom: 10px;';
    loadingDiv.innerHTML = '<p style="margin: 0; color: #666;">🤔 Thinking...</p>';
    chatMessages.appendChild(loadingDiv);
    chatMessages.scrollTop = chatMessages.scrollHeight;

    // Call API
    const result = await api.post('/knowledge-base/chat', { question });

    // Remove loading message
    chatMessages.removeChild(loadingDiv);

    if (result && result.answer) {
        addInlineChatMessage('bot', result.answer);
    } else {
        addInlineChatMessage('bot', 'Sorry, I encountered an error. Please make sure your API key is configured in My Settings.');
    }
}

function addInlineChatMessage(role, message) {
    const chatMessages = document.getElementById('inline-chat-messages');
    const messageDiv = document.createElement('div');
    messageDiv.className = `chat-message ${role}-message`;

    if (role === 'user') {
        messageDiv.style.cssText = 'background: #4CAF50; color: white; padding: 12px; border-radius: 8px; margin-bottom: 10px; margin-left: 40px;';
        messageDiv.innerHTML = `
            <strong>You:</strong>
            <p style="margin: 5px 0 0 0;">${escapeHtml(message)}</p>
        `;
    } else {
        messageDiv.style.cssText = 'background: #e3f2fd; padding: 12px; border-radius: 8px; margin-bottom: 10px;';
        messageDiv.innerHTML = `
            <strong style="color: #1976d2;">Security Assistant:</strong>
            <p style="margin: 5px 0 0 0; color: #333;">${escapeHtml(message)}</p>
        `;
    }

    chatMessages.appendChild(messageDiv);
    chatMessages.scrollTop = chatMessages.scrollHeight;
}

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

// User Settings Functions
async function loadUserSettings() {
    // Simple settings load - matches admin page style
    const messageDiv = document.getElementById('user-settings-message');
    if (messageDiv) {
        messageDiv.innerHTML = '';
    }
}

async function saveUserAPIKey() {
    const apiKeyInput = document.getElementById('user-api-key-input');
    const apiKey = apiKeyInput.value.trim();
    const messageDiv = document.getElementById('user-settings-message');

    if (!apiKey) {
        messageDiv.innerHTML = '<p style="color: #d32f2f;">Please enter a valid API key</p>';
        return;
    }

    const btn = document.getElementById('save-api-key-btn');
    btn.disabled = true;
    btn.textContent = 'Saving...';
    messageDiv.innerHTML = '<p style="color: #666;">Saving API key...</p>';

    const result = await api.post('/user/settings/api-key', { api_key: apiKey });

    if (result && result.success) {
        messageDiv.innerHTML = '<p style="color: #4CAF50;">✓ API key saved successfully!</p>';
        apiKeyInput.value = ''; // Clear input for security
    } else {
        messageDiv.innerHTML = '<p style="color: #d32f2f;">Failed to save API key. Please try again.</p>';
    }

    btn.disabled = false;
    btn.textContent = 'Save API Key';
}

async function testUserAPIKey() {
    const testBtn = document.getElementById('test-api-key-btn');
    const messageDiv = document.getElementById('user-settings-message');

    testBtn.disabled = true;
    testBtn.textContent = 'Testing...';
    messageDiv.innerHTML = '<p style="color: #666;">Testing API connection...</p>';

    // Test by asking a simple question to the knowledge base
    const result = await api.post('/knowledge-base/chat', {
        question: 'What is phishing?'
    });

    if (result && result.answer) {
        messageDiv.innerHTML = '<p style="color: #4CAF50;">✓ API key is working correctly!</p>';
    } else {
        messageDiv.innerHTML = '<p style="color: #d32f2f;">API key test failed. Please check your configuration.</p>';
    }

    testBtn.disabled = false;
    testBtn.textContent = 'Test Connection';
}

// Initialize event listeners for user settings
document.addEventListener('DOMContentLoaded', () => {
    const saveApiKeyBtn = document.getElementById('save-api-key-btn');
    if (saveApiKeyBtn) {
        saveApiKeyBtn.addEventListener('click', saveUserAPIKey);
    }

    const testApiKeyBtn = document.getElementById('test-api-key-btn');
    if (testApiKeyBtn) {
        testApiKeyBtn.addEventListener('click', testUserAPIKey);
    }
});
