OutFile "photounloader-1.0.1.exe"

InstallDir $PROGRAMFILES64\photounloader

LicenseData LICENSE

Page license 
Page directory
Page instfiles
UninstPage uninstConfirm
UninstPage instfiles

Section 
    SetOutPath $INSTDIR
    File "photounloader.exe"
    EnVar::AddValue PATH $INSTDIR
    WriteUninstaller $INSTDIR\photounloader-uninstaller.exe 
SectionEnd

Section "Uninstall"
    Delete $INSTDIR\photounloader.exe
    Delete $INSTDIR\photounloader-uninstaller.exe
    RMDir /r /REBOOTOK $INSTDIR
SectionEnd
