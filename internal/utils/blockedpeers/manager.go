package blockedpeers

import (
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/zennittians/intelchain/internal/utils/lrucache"
)

type Manager struct {
	internal *lrucache.Cache[peer.ID, time.Time]
}

func NewManager(size int) *Manager {
	return &Manager{
		internal: lrucache.NewCache[peer.ID, time.Time](size),
	}
}

func (m *Manager) IsBanned(key peer.ID, now time.Time) bool {
	future, ok := m.internal.Get(key)

	if ok {
		return future.After(now) // future > now
	}
	return ok
}

func (m *Manager) Ban(key peer.ID, future time.Time) {
	m.internal.Set(key, future)
}

func (m *Manager) Contains(key peer.ID) bool {
	return m.internal.Contains(key)
}

func (m *Manager) Len() int {
	return m.internal.Len()
}

func (m *Manager) Keys() []peer.ID {
	return m.internal.Keys()
}
