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
            case 'awareness': loadAwarenessAssessments(); break;
            case 'my-results': loadMyResults(); break;
            case 'knowledge-base': initKnowledgeBase(); break;
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
                        `<button class="btn btn-secondary" onclick="viewMyResult(${a.attempt_id})">View Result</button>`
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
                        <label style="display: flex; align-items: center; padding: 15px; border: 2px solid #ddd; border-radius: 5px; cursor: pointer;">
                            <input type="radio" name="answer" value="${opt.id}" required style="margin-right: 10px;">
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

    showModal(`Taking Assessment: ${assessment.title}`, renderQuestion());

    // Handle form submission
    document.getElementById('question-form').addEventListener('submit', async (e) => {
        e.preventDefault();
        const formData = new FormData(e.target);
        const selectedOption = parseInt(formData.get('answer'));

        const q = assessment.questions[currentQuestionIndex];
        userAnswers[q.id] = selectedOption;

        // Save the response to backend
        await api.post(`/user/assessments/attempt/${attemptId}/submit`, {
            question_id: q.id,
            selected_option_id: selectedOption
        });

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
                        <h2 style="color: #4CAF50;">Congratulations!</h2>
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
                        <button class="btn btn-primary" onclick="viewMyResult(${result.attempt_id})">View Detailed Results</button>
                        <button class="btn btn-secondary" onclick="closeModal(); loadAwarenessAssessments();">Close</button>
                    </div>
                `);
            }
        }
    });

    function setupQuestionHandlers() {
        document.getElementById('question-form').addEventListener('submit', async (e) => {
            e.preventDefault();
            const formData = new FormData(e.target);
            const selectedOption = parseInt(formData.get('answer'));

            const q = assessment.questions[currentQuestionIndex];
            userAnswers[q.id] = selectedOption;

            await api.post(`/user/assessments/attempt/${attemptId}/submit`, {
                question_id: q.id,
                selected_option_id: selectedOption
            });

            if (currentQuestionIndex < assessment.questions.length - 1) {
                currentQuestionIndex++;
                showModal(`Taking Assessment: ${assessment.title}`, renderQuestion());
                setupQuestionHandlers();
            } else {
                const result = await api.post(`/user/assessments/attempt/${attemptId}/complete`);
                if (result && result.success) {
                    showModal('Assessment Completed!', `
                        <div style="text-align: center;">
                            <h2 style="color: #4CAF50;">Congratulations!</h2>
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
                            <button class="btn btn-primary" onclick="viewMyResult(${result.attempt_id})">View Detailed Results</button>
                            <button class="btn btn-secondary" onclick="closeModal(); loadAwarenessAssessments();">Close</button>
                        </div>
                    `);
                }
            }
        });

        const prevBtn = document.getElementById('prev-btn');
        if (prevBtn) {
            prevBtn.addEventListener('click', () => {
                currentQuestionIndex--;
                showModal(`Taking Assessment: ${assessment.title}`, renderQuestion());
                setupQuestionHandlers();
            });
        }
    }

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
    `);
}

function downloadResultPDF(attemptId) {
    // Open PDF download in new window
    window.open(`/api/user/results/${attemptId}/pdf`, '_blank');
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
            <div class="info-box" style="background: #e3f2fd; padding: 15px; margin: 15px 0; border-radius: 5px;">
                <h4 style="margin-top: 0;">Role Permissions (Least Privilege Principle)</h4>
                <p><strong>Admin:</strong></p>
                <ul>
                    <li>Full access to phishing simulation campaigns</li>
                    <li>Create and manage assessments</li>
                    <li>Manage users</li>
                    <li>View all statistics and reports</li>
                </ul>
                <p><strong>User:</strong></p>
                <ul>
                    <li>Access to awareness training only</li>
                    <li>Take published assessments</li>
                    <li>View own results</li>
                    <li>NO access to phishing simulation or admin features</li>
                </ul>
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
            <div class="info-box" style="background: #fff3cd; padding: 15px; margin: 15px 0; border-radius: 5px; border-left: 4px solid #ffc107;">
                <h4 style="margin-top: 0;">🔒 Least Privilege Principle</h4>
                <p>By default, create users with <strong>"User"</strong> role. Only assign <strong>"Admin"</strong> role when necessary.</p>
                <p><strong>User role grants:</strong></p>
                <ul>
                    <li>✅ Access to security awareness training</li>
                    <li>✅ Ability to take assessments</li>
                    <li>✅ View own results and performance</li>
                    <li>❌ NO access to phishing simulation features</li>
                    <li>❌ NO admin capabilities</li>
                </ul>
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

    // Get Gemini API key from user if not set
    let apiKey = localStorage.getItem('gemini_api_key');
    if (!apiKey) {
        apiKey = prompt('Please enter your Gemini API Key (get it from https://aistudio.google.com/app/apikey):');
        if (apiKey) {
            localStorage.setItem('gemini_api_key', apiKey);
        } else {
            chatMessages.removeChild(loadingDiv);
            addChatMessage('bot', 'API key is required to use the chatbot. Please refresh and provide your Gemini API key.');
            return;
        }
    }

    // Call API
    const result = await api.call('/knowledge-base/chat', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'X-Gemini-API-Key': apiKey
        },
        body: JSON.stringify({ question })
    });

    // Remove loading message
    chatMessages.removeChild(loadingDiv);

    if (result && result.answer) {
        addChatMessage('bot', result.answer);
    } else {
        addChatMessage('bot', 'Sorry, I encountered an error. Please try again or check your API key.');
        if (result?.error?.includes('API key')) {
            localStorage.removeItem('gemini_api_key');
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
    let apiKey = localStorage.getItem('gemini_api_key');
    if (!apiKey) {
        apiKey = prompt('Please enter your Gemini API Key (get it from https://aistudio.google.com/app/apikey):');
        if (apiKey) {
            localStorage.setItem('gemini_api_key', apiKey);
        } else {
            alert('API key is required to generate templates');
            return;
        }
    }

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

    const result = await api.get(`/templates/generate-random?api_key=${apiKey}`);

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
                html_content: formData.get('html'),
                text_content: formData.get('text')
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
        alert(errorMsg);

        // If API key error, clear it so user can re-enter
        if (errorMsg.includes('API key') || errorMsg.includes('api_key')) {
            localStorage.removeItem('gemini_api_key');
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
