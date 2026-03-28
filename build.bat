@echo off

echo.
echo  Compilation du NVIDIA NIM CLI pour Windows...
echo.

echo  [1/3] Telechargement des dependances...
go mod tidy
if %ERRORLEVEL% neq 0 (
    echo  ERREUR : go mod tidy a échoué
    pause
    exit /b 1
)

echo  [2/3] Compilation...
go build -ldflags="-s -w" -o nim.exe .
if %ERRORLEVEL% neq 0 (
    echo  ERREUR : La compilation a échoué
    pause
    exit /b 1
)

echo  [3/3] Verification...
if exist nim.exe (
    echo.
    echo  Compilation réussie ! Binaire : nim.exe
    echo.
    echo  Pour installer globalement (ajouter au PATH) :
    echo    move nim.exe C:\Windows\System32\
    echo.
    echo  Ou ajoutez le dossier courant à votre PATH Windows.
    echo.
    echo  Première utilisation :
    echo    nim auth
    echo    nim chat
    echo.
) else (
    echo  ERREUR : nim.exe introuvable après compilation
)

pause
