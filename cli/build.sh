name="download"
GOOS=windows GOARCH=amd64 go build -v -ldflags="-w -s -X main.Debug=false" -o ./$name.exe
echo "Windows编译完成..."
sleep 10
