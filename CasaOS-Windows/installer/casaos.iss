; Installer of CasaOS for Windows. Built with Inno Setup by build.ps1.
;
; It installs for the current user, with no administrator rights: nothing is added to
; the startup programs and no Windows service is created. The server only runs when the
; user presses the button in the window of CasaOS.

#ifndef AppVersion
  #define AppVersion "0.4.15"
#endif
#ifndef SourceDir
  #define SourceDir "..\build\app"
#endif
#ifndef OutputDir
  #define OutputDir "..\dist"
#endif

[Setup]
AppId={{7E1D5C42-8E7B-4C55-9B1A-2F6A5D0C4A11}
AppName=CasaOS
AppVersion={#AppVersion}
AppVerName=CasaOS {#AppVersion} per Windows
AppPublisher=Porting non ufficiale di CasaOS (IceWhale Technology)
AppPublisherURL=https://github.com/IceWhaleTech/CasaOS
DefaultDirName={autopf}\CasaOS
DefaultGroupName=CasaOS
DisableProgramGroupPage=yes
PrivilegesRequired=lowest
PrivilegesRequiredOverridesAllowed=dialog
ArchitecturesAllowed=x64compatible
ArchitecturesInstallIn64BitMode=x64compatible
MinVersion=10.0.17763
OutputDir={#OutputDir}
OutputBaseFilename=CasaOS-Setup-{#AppVersion}
SetupIconFile=..\launcher\assets\icon.ico
UninstallDisplayIcon={app}\CasaOS.exe
UninstallDisplayName=CasaOS
WizardStyle=modern
WizardImageFile=..\launcher\assets\wizard.bmp,..\launcher\assets\wizard-2x.bmp
WizardSmallImageFile=..\launcher\assets\wizard-small.bmp,..\launcher\assets\wizard-small-2x.bmp
LicenseFile=..\..\CasaOS\LICENSE
Compression=lzma2/ultra64
SolidCompression=yes
ShowLanguageDialog=auto
CloseApplications=no
RestartApplications=no

[Languages]
Name: "it"; MessagesFile: "compiler:Languages\Italian.isl"
Name: "en"; MessagesFile: "compiler:Default.isl"

[CustomMessages]
it.DataPageTitle=Cartella dei tuoi file
it.DataPageDesc=Dove vuoi tenere i file e i dati delle app di CasaOS?
it.DataPageInfo=Questa è la cartella che in CasaOS si chiama DATA, con Documenti, Download, Galleria, Media e i dati delle app. Se pensi di tenerci molti file, puoi sceglierne una su un altro disco.
it.DataPageLabel=Cartella:
it.Preparing=Preparazione dei servizi di CasaOS...
it.RunAfter=Apri CasaOS
it.DesktopIcon=Crea un'icona sul desktop
it.RemoveData=Vuoi eliminare anche le impostazioni e l'account di CasaOS?%n%nI tuoi file nella cartella DATA NON vengono toccati in nessun caso.
en.DataPageTitle=Folder of your files
en.DataPageDesc=Where do you want to keep the files and the app data of CasaOS?
en.DataPageInfo=This is the folder called DATA in CasaOS, with Documents, Downloads, Gallery, Media and the data of the apps. If you plan to keep many files there, you can choose one on another disk.
en.DataPageLabel=Folder:
en.Preparing=Preparing the services of CasaOS...
en.RunAfter=Open CasaOS
en.DesktopIcon=Create a desktop icon
en.RemoveData=Do you also want to delete the settings and the account of CasaOS?%n%nYour files in the DATA folder are NOT touched in any case.

[Tasks]
Name: "desktopicon"; Description: "{cm:DesktopIcon}"

[Files]
Source: "{#SourceDir}\CasaOS.exe"; DestDir: "{app}"; Flags: ignoreversion
Source: "{#SourceDir}\ui-message-bus.json"; DestDir: "{app}"; Flags: ignoreversion
Source: "{#SourceDir}\bin\*"; DestDir: "{app}\bin"; Flags: ignoreversion recursesubdirs createallsubdirs
Source: "{#SourceDir}\www\*"; DestDir: "{app}\www"; Flags: ignoreversion recursesubdirs createallsubdirs
Source: "..\..\CasaOS\LICENSE"; DestDir: "{app}"; DestName: "LICENSE.txt"; Flags: ignoreversion

[Icons]
Name: "{group}\CasaOS"; Filename: "{app}\CasaOS.exe"
Name: "{autodesktop}\CasaOS"; Filename: "{app}\CasaOS.exe"; Tasks: desktopicon

[Run]
Filename: "{app}\CasaOS.exe"; Parameters: "--prepare"; StatusMsg: "{cm:Preparing}"; Flags: runhidden nowait
Filename: "{app}\CasaOS.exe"; Description: "{cm:RunAfter}"; Flags: nowait postinstall skipifsilent

[UninstallRun]
Filename: "{app}\CasaOS.exe"; Parameters: "--quit"; Flags: runhidden waituntilterminated; RunOnceId: "QuitCasaOS"

[UninstallDelete]
Type: files; Name: "{app}\data-path.txt"
Type: dirifempty; Name: "{app}"

[Code]
var
  DataPage: TInputDirWizardPage;

function ExistingDataPath(): String;
var
  Lines: TArrayOfString;
begin
  Result := '';
  if LoadStringsFromFile(ExpandConstant('{app}\data-path.txt'), Lines) and (GetArrayLength(Lines) > 0) then
    Result := Trim(Lines[0]);
end;

procedure InitializeWizard();
begin
  DataPage := CreateInputDirPage(wpSelectDir, CustomMessage('DataPageTitle'), CustomMessage('DataPageDesc'),
    CustomMessage('DataPageInfo'), False, 'CasaOS');
  DataPage.Add(CustomMessage('DataPageLabel'));
  DataPage.Values[0] := ExpandConstant('{%USERPROFILE}\CasaOS');
end;

procedure CurPageChanged(CurPageID: Integer);
var
  Existing: String;
begin
  // an update keeps the folder chosen the first time
  if CurPageID = DataPage.ID then
  begin
    Existing := ExistingDataPath();
    if Existing <> '' then
      DataPage.Values[0] := Existing;
  end;
end;

// The server of a previous version must be stopped before its files are replaced.
function PrepareToInstall(var NeedsRestart: Boolean): String;
var
  ResultCode: Integer;
begin
  Result := '';
  if FileExists(ExpandConstant('{app}\CasaOS.exe')) then
    Exec(ExpandConstant('{app}\CasaOS.exe'), '--quit', '', SW_HIDE, ewWaitUntilTerminated, ResultCode);
end;

procedure CurStepChanged(CurStep: TSetupStep);
var
  Lines: TArrayOfString;
begin
  // in UTF-8: the path can contain the accents of the user name
  if CurStep = ssPostInstall then
  begin
    SetArrayLength(Lines, 1);
    Lines[0] := DataPage.Values[0];
    SaveStringsToUTF8File(ExpandConstant('{app}\data-path.txt'), Lines, False);
  end;
end;

procedure CurUninstallStepChanged(CurUninstallStep: TUninstallStep);
var
  Root: String;
begin
  // only the settings: the DATA folder holds the files of the user and it is never deleted
  if (CurUninstallStep = usPostUninstall) and not UninstallSilent() then
    if MsgBox(CustomMessage('RemoveData'), mbConfirmation, MB_YESNO or MB_DEFBUTTON2) = IDYES then
    begin
      Root := ExpandConstant('{localappdata}\CasaOS');
      DelTree(Root + '\root', True, True, True);
      RemoveDir(Root);
    end;
end;
