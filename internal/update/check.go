package update

import (
	"context"
	"fmt"
)

// CheckAndPrint fetches the feed and prints whether the current build is up
// to date. Used by the CLI to surface the same logic as the GUI.
func CheckAndPrint(current, feed string) error {
	s := &Service{FeedURL: feed}
	v, ok, err := s.Check(current)
	if err != nil {
		return err
	}
	if ok {
		fmt.Printf("up to date: %s\n", current)
		return nil
	}
	fmt.Printf("new version available: %s (current %s)\n", v, current)
	return nil
}

func LatestVersion(ctx context.Context, feed string) (string, error) {
	rel, err := Latest(ctx, feed)
	if err != nil {
		return "", err
	}
	return rel.Version, nil
}
