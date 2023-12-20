# DO NOT EDIT - Generated automatically by `wails build`

!include "x64.nsh"
!include "WinVer.nsh"
!include "FileFunc.nsh"


# Install webview2 by launching the bootstrapper
# See https://docs.microsoft.com/en-us/microsoft-edge/webview2/concepts/distribution#online-only-deployment
!macro wails.webview2runtime
    !ifndef WAILS_INSTALL_WEBVIEW_DETAILPRINT
        !define WAILS_INSTALL_WEBVIEW_DETAILPRINT "Installing: WebView2 Runtime"
    !endif

    SetRegView 64
	# If the admin key exists and is not empty then webview2 is already installed
	ReadRegStr $0 HKLM "SOFTWARE\WOW6432Node\Microsoft\EdgeUpdate\Clients\{F3017226-FE2A-4295-8BDF-00C3A9A7E4C5}" "pv"
    ${If} $0 != ""
        Goto ok
    ${EndIf}

    ; ${If} ${REQUEST_EXECUTION_LEVEL} == "user"
        # If the installer is run in user level, check the user specific key exists and is not empty then webview2 is already installed
	    ReadRegStr $0 HKCU "Software\Microsoft\EdgeUpdate\Clients{F3017226-FE2A-4295-8BDF-00C3A9A7E4C5}" "pv"
        ${If} $0 != ""
            Goto ok
        ${EndIf}
    ;  ${EndIf}
    
	SetDetailsPrint both
    DetailPrint "${WAILS_INSTALL_WEBVIEW_DETAILPRINT}"
    SetDetailsPrint listonly
    
    InitPluginsDir
    CreateDirectory "$pluginsdir\webview2bootstrapper"
    SetOutPath "$pluginsdir\webview2bootstrapper"
    File "tmp\MicrosoftEdgeWebview2Setup.exe"
    ExecWait '"$pluginsdir\webview2bootstrapper\MicrosoftEdgeWebview2Setup.exe" /silent /install'
    
    SetDetailsPrint both
    ok:
!macroend