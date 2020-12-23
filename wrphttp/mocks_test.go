package wrphttp

import (
	"github.com/stretchr/testify/mock"
)

type mockReadCloser struct {
	mock.Mock
}

func (m *mockReadCloser) Read(p []byte) (int, error) {
	arguments := m.Called(p)
	return arguments.Int(0), arguments.Error(1)
}

func (m *mockReadCloser) Close() error {
	return m.Called().Error(0)
}
