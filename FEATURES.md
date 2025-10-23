# ClaripHish - Security Awareness Features

## New Features Added

### 1. Role-Based Access Control (RBAC)

The application now supports two user roles:

- **Admin Role**: Full access to all features including phishing simulation and assessment management
- **User Role**: Limited access to awareness training platform only

### 2. Security Awareness Training Platform

#### For Administrators:
- **Create Assessments**: Build custom MCQ-based security awareness assessments
- **Manage Questions**: Add, edit, and delete questions with multiple choice options
- **Set Deadlines**: Configure assessment deadlines to ensure timely completion
- **Publish Assessments**: Control when assessments become available to users
- **View Statistics**: Monitor completion rates, average scores, and pass rates
- **Track User Progress**: See which users have completed, are in progress, or haven't started

#### For Users:
- **Take Assessments**: Complete published security awareness assessments
- **Progress Tracking**: Resume incomplete assessments
- **View Results**: See detailed results with correct/incorrect answers
- **Download Reports**: Generate PDF reports of assessment performance
- **Performance Dashboard**: Track all completed assessments and scores

### 3. Assessment Features

#### Question Management:
- Multiple choice questions (4 options per question)
- Customizable point values per question
- Question ordering
- Correct answer marking

#### Assessment Configuration:
- Title and description
- Optional deadline setting
- Draft/Published status
- Automatic grading

#### Reporting:
- Individual user performance
- Assessment-wide statistics
- Pass/Fail thresholds (70% default)
- Detailed question review
- PDF export capability (placeholder for future implementation)

### 4. Gemini AI Integration

Administrators can use Google's Gemini AI to generate:
- Phishing email templates
- Security awareness questions
- Educational content

**Usage**: Add `?api_key=YOUR_GEMINI_API_KEY` to the generate template endpoint

### 5. Database Schema

New tables added:
- `assessments` - Stores assessment metadata
- `questions` - MCQ questions linked to assessments
- `answer_options` - Answer choices for each question
- `user_assessment_attempts` - Tracks user assessment attempts
- `user_responses` - Records user answers

### 6. Security Enhancements

- Session-based authentication with role verification
- Role-specific middleware (RequireAdmin, RequireAuth)
- User data isolation
- Migration support for existing databases

## API Endpoints

### Admin Endpoints (Require Admin Role):

**Campaigns, Templates, Pages, Groups, SMTP** - All existing phishing simulation endpoints

**Assessments:**
- `GET /api/assessments` - List all assessments
- `POST /api/assessments` - Create new assessment
- `GET /api/assessments/:id` - Get assessment details
- `PUT /api/assessments/:id` - Update assessment
- `DELETE /api/assessments/:id` - Delete assessment
- `POST /api/assessments/:id/publish` - Publish assessment
- `GET /api/assessments/:id/stats` - Get assessment statistics
- `GET /api/assessments/:id/results` - Get all user results

**Gemini Integration:**
- `POST /api/templates/generate?api_key=KEY` - Generate template with Gemini AI

### User Endpoints (All Authenticated Users):

- `GET /api/user/assessments` - List available assessments
- `GET /api/user/assessments/:id` - Get assessment details
- `POST /api/user/assessments/:id/start` - Start an assessment
- `POST /api/user/assessments/attempt/:attemptId/submit` - Submit answer
- `POST /api/user/assessments/attempt/:attemptId/complete` - Complete assessment
- `GET /api/user/results` - Get user's completed assessments
- `GET /api/user/results/:attemptId` - Get detailed result
- `GET /api/user/results/:attemptId/pdf` - Download PDF report

## Default Credentials

- **Username**: admin
- **Password**: admin
- **Role**: admin

**Important**: Change the default password after first login!

## Creating Additional Users

To create additional users, you'll need to add them directly to the database or create a user management interface. Here's an example SQL:

```sql
INSERT INTO users (username, password_hash, role, created_at)
VALUES ('testuser', '<bcrypt_hash>', 'user', datetime('now'));
```

## Usage Guide

### For Administrators:

1. **Login** with admin credentials
2. **Navigate** to "Assessments" in the menu
3. **Create** a new assessment:
   - Add title and description
   - Set optional deadline
   - Add questions with 4 answer options each
   - Mark the correct answer
4. **Publish** the assessment to make it available to users
5. **Monitor** user progress in the Statistics view

### For Users:

1. **Login** with user credentials
2. **View** available assessments in "Awareness Training"
3. **Start** an assessment to begin
4. **Answer** questions one by one
5. **Submit** to see your score
6. **Review** results in "My Results"
7. **Download** PDF report if needed

## Technical Implementation

### Frontend:
- Vanilla JavaScript
- Role-based UI rendering
- Modal-based forms
- Progress tracking
- Responsive design

### Backend:
- Go with Gorilla Mux router
- SQLite database
- Session-based auth
- RESTful API
- Transaction support for data integrity

### Security:
- Bcrypt password hashing
- Role-based middleware
- SQL injection prevention (parameterized queries)
- Session management
- CSRF protection considerations

## Future Enhancements

- [ ] PDF generation implementation (currently placeholder)
- [ ] Email notifications for assessment deadlines
- [ ] User management interface for admins
- [ ] Assessment templates
- [ ] Question bank
- [ ] Advanced reporting and analytics
- [ ] Certificate generation
- [ ] Bulk user import
- [ ] Integration with LDAP/Active Directory
- [ ] Two-factor authentication

## Where to Add Gemini API Key

The user mentioned wanting to add the Gemini API key for template generation. The button is now available in the Templates section:

1. Go to **Email Templates** (Admin only)
2. Click **New Template**
3. Look for the **"Generate with AI"** button next to the template fields
4. The system will prompt for the API key or you can configure it in the environment

**Note**: For production use, store the Gemini API key as an environment variable rather than passing it in the URL.

## Architecture Decisions

1. **Role stored in session**: Fast access without additional database queries
2. **SQLite**: Simple deployment, suitable for small to medium deployments
3. **Transaction-based creation**: Ensures data consistency when creating assessments
4. **No password in user responses**: Security by not exposing correct answers until completion
5. **Soft separation**: Users can't access admin endpoints, but admins can still use user endpoints

## Database Migrations

The application includes automatic migrations that:
- Add `role` column to existing `users` table
- Set existing admin user's role to 'admin'
- Create new assessment-related tables
- Are safe to run multiple times (idempotent)
