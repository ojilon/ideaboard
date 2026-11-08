; IdeaBoard — Inno Setup installer script
; Build (from the repository root, after building ideaboard.exe):
;   iscc installer\install.iss
; Output: installer\output\IdeaBoard-<version>-setup.exe
;
; The installer:
;   - installs per user, with no admin prompt, into
;     %LOCALAPPDATA%\Programs\IdeaBoard (the user can pick another folder)
;   - copies ideaboard.exe (single file, no external dependencies)
;   - creates a Start Menu shortcut and an uninstaller entry
;   - optionally creates a desktop icon (off by default)
; Paths in this script are relative to this file (installer\).

#define MyAppName "IdeaBoard"
#define MyAppVersion "0.1.0"
#define MyAppPublisher "ojilon"
#define MyAppExeName "ideaboard.exe"

[Setup]
AppId={{7C8F1A2D-3E4B-4C9D-8F2A-1B2C3D4E5F6A}}
AppName={#MyAppName}
AppVersion={#MyAppVersion}
AppPublisher={#MyAppPublisher}
DefaultDirName={localappdata}\Programs\{#MyAppName}
DefaultGroupName={#MyAppName}
AllowNoIcons=yes
OutputDir=output
OutputBaseFilename=IdeaBoard-{#MyAppVersion}-setup
Compression=lzma
SolidCompression=yes
PrivilegesRequired=lowest
UninstallDisplayIcon={app}\{#MyAppExeName}

[Languages]
Name: "english"; MessagesFile: "compiler:Default.isl"

[Tasks]
Name: "desktopicon"; Description: "Create a &desktop icon"; GroupDescription: "Additional icons:"; Flags: unchecked

[Files]
; Build the app first from the repository root:
;   go build -trimpath -o ideaboard.exe .
Source: "..\ideaboard.exe"; DestDir: "{app}"; Flags: ignoreversion

[Icons]
Name: "{group}\{#MyAppName}"; Filename: "{app}\{#MyAppExeName}"
Name: "{group}\{cm:UninstallProgram,#MyAppName}"; Filename: "{uninstallexe}"
Name: "{autodesktop}\{#MyAppName}"; Filename: "{app}\{#MyAppExeName}"; Tasks: desktopicon

[Run]
Filename: "{app}\{#MyAppExeName}"; Description: "Launch {#MyAppName}"; Flags: nowait postinstall skipifsilent
