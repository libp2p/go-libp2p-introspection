package introspection

import (
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/imdario/mergo"
	"github.com/libp2p/go-libp2p-core/introspect"
	"github.com/pkg/errors"
	"sync"
	"time"
)

var _ introspect.Introspector = (*DefaultIntrospector)(nil)

// DefaultIntrospector is a registry of subsystem data/metrics providers and also allows
// clients to inspect the system state by calling all the providers registered with it
type DefaultIntrospector struct {
	treeMu sync.RWMutex
	tree   *introspect.ProvidersMap
	addr   string
}

func NewDefaultIntrospector() *DefaultIntrospector {
	return &DefaultIntrospector{tree: &introspect.ProvidersMap{}}
}

func (d *DefaultIntrospector) RegisterProviders(provs *introspect.ProvidersMap) error {
	d.treeMu.Lock()
	defer d.treeMu.Unlock()

	if err := mergo.Merge(d.tree, provs); err != nil {
		return err
	}

	return nil
}

// TODO Get this to work
func (d *DefaultIntrospector) ListenAddress() string {
	return ""
}

func (d *DefaultIntrospector) FetchCurrentState() (*introspect.State, error) {
	d.treeMu.RLock()
	defer d.treeMu.RUnlock()

	s := &introspect.State{}

	// subsystems
	s.Subsystems = &introspect.Subsystems{}

	// version
	s.Version = &introspect.Version{Number: introspect.ProtoVersion}

	// runtime
	if d.tree.Runtime != nil {
		r, err := d.tree.Runtime()
		if err != nil {
			return nil, errors.Wrap(err, "failed to fetch runtime info")
		}
		s.Runtime = r
	}

	// timestamps
	s.InstantTs = &timestamp.Timestamp{Seconds: time.Now().Unix()}
	// TODO Figure out the other two timestamp fields

	// connections
	if d.tree.Connection != nil {
		conns, err := d.tree.Connection(introspect.ConnectionQueryInput{Type: introspect.ConnListQueryTypeAll, StreamOutputType: introspect.QueryOutputTypeFull})
		if err != nil {
			return nil, errors.Wrap(err, "failed to fetch connections")
		}
		s.Subsystems.Connections = conns
	}

	return nil, nil
}