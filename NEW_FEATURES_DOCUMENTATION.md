# New Features Documentation

This document describes the new features added to PhishSimAI.

## 📧 Email Phishing Analyzer

### Overview
A powerful AI-powered tool that allows both admins and users to analyze suspicious emails to determine if they are legitimate or phishing attempts.

### Access
- **Available to:** All authenticated users (both admins and normal users)
- **Location:** Click on "Email Analyzer" in the navigation menu

### How to Use

1. **Navigate to Email Analyzer**
   - After logging in, click on "Email Analyzer" in the top navigation

2. **Paste Email Content**
   - Copy the entire suspicious email (including subject, sender, body, and any links)
   - Paste it into the text area

3. **Analyze**
   - Click the "🔍 Analyze Email" button
   - The AI will process the email (this takes a few seconds)

4. **Review Results**
   The analysis provides:
   - **Verdict**: Whether the email is legitimate or phishing
   - **Confidence Score**: AI's confidence in the assessment (0-100%)
   - **Risk Level**: Low, Medium, High, or Critical
   - **Phishing Indicators**: Specific red flags found in the email
   - **Detailed Analysis**: Explanation of the assessment
   - **Recommendations**: What actions to take

### Features

#### AI-Powered Analysis
The tool uses Google's Gemini AI to analyze emails for:
- Urgency or threatening language
- Requests for personal/financial information
- Suspicious sender information
- Grammar and spelling errors
- Unusual links or attachments
- Impersonation attempts
- Too-good-to-be-true offers
- Mismatched URLs or domains

#### Visual Risk Indicators
- **Green (✓ LEGITIMATE)**: Email appears safe
- **Yellow (⚠️ SUSPICIOUS)**: Some concerning elements
- **Orange (⚠️ LIKELY PHISHING)**: High probability of phishing
- **Red (⚠️ PHISHING - HIGH RISK)**: Definite phishing attempt

### Best Practices

1. **Include Complete Email**: The more information you provide (subject, sender, full body, links), the more accurate the analysis
2. **Use for Training**: Great educational tool for learning to spot phishing emails
3. **Verify with IT**: Always report suspicious emails to your IT security team
4. **Regular Use**: Analyze any email that seems unusual or suspicious

### Example Use Cases

- Employee receives urgent email requesting password reset
- Suspicious invoice or payment request
- Unexpected package delivery notification
- Unusual request from "management"
- Prize or lottery winning notifications
- Banking security alerts

---

## 📄 Assessment Report PDF Download

### Overview
Users can now download detailed PDF reports of their completed assessment results.

### Access
- **Available to:** All authenticated users
- **Location:** "My Results" page

### How to Use

1. **Navigate to My Results**
   - Click on "My Results" in the navigation menu

2. **Find Completed Assessment**
   - Locate the assessment you want to download

3. **Download PDF**
   - Click the "Download PDF" button next to the assessment
   - The PDF will open in a new tab/window and download automatically

### PDF Contents

The generated PDF includes:

1. **Header Information**
   - Assessment title
   - Student name
   - Completion date and time

2. **Score Summary**
   - Total score achieved
   - Total possible points
   - Percentage score
   - Performance evaluation

3. **Performance Assessment**
   - Excellent (90%+)
   - Good (70-89%)
   - Fair (50-69%)
   - Needs Improvement (<50%)

4. **Detailed Results**
   - Each question with your answer
   - Correct/incorrect indication
   - Points earned per question
   - Correct answers shown for missed questions

5. **Footer**
   - Generation timestamp
   - Platform branding

### Features

#### Professional Format
- Clean, professional PDF layout
- Easy to read and print
- Suitable for records and reporting

#### Comprehensive Results
- Complete question-by-question breakdown
- Clear indication of correct/incorrect answers
- Educational feedback for learning

#### Instant Generation
- PDFs generated on-demand
- No waiting or processing time
- Always up-to-date with latest results

### Use Cases

- **Personal Records**: Keep a record of your training progress
- **Compliance**: Provide proof of completed security training
- **Study Material**: Review questions you got wrong
- **Performance Tracking**: Track improvement over time
- **Reporting**: Submit to supervisors or HR

---

## 🔧 Technical Details

### API Endpoints

#### Email Analysis
```
POST /api/analyze-email
Body: { "email_text": "full email content here" }
Response: {
  "success": true,
  "result": {
    "is_phishing": boolean,
    "confidence_score": 0-100,
    "risk_level": "low|medium|high|critical",
    "indicators": ["array of red flags"],
    "explanation": "detailed analysis",
    "recommendations": ["array of recommendations"]
  }
}
```

#### PDF Download
```
GET /api/user/results/{attemptId}/pdf
Response: PDF file download
```

### Dependencies

- **Gemini AI API**: Required for email analysis
  - Get your API key from: https://aistudio.google.com/app/apikey
  - Configure in Settings (Admin only)

- **gofpdf Library**: Used for PDF generation (already included)

### Configuration

#### Setting up Gemini API
1. Log in as admin
2. Navigate to Settings
3. Enter your Gemini API key
4. Click "Save API Key"
5. (Optional) Click "Test Connection" to verify

---

## 🎯 User Guide

### For Normal Users

1. **Complete Assessments**: Take security awareness assessments
2. **View Results**: Check your scores and detailed results
3. **Download Reports**: Save PDF copies of your results
4. **Analyze Suspicious Emails**: Use the Email Analyzer to check suspicious emails
5. **Learn Continuously**: Review wrong answers and learn from mistakes

### For Administrators

1. **Create Assessments**: Build security awareness assessments
2. **Monitor Progress**: Track user completion and scores
3. **Configure AI**: Set up Gemini API for AI features
4. **Generate Reports**: Export campaign and assessment reports
5. **Manage Users**: Create and manage user accounts

---

## 🚀 Getting Started

### First Time Setup

1. **Start the Application**
   ```bash
   ./phishSimAI.exe
   ```
   or double-click `START.bat`

2. **Login**
   - Default admin: username: `admin`, password: `changeme`
   - Change password after first login

3. **Configure AI (Admin)**
   - Go to Settings
   - Add Gemini API key
   - Test connection

4. **Start Using**
   - Users can immediately access Email Analyzer
   - PDF downloads work automatically for completed assessments

### Testing the Features

#### Test Email Analyzer
1. Navigate to Email Analyzer
2. Paste this sample phishing email:
   ```
   Subject: URGENT: Your account will be suspended!

   Dear User,

   We have detected suspicious activity on your account. You MUST verify your
   information immediately or your account will be permanently suspended within 24 hours.

   Click here to verify: http://fake-bank-login.suspicious.com

   Enter your username, password, and social security number to confirm your identity.

   Regards,
   Bank Security Team
   ```
3. Click Analyze
4. Review the AI's assessment

#### Test PDF Download
1. Complete an assessment as a user
2. Go to "My Results"
3. Click "Download PDF" on any completed assessment
4. Verify the PDF downloads correctly

---

## 📝 Notes

- Email Analyzer requires internet connection (calls Gemini API)
- PDF generation works offline
- Email analysis is for educational purposes only
- Always verify suspicious emails with your IT security team
- PDFs are generated in real-time and are not stored on the server

---

## 🆘 Troubleshooting

### Email Analyzer Issues

**Problem**: "Gemini API key not configured" error
- **Solution**: Admin must configure API key in Settings

**Problem**: Analysis fails or returns errors
- **Solution**:
  1. Check internet connection
  2. Verify API key is valid (Settings → Test Connection)
  3. Ensure you have API quota remaining

**Problem**: Results don't display
- **Solution**: Check browser console for errors, refresh page

### PDF Download Issues

**Problem**: PDF doesn't download
- **Solution**:
  1. Check popup blocker settings
  2. Ensure assessment is completed
  3. Try different browser

**Problem**: PDF shows incorrect data
- **Solution**: Refresh the page and try again

---

## 🔒 Security Considerations

- Email content sent to Gemini API for analysis (review privacy policy)
- PDFs contain assessment results - handle according to your privacy policy
- API keys should be kept secure (admin access only)
- Email Analyzer is a tool to help identify phishing, not a guarantee

---

## 📊 Feature Summary

| Feature | Available To | Purpose | Status |
|---------|-------------|---------|--------|
| Email Analyzer | All Users | Analyze suspicious emails | ✅ Active |
| PDF Reports | All Users | Download assessment results | ✅ Active |
| AI Configuration | Admin Only | Set up Gemini API | ✅ Active |

---

**Version**: 1.1.0
**Last Updated**: November 13, 2024
**Generated by**: Claude Code
