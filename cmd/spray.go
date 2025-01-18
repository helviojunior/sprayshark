package cmd

import (
    "errors"
    "log/slog"
    "net/mail"
    "os"

    "github.com/helviojunior/sprayshark/internal/ascii"
    "github.com/helviojunior/sprayshark/internal/islazy"
    "github.com/helviojunior/sprayshark/pkg/log"
    "github.com/helviojunior/sprayshark/pkg/runner"
    "github.com/helviojunior/sprayshark/pkg/database"
    driver "github.com/helviojunior/sprayshark/pkg/runner/drivers"
    "github.com/helviojunior/sprayshark/pkg/writers"
    "github.com/helviojunior/sprayshark/pkg/readers"
    "github.com/spf13/cobra"
)

var fileCmdOptions = &readers.FileReaderOptions{}
var sprayOptions = runner.UserOptions{}

var scanWriters = []writers.Writer{}
var scanDriver runner.Driver
var scanRunner *runner.Runner

var scanCmd = &cobra.Command{
    Use:   "spray",
    Short: "Perform password spray",
    Long: ascii.LogoHelp(ascii.Markdown(`
# spray

Perform password spray.

By default, sprayshark will only show information regarding the spray process. 
However, that is only half the fun! You can add multiple _writers_ that will 
collect information such as response codes, content, and more. You can specify 
multiple writers using the _--writer-*_ flags (see --help).
`)),
    Example: `
   - sprayshark spray -u test@helviojunior.com.br -p Test@123 --write-jsonl
   - sprayshark spray -U emails.txt -p Test@123 --save-content --write-db
   - sprayshark spray -U emails.txt -P passwords.txt
   - cat targets.txt | sprayshark spray usernames - -p Test@123 --write-db --write-jsonl`,
    PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
        var err error

        // Annoying quirk, but because I'm overriding PersistentPreRun
        // here which overrides the parent it seems.
        // So we need to explicitly call the parent's one now.
        if err = rootCmd.PersistentPreRunE(cmd, args); err != nil {
            return err
        }

        // An slog-capable logger to use with drivers and runners
        logger := slog.New(log.Logger)

        // Configure the driver
        switch opts.Scan.Driver {
        case "chromedp":
            scanDriver, err = driver.NewChromedp(logger, *opts)
            if err != nil {
                return err
            }
        default:
            return errors.New("invalid scan driver chosen")
        }

        log.Debug("scanning driver started", "driver", opts.Scan.Driver)

        // Configure writers that subcommand scanners will pass to
        // a runner instance.

        //The first one is the general writer (global user)
        w, err := writers.NewDbWriter("sqlite://" + opts.Writer.UserPath +"/.sprayshark.db", false)
        if err != nil {
            return err
        }
        scanWriters = append(scanWriters, w)

        //The second one is the STDOut
        if opts.Logging.Silence != true {
            w, err := writers.NewStdoutWriter()
            if err != nil {
                return err
            }
            scanWriters = append(scanWriters, w)
        }
    
        if opts.Writer.Jsonl {
            w, err := writers.NewJsonWriter(opts.Writer.JsonlFile)
            if err != nil {
                return err
            }
            scanWriters = append(scanWriters, w)
        }

        if opts.Writer.Db {
            w, err := writers.NewDbWriter(opts.Writer.DbURI, opts.Writer.DbDebug)
            if err != nil {
                return err
            }
            scanWriters = append(scanWriters, w)
        }

        if opts.Writer.Csv {
            w, err := writers.NewCsvWriter(opts.Writer.CsvFile)
            if err != nil {
                return err
            }
            scanWriters = append(scanWriters, w)
        }

        if opts.Writer.None {
            w, err := writers.NewNoneWriter()
            if err != nil {
                return err
            }
            scanWriters = append(scanWriters, w)
        }

        if len(scanWriters) == 0 {
            log.Warn("no writers have been configured. to persist probe results, add writers using --write-* flags")
        }

        // Get the runner up. Basically, all of the subcommands will use this.
        scanRunner, err = runner.NewRunner(logger, scanDriver, *opts, scanWriters)
        if err != nil {
            return err
        }

        return nil
    },
    PreRunE: func(cmd *cobra.Command, args []string) error {
        if sprayOptions.Username == "" && fileCmdOptions.UserFile == "" {
            return errors.New("a username or username file must be specified")
        }

        if fileCmdOptions.UserFile != "" {
            if fileCmdOptions.UserFile != "-" && !islazy.FileExists(fileCmdOptions.UserFile) {
                return errors.New("usernames file is not readable")
            }
        }

        if sprayOptions.Password == "" && fileCmdOptions.PassFile == "" {
            return errors.New("a password or password file must be specified")
        }

        if fileCmdOptions.PassFile != "" {
            if !islazy.FileExists(fileCmdOptions.PassFile) {
                return errors.New("passwords file is not readable")
            }
        }

        return nil
    },
    Run: func(cmd *cobra.Command, args []string) {
        log.Debug("starting spray")

        users := make(chan string)
        passwords := make(chan string)
        reader := readers.NewFileReader(fileCmdOptions)

        go func() {
            if fileCmdOptions.UserFile != "" {
                if err := reader.ReadEmails(users); err != nil {
                    log.Error("error in reader.Read", "err", err)
                    os.Exit(2)
                }
            }else{
                m, err := mail.ParseAddress(sprayOptions.Username)
                if err != nil {
                    log.Error("invalid user email", "e-mail", sprayOptions.Username, "err", err)
                    os.Exit(2)
                }
                defer close(users)
                users <- m.Address
            }
        }()
    
        go func() {
            if fileCmdOptions.PassFile != "" {
                if err := reader.ReadPasswords(passwords); err != nil {
                    log.Error("error in reader.Read", "err", err)
                    os.Exit(2)
                }   
            }else{
                defer close(passwords)
                passwords <- sprayOptions.Password
            }
        }()

        // Check runned items
        conn, _ := database.Connection("sqlite://" + opts.Writer.UserPath +"/.sprayshark.db", true, false)

        go func() {
            defer close(scanRunner.Targets)
            for p := range passwords {
                for u := range users {

                    i := true
                    if conn != nil {
                        response := conn.Raw("SELECT count(id) as count from results WHERE failed = 0 AND user = ? AND password = ?", u, p)
                        if response != nil {
                            var cnt int
                            _ = response.Row().Scan(&cnt)
                            i = (cnt == 0)
                            if cnt >= 0 {
                                log.Info("[already tested, same password]", "user", u, "pass", p)
                            }
                        }
                        if i {
                            response := conn.Raw("SELECT count(id) as count from results WHERE failed = 0 AND user = ? AND user_exists = 0", u)
                            if response != nil {
                                var cnt int
                                _ = response.Row().Scan(&cnt)
                                i = (cnt == 0)
                                if cnt >= 0 {
                                    log.Info("[already tested, user not found]", "user", u)
                                }
                            }
                        }
                    }

                    if i {
                        scanRunner.Targets <- runner.Credential{
                            Username: u,
                            Password: p,
                        }
                    }
                }
            }
        }()

        scanRunner.Run()
        scanRunner.Close()
    },
}

func init() {
    rootCmd.AddCommand(scanCmd)

    //Username & password Options
    scanCmd.Flags().StringVarP(&sprayOptions.Username, "username", "u", "", "Single username")
    scanCmd.Flags().StringVarP(&sprayOptions.Password, "password", "p", "", "Single password")
    scanCmd.Flags().StringVarP(&fileCmdOptions.UserFile, "usernames", "U", "", "File containing usernames")
    scanCmd.Flags().StringVarP(&fileCmdOptions.PassFile, "passwords", "P", "", "File containing passwords")
        
    // Logging control for subcommands
    scanCmd.Flags().BoolVar(&opts.Logging.LogScanErrors, "log-scan-errors", false, "Log scan errors (timeouts, DNS errors, etc.) to stderr (warning: can be verbose!)")

    // "Threads" & other
    scanCmd.Flags().StringVarP(&opts.Scan.Driver, "driver", "", "chromedp", "The scan driver to use. Can be one of [gorod, chromedp]")
    scanCmd.Flags().IntVarP(&opts.Scan.Threads, "threads", "t", 6, "Number of concurrent threads (goroutines) to use")
    scanCmd.Flags().IntVarP(&opts.Scan.Timeout, "timeout", "T", 60, "Number of seconds before considering a page timed out")
    scanCmd.Flags().IntVar(&opts.Scan.Delay, "delay", 3, "Number of seconds delay between navigation and screenshotting")
    scanCmd.Flags().StringVarP(&opts.Scan.ScreenshotPath, "screenshot-path", "s", "./screenshots", "Path to store screenshots")
    scanCmd.Flags().StringVar(&opts.Scan.ScreenshotFormat, "screenshot-format", "jpeg", "Format to save screenshots as. Valid formats are: jpeg, png")
    scanCmd.Flags().BoolVar(&opts.Scan.ScreenshotFullPage, "screenshot-fullpage", false, "Do full-page screenshots, instead of just the viewport")
    scanCmd.Flags().BoolVar(&opts.Scan.ScreenshotSkipSave, "screenshot-skip-save", false, "Do not save screenshots to the screenshot-path (useful together with --write-screenshots)")
    scanCmd.Flags().BoolVar(&opts.Scan.SaveContent, "save-content", false, "Save content from network requests to the configured writers. WARNING: This flag has the potential to make your storage explode in size")
    scanCmd.Flags().BoolVar(&opts.Scan.SkipHTML, "skip-html", false, "Don't include the first request's HTML response when writing results")
    scanCmd.Flags().BoolVar(&opts.Scan.ScreenshotToWriter, "write-screenshots", false, "Store screenshots with writers in addition to filesystem storage")

    // Chrome options
    scanCmd.Flags().StringVar(&opts.Chrome.Path, "chrome-path", "", "The path to a Google Chrome binary to use (downloads a platform-appropriate binary by default)")
    scanCmd.Flags().StringVar(&opts.Chrome.WSS, "chrome-wss-url", "", "A websocket URL to connect to a remote, already running Chrome DevTools instance (i.e., Chrome started with --remote-debugging-port)")
    scanCmd.Flags().StringVar(&opts.Chrome.UserAgent, "chrome-user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.6778.140 Safari/537.36", "The user-agent string to use")
    scanCmd.Flags().IntVar(&opts.Chrome.WindowX, "chrome-window-x", 1024, "The Chrome browser window width, in pixels")
    scanCmd.Flags().IntVar(&opts.Chrome.WindowY, "chrome-window-y", 768, "The Chrome browser window height, in pixels")
    scanCmd.Flags().StringSliceVar(&opts.Chrome.Headers, "chrome-header", []string{}, "Extra headers to add to requests. Supports multiple --header flags")

    // Write options for scan subcommands
    scanCmd.Flags().BoolVar(&opts.Writer.Db, "write-db", false, "Write results to a SQLite database")
    scanCmd.Flags().StringVar(&opts.Writer.DbURI, "write-db-uri", "sqlite://sprayshark.sqlite3", "The database URI to use. Supports SQLite, Postgres, and MySQL (e.g., postgres://user:pass@host:port/db)")
    scanCmd.Flags().BoolVar(&opts.Writer.DbDebug, "write-db-enable-debug", false, "Enable database query debug logging (warning: verbose!)")
    scanCmd.Flags().BoolVar(&opts.Writer.Csv, "write-csv", false, "Write results as CSV (has limited columns)")
    scanCmd.Flags().StringVar(&opts.Writer.CsvFile, "write-csv-file", "sprayshark.csv", "The file to write CSV rows to")
    scanCmd.Flags().BoolVar(&opts.Writer.Jsonl, "write-jsonl", false, "Write results as JSON lines")
    scanCmd.Flags().StringVar(&opts.Writer.JsonlFile, "write-jsonl-file", "sprayshark.jsonl", "The file to write JSON lines to")
    scanCmd.Flags().BoolVar(&opts.Writer.None, "write-none", false, "Use an empty writer to silence warnings")
}