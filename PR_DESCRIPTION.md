# Add Complete RBAC System, Security Awareness Training, and Advanced Features

## Summary

This PR adds a comprehensive role-based access control (RBAC) system and security awareness training platform to clariphish, along with several advanced features powered by AI.

## Major Features Added

### 1. Role-Based Access Control (RBAC) ✅
- **Two Roles**: Admin (full access) and User (awareness training only)
- **Admin Capabilities**:
  - Full access to phishing simulation features
  - Create and manage campaigns, templates, landing pages
  - User management (create, edit, delete users)
  - Assessment creation and publishing
  - View all statistics and reports
- **User Capabilities**:
  - Access to security awareness training only
  - Take published assessments
  - View own performance and results
  - Access knowledge base chatbot
  - Generate PDF reports of results

### 2. Security Awareness Training Platform 📚
- **Assessment Management** (Admin):
  - Create MCQ-based security assessments
  - Add multiple questions with answer options
  - Set deadlines for assessments
  - Publish/unpublish assessments
  - View completion statistics and results
- **Assessment Taking** (User):
  - View available assessments
  - Start and complete assessments
  - View scores and performance feedback
  - Track progress and history

### 3. User Management Interface 👥
- Create new users with specific roles
- Assign Admin or User role based on least privilege principle
- Change user passwords
- Delete users (with safety checks)
- Cannot delete last admin or self
- UI warnings about least privilege

### 4. Landing Page Cloner 🌐 (NEW)
- Clone any website as a landing page with one click
- Paste URL to clone (e.g., Facebook login, Google login)
- Automatically converts relative URLs to absolute
- Adds tracking pixel automatically

### 5. Knowledge Base Chatbot 🤖 (NEW)
- AI-powered security assistant for users
- Uses Gemini API to answer security questions
- Helps with phishing awareness, password security, MFA, etc.
- Available to all users (not just admins)
- Conversational chat interface

### 6. AI Template Generator ✨ (NEW)
- One-click phishing email template generation
- Uses Gemini API to create realistic scenarios
- 15 pre-defined scenarios (password reset, package delivery, etc.)
- Generates subject line, HTML content, and text content
- Admin can review and edit before saving

### 7. PDF Report Generation 📄 (NEW)
- **For Users**: Download PDF reports of assessment results
- **For Admins**:
  - Download campaign overview PDFs with statistics
  - Download assessment overview PDFs with completion stats
  - Detailed analytics and performance metrics

## Technical Implementation

### Backend Changes
- **Database Schema**: Added role column to users, 5 new tables for assessments
- **API Endpoints**: 30+ new endpoints for assessments, user management, AI features
- **Middleware**: RequireAuth and RequireAdmin for authorization
- **Security**: bcrypt password hashing, session-based authentication
- **New API Files**:
  - `api/users.go` - User management CRUD
  - `api/assessment.go` - Assessment management
  - `api/user_assessment.go` - User assessment taking
  - `api/page_cloner.go` - Landing page cloning
  - `api/knowledge_base.go` - AI chatbot
  - `api/pdf_reports.go` - PDF generation
  - `api/gemini.go` - Enhanced with random template generation

### Frontend Changes
- Role-based navigation and UI
- Assessment creation and taking interfaces
- User management interface
- Knowledge base chat interface
- Landing page cloner UI
- PDF download buttons
- Updated CSS with chat interface styling

### Dependencies Added
- `github.com/jung-kurt/gofpdf` - PDF generation
- `golang.org/x/net` - HTML parsing for page cloning

## Database Migrations
- Automatic migration system for existing databases
- Adds role column to users table
- Creates assessment-related tables
- Sets default admin user role

## Documentation
- Updated README.md with all new features
- Added QUICK_START.md for easy setup
- Added DOWNLOAD_INSTRUCTIONS.txt
- Created START.bat and START.sh for one-click startup
- Renamed all "ClaripHish" to "clariphish" (lowercase)

## Testing
- Build tested successfully
- All dependencies installed
- No compilation errors

## How to Use New Features

### For Admins:
1. **Landing Page Cloner**: Go to Landing Pages → "Clone from URL" button
2. **AI Template Generator**: Go to Email Templates → "🤖 Generate with AI" button
3. **Create Assessments**: Go to Assessments → "Create Assessment"
4. **Manage Users**: Go to User Management → "Add User"
5. **Download Reports**: Click PDF buttons on campaigns/assessments

### For Users:
1. **Take Assessments**: Go to Awareness Training → Select assessment
2. **Ask Questions**: Go to Knowledge Base → Type security questions
3. **Download Results**: Go to My Results → Click "Download PDF"

### Gemini API Key:
- Get free API key from: https://aistudio.google.com/app/apikey
- App will prompt for key when using AI features
- Stored in browser localStorage for convenience

## Breaking Changes
None - Fully backward compatible with existing data

## Security Considerations
- All AI features use user-provided API keys (no server-side key storage)
- RBAC enforced at both API and UI levels
- Password strength validation
- Session-based authentication
- Cannot delete last admin (prevents lockout)
- Least privilege principle in user creation

---

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>
