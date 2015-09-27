package elevator_test

import (
	"github.com/fordaz/elevator-kata/elevator"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"time"
)

var _ = Describe("Elevator", func() {
	Describe("RequestDestinationFloor", func() {
		var (
			ele1 *elevator.Elevator
		)
		BeforeEach(func() {
			ele1 = elevator.New(1, 10)
			ele1.Start()
		})

		AfterEach(func() {
			ele1.Shutdown()
		})

		checkVisitedFloor := func(expectedFloor int, e1 *elevator.Elevator, timeoutMillis time.Duration) {
			select {
			case visitedFloor := <-e1.GetVisitedFloors():
				Expect(visitedFloor).To(Equal(expectedFloor))
			case <-time.After(time.Millisecond * timeoutMillis):
			}
		}

		Context("with no other floor request", func() {
			It("eventually stops in the requested floor", func() {
				accepted := ele1.RequestDestinationFloor(1)

				Expect(accepted).Should(Equal(true))

				checkVisitedFloor(1, ele1, 20)
			})
		})
		Context("with two subsequent floor requests", func() {
			It("eventually stops in each requested floor", func() {
				accepted := ele1.RequestDestinationFloor(3)
				Expect(accepted).Should(Equal(true))

				accepted = ele1.RequestDestinationFloor(6)
				Expect(accepted).Should(Equal(true))

				checkVisitedFloor(3, ele1, 40)
				checkVisitedFloor(6, ele1, 40)
			})
		})
		Context("with two subsequent floor requests", func() {
			It("eventually stops in each requested floor", func() {
				accepted := ele1.RequestDestinationFloor(3)
				Expect(accepted).Should(Equal(true))

				accepted = ele1.RequestDestinationFloor(6)
				Expect(accepted).Should(Equal(true))

				checkVisitedFloor(3, ele1, 40)

				accepted = ele1.RequestDestinationFloor(2)
				Expect(accepted).Should(Equal(false))

				checkVisitedFloor(6, ele1, 40)

			})
		})
	})
})
