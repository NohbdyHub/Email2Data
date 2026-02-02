package util

import (
	"fmt"
	"time"

	progressbar "github.com/schollz/progressbar/v3"
)

func Spinner(description, completion string, options ...progressbar.Option) *progressbar.ProgressBar {
	opt := []progressbar.Option{
		progressbar.OptionSetDescription(description),
		progressbar.OptionSetTheme(progressbar.ThemeASCII),
		progressbar.OptionSpinnerCustom([]string{"[ ฅ₍^.  ̫ .^₎ ]", "[ ₍^. ̫  .^₎ฅ ]"}),
		progressbar.OptionSetRenderBlankState(true),
		progressbar.OptionSetSpinnerChangeInterval(time.Millisecond * 333),
		progressbar.OptionClearOnFinish(),
		progressbar.OptionOnCompletion(func() { fmt.Printf("[ ฅ₍^. ̫ .^₎ฅ ] %s\n", completion) }),
	}

	opt = append(options, opt...)

	return progressbar.NewOptions(-1, opt...)

}
