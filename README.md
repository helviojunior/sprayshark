# SprayShark

SprayShark is a modular password sprayer with threading! 

Available modules:

1. Enumeration
2. Spray

## Some amazing features

* [x] You can stop and resume anytime
* [x] Control user/password that already tested
* [x] Control if already tested user does not exists
* [x] Multi threading
* [x] Save screenshots of valid authentications
* [x] Much more...

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

# Utilization

```
$ sprayshark -h


                                                    #
                                                   .#++#
                                                    ###++#
  .                                                 ####+++#
  #++#                                              #####+++#
   #+++#                                           -#######+++#
    -#+++#                                         ##########+++##
      ##+++#                                -###+++++++++++++++++++++++++++#####+.
       ##+++#.                       .###++++++++++++++++++++++++++++++++   -++++   .+++    #-
        ###+++#             +++###+++++++-.++++++++++++          -+++++   +.   +#+..     +++++++++##
        ####+++#      .###.          +++.          +++++++++++   ++++   ++-.     ++++   ++++++++#++++++.
         ####+++++++++++--  .+++#################  .+.          -+++   +     ..-#++++#..............#.
         ##################.         -##          ###-  ##+++++####+++++++++#.#...............##+
        ########      +##.            ##  .#####################++++#######...................#
       #######         +#.    .###-......+###################.#++++++++++#.............+##.
       #####                             ###########+-........##+++++++++-.-####+.
      ####                              .##                  --###++++++#
     #                                   SprayShark          #.-##++++++.
                                                              ..##+++++#
                                                             -.#-##++#-
                                                             +#  #+#

Usage:
  sprayshark [command]

Available Commands:
  help        Help about any command
  spray       Perform password spray
  version     Get the sprayshark version

Flags:
  -D, --debug-log           Enable debug logging
  -h, --help                help for sprayshark
  -X, --proxy string        Proxy to pass traffic through: <scheme://ip:port>
      --proxy-pass string   Proxy Password
      --proxy-user string   Proxy User
  -q, --quiet               Silence (almost all) logging
  -K, --ssl-insecure        SSL Insecure (default true)

Use "sprayshark [command] --help" for more information about a command.
```

### Note

If you are using a proxy with a protocol other than HTTP, you should specify the schema like `socks5://127.0.0.1:9050`.

## Proxy recomendation

I recomend to use a kind of proxy to work with password spray. An amazing project is [audibleblink/doxycannon](https://github.com/audibleblink/doxycannon)

You can start the doxycannon and use the `sprayshark` with parameter `--proxy socks4://127.0.0.1:1337`

## Acknowledgments

* This project was heavily inspired by [y0k4i-1337/gsprayer](https://github.com/y0k4i-1337/gsprayer) and [sensepost/gowitness](https://github.com/sensepost/gowitness)


## Disclaimer

This tool is intended for educational purpose or for use in environments where you have been given explicit/legal authorization to do so.