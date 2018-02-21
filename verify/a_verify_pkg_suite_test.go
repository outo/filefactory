package verify_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestVerify(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Verify pkg Suite")
}
