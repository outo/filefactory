package def_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestDef(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Def pkg Suite")
}
