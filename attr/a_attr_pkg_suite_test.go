package attr_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestAttr(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Attr pkg Suite")
}
