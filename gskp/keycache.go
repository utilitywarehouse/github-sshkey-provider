package gskp

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/utilitywarehouse/github-sshkey-provider/gskp/simplelog"
)

// KeyCache wraps around the KeyCollector to provide a simple caching mechanism
// for retrieved SSH keys.
type KeyCache struct {
	cache        map[string]cacheEntry
	collector    *KeyCollector
	mutex        *sync.Mutex
	organisation string
	TTL          time.Duration
	Updates      chan string
}

type cacheEntry struct {
	TeamID    int
	JSON      []byte
	UpdatedAt time.Time
}

// NewKeyCache creates a new Cache for the specified GitHub organisation, using
// the provided GitHub access token and TTL.
func NewKeyCache(githubOrg string, githubAccessToken string, ttl time.Duration) *KeyCache {
	return &KeyCache{
		cache:        map[string]cacheEntry{},
		collector:    NewKeyCollector(githubAccessToken),
		mutex:        &sync.Mutex{},
		organisation: githubOrg,
		TTL:          ttl,
		Updates:      make(chan string, 5),
	}
}

// Get returns the user SSH keys for the specified team. It will update if
// there are no keys for this team in the cache or if they are older than
// the Cache's TTL.
func (c *KeyCache) Get(teamName string) ([]byte, error) {
	if keys, exists := c.cache[teamName]; exists && time.Since(keys.UpdatedAt) < c.TTL {
		simplelog.Debugf("found recent keys in the cache")
		return keys.JSON, nil
	}

	simplelog.Debugf("keys not found in cache, updating...")
	if err := c.updateSnippet(teamName); err != nil {
		return nil, err
	}

	return c.cache[teamName].JSON, nil
}

func (c *KeyCache) updateSnippet(teamName string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if _, exists := c.cache[teamName]; !exists {
		c.cache[teamName] = cacheEntry{}
	}

	keys := c.cache[teamName]

	// it could be that this was updating while we were waiting to acquire a lock
	if time.Since(keys.UpdatedAt) < c.TTL {
		simplelog.Debugf("keys are already up to date, won't update")
		return nil
	}

	if keys.TeamID == 0 {
		id, err := c.collector.GetTeamID(c.organisation, teamName)
		if err != nil {
			return err
		}
		keys.TeamID = id
	}

	data, err := c.collector.GetTeamMemberInfo(keys.TeamID)
	if err != nil {
		return err
	}

	jsonText, err := json.Marshal(map[string][]UserInfo{"keys": data})
	if err != nil {
		return err
	}
	keys.JSON = jsonText

	keys.UpdatedAt = time.Now()

	c.cache[teamName] = keys

	select {
	case c.Updates <- teamName:
		simplelog.Debugf("sent an update for team '%s' to the channel", teamName)
	default:
		simplelog.Debugf("could not send an update to the channel")
	}

	return nil
}
