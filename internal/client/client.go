package client

// ClientID uniquely identifies an AI coding client.
type ClientID string

const (
	Claude   ClientID = "claude"
	Gemini   ClientID = "gemini"
	Codex    ClientID = "codex"
	Copilot  ClientID = "copilot"
	Cursor   ClientID = "cursor"
	Windsurf ClientID = "windsurf"
)

// AllClientIDs lists all supported clients in display order.
var AllClientIDs = []ClientID{Claude, Gemini, Codex, Copilot, Cursor, Windsurf}

// Client represents a detected AI coding assistant.
type Client struct {
	ID              ClientID
	Name            string // human-readable name
	Detected        bool
	GlobalPath      string // resolved global install path
	ProjectPath     string // relative project install path template
	SupportsGlobal  bool
	SupportsProject bool
}

// Registry holds all known clients.
type Registry struct {
	clients map[ClientID]*Client
}

// NewRegistry creates a registry with all supported clients (not yet detected).
func NewRegistry() *Registry {
	return &Registry{
		clients: map[ClientID]*Client{
			Claude:   {ID: Claude, Name: "Claude Code", SupportsGlobal: true, SupportsProject: true},
			Gemini:   {ID: Gemini, Name: "Gemini CLI", SupportsGlobal: true, SupportsProject: true},
			Codex:    {ID: Codex, Name: "Codex CLI", SupportsGlobal: true, SupportsProject: true},
			Copilot:  {ID: Copilot, Name: "VS Code Copilot", SupportsGlobal: false, SupportsProject: true},
			Cursor:   {ID: Cursor, Name: "Cursor", SupportsGlobal: false, SupportsProject: true},
			Windsurf: {ID: Windsurf, Name: "Windsurf", SupportsGlobal: true, SupportsProject: true},
		},
	}
}

// Get returns a client by ID.
func (r *Registry) Get(id ClientID) *Client {
	return r.clients[id]
}

// All returns all clients in display order.
func (r *Registry) All() []*Client {
	result := make([]*Client, 0, len(AllClientIDs))
	for _, id := range AllClientIDs {
		result = append(result, r.clients[id])
	}
	return result
}

// Detected returns only clients that were detected on the system.
func (r *Registry) Detected() []*Client {
	var result []*Client
	for _, id := range AllClientIDs {
		c := r.clients[id]
		if c.Detected {
			result = append(result, c)
		}
	}
	return result
}

// ParseClientID parses a string to ClientID, returns empty string if invalid.
func ParseClientID(s string) ClientID {
	switch ClientID(s) {
	case Claude, Gemini, Codex, Copilot, Cursor, Windsurf:
		return ClientID(s)
	default:
		return ""
	}
}
