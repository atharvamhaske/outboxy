package main

import (
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Dispatcher", func() {
	Describe("OutboxMsg", func() {
		It("holds ID, Topic and Message fields", func() {
			id := uuid.New()
			msg := OutboxMsg{
				ID:      id,
				Topic:   "orders.created",
				Message: []byte(`{"foo":"bar"}`),
			}

			Expect(msg.ID).To(Equal(id))
			Expect(msg.Topic).To(Equal("orders.created"))
			Expect(msg.Message).To(Equal([]byte(`{"foo":"bar"}`)))
		})
	})
})
