package render_dom

import (
	"context"
	"fmt"
	"github.com/LubyRuffy/chrome_proxy/chrome_action"
	"github.com/LubyRuffy/chrome_proxy/models"
	"github.com/chromedp/chromedp"
	"log"
	"os"
	"time"
)

// RenderDom 生成单个url的 dom html
func RenderDom(options *models.ChromeParam) (*models.RenderDomOutput, error) {
	log.Println("RenderDom of url:", options.URL)

	var preActions []chromedp.Action
	var actions []chromedp.Action

	var title string
	actions = append(actions, chromedp.Title(&title))
	var location string
	actions = append(actions, chromedp.Location(&location))

	// set user-agent
	if options.UserAgent == "" {
		options.UserAgent = models.DefaultUserAgent
	}

	// prepare the chrome options
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("incognito", true), // 隐身模式
		chromedp.Flag("ignore-certificate-errors", true),
		chromedp.Flag("enable-automation", true),
		chromedp.Flag("disable-extensions", true),
		chromedp.Flag("disable-setuid-sandbox", true),
		chromedp.Flag("disable-web-security", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.IgnoreCertErrors,
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		chromedp.NoSandbox,
		chromedp.DisableGPU,
		chromedp.UserAgent(options.UserAgent), // chromedp.Flag("user-agent", defaultUserAgent)
		chromedp.WindowSize(1024, 768),
	)

	// set proxy if exists
	if options.Proxy != "" {
		opts = append(opts, chromedp.ProxyServer(options.Proxy))
	}

	if models.Debug {
		opts = append(chromedp.DefaultExecAllocatorOptions[:2],
			chromedp.DefaultExecAllocatorOptions[3:]...)
		opts = append(opts, chromedp.Flag("auto-open-devtools-for-tabs", true))
	}

	allocCtx, bcancel := chromedp.NewExecAllocator(context.TODO(), opts...)
	defer func() {
		bcancel()
		b := chromedp.FromContext(allocCtx).Browser
		if b != nil && b.Process() != nil {
			b.Process().Signal(os.Kill)
		}
	}()

	ctx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(func(s string, i ...interface{}) {}))
	defer cancel()
	ctx, cancel = context.WithTimeout(ctx, time.Duration(options.Timeout)*time.Second)
	defer cancel()

	var out models.ChromeActionOutput
	err := chrome_action.ChromeActions(ctx, options.ChromeActionInput, &out, preActions, actions...)

	if err != nil {
		return nil, fmt.Errorf("RenderDom failed(%w): %s", err, options.URL)
	}

	return &models.RenderDomOutput{
		Html:     out.OutHtml,
		Title:    title,
		Location: location,
	}, nil
}
