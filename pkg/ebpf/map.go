// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Cilium

package ebpf

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	ciliumebpf "github.com/cilium/ebpf"

	"github.com/cilium/cilium/api/v1/models"
	"github.com/cilium/cilium/pkg/bpf"
	"github.com/cilium/cilium/pkg/lock"
	"github.com/cilium/cilium/pkg/logging/logfields"
	"github.com/cilium/cilium/pkg/metrics"
)

type MapSpec = ciliumebpf.MapSpec

type PinType = ciliumebpf.PinType

const (
	Hash       = ciliumebpf.Hash
	PerCPUHash = ciliumebpf.PerCPUHash
	Array      = ciliumebpf.Array
	HashOfMaps = ciliumebpf.HashOfMaps
	LPMTrie    = ciliumebpf.LPMTrie
	LRUHash    = ciliumebpf.LRUHash
	LRUCPUHash = ciliumebpf.LRUCPUHash
	RingBuf    = ciliumebpf.RingBuf

	PinNone   = ciliumebpf.PinNone
	PinByName = ciliumebpf.PinByName
)

var (
	ErrKeyNotExist = ciliumebpf.ErrKeyNotExist
)

// IterateCallback represents the signature of the callback function expected by
// the IterateWithCallback method, which in turn is used to iterate all the
// keys/values of a map.
type IterateCallback func(key, value any)

// Map represents an eBPF map.
type Map struct {
	logger *slog.Logger
	lock   lock.RWMutex
	*ciliumebpf.Map

	spec *MapSpec
	path string
}

// NewMap creates a new Map object.
func NewMap(logger *slog.Logger, spec *MapSpec) *Map {
	return &Map{
		logger: logger,
		spec:   spec,
	}
}

// LoadRegisterMap loads the specified map from a bpffs pin path and registers
// its handle in the package-global map register.
func LoadRegisterMap(logger *slog.Logger, mapName string) (*Map, error) {
	path := bpf.MapPath(logger, mapName)

	m, err := LoadPinnedMap(logger, path)
	if err != nil {
		return nil, err
	}

	registerMap(m)

	return m, nil
}

// LoadPinnedMap wraps cilium/ebpf's LoadPinnedMap.
func LoadPinnedMap(logger *slog.Logger, fileName string) (*Map, error) {
	m, err := ciliumebpf.LoadPinnedMap(fileName, nil)
	if err != nil {
		return nil, err
	}

	return &Map{
		logger: logger,
		Map:    m,
		path:   fileName,
	}, nil
}

func MapFromID(logger *slog.Logger, id int) (*Map, error) {
	newMap, err := ciliumebpf.NewMapFromID(ciliumebpf.MapID(id))
	if err != nil {
		return nil, err
	}

	return &Map{
		logger: logger,
		Map:    newMap,
	}, nil
}

// OpenOrCreate tries to open or create the eBPF map identified by the spec in
// the Map object.
func (m *Map) OpenOrCreate() error {
	m.lock.Lock()
	defer m.lock.Unlock()

	if m.Map != nil {
		return nil
	}

	if m.spec == nil {
		return fmt.Errorf("cannot create map: nil map spec")
	}

	opts := ciliumebpf.MapOptions{
		PinPath: bpf.TCGlobalsPath(),
	}

	m.spec.Flags |= bpf.GetMapMemoryFlags(m.spec.Type)

	path := bpf.MapPath(m.logger, m.spec.Name)

	if m.spec.Pinning == ciliumebpf.PinByName {
		mapDir := filepath.Dir(path)

		if _, err := os.Stat(mapDir); os.IsNotExist(err) {
			if err = os.MkdirAll(mapDir, 0755); err != nil {
				return &os.PathError{
					Op:   "Unable create map base directory",
					Path: path,
					Err:  err,
				}
			}
		}
	}

	newMap, err := ciliumebpf.NewMapWithOptions(m.spec, opts)
	if err != nil {
		if !errors.Is(err, ciliumebpf.ErrMapIncompatible) {
			return fmt.Errorf("unable to create map: %w", err)
		}

		// There already exists a pinned map but it has a different
		// configuration (e.g different type, k/v size or flags).
		// Try to delete and recreate it.

		m.logger.Warn(
			"Removing map to allow for property upgrade (expect map data loss)",
			logfields.Error, err,
			logfields.BPFMapName, m.spec.Name,
		)

		oldMap, err := ciliumebpf.LoadPinnedMap(path, &opts.LoadPinOptions)
		if err != nil {
			return fmt.Errorf("cannot load pinned map %s: %w", m.spec.Name, err)
		}
		defer func() {
			if err := oldMap.Close(); err != nil {
				m.logger.Warn(
					"Cannot close map",
					logfields.Error, err,
					logfields.BPFMapName, m.spec.Name,
				)
			}
		}()

		if err = oldMap.Unpin(); err != nil {
			return fmt.Errorf("cannot unpin map %s: %w", m.spec.Name, err)
		}

		newMap, err = ciliumebpf.NewMapWithOptions(m.spec, opts)
		if err != nil {
			return fmt.Errorf("unable to create map: %w", err)
		}
	}

	m.Map = newMap
	m.path = path

	registerMap(m)
	metrics.UpdateMapCapacity(m.spec.Name, m.spec.MaxEntries)
	return nil
}

// IterateWithCallback iterates through all the keys/values of a map, passing
// each key/value pair to the cb callback.
func (m *Map) IterateWithCallback(key, value any, cb IterateCallback) error {
	if m.Map == nil {
		if err := m.OpenOrCreate(); err != nil {
			return err
		}
	}

	m.lock.RLock()
	defer m.lock.RUnlock()

	entries := m.Iterate()
	for entries.Next(key, value) {
		cb(key, value)
	}

	return nil
}

// GetModel returns a BPF map in the representation served via the API.
func (m *Map) GetModel() *models.BPFMap {
	m.lock.RLock()
	defer m.lock.RUnlock()

	mapModel := &models.BPFMap{
		Path: m.path,
	}

	// TODO: handle map cache. See pkg/bpf/map_linux.go:GetModel()

	return mapModel
}

func (m *Map) IsEmpty() bool {
	m.lock.RLock()
	defer m.lock.RUnlock()
	var key, value any
	return !m.Iterate().Next(key, value)
}
