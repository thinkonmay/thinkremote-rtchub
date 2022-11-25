package adaptive

import "fmt"

type AdsContext struct {
	In chan string
}

func NewAdsContext(InChan chan string) *AdsContext {
	ret := &AdsContext{}

	ret.In = InChan
	go func() {
		for {
			metricRaw := <-ret.In
			fmt.Printf("%s\n",metricRaw);
		}
	}()

	return ret
}