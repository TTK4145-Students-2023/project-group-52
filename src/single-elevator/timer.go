package single_elevator

import (
	"time"
)

var timer_start = make(chan bool)

func Timer_main(timer_timeout chan<- bool){
	for {
		<- timer_start
		time.Sleep(3*time.Second)
		timer_timeout <- true
	}
}

func Timer_start(){
	timer_start <- true
}