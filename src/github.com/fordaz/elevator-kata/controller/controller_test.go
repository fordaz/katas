package controller_test

import (
	"github.com/fordaz/elevator-kata/controller"
	"github.com/fordaz/elevator-kata/elevator"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Controller", func() {
	var (
		c controller.ElevatorController
	)
	BeforeEach(func() {
		c = controller.New(2, 10)
		c.Start()
	})
	AfterEach(func() {
		c.Shutdown()
	})

	Context("when single request is sent", func() {
		BeforeEach(func() {
			c.RequestElevator(2, elevator.Down)
		})
		It("sends an elevator to the requesting floor", func() {
			Expect(c.CheckVisitedFloor(2, 30)).To(BeTrue())
		})
	})

	Context("when two requests are sent", func() {
		BeforeEach(func() {
			c.RequestElevator(5, elevator.Down)
			c.RequestElevator(2, elevator.Down)
		})
		It("sends an elevator to the requesting floor", func() {
			Expect(c.CheckVisitedFloor(2, 30)).To(BeTrue())
			Expect(c.CheckVisitedFloor(5, 60)).To(BeTrue())
		})
	})

	Context("when roundtripping", func() {
		BeforeEach(func() {
			c.RequestElevator(5, elevator.Down)
			c.RequestElevator(2, elevator.Down)
		})
		It("sends an elevator to the requesting floor", func() {
			Expect(c.CheckVisitedFloor(2, 30)).To(BeTrue())
			Expect(c.CheckVisitedFloor(5, 60)).To(BeTrue())
			c.RequestElevator(3, elevator.Down)
			Expect(c.CheckVisitedFloor(3, 40)).To(BeTrue())
		})
	})

})
