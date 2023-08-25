package chrome_action

import (
	"context"
	"github.com/LubyRuffy/chrome_proxy/models"
	"github.com/chromedp/chromedp"
	"log"
	"strings"
	"time"
)

// ChromeActions 完成chrome的headless操作
func ChromeActions(ctx context.Context, in models.ChromeActionInput, out *models.ChromeActionOutput, preActions []chromedp.Action, actions ...chromedp.Action) error {
	var err error

	realActions := []chromedp.Action{
		// 同步处理，如果用异步会导致后面的title、location获取出现问题
		chromedp.ActionFunc(func(ctx context.Context) error {
			// 等待完成，要么是body出来了，要么是资源加载完成
			err := chromedp.Navigate(in.URL).Do(ctx)
			if err != nil {
				return err
			}
			err = chromedp.WaitReady("body", chromedp.ByQuery).Do(ctx)
			if err == nil {
				if err2 := chromedp.OuterHTML("html", &out.OutHtml).Do(ctx); err2 != nil {
					log.Println("[DEBUG] fetch html failed:", err2)
					return err
				}
			}

			// 20211219发现如果存在JS前端框架 (如vue, react...) 执行等待读取.
			html2Low := strings.ToLower(out.OutHtml)

			if strings.Contains(html2Low, "javascript") || strings.Contains(html2Low, "</script>") {
				// 这里不是100%会出现div，所以会导致 context deadline exceeded，要把这个错误包含并获取html渲染内容
				newCtx, cancel := context.WithTimeout(ctx, time.Duration(in.Sleep+2)*time.Second)
				defer cancel()

				// 将休眠移至当前位置，兼容title & location & html 的获取
				if err = chromedp.Sleep(time.Duration(in.Sleep) * time.Second).Do(newCtx); err != nil {
					return err
				}
				if err2 := chromedp.OuterHTML("html", &out.OutHtml).Do(newCtx); err2 != nil {
					// extra error, doesnt affect anything else
					log.Println("[DEBUG] fetch htmlDOM failed:", err2)
					return nil
				}
			}

			return nil
		}),
	}

	// 用于RenderDOM中的actions执行
	if preActions == nil {
		preActions = []chromedp.Action{}
	}

	realActions = append(preActions, realActions...)
	realActions = append(realActions, actions...)

	// run task list
	err = chromedp.Run(ctx, realActions...)

	// 以 finished 作为当前任务结尾
	if err != nil && (err.Error() == "finished" || err.Error() == "context deadline exceeded") {
		return nil
	}

	return err
}
