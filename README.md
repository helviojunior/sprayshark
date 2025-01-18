# SprayShark

SprayShark is a modular password sprayer with threading! 

Available modules:

1. Enumeration
2. Spray

# Build

Clone the repository and build the project with Golang:

```
git clone https://github.com/helviojunior/sprayshark.git
cd sprayshark
go get ./...
go build
```

If you want to update go.sum file just run the command `go mod tidy`.

# Installing system wide

After build run the commands bellow

```
go install .
ln -s /root/go/bin/sprayshark /usr/bin/sprayshark
```

### Note

If you are using a proxy with a protocol other than HTTP, you should specify the schema like `socks5://127.0.0.1:9050`.

## Acknowledgments

* This project was heavily inspired by [y0k4i-1337/gsprayer](https://github.com/y0k4i-1337/gsprayer) and [sensepost/gowitness](https://github.com/sensepost/gowitness)


## Disclaimer

This tool is intended for educational purpose or for use in environments where you have been given explicit/legal authorization to do so.