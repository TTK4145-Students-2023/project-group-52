package network_control

import(
	"project/single-elevator/elevio"
	elev "project/single-elevator"
)

func Quick_fix(requests_chan chan<-[elev.N_FLOORS][elev.N_BUTTONS]bool){
	drv_buttons := make(chan elevio.ButtonEvent)
	go elevio.PollButtons(drv_buttons)

	requests := [elev.N_FLOORS][elev.N_BUTTONS]bool{}

	for {
		select {
		case btn := <-drv_buttons:
			requests[btn.Floor][btn.Button] = true
			requests_chan <- requests
			requests[btn.Floor][btn.Button] = false
		}
	}
}