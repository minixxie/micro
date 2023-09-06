package lib

import (
	"log"
	"net/http"
)

type VirtualHosts struct {
	Handlers map[string]http.Handler
	R301     map[string]string
}

func NewVirtualHosts() *VirtualHosts {
	vh := VirtualHosts{}
	vh.Handlers = make(map[string]http.Handler)
	vh.R301 = make(map[string]string)
	return &vh
}

// Implement the ServerHTTP method on our new type
func (vh VirtualHosts) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("HostSwitch: r.Host=", r.Host)
	log.Printf("vh.Handlers: %v", vh.Handlers)
	// log.Printf("HostSwitch: vh.R301=%v", vh.R301)
	// log.Printf("HostSwitch: r.Host redirect:", vh.R301[r.Host])
	if r301, ok := vh.R301[r.Host]; ok {
		// log.Printf("R301 to //" + r301)
		http.Redirect(w, r, "//"+r301, 301)
	} else {
		// Check if a http.Handler is registered for the given host.
		// If yes, use it to handle the request.
		if handler := vh.Handlers[r.Host]; handler != nil {
			handler.ServeHTTP(w, r)
		} else {
			// Handle host names for wich no handler is registered
			http.Error(w, "Forbidden", 403) // Or Redirect?
		}
	}
}
