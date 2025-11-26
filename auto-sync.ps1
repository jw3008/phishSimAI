# PhishSimAI Auto-Sync Script
# Automatically commits and pushes changes to GitHub

$repoPath = "C:\Users\User\phishSimAI"
$debounceSeconds = 10  # Wait 10 seconds after last change before committing

Write-Host "======================================" -ForegroundColor Cyan
Write-Host "PhishSimAI Auto-Sync Started" -ForegroundColor Green
Write-Host "======================================" -ForegroundColor Cyan
Write-Host "Watching: $repoPath" -ForegroundColor Yellow
Write-Host "Changes will be auto-committed and pushed to GitHub" -ForegroundColor Yellow
Write-Host "Press Ctrl+C to stop" -ForegroundColor Yellow
Write-Host "======================================`n" -ForegroundColor Cyan

# Change to repo directory
Set-Location $repoPath

# Create FileSystemWatcher
$watcher = New-Object System.IO.FileSystemWatcher
$watcher.Path = $repoPath
$watcher.IncludeSubdirectories = $true
$watcher.EnableRaisingEvents = $true

# Filters - watch all files
$watcher.NotifyFilter = [System.IO.NotifyFilters]'FileName,DirectoryName,LastWrite'

# Debounce timer
$timer = $null
$lastChange = Get-Date

# Function to commit and push changes
function Sync-Changes {
    try {
        Write-Host "`n[$(Get-Date -Format 'HH:mm:ss')] Syncing changes..." -ForegroundColor Cyan

        # Check if there are any changes
        $status = git status --porcelain
        if ([string]::IsNullOrEmpty($status)) {
            Write-Host "[$(Get-Date -Format 'HH:mm:ss')] No changes to sync" -ForegroundColor Gray
            return
        }

        # Stage all changes
        git add -A

        # Create commit message with timestamp
        $timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
        $commitMessage = @"
Auto-sync: Changes detected at $timestamp

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>
"@

        # Commit
        git commit -m $commitMessage 2>&1 | Out-Null

        # Push to GitHub
        Write-Host "[$(Get-Date -Format 'HH:mm:ss')] Pushing to GitHub..." -ForegroundColor Yellow
        $pushResult = git push 2>&1

        if ($LASTEXITCODE -eq 0) {
            Write-Host "[$(Get-Date -Format 'HH:mm:ss')] ✓ Successfully synced to GitHub!" -ForegroundColor Green
        } else {
            Write-Host "[$(Get-Date -Format 'HH:mm:ss')] ✗ Push failed: $pushResult" -ForegroundColor Red
        }
    }
    catch {
        Write-Host "[$(Get-Date -Format 'HH:mm:ss')] Error: $_" -ForegroundColor Red
    }
}

# Debounced change handler
$onChange = {
    $path = $Event.SourceEventArgs.FullPath
    $changeType = $Event.SourceEventArgs.ChangeType

    # Ignore .git directory changes
    if ($path -like "*\.git\*") {
        return
    }

    # Ignore PowerShell script itself
    if ($path -like "*\auto-sync.ps1") {
        return
    }

    Write-Host "[$(Get-Date -Format 'HH:mm:ss')] Detected: $changeType - $($Event.SourceEventArgs.Name)" -ForegroundColor Gray

    # Update last change time
    $global:lastChange = Get-Date

    # Reset/create timer
    if ($null -ne $global:timer) {
        $global:timer.Stop()
        $global:timer.Dispose()
    }

    $global:timer = New-Object System.Timers.Timer
    $global:timer.Interval = $debounceSeconds * 1000
    $global:timer.AutoReset = $false

    Register-ObjectEvent -InputObject $global:timer -EventName Elapsed -Action {
        Sync-Changes
    } | Out-Null

    $global:timer.Start()
}

# Register event handlers
$handlers = @()
$handlers += Register-ObjectEvent $watcher "Created" -Action $onChange
$handlers += Register-ObjectEvent $watcher "Changed" -Action $onChange
$handlers += Register-ObjectEvent $watcher "Deleted" -Action $onChange
$handlers += Register-ObjectEvent $watcher "Renamed" -Action $onChange

try {
    # Keep script running
    while ($true) {
        Start-Sleep -Seconds 1
    }
}
finally {
    # Cleanup
    Write-Host "`n`nStopping auto-sync..." -ForegroundColor Yellow
    $watcher.EnableRaisingEvents = $false
    $watcher.Dispose()

    if ($null -ne $timer) {
        $timer.Stop()
        $timer.Dispose()
    }

    $handlers | ForEach-Object { Unregister-Event -SourceIdentifier $_.Name }
    Write-Host "Auto-sync stopped." -ForegroundColor Red
}
