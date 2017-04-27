package coreaffinity

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCoreAffinityIsolator_Name(t *testing.T) {
	testCases := []string{
		"foo",
		"bar",
		"⛱⛱⛱⛱",
		"\x00\xA0\x00\xA0\x00\xA0",
		"\xbd\xb2\x3d\xbc\x20\xe2\x8c\x98",
	}
	for _, testCase := range testCases {
		cfa, err := New(testCase)
		assert.Nil(t, err, fmt.Sprintf("Cannot create coreAffinityIsolator: %q", err.Error()))
		name := cfa.Name()
		assert.Equal(t, testCase, name, fmt.Sprintf("Isolator name %q is not as expected name %q", name, testCase))
	}
}

type mockedOpaque struct {
	mock.Mock
}

func (m *mockedOpaque) AdvertiseOpaqueResource(name string, value int) error {
	args := m.Called(name, value)
	return args.Error(0)
}

func TestCoreAffinityIsolator_ShutDown(t *testing.T) {
	testObj :=new(mockedOpaque)
}
