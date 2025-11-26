# Auto-Sync Setup for PhishSimAI

This repository is now configured with automatic GitHub synchronization!

## How It Works

The auto-sync feature monitors your project directory for any file changes and automatically:
1. Detects when you create, modify, or delete files
2. Waits 10 seconds after the last change (to avoid committing on every keystroke)
3. Stages all changes with `git add -A`
4. Creates a commit with timestamp
5. Pushes to your GitHub repository

## How to Use

### Starting Auto-Sync

**Option 1: Double-click the batch file**
- Simply double-click `START_AUTO_SYNC.bat` in your project folder
- A window will open showing the auto-sync status
- Keep this window open while you work

**Option 2: Run from PowerShell**
```powershell
cd C:\Users\User\phishSimAI
powershell -ExecutionPolicy Bypass -File auto-sync.ps1
```

### Stopping Auto-Sync

- Press `Ctrl+C` in the auto-sync window
- Or simply close the window

## What Gets Synced

- All code changes (.go, .js, .html, .css files)
- New files and folders
- Deleted files
- Everything except `.git` directory and the auto-sync script itself

## Tips

- **Keep the window open**: Auto-sync only works while the script is running
- **Be patient**: Changes are committed 10 seconds after you stop editing (debounce time)
- **Check the window**: You'll see real-time updates of what's being synced
- **Manual sync still works**: You can still use `git add`, `git commit`, and `git push` manually if needed

## Debounce Time

The script waits **10 seconds** after your last file change before committing. This prevents:
- Committing on every single keystroke
- Creating too many small commits
- Overwhelming GitHub with pushes

If you want to change this, edit the `$debounceSeconds` variable in `auto-sync.ps1`

## Troubleshooting

**PowerShell execution policy error:**
- The batch file handles this automatically
- If issues persist, run: `Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser`

**Push failures:**
- Check your internet connection
- Make sure you're authenticated with GitHub
- Check the auto-sync window for error messages

**Too many commits:**
- Increase the `$debounceSeconds` value in `auto-sync.ps1`

## Git Configuration

Your Git is configured with:
- **Name**: jw3008
- **Email**: jw3008@users.noreply.github.com
- **Remote**: https://github.com/jw3008/phishSimAI.git

---

**Happy coding! Your changes will automatically sync to GitHub!** 🚀
