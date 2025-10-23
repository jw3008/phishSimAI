// Main App State
const app = {
    user: null,
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
        document.getElementById('current-user').textContent = result.user;
        showPage('dashboard');
        loadCampaigns();
    } else {
        errorDiv.textContent = 'Invalid credentials';
    }
});

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
                    <div class="stat-value">${c.stats.opened}</div>
                    <div class="stat-label">Opened</div>
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
                    <th>Opened</th>
                    <th>Clicked</th>
                    <th>Submitted</th>
                </tr>
            </thead>
            <tbody>
                ${campaign.results.map(r => `
                    <tr>
                        <td>${r.first_name} ${r.last_name}</td>
                        <td>${r.email}</td>
                        <td><span class="badge badge-${r.status}">${r.status}</span></td>
                        <td>${r.send_date ? new Date(r.send_date).toLocaleString() : '-'}</td>
                        <td>${r.open_date ? new Date(r.open_date).toLocaleString() : '-'}</td>
                        <td>${r.click_date ? new Date(r.click_date).toLocaleString() : '-'}</td>
                        <td>${r.submit_date ? new Date(r.submit_date).toLocaleString() : '-'}</td>
                    </tr>
                `).join('')}
            </tbody>
        </table>
    ` : '<p>No results yet</p>';

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
                <div class="stat-value">${campaign.stats.open_rate}%</div>
                <div class="stat-label">Open Rate</div>
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
                <small>Variables: {{.FirstName}}, {{.LastName}}, {{.URL}}</small>
            </div>
            <div class="form-group">
                <label>HTML Content</label>
                <textarea name="html" rows="10" required>${template?.html || ''}</textarea>
                <small>Variables: {{.FirstName}}, {{.LastName}}, {{.URL}}</small>
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
                    <div class="card" style="padding: 10px; margin-bottom: 10px;">
                        <input type="text" placeholder="First Name" value="${t.first_name}" data-field="first_name" data-index="${i}" style="margin-bottom: 5px;">
                        <input type="text" placeholder="Last Name" value="${t.last_name}" data-field="last_name" data-index="${i}" style="margin-bottom: 5px;">
                        <input type="email" placeholder="Email" value="${t.email}" data-field="email" data-index="${i}" required style="margin-bottom: 5px;">
                        <input type="text" placeholder="Position" value="${t.position}" data-field="position" data-index="${i}">
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
    div.innerHTML = `
        <input type="text" placeholder="First Name" data-field="first_name" data-index="${index}" style="margin-bottom: 5px;">
        <input type="text" placeholder="Last Name" data-field="last_name" data-index="${index}" style="margin-bottom: 5px;">
        <input type="email" placeholder="Email" data-field="email" data-index="${index}" required style="margin-bottom: 5px;">
        <input type="text" placeholder="Position" data-field="position" data-index="${index}">
    `;
    container.appendChild(div);
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
            <div class="form-group">
                <div class="checkbox-group">
                    <input type="checkbox" name="ignore_cert_errors" id="ignore_cert" ${smtp?.ignore_cert_errors ? 'checked' : ''}>
                    <label for="ignore_cert">Ignore Certificate Errors</label>
                </div>
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
            from_address: formData.get('from_address'),
            ignore_cert_errors: formData.get('ignore_cert_errors') === 'on'
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

// Initialize
(async function() {
    const user = await api.get('/user');
    if (user && user.id) {
        app.user = user.username;
        document.getElementById('current-user').textContent = user.username;
        showPage('dashboard');
        loadCampaigns();
    } else {
        showPage('login');
    }
})();
