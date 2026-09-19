package ingest

import "github.com/saeedbazmi/pharmacy/backend/internal/platform/textfa"

// ProductRef is the minimum identity matching needs.
type ProductRef struct {
	ID   int64
	Slug string
}

// Decision is the first-pass matching result. Fuzzy similarity is out of
// scope: M1 either links on a unique key, bootstraps a new product, or queues.
type Decision struct {
	Action    string // "link", "create", "queue"
	ProductID int64
	Reason    string
}

const (
	ActionLink   = "link"
	ActionCreate = "create"
	ActionQueue  = "queue"
)

// DecideMatch is a pure function: unique code beats exact normalised name.
// Speculative (fuzzy) merges are never produced.
func DecideMatch(item RawItem, byGTIN, byIRC, byName *ProductRef) Decision {
	if item.GTIN != "" && byGTIN != nil {
		return Decision{Action: ActionLink, ProductID: byGTIN.ID, Reason: "gtin"}
	}
	if item.IRC != "" && byIRC != nil {
		return Decision{Action: ActionLink, ProductID: byIRC.ID, Reason: "irc"}
	}
	if textfa.Normalize(item.NameFa) != "" && byName != nil {
		return Decision{Action: ActionLink, ProductID: byName.ID, Reason: "exact_name"}
	}
	if item.GTIN == "" && item.IRC == "" && byName == nil {
		// First observation of this item in a single-source catalogue.
		return Decision{Action: ActionCreate, Reason: "first_seen"}
	}
	return Decision{Action: ActionQueue, Reason: "unmatched"}
}
