package bookmarks

import (
	"crypto/md5"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"dhiarhome/internal/config"
)

// DisplayLink represents a processed bookmark link ready for rendering.
type DisplayLink struct {
	Name        string
	URL         string
	Icon        string // Resolved icon: SVG path, image URL, or favicon URL
	IconType    string // "svg", "image", or "favicon"
	Description string
	NewTab      bool
}

// DisplayGroup represents a processed bookmark group.
type DisplayGroup struct {
	Group string
	Links []DisplayLink
}

// Store manages bookmarks and favicon caching.
type Store struct {
	mu         sync.RWMutex
	groups     []DisplayGroup
	cacheDir   string
	httpClient *http.Client
}

// NewStore creates a new bookmarks store from config.
func NewStore(bookmarkGroups []config.BookmarkGroup, cacheDir string) *Store {
	s := &Store{
		cacheDir: cacheDir,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}

	// Ensure cache directory exists
	if cacheDir != "" {
		os.MkdirAll(cacheDir, 0755)
	}

	// Process bookmark groups
	s.groups = s.processGroups(bookmarkGroups)

	return s
}

// GetGroups returns all bookmark groups.
func (s *Store) GetGroups() []DisplayGroup {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.groups
}

// processGroups converts config bookmark groups to display groups.
func (s *Store) processGroups(groups []config.BookmarkGroup) []DisplayGroup {
	var result []DisplayGroup

	for _, g := range groups {
		dg := DisplayGroup{Group: g.Group}

		for _, link := range g.Links {
			dl := DisplayLink{
				Name:        link.Name,
				URL:         link.URL,
				Description: link.Description,
				NewTab:      link.NewTab,
			}

			// Resolve icon
			dl.Icon, dl.IconType = s.resolveIcon(link.Icon, link.URL)

			dg.Links = append(dg.Links, dl)
		}

		result = append(result, dg)
	}

	return result
}

// resolveIcon determines the icon type and resolves it.
func (s *Store) resolveIcon(icon string, linkURL string) (string, string) {
	// Empty icon: use favicon
	if icon == "" || icon == "favicon" {
		return s.fetchFavicon(linkURL)
	}

	// Check if it's a file path (starts with / or ./)
	if strings.HasPrefix(icon, "/") || strings.HasPrefix(icon, "./") {
		return icon, "image"
	}

	// Check if it's a known Lucide icon name (return as-is, template will render SVG)
	if isLucideIcon(icon) {
		return icon, "svg"
	}

	// Fallback to favicon
	return s.fetchFavicon(linkURL)
}

// fetchFavicon fetches and caches the favicon for a URL.
func (s *Store) fetchFavicon(linkURL string) (string, string) {
	if linkURL == "" {
		return "", "none"
	}

	// Parse the URL to get the base URL
	parsed, err := url.Parse(linkURL)
	if err != nil {
		return "", "none"
	}

	baseURL := fmt.Sprintf("%s://%s", parsed.Scheme, parsed.Host)
	faviconURL := baseURL + "/favicon.ico"

	// Generate cache filename from URL hash
	hash := md5.Sum([]byte(baseURL))
	cacheFile := fmt.Sprintf("%x.ico", hash)
	cachePath := filepath.Join(s.cacheDir, cacheFile)

	// Check if cached file exists
	if s.cacheDir != "" {
		if _, err := os.Stat(cachePath); err == nil {
			return "/bookmarks/icons/" + cacheFile, "favicon"
		}

		// Fetch and cache favicon
		go s.fetchAndCache(faviconURL, cachePath)
	}

	// Return the favicon URL (will be fetched asynchronously)
	return faviconURL, "favicon"
}

// fetchAndCache downloads the favicon and saves it to the cache.
func (s *Store) fetchAndCache(faviconURL, cachePath string) {
	// Validate URL to prevent SSRF attacks
	parsed, err := url.Parse(faviconURL)
	if err != nil {
		return
	}

	// Only allow http/https schemes
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return
	}

	// Resolve hostname to check for private/internal IPs
	host := parsed.Hostname()
	addrs, err := net.LookupHost(host)
	if err != nil {
		return
	}
	for _, addr := range addrs {
		ip := net.ParseIP(addr)
		if ip != nil && isPrivateIP(ip) {
			return // Block requests to private/internal networks
		}
	}

	resp, err := s.httpClient.Get(faviconURL)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return
	}

	// Limit size to 1MB
	limitedReader := io.LimitReader(resp.Body, 1024*1024)

	data, err := io.ReadAll(limitedReader)
	if err != nil {
		return
	}

	os.WriteFile(cachePath, data, 0644)
}

// isPrivateIP checks if an IP address is private, loopback, or link-local.
func isPrivateIP(ip net.IP) bool {
	privateRanges := []string{
		"127.0.0.0/8",    // Loopback
		"10.0.0.0/8",     // Private
		"172.16.0.0/12",  // Private
		"192.168.0.0/16", // Private
		"169.254.0.0/16", // Link-local (AWS metadata, etc.)
		"::1/128",        // IPv6 loopback
		"fc00::/7",       // IPv6 unique local
		"fe80::/10",      // IPv6 link-local
	}
	for _, cidr := range privateRanges {
		_, network, err := net.ParseCIDR(cidr)
		if err != nil {
			continue
		}
		if network.Contains(ip) {
			return true
		}
	}
	return false
}

// isLucideIcon checks if the icon name is a known Lucide icon.
// This is a simple check for common icons; expand as needed.
func isLucideIcon(name string) bool {
	commonIcons := []string{
		"globe", "server", "monitor", "cpu", "hard-drive", "database",
		"cloud", "shield", "lock", "key", "settings", "home",
		"play-circle", "tv", "film", "music", "book", "file",
		"folder", "inbox", "mail", "message-circle", "bell",
		"calendar", "clock", "map", "navigation", "compass",
		"box", "package", "archive", "download", "upload",
		"link", "external-link", "wifi", "bluetooth", "radio",
		"container", "layers", "grid", "layout", "sidebar",
		"terminal", "code", "git-branch", "github", "docker",
	}

	name = strings.ToLower(name)
	for _, icon := range commonIcons {
		if icon == name {
			return true
		}
	}
	return false
}
