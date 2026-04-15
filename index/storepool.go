package index

import "sync"

// Process-scoped store cache. Each unique dbPath is opened once and reused
// for the lifetime of the process. Call CloseAll in cobra PersistentPostRun.
var (
	poolMu    sync.Mutex
	poolCache = map[string]*Store{}
)

// openCached returns a shared Store for dbPath, opening it on first call.
// Callers must not call Close on the returned store.
func openCached(dbPath string) (*Store, error) {
	poolMu.Lock()
	defer poolMu.Unlock()
	if s, ok := poolCache[dbPath]; ok {
		return s, nil
	}
	s, err := OpenStore(dbPath)
	if err != nil {
		return nil, err
	}
	poolCache[dbPath] = s
	return s, nil
}

// CloseAll closes every cached store. Call once at process exit.
func CloseAll() {
	poolMu.Lock()
	defer poolMu.Unlock()
	for k, s := range poolCache {
		s.Close()
		delete(poolCache, k)
	}
}
