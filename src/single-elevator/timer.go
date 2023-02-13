package single_elevator

import (
	"time"
)

var timer_control = make(chan bool)

func Timer_main(timer_timeout chan<- bool){
	counter := 0
	counting := false
	for {
		select {
		case control := <- timer_control:
			if control {
				counting = true
			} else {
				counting = false
				counter = 0
			}
		default:
			if counter == 3000 {
				counter = 0
				counting = false
				timer_timeout <- true
			}
			
			if counting {
				counter++
				time.Sleep(time.Millisecond)
			}
		}
	}
}

func Timer_start(){
	timer_control <- true
}

func Timer_kill(){
	timer_control <- false
}