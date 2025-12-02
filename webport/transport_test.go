package webport

import (
	"io"
	"net"
	"testing"
	"time"
)

func TestPumpBiDirection(t *testing.T) {
	// Create a pair of in-memory connections
	localServer, localClient := net.Pipe()
	remoteServer, remoteClient := net.Pipe()

	done := make(chan struct{})

	// Start PumpBiDirectional in a goroutine
	go PumpBiDirectional(localClient, remoteClient, done)

	// Test: local -> remote
	go func() {
		message := []byte("Hello from local")
		_, err := localServer.Write(message)
		if err != nil {
			t.Errorf("Failed to write to local: %v", err)
		}
	}()

	// Read from remote side
	buffer := make([]byte, 1024)
	remoteServer.SetReadDeadline(time.Now().Add(2 * time.Second))
	n, err := remoteServer.Read(buffer)
	if err != nil {
		t.Fatalf("Failed to read from remote: %v", err)
	}
	if string(buffer[:n]) != "Hello from local" {
		t.Errorf("Expected 'Hello from local', got '%s'", string(buffer[:n]))
	}

	// Test: remote -> local
	go func() {
		message := []byte("Hello from remote")
		_, err := remoteServer.Write(message)
		if err != nil {
			t.Errorf("Failed to write to remote: %v", err)
		}
	}()

	// Read from local side
	localServer.SetReadDeadline(time.Now().Add(2 * time.Second))
	n, err = localServer.Read(buffer)
	if err != nil {
		t.Fatalf("Failed to read from local: %v", err)
	}
	if string(buffer[:n]) != "Hello from remote" {
		t.Errorf("Expected 'Hello from remote', got '%s'", string(buffer[:n]))
	}

	// Test: graceful shutdown via done channel
	close(done)
	time.Sleep(100 * time.Millisecond) // Give goroutines time to finish

	// Verify connections are closed
	localServer.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
	_, err = localServer.Read(buffer)
	if err != io.EOF && !isClosedError(err) {
		t.Errorf("Expected connection to be closed, got error: %v", err)
	}

	remoteServer.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
	_, err = remoteServer.Read(buffer)
	if err != io.EOF && !isClosedError(err) {
		t.Errorf("Expected connection to be closed, got error: %v", err)
	}
}

func TestOrChans(t *testing.T) {
	tests := []struct {
		name        string
		numChans    int
		closeIndex  int // which channel to close (0-indexed)
		expectClose bool
	}{
		{
			name:        "no channels",
			numChans:    0,
			closeIndex:  -1,
			expectClose: false,
		},
		{
			name:        "single channel closed",
			numChans:    1,
			closeIndex:  0,
			expectClose: true,
		},
		{
			name:        "two channels, first closed",
			numChans:    2,
			closeIndex:  0,
			expectClose: true,
		},
		{
			name:        "two channels, second closed",
			numChans:    2,
			closeIndex:  1,
			expectClose: true,
		},
		{
			name:        "multiple channels, middle one closed",
			numChans:    5,
			closeIndex:  2,
			expectClose: true,
		},
		{
			name:        "multiple channels, last one closed",
			numChans:    5,
			closeIndex:  4,
			expectClose: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.numChans == 0 {
				result := OrChans[struct{}]()
				if result != nil {
					t.Errorf("Expected nil for zero channels, got non-nil")
				}
				return
			}

			// Create channels
			chans := make([]chan struct{}, tt.numChans)
			for i := 0; i < tt.numChans; i++ {
				chans[i] = make(chan struct{})
			}

			// Call OrChans
			subChans := make([]<-chan struct{}, tt.numChans)
			for i, ch := range chans {
				subChans[i] = ch
			}
			orDone := OrChans(subChans...)

			// Close the specified channel after a short delay
			go func() {
				time.Sleep(50 * time.Millisecond)
				close(chans[tt.closeIndex])
			}()

			// Wait for orDone to close with timeout
			select {
			case <-orDone:
				if !tt.expectClose {
					t.Errorf("Expected orDone to remain open, but it closed")
				}
			case <-time.After(200 * time.Millisecond):
				if tt.expectClose {
					t.Errorf("Expected orDone to close, but it remained open")
				}
			}
		})
	}
}

// Helper function to check if error is a closed connection error
func isClosedError(err error) bool {
	if err == nil {
		return false
	}
	return err.Error() == "io: read/write on closed pipe" ||
		err.Error() == "use of closed network connection"
}
