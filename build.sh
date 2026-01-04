#!/bin/bash

set -e  # Exit on error

echo "🔨 Building Cursor Bar..."

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Build the Go binary
echo -e "${BLUE}Step 1: Building Go binary...${NC}"
go build -o cursor-bar

echo -e "${GREEN}✓ Go binary built successfully${NC}"

# Create macOS app bundle structure
echo -e "${BLUE}Step 2: Creating macOS app bundle...${NC}"

APP_NAME="CursorBar.app"
APP_DIR="$APP_NAME/Contents"
MACOS_DIR="$APP_DIR/MacOS"
RESOURCES_DIR="$APP_DIR/Resources"

# Clean up old app if it exists
rm -rf "$APP_NAME"

# Create directory structure
mkdir -p "$MACOS_DIR"
mkdir -p "$RESOURCES_DIR"

# Copy binary
cp cursor-bar "$MACOS_DIR/CursorBar"
chmod +x "$MACOS_DIR/CursorBar"

# Create app icon (.icns) from PNG
ICON_SOURCE="cursor-brand-assets/App Icons/PNG/APP_ICON_25D_DARK.png"
if [ -f "$ICON_SOURCE" ]; then
    echo -e "${BLUE}Step 2a: Creating app icon...${NC}"
    
    # Create temporary iconset directory
    ICONSET_DIR="cursor-icon.iconset"
    mkdir -p "$ICONSET_DIR"
    
    # Generate various icon sizes using sips (macOS built-in)
    sips -z 16 16     "$ICON_SOURCE" --out "$ICONSET_DIR/icon_16x16.png" > /dev/null 2>&1
    sips -z 32 32     "$ICON_SOURCE" --out "$ICONSET_DIR/icon_16x16@2x.png" > /dev/null 2>&1
    sips -z 32 32     "$ICON_SOURCE" --out "$ICONSET_DIR/icon_32x32.png" > /dev/null 2>&1
    sips -z 64 64     "$ICON_SOURCE" --out "$ICONSET_DIR/icon_32x32@2x.png" > /dev/null 2>&1
    sips -z 128 128   "$ICON_SOURCE" --out "$ICONSET_DIR/icon_128x128.png" > /dev/null 2>&1
    sips -z 256 256   "$ICON_SOURCE" --out "$ICONSET_DIR/icon_128x128@2x.png" > /dev/null 2>&1
    sips -z 256 256   "$ICON_SOURCE" --out "$ICONSET_DIR/icon_256x256.png" > /dev/null 2>&1
    sips -z 512 512   "$ICON_SOURCE" --out "$ICONSET_DIR/icon_256x256@2x.png" > /dev/null 2>&1
    sips -z 512 512   "$ICON_SOURCE" --out "$ICONSET_DIR/icon_512x512.png" > /dev/null 2>&1
    sips -z 1024 1024 "$ICON_SOURCE" --out "$ICONSET_DIR/icon_512x512@2x.png" > /dev/null 2>&1
    
    # Convert to .icns
    iconutil -c icns "$ICONSET_DIR" -o "$RESOURCES_DIR/AppIcon.icns"
    
    # Clean up
    rm -rf "$ICONSET_DIR"
    
    echo -e "${GREEN}✓ App icon created${NC}"
else
    echo -e "${BLUE}Note: Cursor brand assets not found, skipping icon creation${NC}"
fi

# Create Info.plist
cat > "$APP_DIR/Info.plist" << 'EOF'
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>CFBundleExecutable</key>
    <string>CursorBar</string>
    <key>CFBundleIconFile</key>
    <string>AppIcon</string>
    <key>CFBundleIdentifier</key>
    <string>com.cursor.bar</string>
    <key>CFBundleName</key>
    <string>Cursor Bar</string>
    <key>CFBundlePackageType</key>
    <string>APPL</string>
    <key>CFBundleShortVersionString</key>
    <string>1.0</string>
    <key>CFBundleVersion</key>
    <string>1</string>
    <key>LSUIElement</key>
    <true/>
    <key>NSHighResolutionCapable</key>
    <true/>
</dict>
</plist>
EOF

echo -e "${GREEN}✓ macOS app bundle created: $APP_NAME${NC}"

# Show final output
echo ""
echo -e "${GREEN}🎉 Build complete!${NC}"
echo ""
echo "Binary: ./cursor-bar"
echo "App:    ./$APP_NAME"
echo ""
echo "To run:"
echo "  ./cursor-bar          (CLI mode)"
echo "  open $APP_NAME        (App mode)"
