package runner

import (
	"context"
	"errors"
	"log/slog"
	//"net/url"
	"net/mail"
	"os"
	"fmt"
	"sync"
	"time"
	//"strings"
	"math/rand/v2"

	"github.com/helviojunior/sprayshark/internal/islazy"
	"github.com/helviojunior/sprayshark/pkg/models"
	"github.com/helviojunior/sprayshark/pkg/writers"
)

type Credential struct {
	Username string
	Password string
}

// Runner is a runner that probes web targets using a driver
type Runner struct {
	Driver     Driver

	// options for the Runner to consider
	options Options
	// writers are the result writers to use
	writers []writers.Writer
	// log handler
	log *slog.Logger

	// Targets to scan.
	Targets chan Credential

	// in case we need to bail
	ctx    context.Context
	cancel context.CancelFunc

	status *Status

	//Test id
	uid string
}

type Status struct {
	Total int
	Tested int
	UserExists int
	NotFound int
	Valid int
	Error int
	Skipped int
	Label string
	Running bool
}

func (st *Status) Print() { 
	switch st.Label {
		case "[=====]":
            st.Label = "[ ====]"
        case  "[ ====]":
            st.Label = "[  ===]"
        case  "[  ===]":
            st.Label = "[=  ==]"
        case "[=  ==]":
            st.Label = "[==  =]"
        case  "[==  =]":
            st.Label = "[===  ]"
        case "[===  ]":
            st.Label = "[==== ]"
        default:
            st.Label = "[=====]"
	}
	fmt.Fprintf(os.Stderr, "%s\n    %s %d/%d, valid %d, exists %d, not found %d, errors %d, skipped %d\r\033[A", 
    	"                                                                                           ",
    	st.Label, st.Tested, st.Total, st.Valid, st.UserExists, st.NotFound, st.Error, st.Skipped)
} 

func (st *Status) AddResult(result *models.Result) { 
    st.Tested += 1
    if st.Tested > st.Total {
    	st.Total = st.Tested
    }
	if result.Failed {
		st.Error += 1
		return
	}
	if result.ValidCredential {
		st.Valid += 1
		return
	}
	if result.UserExists {
		st.UserExists += 1
		return
	}
	st.NotFound += 1
} 


// New gets a new Runner ready for probing.
// It's up to the caller to call Close() on the runner
func NewRunner(logger *slog.Logger, driver Driver, opts Options, writers []writers.Writer) (*Runner, error) {
	if !opts.Scan.ScreenshotSkipSave {
		screenshotPath, err := islazy.CreateDir(opts.Scan.ScreenshotPath)
		if err != nil {
			return nil, err
		}
		opts.Scan.ScreenshotPath = screenshotPath
		logger.Debug("final screenshot path", "screenshot-path", opts.Scan.ScreenshotPath)
	} else {
		logger.Debug("not saving screenshots to disk")
	}

	// screenshot format check
	if !islazy.SliceHasStr([]string{"jpeg", "png"}, opts.Scan.ScreenshotFormat) {
		return nil, errors.New("invalid screenshot format")
	}

	//

	ctx, cancel := context.WithCancel(context.Background())

	return &Runner{
		Driver:     driver,
		options:    opts,
		writers:    writers,
		Targets:    make(chan Credential),
		log:        logger,
		ctx:        ctx,
		cancel:     cancel,
		uid: 		string(time.Now().UnixMilli()),
		status:     &Status{
			Total: 0,
			Tested: 0,
			UserExists: 0,
			Valid: 0,
			Error: 0,
			Skipped: 0,
			NotFound: 0,
			Label: "[=====]",
			Running: true,
		},
	}, nil
}

// runWriters takes a result and passes it to writers
func (run *Runner) runWriters(result *models.Result) error {
	for _, writer := range run.writers {
		if err := writer.Write(result); err != nil {
			return err
		}
	}

	return nil
}

// checkUrl ensures a e-mail is valid
func (run *Runner) checkEmail(target string) error {
	_, err := mail.ParseAddress(target)
	if err != nil {
		return err
	}

	return nil
}

func (run *Runner) AddSkipped() {
	run.status.Skipped += 1
	run.status.Tested += 1
}

// Run executes the runner, processing targets as they arrive
// in the Targets channel
func (run *Runner) Run(total int, enumOnly_optional ...bool) Status {
	wg := sync.WaitGroup{}
	swg := sync.WaitGroup{}

	run.status.Total = total

	enumOnly := false
	if len(enumOnly_optional) > 0 {
		enumOnly = enumOnly_optional[0]
	}

	if !run.options.Logging.Silence {
		swg.Add(1)
		go func() {
	        defer swg.Done()
			for run.status.Running {
				select {
					case <-run.ctx.Done():
						return
					default:
			        	run.status.Print()
			        	time.Sleep(time.Duration(time.Second/2))
			    }
	        }
	    }()
	}

	// will spawn Scan.Theads number of "workers" as goroutines
	for w := 0; w < run.options.Scan.Threads; w++ {
		wg.Add(1)

		// start a worker
		go func() {
			defer wg.Done()
			for {
				select {
				case <-run.ctx.Done():
					return
				case credential, ok := <-run.Targets:
					if !ok {
						return
					}
					logger := run.log.With("user", credential.Username)
					logger.Debug("spraying ", "pass", credential.Password)

					// validate the target
					if err := run.checkEmail(credential.Username); err != nil {
						if run.options.Logging.LogScanErrors {
							logger.Error("invalid user email", "err", err)
						}
						logger.Debug("invalid user email", "err", err)
						continue
					}

					// Wait random time to not start everything at same time
					time.Sleep(time.Duration(rand.IntN(10)) * time.Second)

					good_to_go := false
					counter := 0
					for good_to_go != true {
						result, err := run.Driver.Check(credential.Username, credential.Password, run, counter * 5, enumOnly)
						if result != nil {
							result.TestId = run.uid
						}

						counter += 1
						good_to_go = (err == nil)

						if err != nil {

							// is this a chrome not found error?
							var chromeErr *ChromeNotFoundError
							if errors.As(err, &chromeErr) {
								logger.Error("no valid chrome intallation found", "err", err)
								run.cancel()
								return
							}

							logger.Debug("Error running checker, trying again...", "err", err)
							time.Sleep(time.Duration(rand.IntN(20)) * time.Second)
						}
					
						if !good_to_go && counter >= 5 {
							if result == nil {
								logger.Debug("Setting result")
								result = &models.Result{
									User: credential.Username,
									Password: credential.Password,
									ProbedAt: time.Now(),
								}
							}
							result.Failed = true
							result.FailedReason = "Cannot run checker"
							good_to_go = true
						}

						if good_to_go {
							run.status.AddResult(result)

							if err := run.runWriters(result); err != nil {
								logger.Error("failed to write result for target", "err", err)
							}
						}


					}

				}
			}

		}()
	}

	wg.Wait()
	run.status.Running = false
	swg.Wait()

	fmt.Fprintf(os.Stderr,  "                                                                                           \r")

	return *run.status
}

func (run *Runner) Close() {
	// close the driver
	run.Driver.Close()
}
