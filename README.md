# SprayShark

SprayShark is a modular password sprayer with threading! 

Available modules:

1. Enumeration
2. Spray

## Some amazing features

* [x] Pause and resume testing at any time.  
* [x] Manage and track tested user/password combinations.  
* [x] Ensure users are not tested multiple times.  
* [x] Utilize multi-threading for faster performance.  
* [x] Save screenshots of successful authentications.  
* [x] And much more!  


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
$ sprayshark spray -h


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
  sprayshark spray [flags]

Examples:

   - sprayshark spray -u test@helviojunior.com.br -p Test@123 --write-jsonl
   - sprayshark spray -U emails.txt -p Test@123 --save-content --write-db
   - sprayshark spray -U emails.txt -P passwords.txt
   - cat targets.txt | sprayshark spray usernames - -p Test@123 --write-db --write-jsonl

Flags:
      --chrome-header strings      Extra headers to add to requests. Supports multiple --header flags
      --chrome-path string         The path to a Google Chrome binary to use (downloads a platform-appropriate binary by default)
      --chrome-user-agent string   The user-agent string to use (default "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.6778.140 Safari/537.36")
      --chrome-window-x int        The Chrome browser window width, in pixels (default 1024)
      --chrome-window-y int        The Chrome browser window height, in pixels (default 768)
      --chrome-wss-url string      A websocket URL to connect to a remote, already running Chrome DevTools instance (i.e., Chrome started with --remote-debugging-port)
      --delay int                  Number of seconds delay between navigation and screenshotting (default 3)
      --driver string              The scan driver to use. Can be one of [gorod, chromedp] (default "chromedp")
  -h, --help                       help for spray
      --log-scan-errors            Log scan errors (timeouts, DNS errors, etc.) to stderr (warning: can be verbose!)
  -p, --password string            Single password
  -P, --passwords string           File containing passwords
      --save-content               Save content from network requests to the configured writers. WARNING: This flag has the potential to make your storage explode in size
      --screenshot-format string   Format to save screenshots as. Valid formats are: jpeg, png (default "jpeg")
      --screenshot-fullpage        Do full-page screenshots, instead of just the viewport
  -s, --screenshot-path string     Path to store screenshots (default "./screenshots")
      --screenshot-skip-save       Do not save screenshots to the screenshot-path (useful together with --write-screenshots)
      --skip-html                  Don't include the first request's HTML response when writing results
  -t, --threads int                Number of concurrent threads (goroutines) to use (default 6)
  -T, --timeout int                Number of seconds before considering a page timed out (default 60)
  -u, --username string            Single username
  -U, --usernames string           File containing usernames
      --write-csv                  Write results as CSV (has limited columns)
      --write-csv-file string      The file to write CSV rows to (default "sprayshark.csv")
      --write-db                   Write results to a SQLite database
      --write-db-enable-debug      Enable database query debug logging (warning: verbose!)
      --write-db-uri string        The database URI to use. Supports SQLite, Postgres, and MySQL (e.g., postgres://user:pass@host:port/db) (default "sqlite://sprayshark.sqlite3")
      --write-jsonl                Write results as JSON lines
      --write-jsonl-file string    The file to write JSON lines to (default "sprayshark.jsonl")
      --write-none                 Use an empty writer to silence warnings
      --write-screenshots          Store screenshots with writers in addition to filesystem storage

Global Flags:
  -D, --debug-log           Enable debug logging
  -X, --proxy string        Proxy to pass traffic through: <scheme://ip:port>
      --proxy-pass string   Proxy Password
      --proxy-user string   Proxy User
  -q, --quiet               Silence (almost all) logging
  -K, --ssl-insecure        SSL Insecure (default true)
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