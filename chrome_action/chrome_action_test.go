package chrome_action

import (
	"context"
	"github.com/LubyRuffy/chrome_proxy/models"
	"github.com/chromedp/chromedp"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

func Test_chromeActions(t *testing.T) {
	buf := []byte{}
	type args struct {
		in      models.ChromeActionInput
		logf    func(string, ...interface{})
		timeout int
		actions []chromedp.Action
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "测试正常screen picture",
			args: args{
				in: models.ChromeActionInput{
					URL: "https://www.baidu.com",
				},
				logf:    func(s string, i ...interface{}) {},
				timeout: 10,
				actions: []chromedp.Action{
					chromedp.Sleep(time.Second * time.Duration(1)),
					chromedp.CaptureScreenshot(&buf),
				},
			},
		},
		{
			name: "测试钓鱼页面proxy & UA",
			args: args{
				in: models.ChromeActionInput{
					URL:       "http://shop.bnuzac.com/articles.php/about/456415?newsid=oaxn1d.html",
					Proxy:     "socks5://127.0.0.1:7890",
					UserAgent: "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4240.198 Safari/537.36",
				},
				logf:    func(s string, i ...interface{}) {},
				timeout: 10,
				actions: []chromedp.Action{
					chromedp.Sleep(time.Second * time.Duration(1)),
					chromedp.CaptureScreenshot(&buf),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// set user-agent
			if tt.args.in.UserAgent == "" {
				tt.args.in.UserAgent = models.DefaultUserAgent
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
				chromedp.UserAgent(tt.args.in.UserAgent), // chromedp.Flag("user-agent", defaultUserAgent)
				chromedp.WindowSize(1024, 768),
			)

			// set proxy if exists
			if tt.args.in.Proxy != "" {
				opts = append(opts, chromedp.ProxyServer(tt.args.in.Proxy))
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
			ctx, cancel = context.WithTimeout(ctx, time.Duration(tt.args.timeout)*time.Second)
			defer cancel()

			var out models.ChromeActionOutput
			err := ChromeActions(ctx, tt.args.in, &out, nil, tt.args.actions...)
			assert.Nil(t, err)
			assert.NotNil(t, buf)
			t.Logf("screenshot result: %s", buf[0:5])
		})
	}
}
