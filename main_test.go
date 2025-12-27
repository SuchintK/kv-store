package main

import (
	"bufio"
	"net"
	"sync"
	"testing"
	"time"
)

func TestConcurrentClients(t *testing.T) {
	// Start server in goroutine
	go func() {
		listener, err := net.Listen("tcp", ":6380") // Use different port for testing
		if err != nil {
			t.Errorf("Failed to start test server: %v", err)
			return
		}
		defer listener.Close()

		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go handleConnection(conn)
		}
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Test with multiple concurrent clients
	numClients := 50
	var wg sync.WaitGroup
	errors := make(chan error, numClients)

	for i := 0; i < numClients; i++ {
		wg.Add(1)
		go func(clientID int) {
			defer wg.Done()

			// Connect to server
			conn, err := net.Dial("tcp", "localhost:6380")
			if err != nil {
				errors <- err
				return
			}
			defer conn.Close()

			// Send PING command
			_, err = conn.Write([]byte("*1\r\n$4\r\nPING\r\n"))
			if err != nil {
				errors <- err
				return
			}

			// Read response
			reader := bufio.NewReader(conn)
			response, err := reader.ReadString('\n')
			if err != nil {
				errors <- err
				return
			}

			// Verify response
			expected := "+PONG\r\n"
			if response != expected {
				t.Errorf("Client %d: expected %q, got %q", clientID, expected, response)
			}
		}(i)
	}

	// Wait for all clients to finish
	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		t.Errorf("Client error: %v", err)
	}
}

func TestConcurrentPingWithMessage(t *testing.T) {
	// Connect to server on test port
	time.Sleep(100 * time.Millisecond)

	numClients := 20
	var wg sync.WaitGroup
	results := make(chan bool, numClients)

	for i := 0; i < numClients; i++ {
		wg.Add(1)
		go func(clientID int) {
			defer wg.Done()

			conn, err := net.Dial("tcp", "localhost:6380")
			if err != nil {
				results <- false
				return
			}
			defer conn.Close()

			// Send PING with custom message
			cmd := "*2\r\n$4\r\nPING\r\n$5\r\nhello\r\n"
			_, err = conn.Write([]byte(cmd))
			if err != nil {
				results <- false
				return
			}

			// Read response (bulk string format)
			reader := bufio.NewReader(conn)

			// Read $5\r\n
			line1, _ := reader.ReadString('\n')
			// Read hello\r\n
			line2, _ := reader.ReadString('\n')

			if line1 == "$5\r\n" && line2 == "hello\r\n" {
				results <- true
			} else {
				results <- false
			}
		}(i)
	}

	wg.Wait()
	close(results)

	// Count successful responses
	successful := 0
	for result := range results {
		if result {
			successful++
		}
	}

	if successful != numClients {
		t.Errorf("Expected %d successful responses, got %d", numClients, successful)
	}
}
