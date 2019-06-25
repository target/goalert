package main

var (
	nextPortCh = make(chan int)
	donePortCh = make(chan int)
	minPort    = 33000
	maxPort    = 34000
)

// NextPort will return the next available port.
func NextPort() int { return <-nextPortCh }

// DonePort will mark a port number as available for re-use.
func DonePort(p int) { donePortCh <- p }

func startPorts() {
	go func() {
		n := minPort
		avail := make([]int, 0, 100)
		var p int
		for {
			if len(avail) == 0 && n <= maxPort {
				avail = append(avail, n)
				n++
			}
			select {
			case nextPortCh <- avail[len(avail)-1]:
				avail = avail[:len(avail)-1]
			case p = <-donePortCh:
				avail = append(avail, p)
			}
		}
	}()
}
