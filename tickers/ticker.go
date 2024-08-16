package tickers

import "time"

type TickerAction func() []error

func RunEvery(frequency time.Duration, action TickerAction) {
	ticker := time.NewTicker(frequency)
	done := make(chan bool)

	go func() {
		for {
			<-ticker.C

			errors := action()
			if len(errors) > 0 {
				continue
			}
		}
	}()

	<-done
}
