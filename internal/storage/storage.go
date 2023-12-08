package storage

import (
	"github.com/ilnsm/mcollector/internal/server/transport"
	memorystorage "github.com/ilnsm/mcollector/internal/storage/memory"
)

func New(FileStoragePath string) (transport.Storage, error) {
	if FileStoragePath != "" {
		s, err := memorystorage.New()
		if err != nil {
			return nil, err
		}
		return s, nil
	}
	// for backward compatibility
	s, err := memorystorage.New()
	if err != nil {
		return nil, err
	}
	return s, nil
}
