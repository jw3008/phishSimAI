#!/bin/bash

# clariphish Startup Script
# This script will set up and run the phishing simulation platform

echo "=================================="
echo "clariphish - Setup & Start"
echo "=================================="
echo ""

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed!"
    echo ""
    echo "Please install Go first:"
    echo "  - Mac: brew install go"
    echo "  - Linux: sudo apt install golang-go"
    echo "  - Windows: Download from https://go.dev/dl/"
    echo ""
    exit 1
fi

echo "✅ Go is installed: $(go version)"
echo ""

# Check if we're in the right directory
if [ ! -f "main.go" ]; then
    echo "❌ main.go not found!"
    echo "Please run this script from the clariphish directory"
    exit 1
fi

echo "✅ Found main.go"
echo ""

# Install dependencies
echo "📦 Installing dependencies..."
go mod download
if [ $? -eq 0 ]; then
    echo "✅ Dependencies installed"
else
    echo "⚠️  Warning: Some dependencies may not have installed correctly"
fi
echo ""

# Clean database (optional - comment out if you want to keep existing data)
if [ -f "clariphish.db" ]; then
    echo "🗑️  Removing old database..."
    rm -f clariphish.db
    echo "✅ Old database removed"
    echo ""
fi

# Build the application
echo "🔨 Building application..."
go build -o clariphish .
if [ $? -eq 0 ]; then
    echo "✅ Build successful"
else
    echo "❌ Build failed"
    exit 1
fi
echo ""

# Start the application
echo "🚀 Starting clariphish..."
echo ""
echo "=================================="
echo "Server will start on: http://localhost:3333"
echo "Default credentials:"
echo "  Username: admin"
echo "  Password: admin"
echo "=================================="
echo ""
echo "Press Ctrl+C to stop the server"
echo ""

./clariphish
