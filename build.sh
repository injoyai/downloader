


name="downloader"
GOOS=windows GOARCH=amd64 go build -v -ldflags="-w -s -H windowsgui" -o ./$name.exe
echo "Windows编译完成..."
echo "开始压缩..."


./upx -9 -k "./$name.exe"
if [ -f "./$name.ex~" ]; then
  rm "./$name.ex~"
fi
if [ -f "./$name.000" ]; then
  rm "./$name.000"
fi

sleep 20s
