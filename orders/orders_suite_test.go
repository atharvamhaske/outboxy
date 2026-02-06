package main

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestOrders(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Orders Suite")
}
