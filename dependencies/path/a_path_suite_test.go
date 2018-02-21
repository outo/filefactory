package path_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestPath(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Path Suite")
}
