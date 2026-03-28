@echo off
setlocal

set TARGET=C:\Windows\System32\CLI
set EXE=%TARGET%\nim.exe
set SKILLS=%TARGET%\skills
set FILES=%TARGET%\files

echo.
echo  Compilation du NVIDIA NIM CLI...
echo.

echo  [1/4] Dependances...
go mod tidy
if %ERRORLEVEL% neq 0 (
    echo  ERREUR : go mod tidy a echoue
    pause
    exit /b 1
)

echo  [2/4] Compilation...
go build -ldflags="-s -w" -o nim.exe .
if %ERRORLEVEL% neq 0 (
    echo  ERREUR : La compilation a echoue
    pause
    exit /b 1
)

echo  [3/4] Installation dans %TARGET%...
if not exist "%TARGET%" mkdir "%TARGET%"
if not exist "%SKILLS%" mkdir "%SKILLS%"
if not exist "%FILES%" mkdir "%FILES%"

move /Y nim.exe "%EXE%"
if %ERRORLEVEL% neq 0 (
    echo  ERREUR : Impossible de copier nim.exe dans %TARGET%
    echo  Relancez ce script en tant qu'Administrateur.
    pause
    exit /b 1
)

echo  [4/4] Verification...
if exist "%EXE%" (
    echo.
    echo  Installation reussie :
    echo    Executable : %EXE%
    echo    Skills     : %SKILLS%
    echo    Fichiers   : %FILES%
    echo.
    echo  Premiere utilisation :
    echo    nim auth
    echo    nim chat
    echo.
) else (
    echo  ERREUR : nim.exe introuvable apres installation
)

endlocal
pause
