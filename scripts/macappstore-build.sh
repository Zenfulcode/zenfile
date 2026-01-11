#!/bin/bash
set -e  # Exit on any error

APP_CERTIFICATE="3rd Party Mac Developer Application: Gustav Haavik (S6EF64ZEMD)"
PKG_CERTIFICATE="3rd Party Mac Developer Installer: Gustav Haavik (S6EF64ZEMD)"
APP_NAME="Converzen"

# wails build -platform darwin/universal -tags appstore -clean

# Remove any stray provisioning profiles and files that shouldn't be in the bundle
rm -f "./build/bin/$APP_NAME.app/Contents/Mac_Appstore_Converzen.provisionprofile"
rm -f "./build/bin/$APP_NAME.app/Contents/embedded.provisionprofile"
rm -f "./build/bin/$APP_NAME.app/Contents/entitlements.plist"
cp ./Mac_Appstore_Converzen.provisionprofile "./build/bin/$APP_NAME.app/Contents/embedded.provisionprofile"

# Sign the main executable with entitlements first
echo "Signing executable..."
codesign --force --timestamp --options=runtime -s "$APP_CERTIFICATE" -v \
    --entitlements ./build/darwin/entitlements.plist \
    "./build/bin/$APP_NAME.app/Contents/MacOS/converzen" 2>&1

# Sign the entire app bundle with entitlements
echo "Signing app bundle..."
codesign --force --timestamp --options=runtime -s "$APP_CERTIFICATE" -v \
    --entitlements ./build/darwin/entitlements.plist \
    "./build/bin/$APP_NAME.app" 2>&1

# Verify the signature and entitlements
echo "Verifying signature..."
codesign --verify --deep --strict --verbose=2 "./build/bin/$APP_NAME.app"

echo "Checking entitlements on executable..."
codesign -d --entitlements - "./build/bin/$APP_NAME.app/Contents/MacOS/converzen"

productbuild --sign "$PKG_CERTIFICATE" --component "./build/bin/$APP_NAME.app" /Applications "./$APP_NAME.pkg"

echo "Done!"