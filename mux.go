// Package pat implements a simple regexp URL pattern muxer
package pat

import (
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

type PatternServeMux struct {
	handlers map[string][]*patHandler
}

// New returns a new PatternServeMux.
func New() *PatternServeMux {
	return &PatternServeMux{make(map[string][]*patHandler)}
}

// ServeHTTP matches r.URL.Path against its routing table using the rules described above.
func (p *PatternServeMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, ph := range p.handlers[r.Method] {
		if params, ok := ph.try(r.URL.Path); ok {
			if len(params) > 0 {
				r.URL.RawQuery = url.Values(params).Encode() + "&" + r.URL.RawQuery
			}
			ph.ServeHTTP(w, r)
			return
		}
	}

	allowed := make([]string, 0, len(p.handlers))
	for meth, handlers := range p.handlers {
		if meth == r.Method {
			continue
		}

		for _, ph := range handlers {
			if _, ok := ph.try(r.URL.Path); ok {
				allowed = append(allowed, meth)
			}
		}
	}

	if len(allowed) == 0 {
		http.NotFound(w, r)
		return
	}

	w.Header().Add("Allow", strings.Join(allowed, ", "))
	http.Error(w, "Method Not Allowed", 405)
}

// Head will register a pattern with a handler for HEAD requests.
func (p *PatternServeMux) Head(pat string, h http.Handler) {
	p.Add("HEAD", pat, h)
}

// Get will register a pattern with a handler for GET requests.
// It also registers pat for HEAD requests. If this needs to be overridden, use
// Head before Get with pat.
func (p *PatternServeMux) Get(pat string, h http.Handler) {
	p.Add("HEAD", pat, h)
	p.Add("GET", pat, h)
}

// Post will register a pattern with a handler for POST requests.
func (p *PatternServeMux) Post(pat string, h http.Handler) {
	p.Add("POST", pat, h)
}

// Put will register a pattern with a handler for PUT requests.
func (p *PatternServeMux) Put(pat string, h http.Handler) {
	p.Add("PUT", pat, h)
}

// Del will register a pattern with a handler for DELETE requests.
func (p *PatternServeMux) Del(pat string, h http.Handler) {
	p.Add("DELETE", pat, h)
}

// Options will register a pattern with a handler for OPTIONS requests.
func (p *PatternServeMux) Options(pat string, h http.Handler) {
	p.Add("OPTIONS", pat, h)
}

// Add will register a pattern with a handler for meth requests.
func (p *PatternServeMux) Add(meth, pat string, h http.Handler) {
	p.handlers[meth] = append(p.handlers[meth], &patHandler{regexp.MustCompile(pat), h})
}

type patHandler struct {
	pat *regexp.Regexp
	http.Handler
}

func (ph *patHandler) try(path string) (url.Values, bool) {
	p := make(url.Values)
	matches := ph.pat.FindStringSubmatch(path)
	if matches != nil {
		for i, name := range ph.pat.SubexpNames() {
			if name != "" {
				p.Add(":"+name, matches[i])
			}
		}
		return p, true
	}
	return nil, false
}
