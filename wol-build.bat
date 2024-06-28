cd /d %~dp0
set GOOS=linux
go build -ldflags "-s -w" -o bin\%~n1 %~1
pause
