package thread

import (
	"fmt"
	"time"
)

func SafeThread(fun func()) {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				fmt.Printf("panic happened in safe thread %v", err)
			}
		}()

		fun()
	}()
}

func SafeWait(pass_condition func() bool, exe func()) chan error {
	ret := make(chan error)
	wait := func() (_break bool) {
		defer func() {
			if err := recover(); err != nil {
				ret <- fmt.Errorf("panic happened in safe compare %v", err)
				_break = false
			}
		}()

		return pass_condition()
	}

	go func() {
		defer func() {
			if err := recover(); err != nil {
				ret <- fmt.Errorf("panic happened in safe wait %v", err)
			}
		}()
		for !wait() {
			time.Sleep(time.Millisecond * 100)
		}

		exe()
	}()
	return ret
}

func TriggerStop(stop chan bool) {
	if len(stop) > 0 {
		return
	}

	stop <- true
}

func SafeLoop(stop chan bool, sleep_period time.Duration, fun func()) {
	loop := func() {
		defer func() {
			if err := recover(); err != nil {
				fmt.Printf("panic happened in safe loop %v", err)
			}
		}()

		fun()
	}

	go func() {
		for len(stop) == 0 {
			loop()
			time.Sleep(sleep_period)
		}
	}()
}

func SafeSelect(stop chan bool, target chan interface{}, fun func(interface{})) {
	loop := func(i interface{}) {
		defer func() {
			if err := recover(); err != nil {
				fmt.Printf("panic happened in safe loop %v", err)
			}
		}()

		fun(i)
	}

	go func() {
		exit := false
		for !exit {
			select {
			case <-stop:
				exit = true
				go func() { stop <- true }()
			case out := <-target:
				loop(out)
			}
		}
	}()
}
