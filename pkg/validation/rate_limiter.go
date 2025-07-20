package validation

import (
	"sync"
	"time"
)

// RateLimiter implements a token bucket rate limiter per client
type RateLimiter struct {
	maxRequests int
	window      time.Duration
	clients     map[string]*clientLimiter
	mu          sync.RWMutex
	cleanupTick *time.Ticker
	done        chan struct{}
}

// clientLimiter tracks rate limiting state for a single client
type clientLimiter struct {
	tokens     int
	lastRefill time.Time
	maxTokens  int
	window     time.Duration
	mu         sync.Mutex
}

// NewRateLimiter creates a new rate limiter with specified limits
func NewRateLimiter(maxRequests int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		maxRequests: maxRequests,
		window:      window,
		clients:     make(map[string]*clientLimiter),
		done:        make(chan struct{}),
	}

	// Start cleanup goroutine to remove inactive clients
	rl.cleanupTick = time.NewTicker(window)
	go rl.cleanup()

	return rl
}

// Allow checks if a request should be allowed for the given client ID
func (rl *RateLimiter) Allow(clientID string) bool {
	rl.mu.RLock()
	limiter, exists := rl.clients[clientID]
	rl.mu.RUnlock()

	if !exists {
		// Create new client limiter
		limiter = &clientLimiter{
			tokens:     rl.maxRequests,
			lastRefill: time.Now(),
			maxTokens:  rl.maxRequests,
			window:     rl.window,
		}
		rl.mu.Lock()
		rl.clients[clientID] = limiter
		rl.mu.Unlock()
	}

	return limiter.consume()
}

// consume attempts to consume a token from the client's bucket
func (cl *clientLimiter) consume() bool {
	cl.mu.Lock()
	defer cl.mu.Unlock()

	now := time.Now()

	// Calculate how many tokens to refill based on elapsed time
	elapsed := now.Sub(cl.lastRefill)
	if elapsed > 0 && cl.tokens < cl.maxTokens {
		// Calculate the fraction of the window that has passed
		windowsPassed := float64(elapsed) / float64(cl.window)
		tokensToAdd := int(float64(cl.maxTokens) * windowsPassed)

		if tokensToAdd > 0 {
			cl.tokens += tokensToAdd
			if cl.tokens > cl.maxTokens {
				cl.tokens = cl.maxTokens
			}
			cl.lastRefill = now
		}
	}

	// Check if we have tokens available
	if cl.tokens > 0 {
		cl.tokens--
		return true
	}

	return false
}

// cleanup removes inactive clients to prevent memory leaks
func (rl *RateLimiter) cleanup() {
	for {
		select {
		case <-rl.cleanupTick.C:
			rl.removeInactiveClients()
		case <-rl.done:
			return
		}
	}
}

// removeInactiveClients removes clients that haven't been active for 2 windows
func (rl *RateLimiter) removeInactiveClients() {
	cutoff := time.Now().Add(-2 * rl.window)

	rl.mu.Lock()
	for clientID, limiter := range rl.clients {
		limiter.mu.Lock()
		if limiter.lastRefill.Before(cutoff) {
			delete(rl.clients, clientID)
		}
		limiter.mu.Unlock()
	}
	rl.mu.Unlock()
}

// Close stops the rate limiter and cleans up resources
func (rl *RateLimiter) Close() {
	close(rl.done)
	rl.cleanupTick.Stop()
}
