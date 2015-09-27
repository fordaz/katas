package elevator

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type Direction string
type State string
type FloorCommand string

const (
	Up   Direction = "Up"
	Down Direction = "Down"
)
const (
	Stopped      State = "Stopped"
	Hold         State = "Hold"
	Panic        State = "Panic"
	OutOfService State = "OutOfService"
	Moving       State = "Moving"
)

type ElevatorHandler interface {
	RequestDestinationFloor(destinationFloor int) bool
	GetVisitedFloors() chan int32
	GetLocation() (int, State, Direction)
	Start()
	Shutdown()
}

type Elevator struct {
	Id           int
	CurrentFloor *int32
	Direction    Direction
	State        State
	numFloors    int

	// a value of 1 in element i-th indicates i-th floor have been scheduled to be visited
	scheduledStops []int
	// number of sheduled stops
	numScheduledStops int

	// synchronize access to scheduledStops, numScheduledStops
	animatingRW sync.RWMutex

	shutdown       chan bool
	startAnimation chan bool
	visitedFloors  chan int
}

func New(id int, numFloors int) *Elevator {
	ground := int32(0)
	return &Elevator{
		Id:                id,
		numFloors:         numFloors,
		CurrentFloor:      &ground,
		Direction:         Up,
		State:             Stopped,
		scheduledStops:    make([]int, numFloors),
		numScheduledStops: 0,
		shutdown:          make(chan bool),
		startAnimation:    make(chan bool),
		visitedFloors:     make(chan int),
	}
}

func (elevator *Elevator) RequestDestinationFloor(destinationFloor int) bool {
	if elevator.canTakeRequest(destinationFloor) {
		elevator.scheduleStop(destinationFloor)
		return true
	}
	return false
}

func (elevator *Elevator) GetLocation() (int, State, Direction) {
	elevator.animatingRW.Lock()
	defer elevator.animatingRW.Unlock()
	currentFloor := atomic.LoadInt32(elevator.CurrentFloor)
	return int(currentFloor), elevator.State, elevator.Direction
}

func (elevator *Elevator) Start() {
	go func(elevator *Elevator) {
		for {
			select {
			case <-elevator.startAnimation:
				go elevator.animate()
			case <-elevator.shutdown:
				fmt.Printf("Shutting down elevator %d\n", elevator.Id)
				return
			}
		}
	}(elevator)
}

func (elevator *Elevator) GetVisitedFloors() chan int {
	return elevator.visitedFloors
}

func (elevator *Elevator) canTakeRequest(destinationFloor int) bool {
	if elevator.State == Stopped {
		return true
	}
	currentFloor := int(atomic.LoadInt32(elevator.CurrentFloor))
	if elevator.State == Moving && elevator.Direction == Up {
		return currentFloor+1 <= destinationFloor
	}
	if elevator.State == Moving && elevator.Direction == Down {
		return currentFloor-1 >= destinationFloor
	}
	return false
}

func (elevator *Elevator) scheduleStop(floor int) {
	elevator.animatingRW.Lock()
	defer elevator.animatingRW.Unlock()

	fmt.Printf("Scheduled stop on floor %d\n", floor)
	elevator.scheduledStops[floor] = 1
	elevator.numScheduledStops = elevator.numScheduledStops + 1

	if elevator.numScheduledStops == 1 {
		elevator.State = Moving
		if int32(floor) >= atomic.LoadInt32(elevator.CurrentFloor) {
			elevator.Direction = Up
		} else {
			elevator.Direction = Down
		}
		elevator.startAnimation <- true
	}
}

func (elevator *Elevator) stopElevator() {
	elevator.animatingRW.Lock()
	defer elevator.animatingRW.Unlock()

	// any pending scheduled stop are cleared out
	elevator.State = Stopped
	elevator.numScheduledStops = 0

	for i := 0; i < len(elevator.scheduledStops); i++ {
		elevator.scheduledStops[i] = 0
	}
}

func (elevator *Elevator) descheduleStop(floor int32) {
	elevator.animatingRW.Lock()
	defer elevator.animatingRW.Unlock()
	elevator.numScheduledStops = elevator.numScheduledStops - 1
	elevator.scheduledStops[floor] = 0
	elevator.visitedFloors <- int(floor)
}

func (elevator *Elevator) getNumScheduledStops() int {
	elevator.animatingRW.Lock()
	defer elevator.animatingRW.Unlock()
	return elevator.numScheduledStops
}

func (elevator *Elevator) Shutdown() {
	fmt.Printf("Shutting down elevator %d\n", elevator.Id)
	elevator.shutdown <- true
}

func (elevator *Elevator) animate() {
	i := 0
	for elevator.getNumScheduledStops() > 0 && i < 100000 {
		fmt.Printf("  Moving %s, current floor %d\n", elevator.Direction, *elevator.CurrentFloor)
		if elevator.Direction == Up {
			if atomic.LoadInt32(elevator.CurrentFloor) < int32(elevator.numFloors-1) {
				atomic.AddInt32(elevator.CurrentFloor, 1)
				time.Sleep(10 * time.Millisecond)
			}
		}
		if elevator.Direction == Down {
			if atomic.LoadInt32(elevator.CurrentFloor) > int32(0) {
				atomic.AddInt32(elevator.CurrentFloor, -1)
				time.Sleep(10 * time.Millisecond)
			}
		}
		if elevator.scheduledStops[*elevator.CurrentFloor] == 1 {
			fmt.Printf("Visiting floor %d\n", *elevator.CurrentFloor)
			elevator.descheduleStop(*elevator.CurrentFloor)
		}
		i = i + 1
	}
	if i >= 100000 {
		panic("Something is really wrong, elevator went cuckoo")
	}

	elevator.stopElevator()
}
