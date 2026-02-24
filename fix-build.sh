#!/bin/bash

# PicoClaw Build Fix Script for macOS/Linux
# This script fixes the workspace embed issue

set -e

echo "🦞 PicoClaw Build Fix Script"
echo "============================"
echo

# Check if we're in the right directory
if [ ! -f "go.mod" ]; then
    echo "❌ Error: Please run this script from the PicoClaw project root"
    echo "   (should contain go.mod file)"
    exit 1
fi

# Check if workspace directory exists
if [ ! -d "workspace" ]; then
    echo "⚠️  Warning: workspace directory not found"
    echo "   Creating basic workspace structure..."
    mkdir -p workspace/memory workspace/skills
    
    # Create basic workspace files
    cat > workspace/USER.md << 'EOF'
# User Information
This file contains your preferences and context information.

## Preferences
- Response style: 
- Topics of interest:
- Communication preferences:

## Context
Add any context information you want the AI to remember about you.
EOF

    cat > workspace/AGENT.md << 'EOF'
# Agent Information
This file defines the AI agent's characteristics.

## Agent Name
PicoClaw

## Personality
- Helpful and efficient
- Friendly and professional
- Focused on accuracy

## Capabilities
- Natural language understanding
- Code generation
- Problem solving
- Creative assistance
EOF

    cat > workspace/IDENTITY.md << 'EOF'
# Identity Configuration
This file defines the system identity and context.

## System
PicoClaw AI Assistant

## Version
Current version information

## Purpose
To provide intelligent assistance for various tasks.
EOF

    cat > workspace/SOUL.md << 'EOF'
# Soul/Character
This file defines the deeper character and interaction style.

## Core Values
- Helpfulness
- Accuracy
- Creativity
- Efficiency

## Interaction Style
- Conversational and natural
- Adaptable to user preferences
- Respectful and professional
EOF

    echo "✅ Basic workspace created"
fi

# Check Go installation
if ! command -v go &> /dev/null; then
    echo "❌ Error: Go is not installed or not in PATH"
    echo "   Please install Go from: https://golang.org/"
    exit 1
fi

echo "✅ Go installation found: $(go version)"

# Check workspace files
echo "📁 Checking workspace files..."
if [ -d "workspace" ]; then
    echo "   - workspace directory: ✅"
    echo "   - USER.md: $([ -f "workspace/USER.md" ] && echo '✅' || echo '❌')"
    echo "   - AGENT.md: $([ -f "workspace/AGENT.md" ] && echo '✅' || echo '❌')"
    echo "   - IDENTITY.md: $([ -f "workspace/IDENTITY.md" ] && echo '✅' || echo '❌')"
    echo "   - SOUL.md: $([ -f "workspace/SOUL.md" ] && echo '✅' || echo '❌')"
    echo "   - memory/ directory: $([ -d "workspace/memory" ] && echo '✅' || echo '❌')"
    echo "   - skills/ directory: $([ -d "workspace/skills" ] && echo '✅' || echo '❌')"
fi

# Clean previous build
echo "🧹 Cleaning previous build..."
if [ -d "build" ]; then
    rm -rf build
    echo "   - Build directory cleaned: ✅"
fi

# Copy workspace to cmd/picoclaw
echo "📋 Copying workspace for embed..."
if [ -d "workspace" ]; then
    rm -rf cmd/picoclaw/workspace
    cp -r workspace cmd/picoclaw/
    echo "   - Workspace copied: ✅"
else
    echo "   - Workspace not found: ❌"
    exit 1
fi

# Download dependencies
echo "📦 Downloading Go dependencies..."
go mod download
go mod verify
echo "   - Dependencies downloaded: ✅"

# Build the project
echo "🔨 Building PicoClaw..."
make build

# Check if build succeeded
if [ $? -eq 0 ]; then
    echo
    echo "🎉 Build completed successfully!"
    echo
    echo "Build output:"
    if [ -f "build/picoclaw" ]; then
        echo "   - Binary: $(ls -lh build/picoclaw | awk '{print $5}')"
        echo "   - Location: $(pwd)/build/picoclaw"
    fi
    if [ -f "build/picoclaw-"* ]; then
        for binary in build/picoclaw-*; do
            echo "   - Binary: $(ls -lh "$binary" | awk '{print $5}')"
            echo "   - Location: $(pwd)/$binary"
        done
    fi
    echo
    echo "🧪 Testing binary..."
    ./build/picoclaw --version
    echo
    echo "🚀 Ready to use!"
    echo
    echo "To install to system:"
    echo "   make install"
    echo
    echo "To run Web interface:"
    echo "   ./build/picoclaw gateway"
else
    echo "❌ Build failed!"
    echo
    echo "Troubleshooting:"
    echo "1. Check Go version: go version"
    echo "2. Check dependencies: go mod tidy"
    echo "3. Try manual build: make build VERBOSE=1"
    exit 1
fi

echo
echo "Build fix complete! 🦞"