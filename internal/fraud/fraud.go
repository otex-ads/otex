package fraud

import (
	"net"
	"strings"
)

type FraudDetector struct {
	// In production, this would load from a database or file
	blockedASNs map[string]bool
	blockedIPs map[string]bool
}

func NewFraudDetector() *FraudDetector {
	return &FraudDetector{
		blockedASNs: make(map[string]bool),
		blockedIPs: make(map[string]bool),
	}
}

// AddBlockedASN adds an ASN to the blocklist (datacenter/hosting ranges)
func (f *FraudDetector) AddBlockedASN(asn string) {
	f.blockedASNs[asn] = true
}

// AddBlockedIP adds an IP or CIDR to the blocklist
func (f *FraudDetector) AddBlockedIP(ip string) {
	f.blockedIPs[ip] = true
}

// IsBlockedIP checks if an IP is blocked
func (f *FraudDetector) IsBlockedIP(ip string) bool {
	// Check exact match
	if f.blockedIPs[ip] {
		return true
	}

	// Check CIDR ranges (simplified - in production use proper CIDR matching)
	for blockedIP := range f.blockedIPs {
		if strings.Contains(blockedIP, "/") {
			_, ipNet, err := net.ParseCIDR(blockedIP)
			if err != nil {
				continue
			}
			if ipNet.Contains(net.ParseIP(ip)) {
				return true
			}
		}
	}

	return false
}

// IsDatacenterIP checks if IP belongs to known datacenter ranges
// In production, this would use a proper ASN lookup service
func (f *FraudDetector) IsDatacenterIP(ip string) bool {
	// Simplified: common datacenter IP ranges
	datacenterRanges := []string{
		"10.0.0.0/8",      // Private
		"172.16.0.0/12",   // Private
		"192.168.0.0/16",  // Private
		// Add known datacenter ASN ranges here
	}

	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}

	for _, cidr := range datacenterRanges {
		_, ipNet, err := net.ParseCIDR(cidr)
		if err != nil {
			continue
		}
		if ipNet.Contains(parsedIP) {
			return true
		}
	}

	return false
}

// CheckUserAgent checks for suspicious user agent patterns
func (f *FraudDetector) CheckUserAgent(ua string) bool {
	// Check for headless browser signatures
	suspiciousPatterns := []string{
		"HeadlessChrome",
		"PhantomJS",
		"SlimerJS",
		"Zombie.js",
		"python-requests",
		"curl",
		"wget",
	}

	uaLower := strings.ToLower(ua)
	for _, pattern := range suspiciousPatterns {
		if strings.Contains(uaLower, strings.ToLower(pattern)) {
			return true
		}
	}

	// Check for missing common headers would be done at request level
	return false
}

// RateLimiter handles rate limiting per IP
type RateLimiter struct {
	requests map[string][]int64 // IP -> timestamps
	maxRequests int
	windowSeconds int64
}

func NewRateLimiter(maxRequests int, windowSeconds int64) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]int64),
		maxRequests: maxRequests,
		windowSeconds: windowSeconds,
	}
}

// CheckRate checks if IP has exceeded rate limit
func (r *RateLimiter) CheckRate(ip string, now int64) bool {
	timestamps, exists := r.requests[ip]
	if !exists {
		r.requests[ip] = []int64{now}
		return true
	}

	// Remove timestamps outside the window
	cutoff := now - r.windowSeconds
	validTimestamps := []int64{}
	for _, ts := range timestamps {
		if ts > cutoff {
			validTimestamps = append(validTimestamps, ts)
		}
	}

	// Check if under limit
	if len(validTimestamps) < r.maxRequests {
		validTimestamps = append(validTimestamps, now)
		r.requests[ip] = validTimestamps
		return true
	}

	r.requests[ip] = validTimestamps
	return false
}

// ClearOldEntries removes old entries to prevent memory leak
func (r *RateLimiter) ClearOldEntries(now int64) {
	cutoff := now - r.windowSeconds
	for ip, timestamps := range r.requests {
		valid := []int64{}
		for _, ts := range timestamps {
			if ts > cutoff {
				valid = append(valid, ts)
			}
		}
		if len(valid) == 0 {
			delete(r.requests, ip)
		} else {
			r.requests[ip] = valid
		}
	}
}

// ClickValidator validates clicks based on timing patterns
type ClickValidator struct {
	// Track impression-to-click time per user
	clickTimes map[string][]int64 // userHash -> click timestamps
}

func NewClickValidator() *ClickValidator {
	return &ClickValidator{
		clickTimes: make(map[string][]int64),
	}
}

// ValidateClick checks if click timing is suspicious
func (cv *ClickValidator) ValidateClick(userHash string, impressionTime int64, clickTime int64) bool {
	timeDiff := clickTime - impressionTime
	
	// Clicks faster than 100ms are suspicious (bots)
	if timeDiff < 100 {
		return false
	}
	
	// Clicks slower than 5 minutes are suspicious (manual refresh)
	if timeDiff > 300000 {
		return false
	}
	
	// Check for rapid clicking pattern
	timestamps, exists := cv.clickTimes[userHash]
	if !exists {
		cv.clickTimes[userHash] = []int64{clickTime}
		return true
	}
	
	// Remove old timestamps (last hour)
	cutoff := clickTime - 3600000
	valid := []int64{}
	for _, ts := range timestamps {
		if ts > cutoff {
			valid = append(valid, ts)
		}
	}
	
	// Check if user clicked more than 10 times in last hour
	if len(valid) >= 10 {
		return false
	}
	
	valid = append(valid, clickTime)
	cv.clickTimes[userHash] = valid
	return true
}

// IPReputationChecker checks IP reputation
type IPReputationChecker struct {
	// In production, this would integrate with external services
	// like AbuseIPDB, IPQualityScore, etc.
	suspiciousIPs map[string]bool
}

func NewIPReputationChecker() *IPReputationChecker {
	return &IPReputationChecker{
		suspiciousIPs: make(map[string]bool),
	}
}

// CheckIPReputation checks if IP has suspicious reputation
func (irc *IPReputationChecker) CheckIPReputation(ip string) bool {
	// Check local blocklist
	if irc.suspiciousIPs[ip] {
		return false
	}
	
	// In production, call external API here
	// For now, return true (allow)
	return true
}

// AddSuspiciousIP adds IP to suspicious list
func (irc *IPReputationChecker) AddSuspiciousIP(ip string) {
	irc.suspiciousIPs[ip] = true
}
