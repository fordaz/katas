package controller

import (
	"fmt"
	"github.com/fordaz/elevator-kata/elevator"
	"time"
)

type ElevatorsGroup struct {
	numElevators           int
	numFloors              int
	elevators              []*elevator.Elevator
	picker                 ElevatorPicker
	queuedElevatorRequests chan ElevatorRequest
	shutdown               chan bool
}

type ElevatorController interface {
	RequestElevator(floor int, direction elevator.Direction)
	CheckVisitedFloor(expectedFloor int, timeout time.Duration) bool
	Start()
	Shutdown()
}

type ElevatorRequest struct {
	floor              int
	requestedDirection elevator.Direction
}

type ElevatorPicker interface {
	PickElevator(request ElevatorRequest, elevators []*elevator.Elevator) *elevator.Elevator
}

type DefaultElevatorPicker struct{}

func New(numElevators, numFloors int) *ElevatorsGroup {
	elevators := make([]*elevator.Elevator, numElevators)
	for i := 0; i < numElevators; i++ {
		elevators[i] = elevator.New(i, numFloors)
		elevators[i].Start()
	}
	return &ElevatorsGroup{
		numElevators:           numElevators,
		numFloors:              numFloors,
		elevators:              elevators,
		picker:                 DefaultElevatorPicker{},
		queuedElevatorRequests: make(chan ElevatorRequest),
		shutdown:               make(chan bool),
	}
}

func (elevatorGroup *ElevatorsGroup) Start() {
	go func() {
		for {
			select {

			case r := <-elevatorGroup.queuedElevatorRequests:

				pickedElevator := elevatorGroup.picker.PickElevator(r, elevatorGroup.elevators)

				pickedElevator.RequestDestinationFloor(r.floor)

			case <-elevatorGroup.shutdown:
				return
			}
		}
	}()
}

func (elevatorGroup *ElevatorsGroup) Shutdown() {
	for _, e := range elevatorGroup.elevators {
		e.Shutdown()
	}
}

func (elevatorGroup *ElevatorsGroup) RequestElevator(floor int, requestedDirection elevator.Direction) {

	request := ElevatorRequest{floor, requestedDirection}

	elevatorGroup.queuedElevatorRequests <- request

}

func (elevatorGroup *ElevatorsGroup) CheckVisitedFloor(expectedFloor int, timeoutMillis time.Duration) bool {

	wasFloorVisited := make(chan bool)

	for _, e := range elevatorGroup.elevators {
		go func(wasFloorVisited chan bool, e1 *elevator.Elevator) {
			select {

			case visitedFloor := <-e1.GetVisitedFloors():
				if visitedFloor == expectedFloor {
					wasFloorVisited <- true
				}

			case <-time.After(time.Millisecond * timeoutMillis):
			}

		}(wasFloorVisited, e)
	}
	visited := false
	select {
	case <-wasFloorVisited:
		visited = true
	case <-time.After(time.Millisecond * (timeoutMillis + 100)):
	}

	return visited

}

func (p DefaultElevatorPicker) PickElevator(request ElevatorRequest, elevators []*elevator.Elevator) *elevator.Elevator {

	floor := request.floor
	requestedDirection := request.requestedDirection

	elevatorRanks := make([]int, len(elevators))

	for i, e := range elevators {

		currentFloor, state, direction := e.GetLocation()

		elevatorMoving := (state == elevator.Moving)

		if elevatorMoving {
			elevatorRanks[i] = elevatorRanks[i] + 1
		}

		floorGap := (currentFloor - floor)
		onSameDirection := requestedDirection == direction
		elevatorAbove := (floorGap >= 1)
		elevatorGoingDown := direction == elevator.Down

		if elevatorMoving && elevatorAbove &&
			onSameDirection && elevatorGoingDown {
			elevatorRanks[i] = elevatorRanks[i] + floorGap
		}

		if elevatorMoving && !elevatorAbove &&
			onSameDirection && !elevatorGoingDown {
			elevatorRanks[i] = elevatorRanks[i] + -1*floorGap
		}

	}

	maxRankedElevator := -1
	maxRanking := -1
	for i, rank := range elevatorRanks {
		if rank > maxRanking {
			maxRankedElevator = i
			maxRanking = rank
		}
	}

	if maxRanking == -1 {
		maxRankedElevator = 0
	}

	fmt.Printf("picking elevator %d, floor %d\n", maxRankedElevator, floor)
	return elevators[maxRankedElevator]
}
