; IdeaBoard — Inno Setup installer script
; Build:  iscc installer/install.iss   (requires Inno Setup 6+)
;
; Produces a Windows installer that:
;   - lets the user choose the install directory
;   - copies ideaboard.exe (single-file, no external dependencies)
;   - creates a Start Menu shortcut
;   - registers an uninstaller

#define MyAppName "IdeaBoard"
#define MyAppVersion "0.1.0"
#define MyAppPublisher "ojilon"
#define MyAppExeName "ideaboard.exe"

[Setup]
AppId={{7C8F1A2D-3E4B-4C9D-8F2A-1B2C3D4E5F6A}}
AppName={#MyAppName}
AppVersion={#MyAppVersion}
AppPublisher={#MyAppPublisher}
DefaultDirName={autpf}\{#MyAppName}
DefaultGroupName={#MyAppName}
AllowNoIcons=yes
LicenseFile=
OutputDir=installer\output
OutputBaseFilename=IdeaBoard-{#MyAppVersion}-setup
SetupIconFile=
Compression=lzma
SolidCompression=yes
PrivilegesRequired=lowest   ; install per-user, no admin elevation needed
ChangesEnvironment=yes
UninstallDisplayIcon={app}\{#MyAppExeName}

[Languages]
Name: "english"; MessagesFile: "compiler:Default.isl"

[Tasks]
Name: "desktopicon"; Description: "Create a &desktop icon"; GroupDescription: "Additional icons:"; Flags: unchecked

[Files]
; The built Go binary goes here. Build the app first:
;   go build -trimpath -o ideaboard.exe .
; Then place ideaboard.exe next to this script (or adjust Source below).
Source: "ideaboard.exe"; DestDir: "{app}"; Flags: ignoreversion

[Icons]
Name: "{group}\{#MyAppName}"; Filename: "{app}\{#MyAppExeName}"
Name: "{group}\{cm:UninstallProgram,#MyAppName}"; Filename: "{uninstallexe}"
Name: "{autodesktop}\{#MyAppName}"; Filename: "{app}\{#MyAppExeName}"; Tasks: desktopicon

[Run]
Filename: "{app}\{#MyAppExeName}"; Description: "Launch {#MyAppName}"; Flags: nowait postinstall skipifsilent

[UninstallDelete]
Type: filesandordirs

[Code]
; Nothing custom required: MyGo is a self-contained pure-Go binary.
