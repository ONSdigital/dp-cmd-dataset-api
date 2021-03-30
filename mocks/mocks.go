// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package mocks

import (
	"context"
	"sync"
)

var (
	lockDownloadsGeneratorMockGenerate sync.RWMutex
)

// DownloadsGeneratorMock is a mock implementation of api.DownloadsGenerator.
//
//     func TestSomethingThatUsesDownloadsGenerator(t *testing.T) {
//
//         // make and configure a mocked api.DownloadsGenerator
//         mockedDownloadsGenerator := &DownloadsGeneratorMock{
//             GenerateFunc: func(ctx context.Context, datasetID string, instanceID string, edition string, version string) error {
// 	               panic("mock out the Generate method")
//             },
//         }
//
//         // use mockedDownloadsGenerator in code that requires api.DownloadsGenerator
//         // and then make assertions.
//
//     }
type DownloadsGeneratorMock struct {
	// GenerateFunc mocks the Generate method.
	GenerateFunc func(ctx context.Context, datasetID string, instanceID string, edition string, version string) error

	// calls tracks calls to the methods.
	calls struct {
		// Generate holds details about calls to the Generate method.
		Generate []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// DatasetID is the datasetID argument value.
			DatasetID string
			// InstanceID is the instanceID argument value.
			InstanceID string
			// Edition is the edition argument value.
			Edition string
			// Version is the version argument value.
			Version string
		}
	}
}

// Generate calls GenerateFunc.
func (mock *DownloadsGeneratorMock) Generate(ctx context.Context, datasetID string, instanceID string, edition string, version string) error {
	if mock.GenerateFunc == nil {
		panic("DownloadsGeneratorMock.GenerateFunc: method is nil but DownloadsGenerator.Generate was just called")
	}
	callInfo := struct {
		Ctx        context.Context
		DatasetID  string
		InstanceID string
		Edition    string
		Version    string
	}{
		Ctx:        ctx,
		DatasetID:  datasetID,
		InstanceID: instanceID,
		Edition:    edition,
		Version:    version,
	}
	lockDownloadsGeneratorMockGenerate.Lock()
	mock.calls.Generate = append(mock.calls.Generate, callInfo)
	lockDownloadsGeneratorMockGenerate.Unlock()
	return mock.GenerateFunc(ctx, datasetID, instanceID, edition, version)
}

// GenerateCalls gets all the calls that were made to Generate.
// Check the length with:
//     len(mockedDownloadsGenerator.GenerateCalls())
func (mock *DownloadsGeneratorMock) GenerateCalls() []struct {
	Ctx        context.Context
	DatasetID  string
	InstanceID string
	Edition    string
	Version    string
} {
	var calls []struct {
		Ctx        context.Context
		DatasetID  string
		InstanceID string
		Edition    string
		Version    string
	}
	lockDownloadsGeneratorMockGenerate.RLock()
	calls = mock.calls.Generate
	lockDownloadsGeneratorMockGenerate.RUnlock()
	return calls
}
