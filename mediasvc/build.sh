output_linux_debug="./_build/linux/mediasvc_debug"
output_linux_production="./_build/linux/mediasvc_production"
output_mac="./_build/mac/mediasvc"
output_win="./_build/win/mediasvc.exe"
output="./_build/"

rm -f "./_build/"*

case $1 in
    linux_debug)
    echo "build for linux_debug"
    CGO_ENABLED=0 GOOS=linux go build -o=$output_linux_debug
    ;;
    linux_prod)
    echo "build for linux_production"
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o=$output_linux_production
    ;;
    mac)
    echo "build for macos"
    CGO_ENABLED=0 GOOS=darwin go build -o=$output_mac
    ;;
    win)
    echo "build for windows"
    CGO_ENABLED=0 GOOS=windows go build -o=$output_win
    ;;
    *)
    echo "build for local"
    go build -o=$output
    ;;
esac

echo "done"