package update

import (
	"context"

	"github.com/menghun3-cn/go-gui-template/internal/server"
)

// Service adapts the update package to the GUI server.
type Service struct {
	FeedURL string
}

func (s *Service) Check(current string) (string, bool, error) {
	rel, err := Latest(context.Background(), s.FeedURL)
	if err != nil {
		return "", true, err
	}
	if Newer(rel.Version, current) {
		return rel.Version, false, nil
	}
	return rel.Version, true, nil
}

var _ server.Updater = (*Service)(nil)

