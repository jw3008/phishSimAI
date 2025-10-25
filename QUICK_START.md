# Quick Start Guide - clariphish

## 🚀 Getting Started in 3 Steps

### Step 1: Install Go (if not already installed)

**Check if Go is installed:**
```bash
go version
```

**If not installed:**
- **Windows**: https://go.dev/dl/ → Download and install
- **Mac**: `brew install go`
- **Linux**: `sudo apt install golang-go`

---

### Step 2: Run the Startup Script

**On Windows:**
```cmd
START.bat
```

**On Mac/Linux:**
```bash
chmod +x START.sh
./START.sh
```

**Or manually:**
```bash
go mod download
go run main.go
```

---

### Step 3: Open Browser

Navigate to: **http://localhost:3333**

**Login:**
- Username: `admin`
- Password: `admin`

---

## 📁 What's Included

This is a complete phishing simulation platform with:
- ✅ Role-based access control (Admin & User roles)
- ✅ Phishing campaign management
- ✅ Security awareness training with MCQ assessments
- ✅ User progress tracking and reporting
- ✅ Gemini AI integration for template generation

---

## 🎯 Quick Test

1. **Login as admin** (admin/admin)
2. **Navigate to "Assessments"**
3. **Click "Create Assessment"**
4. **Add a few questions**
5. **Publish it**
6. **View statistics**

---

## 🔧 Troubleshooting

### "Port 3333 already in use"
```bash
PORT=8080 go run main.go
# Then access: http://localhost:8080
```

### "Permission denied" (Mac/Linux)
```bash
chmod +x START.sh
chmod +x clariphish
```

### "Database locked"
```bash
rm -f clariphish.db
go run main.go
```

### Windows: "CGO error"
- Download TDM-GCC: https://jmeubank.github.io/tdm-gcc/
- Or use: `go build -tags="sqlite_omit_load_extension"`

---

## 📖 Full Documentation

- **FEATURES.md** - Complete feature list and API documentation
- **IMPLEMENTATION_SUMMARY.md** - Detailed implementation guide

---

## 🆘 Need Help?

If you see an error:
1. Copy the exact error message
2. Check the troubleshooting section above
3. Make sure Go is properly installed (`go version`)
4. Try deleting `clariphish.db` and running again

---

## 🎉 You're All Set!

The application should now be running at **http://localhost:3333**

**Default admin credentials:**
- Username: admin
- Password: admin

**⚠️ Change the password after first login!**
