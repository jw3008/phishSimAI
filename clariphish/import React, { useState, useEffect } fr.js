import React, { useState, useEffect } from 'react';
import { Mail, Users, BarChart3, Settings, Plus, Eye, Edit, Trash2, Play, AlertCircle, CheckCircle, XCircle, Clock, Send } from 'lucide-react';

const ClariPhish = () => {
  const [activeTab, setActiveTab] = useState('dashboard');
  const [campaigns, setCampaigns] = useState([]);
  const [templates, setTemplates] = useState([]);
  const [groups, setGroups] = useState([]);
  const [loading, setLoading] = useState(true);
  const [showNewCampaign, setShowNewCampaign] = useState(false);
  const [showNewTemplate, setShowNewTemplate] = useState(false);
  const [showNewGroup, setShowNewGroup] = useState(false);

  const [newCampaign, setNewCampaign] = useState({
    name: '',
    template: '',
    group: '',
    url: ''
  });

  const [newTemplate, setNewTemplate] = useState({
    name: '',
    subject: '',
    body: '',
    category: 'credential'
  });

  const [newGroup, setNewGroup] = useState({
    name: '',
    emails: ''
  });

  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    setLoading(true);
    try {
      const campaignsData = await window.storage.get('campaigns');
      if (campaignsData) setCampaigns(JSON.parse(campaignsData.value));
      
      const templatesData = await window.storage.get('templates');
      if (templatesData) {
        setTemplates(JSON.parse(templatesData.value));
      } else {
        const defaults = [
          { id: 1, name: 'Password Reset', subject: 'Urgent: Password Reset Required', body: '<p>Reset your password</p>', category: 'credential' },
          { id: 2, name: 'IT Support', subject: 'Action Required', body: '<p>Verify account</p>', category: 'support' }
        ];
        setTemplates(defaults);
        await window.storage.set('templates', JSON.stringify(defaults));
      }
      
      const groupsData = await window.storage.get('groups');
      if (groupsData) setGroups(JSON.parse(groupsData.value));
    } catch (e) {
      console.log('Starting fresh');
    }
    setLoading(false);
  };

  const saveCampaigns = async (data) => {
    await window.storage.set('campaigns', JSON.stringify(data));
  };

  const saveTemplates = async (data) => {
    await window.storage.set('templates', JSON.stringify(data));
  };

  const saveGroups = async (data) => {
    await window.storage.set('groups', JSON.stringify(data));
  };

  const handleCreateCampaign = async () => {
    if (!newCampaign.name || !newCampaign.template || !newCampaign.group) return;
    
    const campaign = {
      id: Date.now(),
      name: newCampaign.name,
      templateId: newCampaign.template,
      groupId: newCampaign.group,
      status: 'scheduled',
      created: new Date().toISOString().split('T')[0],
      sent: 0,
      opened: 0,
      clicked: 0,
      submitted: 0
    };
    
    const updated = [...campaigns, campaign];
    setCampaigns(updated);
    await saveCampaigns(updated);
    setNewCampaign({ name: '', template: '', group: '', url: '' });
    setShowNewCampaign(false);
  };

  const handleLaunchCampaign = async (id) => {
    const campaign = campaigns.find(c => c.id === id);
    const group = groups.find(g => g.id === parseInt(campaign.groupId));
    if (!group) return;
    
    const count = group.emails.split('\n').filter(e => e.trim()).length;
    const updated = campaigns.map(c => c.id === id ? { ...c, status: 'running', sent: count } : c);
    setCampaigns(updated);
    await saveCampaigns(updated);
    alert('Campaign launched!');
  };

  const handleCreateTemplate = async () => {
    if (!newTemplate.name || !newTemplate.subject) return;
    
    const template = {
      id: Date.now(),
      name: newTemplate.name,
      subject: newTemplate.subject,
      body: newTemplate.body,
      category: newTemplate.category
    };
    
    const updated = [...templates, template];
    setTemplates(updated);
    await saveTemplates(updated);
    setNewTemplate({ name: '', subject: '', body: '', category: 'credential' });
    setShowNewTemplate(false);
  };

  const handleCreateGroup = async () => {
    if (!newGroup.name || !newGroup.emails) return;
    
    const count = newGroup.emails.split('\n').filter(e => e.trim()).length;
    const group = {
      id: Date.now(),
      name: newGroup.name,
      emails: newGroup.emails,
      members: count
    };
    
    const updated = [...groups, group];
    setGroups(updated);
    await saveGroups(updated);
    setNewGroup({ name: '', emails: '' });
    setShowNewGroup(false);
  };

  const deleteCampaign = async (id) => {
    const updated = campaigns.filter(c => c.id !== id);
    setCampaigns(updated);
    await saveCampaigns(updated);
  };

  const deleteTemplate = async (id) => {
    const updated = templates.filter(t => t.id !== id);
    setTemplates(updated);
    await saveTemplates(updated);
  };

  const deleteGroup = async (id) => {
    const updated = groups.filter(g => g.id !== id);
    setGroups(updated);
    await saveGroups(updated);
  };

  const getStatusIcon = (status) => {
    if (status === 'completed') return <CheckCircle className="w-4 h-4 text-green-500" />;
    if (status === 'running') return <Clock className="w-4 h-4 text-blue-500" />;
    if (status === 'scheduled') return <Clock className="w-4 h-4 text-yellow-500" />;
    return <XCircle className="w-4 h-4 text-gray-500" />;
  };

  const calculateStats = () => {
    const total = campaigns.reduce((acc, c) => ({
      sent: acc.sent + c.sent,
      opened: acc.opened + c.opened,
      clicked: acc.clicked + c.clicked,
      submitted: acc.submitted + c.submitted
    }), { sent: 0, opened: 0, clicked: 0, submitted: 0 });

    return {
      openRate: total.sent > 0 ? ((total.opened / total.sent) * 100).toFixed(1) : 0,
      clickRate: total.sent > 0 ? ((total.clicked / total.sent) * 100).toFixed(1) : 0,
      submitRate: total.sent > 0 ? ((total.submitted / total.sent) * 100).toFixed(1) : 0
    };
  };

  const stats = calculateStats();

  if (loading) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto mb-4"></div>
          <p className="text-gray-600">Loading...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="bg-blue-600 text-white p-4 shadow-lg">
        <div className="max-w-7xl mx-auto flex items-center justify-between">
          <div className="flex items-center gap-3">
            <Mail className="w-8 h-8" />
            <h1 className="text-2xl font-bold">ClariPhish</h1>
          </div>
          <div className="text-sm">Educational Phishing Simulation Platform</div>
        </div>
      </div>

      <div className="bg-white border-b shadow-sm">
        <div className="max-w-7xl mx-auto">
          <nav className="flex gap-1 p-2">
            {[
              { id: 'dashboard', label: 'Dashboard', icon: BarChart3 },
              { id: 'campaigns', label: 'Campaigns', icon: Send },
              { id: 'templates', label: 'Templates', icon: Mail },
              { id: 'groups', label: 'Groups', icon: Users },
              { id: 'settings', label: 'Settings', icon: Settings }
            ].map(tab => {
              const Icon = tab.icon;
              return (
                <button
                  key={tab.id}
                  onClick={() => setActiveTab(tab.id)}
                  className={`flex items-center gap-2 px-4 py-2 rounded transition-colors ${
                    activeTab === tab.id ? 'bg-blue-50 text-blue-600 font-medium' : 'text-gray-600 hover:bg-gray-50'
                  }`}
                >
                  <Icon className="w-4 h-4" />
                  {tab.label}
                </button>
              );
            })}
          </nav>
        </div>
      </div>

      <div className="max-w-7xl mx-auto p-6">
        {activeTab === 'dashboard' && (
          <div className="space-y-6">
            <h2 className="text-2xl font-bold text-gray-800">Dashboard</h2>
            
            <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
              <div className="bg-white p-6 rounded-lg shadow border border-gray-200">
                <div className="text-sm text-gray-500 mb-1">Total Campaigns</div>
                <div className="text-3xl font-bold text-gray-800">{campaigns.length}</div>
              </div>
              <div className="bg-white p-6 rounded-lg shadow border border-gray-200">
                <div className="text-sm text-gray-500 mb-1">Email Open Rate</div>
                <div className="text-3xl font-bold text-blue-600">{stats.openRate}%</div>
              </div>
              <div className="bg-white p-6 rounded-lg shadow border border-gray-200">
                <div className="text-sm text-gray-500 mb-1">Click Rate</div>
                <div className="text-3xl font-bold text-yellow-600">{stats.clickRate}%</div>
              </div>
              <div className="bg-white p-6 rounded-lg shadow border border-gray-200">
                <div className="text-sm text-gray-500 mb-1">Data Submitted</div>
                <div className="text-3xl font-bold text-red-600">{stats.submitRate}%</div>
              </div>
            </div>

            <div className="bg-white rounded-lg shadow border border-gray-200">
              <div className="p-4 border-b border-gray-200">
                <h3 className="text-lg font-semibold text-gray-800">Recent Campaigns</h3>
              </div>
              {campaigns.length === 0 ? (
                <div className="p-8 text-center text-gray-500">No campaigns yet. Create one to get started!</div>
              ) : (
                <div className="overflow-x-auto">
                  <table className="w-full">
                    <thead className="bg-gray-50 border-b border-gray-200">
                      <tr>
                        <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Campaign</th>
                        <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Status</th>
                        <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Sent</th>
                        <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Opened</th>
                        <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Clicked</th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-gray-200">
                      {campaigns.slice(0, 5).map(campaign => (
                        <tr key={campaign.id} className="hover:bg-gray-50">
                          <td className="px-4 py-3 text-sm font-medium text-gray-800">{campaign.name}</td>
                          <td className="px-4 py-3">
                            <div className="flex items-center gap-2">
                              {getStatusIcon(campaign.status)}
                              <span className="text-sm capitalize">{campaign.status}</span>
                            </div>
                          </td>
                          <td className="px-4 py-3 text-sm text-gray-600">{campaign.sent}</td>
                          <td className="px-4 py-3 text-sm text-gray-600">{campaign.opened}</td>
                          <td className="px-4 py-3 text-sm text-gray-600">{campaign.clicked}</td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              )}
            </div>
          </div>
        )}

        {activeTab === 'campaigns' && (
          <div className="space-y-6">
            <div className="flex justify-between items-center">
              <h2 className="text-2xl font-bold text-gray-800">Campaigns</h2>
              <button
                onClick={() => setShowNewCampaign(true)}
                className="flex items-center gap-2 bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700"
              >
                <Plus className="w-4 h-4" />
                New Campaign
              </button>
            </div>

            {showNewCampaign && (
              <div className="bg-white rounded-lg shadow border border-gray-200 p-6">
                <h3 className="text-lg font-semibold mb-4">Create New Campaign</h3>
                <div className="space-y-4">
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">Campaign Name</label>
                    <input
                      type="text"
                      value={newCampaign.name}
                      onChange={(e) => setNewCampaign({...newCampaign, name: e.target.value})}
                      className="w-full px-3 py-2 border border-gray-300 rounded-lg"
                      placeholder="Q4 Security Training"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">Email Template</label>
                    <select
                      value={newCampaign.template}
                      onChange={(e) => setNewCampaign({...newCampaign, template: e.target.value})}
                      className="w-full px-3 py-2 border border-gray-300 rounded-lg"
                    >
                      <option value="">Select Template</option>
                      {templates.map(t => (
                        <option key={t.id} value={t.id}>{t.name}</option>
                      ))}
                    </select>
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">Target Group</label>
                    <select
                      value={newCampaign.group}
                      onChange={(e) => setNewCampaign({...newCampaign, group: e.target.value})}
                      className="w-full px-3 py-2 border border-gray-300 rounded-lg"
                    >
                      <option value="">Select Group</option>
                      {groups.map(g => (
                        <option key={g.id} value={g.id}>{g.name} ({g.members} members)</option>
                      ))}
                    </select>
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">Landing Page URL</label>
                    <input
                      type="text"
                      value={newCampaign.url}
                      onChange={(e) => setNewCampaign({...newCampaign, url: e.target.value})}
                      className="w-full px-3 py-2 border border-gray-300 rounded-lg"
                      placeholder="https://example.com/landing"
                    />
                  </div>
                  <div className="flex gap-2">
                    <button
                      onClick={handleCreateCampaign}
                      className="bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700"
                    >
                      Create Campaign
                    </button>
                    <button
                      onClick={() => setShowNewCampaign(false)}
                      className="bg-gray-200 text-gray-700 px-4 py-2 rounded-lg hover:bg-gray-300"
                    >
                      Cancel
                    </button>
                  </div>
                </div>
              </div>
            )}

            <div className="bg-white rounded-lg shadow border border-gray-200">
              {campaigns.length === 0 ? (
                <div className="p-8 text-center text-gray-500">No campaigns yet. Create one to get started!</div>
              ) : (
                <div className="overflow-x-auto">
                  <table className="w-full">
                    <thead className="bg-gray-50 border-b border-gray-200">
                      <tr>
                        <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Name</th>
                        <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Status</th>
                        <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Created</th>
                        <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Stats</th>
                        <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Actions</th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-gray-200">
                      {campaigns.map(campaign => (
                        <tr key={campaign.id} className="hover:bg-gray-50">
                          <td className="px-4 py-3 text-sm font-medium text-gray-800">{campaign.name}</td>
                          <td className="px-4 py-3">
                            <div className="flex items-center gap-2">
                              {getStatusIcon(campaign.status)}
                              <span className="text-sm capitalize">{campaign.status}</span>
                            </div>
                          </td>
                          <td className="px-4 py-3 text-sm text-gray-600">{campaign.created}</td>
                          <td className="px-4 py-3 text-sm text-gray-600">
                            {campaign.sent} sent, {campaign.clicked} clicked
                          </td>
                          <td className="px-4 py-3">
                            <div className="flex gap-2">
                              {campaign.status === 'scheduled' && (
                                <button 
                                  onClick={() => handleLaunchCampaign(campaign.id)}
                                  className="p-1 hover:bg-gray-100 rounded" 
                                  title="Launch"
                                >
                                  <Play className="w-4 h-4 text-green-600" />
                                </button>
                              )}
                              <button className="p-1 hover:bg-gray-100 rounded" title="View">
                                <Eye className="w-4 h-4 text-gray-600" />
                              </button>
                              <button 
                                onClick={() => deleteCampaign(campaign.id)}
                                className="p-1 hover:bg-gray-100 rounded" 
                                title="Delete"
                              >
                                <Trash2 className="w-4 h-4 text-red-600" />
                              </button>
                            </div>
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              )}
            </div>
          </div>
        )}

        {activeTab === 'templates' && (
          <div className="space-y-6">
            <div className="flex justify-between items-center">
              <h2 className="text-2xl font-bold text-gray-800">Email Templates</h2>
              <button
                onClick={() => setShowNewTemplate(true)}
                className="flex items-center gap-2 bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700"
              >
                <Plus className="w-4 h-4" />
                New Template
              </button>
            </div>

            {showNewTemplate && (
              <div className="bg-white rounded-lg shadow border border-gray-200 p-6">
                <h3 className="text-lg font-semibold mb-4">Create New Template</h3>
                <div className="space-y-4">
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">Template Name</label>
                    <input
                      type="text"
                      value={newTemplate.name}
                      onChange={(e) => setNewTemplate({...newTemplate, name: e.target.value})}
                      className="w-full px-3 py-2 border border-gray-300 rounded-lg"
                      placeholder="Password Reset"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">Email Subject</label>
                    <input
                      type="text"
                      value={newTemplate.subject}
                      onChange={(e) => setNewTemplate({...newTemplate, subject: e.target.value})}
                      className="w-full px-3 py-2 border border-gray-300 rounded-lg"
                      placeholder="Action Required"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">Email Body (HTML)</label>
                    <textarea
                      value={newTemplate.body}
                      onChange={(e) => setNewTemplate({...newTemplate, body: e.target.value})}
                      rows={6}
                      className="w-full px-3 py-2 border border-gray-300 rounded-lg font-mono text-sm"
                      placeholder="<html><body><p>Email content...</p></body></html>"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">Category</label>
                    <select
                      value={newTemplate.category}
                      onChange={(e) => setNewTemplate({...newTemplate, category: e.target.value})}
                      className="w-full px-3 py-2 border border-gray-300 rounded-lg"
                    >
                      <option value="credential">Credential Harvesting</option>
                      <option value="support">IT Support</option>
                      <option value="executive">Executive Communication</option>
                      <option value="social">Social Engineering</option>
                    </select>
                  </div>
                  <div className="flex gap-2">
                    <button
                      onClick={handleCreateTemplate}
                      className="bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700"
                    >
                      Create Template
                    </button>
                    <button
                      onClick={() => setShowNewTemplate(false)}
                      className="bg-gray-200 text-gray-700 px-4 py-2 rounded-lg hover:bg-gray-300"
                    >
                      Cancel
                    </button>
                  </div>
                </div>
              </div>
            )}

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              {templates.map(template => (
                <div key={template.id} className="bg-white rounded-lg shadow border border-gray-200 p-4">
                  <div className="flex justify-between items-start mb-2">
                    <div>
                      <h3 className="font-semibold text-gray-800">{template.name}</h3>
                      <p className="text-sm text-gray-500 mt-1">{template.subject}</p>
                    </div>
                    <div className="flex gap-1">
                      <button className="p-1 hover:bg-gray-100 rounded" title="Edit">
                        <Edit className="w-4 h-4 text-gray-600" />
                      </button>
                      <button 
                        onClick={() => deleteTemplate(template.id)}
                        className="p-1 hover:bg-gray-100 rounded" 
                        title="Delete"
                      >
                        <Trash2 className="w-4 h-4 text-red-600" />
                      </button>
                    </div>
                  </div>
                  <div className="mt-3">
                    <span className="inline-block px-2 py-1 text-xs font-medium text-blue-700 bg-blue-100 rounded">
                      {template.category}
                    </span>
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}

        {activeTab === 'groups' && (
          <div className="space-y-6">
            <div className="flex justify-between items-center">
              <h2 className="text-2xl font-bold text-gray-800">User Groups</h2>
              <button
                onClick={() => setShowNewGroup(true)}
                className="flex items-center gap-2 bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700"
              >
                <Plus className="w-4 h-4" />
                New Group
              </button>
            </div>

            {showNewGroup && (
              <div className="bg-white rounded-lg shadow border border-gray-200 p-6">
                <h3 className="text-lg font-semibold mb-4">Create New Group</h3>
                <div className="space-y-4">
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">Group Name</label>
                    <input
                      type="text"
                      value={newGroup.name}
                      onChange={(e) => setNewGroup({...newGroup, name: e.target.value})}
                      className="w-full px-3 py-2 border border-gray-300 rounded-lg"
                      placeholder="Engineering Team"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">Email Addresses (one per line)</label>
                    <textarea
                      value={newGroup.emails}
                      onChange={(e) => setNewGroup({...newGroup, emails: e.target.value})}
                      rows={6}
                      className="w-full px-3 py-2 border border-gray-300 rounded-lg font-mono text-sm"
                      placeholder="user1@example.com&#10;user2@example.com"
                    />
                  </div>
                  <div className="flex gap-2">
                    <button
                      onClick={handleCreateGroup}
                      className="bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700"
                    >
                      Create Group
                    </button>
                    <button
                      onClick={() => setShowNewGroup(false)}
                      className="bg-gray-200 text-gray-700 px-4 py-2 rounded-lg hover:bg-gray-300"
                    >
                      Cancel
                    </button>
                  </div>
                </div>
              </div>
            )}

            {groups.length === 0 ? (
              <div className="bg-white rounded-lg shadow border border-gray-200 p-8 text-center text-gray-500">
                No groups yet. Create one to get started!
              </div>
            ) : (
              <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                {groups.map(group => (
                  <div key={group.id} className="bg-white rounded-lg shadow border border-gray-200 p-4">
                    <div className="flex justify-between items-start mb-2">
                      <div className="flex items-center gap-2">
                        <Users className="w-5 h-5 text-blue-600" />
                        <h3 className="font-semibold text-gray-800">{group.name}</h3>
                      </div>
                      <div className="flex gap-1">
                        <button className="p-1 hover:bg-gray-100 rounded" title="Edit">
                          <Edit className="w-4 h-4 text-gray-600" />
                        </button>
                        <button 
                          onClick={() => deleteGroup(group.id)}
                          className="p-1 hover:bg-gray-100 rounded" 
                          title="Delete"
                        >
                          <Trash2 className="w-4 h-4 text-red-600" />
                        </button>
                      </div>
                    </div>
                    <p className="text-sm text-gray-600">{group.members} members</p>
                  </div>
                ))}
              </div>
            )}
          </div>
        )}

        {activeTab === 'settings' && (
          <div className="space-y-6">
            <h2 className="text-2xl font-bold text-gray-800">Settings</h2>
            
            <div className="bg-white rounded-lg shadow border border-gray-200 p-6">
              <h3 className="text-lg font-semibold mb-4">SMTP Configuration</h3>
              <div className="space-y-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">SMTP Host</label>
                  <input
                    type="text"
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg"
                    placeholder="smtp.example.com"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">SMTP Port</label>
                  <input
                    type="text"
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg"
                    placeholder="587"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Username</label>
                  <input
                    type="text"
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg"
                    placeholder="notifications@example.com"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Password</label>
                  <input
                    type="password"
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg"
                  />
                </div>
                <button className="bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700">
                  Save Settings
                </button>
              </div>
            </div>

            <div className="bg-white rounded-lg shadow border border-gray-200 p-6">
              <h3 className="text-lg font-semibold mb-4">Landing Page Configuration</h3>
              <div className="space-y-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Landing Page URL</label>
                  <input
                    type="text"
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg"
                    placeholder="https://yourserver.com"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Redirect URL (after submission)</label>
                  <input
                    type="text"
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg"
                    placeholder="https://example.com"
                  />
                </div>
                <button className="bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700">
                  Save Configuration
                </button>
              </div>
            </div>

            <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-4">
              <div className="flex gap-2">
                <AlertCircle className="w-5 h-5 text-yellow-600 flex-shrink-0 mt-0.5" />
                <div>
                  <h4 className="font-semibold text-yellow-800 mb-1">Educational Use Only</h4>
                  <p className="text-sm text-yellow-700">
                    This platform is designed for authorized security awareness training and testing only. 
                    Ensure you have proper authorization before conducting any phishing simulation campaigns.
                  </p>
                </div>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

export default ClariPhish;1