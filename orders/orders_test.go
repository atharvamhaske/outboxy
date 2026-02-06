package main

import (
	"encoding/json"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Orders", func() {
	Describe("OrderEvent JSON structure", func() {
		It("marshals an OrderEvent with the expected fields", func() {
			id := uuid.New()
			event := OrderEvent{
				OrderID: id,
				Product: "Widget",
			}

			data, err := json.Marshal(event)
			Expect(err).NotTo(HaveOccurred())

			var decoded map[string]any
			Expect(json.Unmarshal(data, &decoded)).To(Succeed())

			Expect(decoded).To(HaveKeyWithValue("order_id", id.String()))
			Expect(decoded).To(HaveKeyWithValue("product", "Widget"))
		})
	})
})
