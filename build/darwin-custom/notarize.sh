# SIGN THE launcher and NicelandVPN binaries
codesign --deep --force --options=runtime --sign "Developer ID Application: Tunnels EHF (K824A68KH8)" --timestamp [BINARY]
# validate signatures
codesign -dv --verbose=4 [BINARY]

# Place launcher and NicelandVPN binaries into NicelandVPN.app/Contents/MacOS/
# notarize the NicelandVPN.app
xcrun notarytool submit NicelandVPN.zip --apple-id "" --team-id "K824A68KH8"  --password "" --wait

# extract NicelandVPN.app
# validate signature
codesign -dv --verbose=4 NicelandVPN.app
