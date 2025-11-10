package blocklist

import "github.com/prguard/prguard/pkg/models"

// BlocklistManager defines the interface for managing the blocklist
type BlocklistManager interface {
	// Block operations
	Block(username, reason, evidenceURL, blockedBy, severity, source string) (*models.BlocklistEntry, error)
	Unblock(username string) error
	IsBlocked(username string) (bool, error)

	// Query operations
	List() ([]*models.BlocklistEntry, error)
	GetByUsername(username string) ([]*models.BlocklistEntry, error)

	// Import/Export operations
	ExportJSON(path string) error
	ExportCSV(path string) error
	ImportJSON(path string) (int, error)
	ImportJSONFromURL(url string) (int, error)
}
