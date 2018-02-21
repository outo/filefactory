package filefactory_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestFileFactory(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "FileFactory pkg Suite")
}
