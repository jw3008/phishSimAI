# ClaripHish - Phishing Simulation Platform

ClaripHish is an open-source phishing simulation platform designed for security awareness training and authorized penetration testing. It provides a complete toolkit for creating, managing, and analyzing phishing campaigns in a controlled environment.

## ⚠️ Educational Use Only

This tool is designed for **authorized security awareness training and penetration testing only**. Only use this tool on systems and networks where you have explicit permission. Unauthorized use is illegal and unethical.

## Features

### Campaign Management
- Create and manage multiple phishing campaigns
- Schedule and launch campaigns
- Real-time campaign monitoring
- Comprehensive campaign analytics

### Email Templates
- Custom HTML and text email templates
- Variable substitution (FirstName, LastName, URL)
- Template library management

### Landing Pages
- Custom HTML landing pages
- Credential capture capabilities
- Automatic form tracking
- Redirect configuration

### Target Groups
- Organize targets into groups
- Import target lists
- Individual target management

### Tracking & Analytics
- Email open tracking
- Link click tracking
- Form submission tracking
- Detailed event timeline
- Statistical analysis (open rate, click rate, submission rate)

### SMTP Management
- Multiple SMTP profile support
- Secure credential storage
- Custom sender configuration

## Technology Stack

- **Backend**: Go (Golang)
- **Database**: SQLite3
- **Frontend**: HTML, CSS, JavaScript
- **Email**: SMTP with tracking
- **Authentication**: Session-based authentication

## Installation

### Prerequisites

- Go 1.21 or higher
- Git

### Setup Instructions

1. **Clone the repository**
```bash
git clone https://github.com/yourusername/clariphish.git
cd clariphish
```

2. **Install dependencies**
```bash
go mod download
```

3. **Build the application**
```bash
go build -o clariphish
```

4. **Run the application**
```bash
./clariphish
```

The server will start on port 3333 by default. Access it at: `http://localhost:3333`

### Default Credentials

- **Username**: admin
- **Password**: admin

**Important**: Change the default password after first login!

## Usage

### 1. Login

Navigate to `http://localhost:3333` and login with your credentials.

### 2. Create an Email Template

1. Go to "Email Templates"
2. Click "New Template"
3. Enter template details with optional variables:
   - `{{.FirstName}}` - Target's first name
   - `{{.LastName}}` - Target's last name
   - `{{.URL}}` - Tracking link

Example template:
```html
<html>
<body>
    <p>Dear {{.FirstName}},</p>
    <p>Please verify your account by clicking the link below:</p>
    <a href="{{.URL}}">Verify Account</a>
</body>
</html>
```

### 3. Create a Landing Page

1. Go to "Landing Pages"
2. Click "New Page"
3. Enter HTML content for your phishing page
4. Configure credential capture settings
5. Optionally set a redirect URL

Example landing page:
```html
<html>
<body>
    <h1>Account Verification</h1>
    <form method="POST">
        <input type="text" name="username" placeholder="Username" required>
        <input type="password" name="password" placeholder="Password" required>
        <button type="submit">Login</button>
    </form>
</body>
</html>
```

### 4. Create a Target Group

1. Go to "User Groups"
2. Click "New Group"
3. Add targets with their information:
   - First Name
   - Last Name
   - Email (required)
   - Position

### 5. Configure SMTP

1. Go to "Sending Profiles"
2. Click "New Profile"
3. Enter SMTP details:
   - Host (e.g., smtp.gmail.com:587)
   - Username
   - Password
   - From Address

**Gmail Example**:
- Host: `smtp.gmail.com:587`
- Username: your-email@gmail.com
- Password: your-app-password (use App Password, not regular password)
- From: your-email@gmail.com

### 6. Create and Launch a Campaign

1. Go to "Campaigns"
2. Click "New Campaign"
3. Configure:
   - Campaign name
   - Email template
   - Landing page
   - Sending profile
   - URL (where ClaripHish is accessible)
   - Target groups
4. Check "Launch immediately" to start the campaign
5. Click "Create Campaign"

### 7. Monitor Results

- Click "View" on any campaign to see detailed results
- Track email opens, link clicks, and form submissions
- Analyze statistics and success rates

## API Endpoints

### Authentication
- `POST /api/login` - Login
- `POST /api/logout` - Logout
- `GET /api/user` - Get current user

### Campaigns
- `GET /api/campaigns` - List campaigns
- `POST /api/campaigns` - Create campaign
- `GET /api/campaigns/{id}` - Get campaign details
- `PUT /api/campaigns/{id}` - Update campaign
- `DELETE /api/campaigns/{id}` - Delete campaign

### Templates
- `GET /api/templates` - List templates
- `POST /api/templates` - Create template
- `GET /api/templates/{id}` - Get template
- `PUT /api/templates/{id}` - Update template
- `DELETE /api/templates/{id}` - Delete template

### Pages
- `GET /api/pages` - List pages
- `POST /api/pages` - Create page
- `GET /api/pages/{id}` - Get page
- `PUT /api/pages/{id}` - Update page
- `DELETE /api/pages/{id}` - Delete page

### Groups
- `GET /api/groups` - List groups
- `POST /api/groups` - Create group
- `GET /api/groups/{id}` - Get group
- `PUT /api/groups/{id}` - Update group
- `DELETE /api/groups/{id}` - Delete group

### SMTP
- `GET /api/smtp` - List SMTP configs
- `POST /api/smtp` - Create SMTP config
- `GET /api/smtp/{id}` - Get SMTP config
- `PUT /api/smtp/{id}` - Update SMTP config
- `DELETE /api/smtp/{id}` - Delete SMTP config

### Tracking (Public)
- `GET /api/track?rid={rid}` - Track email open
- `GET /api/click?rid={rid}` - Track link click & serve landing page
- `POST /api/report` - Track form submission

## Configuration

### Environment Variables

- `PORT` - Server port (default: 3333)

### Database

ClaripHish uses SQLite3 for data storage. The database file is created automatically at `./clariphish.db`

## Security Considerations

1. **Change Default Password**: Always change the default admin password
2. **Use HTTPS**: In production, use HTTPS to protect credentials
3. **Network Security**: Run on isolated networks for testing
4. **Access Control**: Restrict access to authorized personnel only
5. **Data Protection**: Secure the database file
6. **SMTP Credentials**: Use app-specific passwords when possible

## Development

### Project Structure

```
clariphish/
├── main.go              # Application entry point
├── db/
│   └── db.go           # Database initialization
├── models/
│   └── models.go       # Data models
├── api/
│   ├── api.go          # API router
│   ├── auth.go         # Authentication handlers
│   ├── campaigns.go    # Campaign handlers
│   ├── templates.go    # Template handlers
│   ├── pages.go        # Page handlers
│   ├── groups.go       # Group handlers
│   ├── smtp.go         # SMTP handlers
│   └── tracking.go     # Tracking handlers
├── mailer/
│   └── mailer.go       # Email sending logic
├── static/
│   ├── index.html      # Main HTML file
│   ├── css/
│   │   └── style.css   # Styles
│   └── js/
│       └── app.js      # Frontend application
└── go.mod              # Go dependencies
```

### Building from Source

```bash
go build -o clariphish
```

### Running Tests

```bash
go test ./...
```

## Troubleshooting

### Email Not Sending

1. Verify SMTP configuration
2. Check SMTP credentials
3. Ensure firewall allows SMTP traffic
4. For Gmail, use App Passwords instead of regular password
5. Check email server logs

### Campaign Not Launching

1. Verify all required fields are filled
2. Check that target group has targets
3. Ensure SMTP profile is configured
4. Verify URL is accessible

### Database Errors

1. Ensure write permissions in application directory
2. Check disk space
3. Delete and recreate database if corrupted

## License

This project is open-source software designed for educational and authorized security testing purposes only.

## Disclaimer

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND. The authors are not responsible for any misuse or damage caused by this program. Use only on systems you own or have explicit permission to test.

## Contributing

Contributions are welcome! Please ensure all contributions maintain the educational and defensive security focus of this project.

## Support

For issues, questions, or contributions, please open an issue on GitHub.

## Acknowledgments

Inspired by GoPhish and other security awareness training platforms.
