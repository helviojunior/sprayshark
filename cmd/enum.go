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

var enumFileOptions = &readers.FileReaderOptions{}
var enumOptions = runner.UserOptions{}

var enumWriters = []writers.Writer{}
var enumDriver runner.Driver
var enumRunner *runner.Runner

var enumCmd = &cobra.Command{
    Use:   "enum",
    Short: "Perform user enumeration",
    Long: ascii.LogoHelp(ascii.Markdown(`
# enum

Perform user enumeration.

By default, sprayshark will only show information regarding the spray process. 
However, that is only half the fun! You can add multiple _writers_ that will 
collect information such as response codes, content, and more. You can specify 
multiple writers using the _--writer-*_ flags (see --help).
`)),
    Example: `
   - sprayshark enum -u test@helviojunior.com.br --write-jsonl
   - sprayshark enum -U emails.txt --save-content --write-db
   - sprayshark enum -U emails.txt
   - sprayshark enum -U emails.txt --proxy socks4://127.0.0.1:1337 --write-all-screenshots
   - cat targets.txt | sprayshark enum usernames - --write-db --write-jsonl`,
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
            enumDriver, err = driver.NewChromedp(logger, *opts)
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
        w, err := writers.NewDbWriter("sqlite:///" + opts.Writer.UserPath +"/.sprayshark.db", false)
        if err != nil {
            return err
        }
        enumWriters = append(enumWriters, w)

        //The second one is the STDOut
        if opts.Logging.Silence != true {
            w, err := writers.NewStdoutWriter()
            if err != nil {
                return err
            }
            enumWriters = append(enumWriters, w)
        }
    
        if opts.Writer.Jsonl {
            w, err := writers.NewJsonWriter(opts.Writer.JsonlFile)
            if err != nil {
                return err
            }
            enumWriters = append(enumWriters, w)
        }

        if opts.Writer.Db {
            w, err := writers.NewDbWriter(opts.Writer.DbURI, opts.Writer.DbDebug)
            if err != nil {
                return err
            }
            enumWriters = append(enumWriters, w)
        }

        if opts.Writer.Csv {
            w, err := writers.NewCsvWriter(opts.Writer.CsvFile)
            if err != nil {
                return err
            }
            enumWriters = append(enumWriters, w)
        }

        if opts.Writer.None {
            w, err := writers.NewNoneWriter()
            if err != nil {
                return err
            }
            enumWriters = append(enumWriters, w)
        }

        if len(enumWriters) == 0 {
            log.Warn("no writers have been configured. to persist probe results, add writers using --write-* flags")
        }

        // Get the runner up. Basically, all of the subcommands will use this.
        enumRunner, err = runner.NewRunner(logger, enumDriver, *opts, enumWriters)
        if err != nil {
            return err
        }

        return nil
    },
    PreRunE: func(cmd *cobra.Command, args []string) error {
        if enumOptions.Username == "" && enumFileOptions.UserFile == "" {
            return errors.New("a username or username file must be specified")
        }

        if enumFileOptions.UserFile != "" {
            if enumFileOptions.UserFile != "-" && !islazy.FileExists(enumFileOptions.UserFile) {
                return errors.New("usernames file is not readable")
            }
        }

        return nil
    },
    Run: func(cmd *cobra.Command, args []string) {
        log.Debug("starting user enumeration")

        users := []string{}
        reader := readers.NewFileReader(enumFileOptions)

        if enumFileOptions.UserFile != "" {
            log.Debugf("Reading users file: %s", enumFileOptions.UserFile)
            if err := reader.ReadEmails(&users); err != nil {
                log.Error("error in reader.Read", "err", err)
                os.Exit(2)
            }
            
        }else{
            m, err := mail.ParseAddress(enumOptions.Username)
            if err != nil {
                log.Error("invalid user email", "e-mail", enumOptions.Username, "err", err)
                os.Exit(2)
            }
            users = append(users, m.Address)
        }
        log.Debugf("Loaded %d user(s)", len(users))

        log.Infof("Enumerating %d user(s)", len(users))

        // Check runned items
        conn, _ := database.Connection("sqlite:///" + opts.Writer.UserPath +"/.sprayshark.db", true, false)

        go func() {
            defer close(enumRunner.Targets)
            for _, u := range users {

                i := true
                if conn != nil {
                    response := conn.Raw("SELECT count(id) as count from results WHERE failed = 0 AND user = ? AND valid_credential = 1", u)
                    if response != nil {
                        var cnt int
                        _ = response.Row().Scan(&cnt)
                        i = (cnt == 0)
                        if cnt > 0 {
                            log.Info("[Credential already found]", "user", u)
                        }
                    }
                    if i {
                        response := conn.Raw("SELECT count(id) as count from results WHERE failed = 0 AND user = ?", u)
                        if response != nil {
                            var cnt int
                            _ = response.Row().Scan(&cnt)
                            i = (cnt == 0)
                            if cnt > 0 {
                                log.Debug("[already enumerated]", "user", u)
                            }
                        }
                    }
                    
                }

                if i {
                    enumRunner.Targets <- runner.Credential{
                        Username: u,
                        Password: "",
                    }
                }else{
                    enumRunner.AddSkipped()
                }
            }
        
        }()

        status := enumRunner.Run(len(users), true)
        enumRunner.Close()

        st := "Execution statistics\n"
        st += "     -> Wordlist total...: %d\n"
        st += "     -> Skipped..........: %d\n"
        st += "     -> Total tested.....: %d\n"
        st += "     -> Existing users...: %d\n"
        st += "     -> User not found...: %d\n"
        st += "     -> Execution error..: %d\n"

        log.Warnf(st, 
             status.Total, 
             status.Skipped,
             status.Total - status.Skipped,
             status.UserExists,
             status.NotFound,
             status.Error,
        )
    },
}

func init() {
    rootCmd.AddCommand(enumCmd)

    //Username & password Options
    enumCmd.Flags().StringVarP(&enumOptions.Username, "username", "u", "", "Single username")
    enumCmd.Flags().StringVarP(&enumFileOptions.UserFile, "usernames", "U", "", "File containing usernames")
        
    // Logging control for subcommands
    enumCmd.Flags().BoolVar(&opts.Logging.LogScanErrors, "log-scan-errors", false, "Log scan errors (timeouts, DNS errors, etc.) to stderr (warning: can be verbose!)")

    // "Threads" & other
    enumCmd.Flags().StringVarP(&opts.Scan.Driver, "driver", "", "chromedp", "The scan driver to use. Can be one of [gorod, chromedp]")
    enumCmd.Flags().IntVarP(&opts.Scan.Threads, "threads", "t", 6, "Number of concurrent threads (goroutines) to use")
    enumCmd.Flags().IntVarP(&opts.Scan.Timeout, "timeout", "T", 60, "Number of seconds before considering a page timed out")
    enumCmd.Flags().IntVar(&opts.Scan.Delay, "delay", 3, "Number of seconds delay between navigation and screenshotting")
    enumCmd.Flags().StringVarP(&opts.Scan.ScreenshotPath, "screenshot-path", "s", "./screenshots", "Path to store screenshots")
    enumCmd.Flags().StringVar(&opts.Scan.ScreenshotFormat, "screenshot-format", "jpeg", "Format to save screenshots as. Valid formats are: jpeg, png")
    enumCmd.Flags().BoolVar(&opts.Scan.ScreenshotFullPage, "screenshot-fullpage", false, "Do full-page screenshots, instead of just the viewport")
    enumCmd.Flags().BoolVar(&opts.Scan.ScreenshotSkipSave, "screenshot-skip-save", false, "Do not save screenshots to the screenshot-path (useful together with --write-screenshots)")
    enumCmd.Flags().BoolVar(&opts.Scan.SaveHTML, "save-html", false, "Include the result request's HTML response when writing results")
    enumCmd.Flags().BoolVar(&opts.Scan.ScreenshotToWriter, "write-screenshots", false, "Store screenshots with writers in addition to filesystem storage")
    enumCmd.Flags().BoolVar(&opts.Scan.ScreenshotSaveAll, "write-all-screenshots", false, "Store all result screenshots to filesystem storage")

    // Chrome options
    enumCmd.Flags().StringVar(&opts.Chrome.Path, "chrome-path", "", "The path to a Google Chrome binary to use (downloads a platform-appropriate binary by default)")
    enumCmd.Flags().StringVar(&opts.Chrome.WSS, "chrome-wss-url", "", "A websocket URL to connect to a remote, already running Chrome DevTools instance (i.e., Chrome started with --remote-debugging-port)")
    enumCmd.Flags().StringVar(&opts.Chrome.UserAgent, "chrome-user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.6778.140 Safari/537.36", "The user-agent string to use")
    enumCmd.Flags().IntVar(&opts.Chrome.WindowX, "chrome-window-x", 1024, "The Chrome browser window width, in pixels")
    enumCmd.Flags().IntVar(&opts.Chrome.WindowY, "chrome-window-y", 768, "The Chrome browser window height, in pixels")
    enumCmd.Flags().StringSliceVar(&opts.Chrome.Headers, "chrome-header", []string{}, "Extra headers to add to requests. Supports multiple --header flags")

    // Write options for scan subcommands
    enumCmd.Flags().BoolVar(&opts.Writer.Db, "write-db", false, "Write results to a SQLite database")
    enumCmd.Flags().StringVar(&opts.Writer.DbURI, "write-db-uri", "sqlite://sprayshark.sqlite3", "The database URI to use. Supports SQLite, Postgres, and MySQL (e.g., postgres://user:pass@host:port/db)")
    enumCmd.Flags().BoolVar(&opts.Writer.DbDebug, "write-db-enable-debug", false, "Enable database query debug logging (warning: verbose!)")
    enumCmd.Flags().BoolVar(&opts.Writer.Csv, "write-csv", false, "Write results as CSV (has limited columns)")
    enumCmd.Flags().StringVar(&opts.Writer.CsvFile, "write-csv-file", "sprayshark.csv", "The file to write CSV rows to")
    enumCmd.Flags().BoolVar(&opts.Writer.Jsonl, "write-jsonl", false, "Write results as JSON lines")
    enumCmd.Flags().StringVar(&opts.Writer.JsonlFile, "write-jsonl-file", "sprayshark.jsonl", "The file to write JSON lines to")
    enumCmd.Flags().BoolVar(&opts.Writer.None, "write-none", false, "Use an empty writer to silence warnings")
}