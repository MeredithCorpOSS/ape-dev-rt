package dependency

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"sync"

	api "github.com/hashicorp/consul/api"
)

// StoreKey represents a single item in Consul's KV store.
type StoreKey struct {
	sync.Mutex

	rawKey     string
	Path       string
	DataCenter string

	defaultValue   string
	defaultGiven   bool
	existenceCheck bool

	stopped bool
	stopCh  chan struct{}
}

// kvGetResponse is a wrapper around the Consul API response.
type kvGetResponse struct {
	pair *api.KVPair
	meta *api.QueryMeta
	err  error
}

// Fetch queries the Consul API defined by the given client and returns string
// of the value to Path.
func (d *StoreKey) Fetch(clients *ClientSet, opts *QueryOptions) (interface{}, *ResponseMetadata, error) {
	d.Lock()
	if d.stopped {
		defer d.Unlock()
		return nil, nil, ErrStopped
	}
	d.Unlock()

	if opts == nil {
		opts = &QueryOptions{}
	}

	consulOpts := opts.consulQueryOptions()
	if d.DataCenter != "" {
		consulOpts.Datacenter = d.DataCenter
	}

	consul, err := clients.Consul()
	if err != nil {
		return nil, nil, fmt.Errorf("store key: error getting client: %s", err)
	}

	dataCh := make(chan *kvGetResponse, 1)

	go func() {
		log.Printf("[DEBUG] (%s) querying consul with %+v", d.Display(), consulOpts)
		pair, meta, err := consul.KV().Get(d.Path, consulOpts)
		resp := &kvGetResponse{pair: pair, meta: meta, err: err}

		select {
		case dataCh <- resp:
		case <-d.stopCh:
		}
	}()

	select {
	case <-d.stopCh:
		return nil, nil, ErrStopped
	case resp := <-dataCh:
		if resp.err != nil {
			return "", nil, fmt.Errorf("store key: error fetching: %s", resp.err)
		}

		rm := &ResponseMetadata{
			LastIndex:   resp.meta.LastIndex,
			LastContact: resp.meta.LastContact,
		}

		if d.existenceCheck {
			return (resp.pair != nil), rm, nil
		}

		if resp.pair == nil {
			if d.defaultGiven {
				log.Printf("[DEBUG] (%s) Consul returned no data (using default of %q)",
					d.Display(), d.defaultValue)
				return d.defaultValue, rm, nil
			}
			return nil, rm, nil
		}

		log.Printf("[DEBUG] (%s) Consul returned %s", d.Display(), resp.pair.Value)

		return string(resp.pair.Value), rm, nil
	}
}

// SetExistenceCheck sets this keys as an existence check instead of a value
// check.
func (d *StoreKey) SetExistenceCheck(b bool) {
	d.existenceCheck = true
}

// SetDefault is used to set the default value.
func (d *StoreKey) SetDefault(s string) {
	d.defaultGiven = true
	d.defaultValue = s
}

// CanShare returns a boolean if this dependency is shareable.
func (d *StoreKey) CanShare() bool {
	return true
}

// HashCode returns a unique identifier.
func (d *StoreKey) HashCode() string {
	if d.existenceCheck {
		return fmt.Sprintf("StoreKeyExists|%s", d.rawKey)
	}

	if d.defaultGiven {
		return fmt.Sprintf("StoreKey|%s|%s", d.rawKey, d.defaultValue)
	}

	return fmt.Sprintf("StoreKey|%s", d.rawKey)
}

// Display prints the human-friendly output.
func (d *StoreKey) Display() string {
	if d.existenceCheck {
		return fmt.Sprintf(`"key_exists(%s)"`, d.rawKey)
	}

	if d.defaultGiven {
		return fmt.Sprintf(`"key_or_default(%s, %q)"`, d.rawKey, d.defaultValue)
	}

	return fmt.Sprintf(`"key(%s)"`, d.rawKey)
}

// Stop halts the dependency's fetch function.
func (d *StoreKey) Stop() {
	d.Lock()
	defer d.Unlock()

	if !d.stopped {
		close(d.stopCh)
		d.stopped = true
	}
}

// ParseStoreKey parses a string of the format a(/b(/c...))
func ParseStoreKey(s string) (*StoreKey, error) {
	if len(s) == 0 {
		return nil, errors.New("cannot specify empty key dependency")
	}

	re := regexp.MustCompile(`\A` +
		`(?P<key>[^@]+)` +
		`(@(?P<datacenter>.+))?` +
		`\z`)
	names := re.SubexpNames()
	match := re.FindAllStringSubmatch(s, -1)

	if len(match) == 0 {
		return nil, errors.New("invalid key dependency format")
	}

	r := match[0]

	m := map[string]string{}
	for i, n := range r {
		if names[i] != "" {
			m[names[i]] = n
		}
	}

	key, datacenter := m["key"], m["datacenter"]

	if key == "" {
		return nil, errors.New("key part is required")
	}

	kd := &StoreKey{
		rawKey:     s,
		Path:       key,
		DataCenter: datacenter,
		stopCh:     make(chan struct{}),
	}

	return kd, nil
}
