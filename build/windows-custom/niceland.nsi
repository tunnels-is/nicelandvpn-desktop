Unicode true

####
## Please note: Template replacements don't work in this file. They are provided with default defines like
## mentioned underneath.
## If the keyword is not defined, "wails_tools.nsh" will populate them with the values from ProjectInfo. 
## If they are defined here, "wails_tools.nsh" will not touch them. This allows to use this project.nsi manually 
## from outside of Wails for debugging and development of the installer.
## 
## For development first make a wails nsis build to populate the "wails_tools.nsh":
## > wails build --target windows/amd64 --nsis
## Then you can call makensis on this file with specifying the path to your binary:
## For a AMD64 only installer:
## > makensis -DARG_WAILS_AMD64_BINARY=..\..\bin\app.exe
## For a ARM64 only installer:
## > makensis -DARG_WAILS_ARM64_BINARY=..\..\bin\app.exe
## For a installer with both architectures:
## > makensis -DARG_WAILS_AMD64_BINARY=..\..\bin\app-amd64.exe -DARG_WAILS_ARM64_BINARY=..\..\bin\app-arm64.exe
####
## The following information is taken from the ProjectInfo file, but they can be overwritten here. 
####
## !define INFO_PROJECTNAME    "MyProject" # Default "{{.Name}}"
## !define INFO_COMPANYNAME    "MyCompany" # Default "{{.Info.CompanyName}}"
## !define INFO_PRODUCTNAME    "MyProduct" # Default "{{.Info.ProductName}}"
## !define INFO_PRODUCTVERSION "1.0.0"     # Default "{{.Info.ProductVersion}}"
## !define INFO_COPYRIGHT      "Copyright" # Default "{{.Info.Copyright}}"
###
## !define PRODUCT_EXECUTABLE  "Application.exe"      # Default "${INFO_PROJECTNAME}.exe"
## !define UNINST_KEY_NAME     "UninstKeyInRegistry"  # Default "${INFO_COMPANYNAME}${INFO_PRODUCTNAME}"
####
## !define REQUEST_EXECUTION_LEVEL "admin"            # Default "admin"  see also https://nsis.sourceforge.io/Docs/Chapter4.html
####
## Include the wails tools
!include "wails_tools.nsh"
####
# The version information for this two must consist of 4 parts
!define PRODUCT_VERSION "1.1.4"

VIProductVersion "${PRODUCT_VERSION}.0"
VIFileVersion    "${PRODUCT_VERSION}.0"

VIAddVersionKey "CompanyName"     "Tunnels EHF"
VIAddVersionKey "FileDescription" "NicelandVPN"
VIAddVersionKey "ProductVersion"  "${PRODUCT_VERSION}"
VIAddVersionKey "FileVersion"     "${PRODUCT_VERSION}"
VIAddVersionKey "LegalCopyright"  "https://nicelandvpn.is"
VIAddVersionKey "ProductName"     "NicelandVPN"

!include "MUI.nsh"

!define MUI_ICON ".\icon.ico"
!define MUI_UNICON ".\icon.ico"
# !define MUI_WELCOMEFINISHPAGE_BITMAP "resources\leftimage.bmp" #Include this to add a bitmap on the left side of the Welcome Page. Must be a size of 164x314
!define MUI_FINISHPAGE_NOAUTOCLOSE # Wait on the INSTFILES page so the user can take a look into the details of the installation steps
!define MUI_ABORTWARNING # This will warn the user if they exit from the installer.

!insertmacro MUI_PAGE_WELCOME # Welcome to the installer page.
# !insertmacro MUI_PAGE_LICENSE "resources\eula.txt" # Adds a EULA page to the installer
!insertmacro MUI_PAGE_DIRECTORY # In which folder install page.
!insertmacro MUI_PAGE_INSTFILES # Installing page.
!insertmacro MUI_PAGE_FINISH # Finished installation page.

!insertmacro MUI_UNPAGE_INSTFILES # Uinstalling page

!insertmacro MUI_LANGUAGE "English" # Set the Language of the installer

## The following two statements can be used to sign the installer and the uninstaller. The path to the binaries are provided in %1
#!uninstfinalize 'signtool --file "%1"'
#!finalize 'signtool --file "%1"'

Name "NicelandVPN"
OutFile ".\NicelandVPN-windows.exe" # Name of the installer's file.
InstallDir "$PROGRAMFILES64\NicelandVPN" # Default installing folder ($PROGRAMFILES is Program Files folder).
ShowInstDetails show # This will always show the installation details.

!define UNINST_KEY "Software\Microsoft\Windows\CurrentVersion\Uninstall\TunnelsEHF-NicelandVPN"

Section
    SetShellVarContext all

    !insertmacro wails.webview2runtime

    SetOutPath $INSTDIR
    File "wintun.dll"
    File "NicelandVPN.exe"
    
    WriteRegStr HKEY_LOCAL_MACHINE "SOFTWARE\Microsoft\Windows NT\CurrentVersion\AppCompatFlags\Layers" "$INSTDIR\NicelandVPN.exe" "RUNASADMIN"

    CreateShortcut "$SMPROGRAMS\NicelandVPN.lnk" "$INSTDIR\NicelandVPN.exe"
    CreateShortCut "$DESKTOP\NicelandVPN.lnk" "$INSTDIR\NicelandVPN.exe"

    # CREATE UNINSTALLER

    WriteUninstaller "$INSTDIR\uninstall.exe"

    SetRegView 64
    WriteRegStr HKLM "${UNINST_KEY}" "Publisher" "Tunnels EHF"
    WriteRegStr HKLM "${UNINST_KEY}" "DisplayName" "NicelandVPN"
    WriteRegStr HKLM "${UNINST_KEY}" "DisplayVersion" "${PRODUCT_VERSION}"
    WriteRegStr HKLM "${UNINST_KEY}" "DisplayIcon" "$INSTDIR\NicelandVPN.exe"
    WriteRegStr HKLM "${UNINST_KEY}" "UninstallString" "$\"$INSTDIR\uninstall.exe$\""
    WriteRegStr HKLM "${UNINST_KEY}" "QuietUninstallString" "$\"$INSTDIR\uninstall.exe$\" /S"

    ${GetSize} "$INSTDIR" "/S=0K" $0 $1 $2
    IntFmt $0 "0x%08X" $0
    WriteRegDWORD HKLM "${UNINST_KEY}" "EstimatedSize" "$0"
SectionEnd

Section "uninstall" 
    SetShellVarContext all

    RMDir /r $INSTDIR

    Delete "$SMPROGRAMS\NicelandVPN.lnk"
    Delete "$DESKTOP\NicelandVPN.lnk"

    # we remove for both NSIS shell contexts just to be sure
    SetShellVarContext current
    RMDir /r "$APPDATA\NicelandVPN.exe" # Remove the WebView2 DataPath
    SetShellVarContext all
    RMDir /r "$APPDATA\NicelandVPN.exe" # Remove the WebView2 DataPath

    Delete "$INSTDIR\uninstall.exe"

    SetRegView 64
    DeleteRegKey HKLM "${UNINST_KEY}"
SectionEnd