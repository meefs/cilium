// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Cilium

package tunnel

import (
	"fmt"
	"strings"

	"github.com/cilium/hive/cell"
	"github.com/spf13/pflag"

	dpcfgdef "github.com/cilium/cilium/pkg/datapath/linux/config/defines"
	"github.com/cilium/cilium/pkg/datapath/linux/safenetlink"
	"github.com/cilium/cilium/pkg/defaults"
)

// EncapProtocol represents the valid types of encapsulation protocols.
type EncapProtocol string

// UnderlayProtocol represents the valid types of underlay protocols for the tunnel.
type UnderlayProtocol string

const (
	// VXLAN specifies VXLAN encapsulation
	VXLAN EncapProtocol = "vxlan"

	// Geneve specifies Geneve encapsulation
	Geneve EncapProtocol = "geneve"

	// Disabled specifies to disable encapsulation
	Disabled EncapProtocol = ""

	IPv4 UnderlayProtocol = "ipv4"
	IPv6 UnderlayProtocol = "ipv6"
)

func (tp EncapProtocol) String() string { return string(tp) }

func (tp EncapProtocol) toDpID() string {
	switch tp {
	case VXLAN:
		return "1"
	case Geneve:
		return "2"
	default:
		return ""
	}
}

// Config represents the materialized tunneling configuration to be used,
// depending on the user configuration and optional overrides required by
// additional features.
type Config struct {
	underlay       UnderlayProtocol
	protocol       EncapProtocol
	port           uint16
	srcPortLow     uint16
	srcPortHigh    uint16
	deviceName     string
	shouldAdaptMTU bool
}

type newConfigIn struct {
	cell.In

	Cfg      userCfg
	Enablers []enabler `group:"request-enable-tunneling"`
}

var (
	configDisabled = Config{
		underlay:       IPv4,
		protocol:       Disabled,
		port:           0,
		srcPortLow:     0,
		srcPortHigh:    0,
		deviceName:     "",
		shouldAdaptMTU: false,
	}
)

func newConfig(in newConfigIn) (Config, error) {
	switch EncapProtocol(in.Cfg.TunnelProtocol) {
	case VXLAN, Geneve:
	default:
		return configDisabled, fmt.Errorf("invalid tunnel protocol %q", in.Cfg.TunnelProtocol)
	}

	switch UnderlayProtocol(in.Cfg.UnderlayProtocol) {
	case IPv4, IPv6:
	default:
		return configDisabled, fmt.Errorf("invalid IP family for underlay %q", in.Cfg.UnderlayProtocol)
	}

	cfg := Config{
		underlay:       UnderlayProtocol(in.Cfg.UnderlayProtocol),
		protocol:       EncapProtocol(in.Cfg.TunnelProtocol),
		port:           in.Cfg.TunnelPort,
		srcPortLow:     0,
		srcPortHigh:    0,
		deviceName:     "",
		shouldAdaptMTU: false,
	}

	if _, err := fmt.Sscanf(in.Cfg.TunnelSourcePortRange, "%d-%d", &cfg.srcPortLow, &cfg.srcPortHigh); err != nil {
		return configDisabled, fmt.Errorf("invalid tunnel source port range %q", in.Cfg.TunnelSourcePortRange)
	}

	var enabled bool
	for _, e := range in.Enablers {
		if e.enable {
			enabled = true
			cfg.shouldAdaptMTU = cfg.shouldAdaptMTU || e.needsMTUAdaptation

			for _, validator := range e.validators {
				if err := validator(cfg.protocol); err != nil {
					return configDisabled, err
				}
			}
		}
	}

	if !enabled {
		return configDisabled, nil
	}

	switch cfg.protocol {
	case VXLAN:
		cfg.deviceName = defaults.VxlanDevice
		if cfg.port == 0 {
			cfg.port = defaults.TunnelPortVXLAN
		}
	case Geneve:
		cfg.deviceName = defaults.GeneveDevice
		if cfg.port == 0 {
			cfg.port = defaults.TunnelPortGeneve
		}
	}

	return cfg, nil
}

// NewTestConfig returns a new TunnelConfig for testing purposes.
func NewTestConfig(proto EncapProtocol) Config {
	//exhaustruct:ignore // Test code can underspecify the default config
	cfg := Config{protocol: proto}

	switch proto {
	case VXLAN:
		cfg.port = defaults.TunnelPortVXLAN
		cfg.deviceName = defaults.VxlanDevice
	case Geneve:
		cfg.port = defaults.TunnelPortGeneve
		cfg.deviceName = defaults.GeneveDevice
	}

	return cfg
}

// EncapProtocol returns the enabled tunnel protocol. The tunnel protocol may be
// set to either VXLAN or Geneve even when the primary mode is native routing, in
// case an additional feature (e.g., egress gateway) may request some traffic to
// be routed through a tunnel.
func (cfg Config) EncapProtocol() EncapProtocol { return cfg.protocol }

func (cfg Config) UnderlayProtocol() UnderlayProtocol { return cfg.underlay }

// Port returns the port used by the tunnel (0 if disabled).
func (cfg Config) Port() uint16 { return cfg.port }

// SrcPortLow returns the lower src port hint to be used by the tunnel (0 if disabled).
func (cfg Config) SrcPortLow() uint16 { return cfg.srcPortLow }

// SrcPortHigh returns the upper src port hint to be used by the tunnel (0 if disabled).
func (cfg Config) SrcPortHigh() uint16 { return cfg.srcPortHigh }

// DeviceName returns the name of the tunnel device (empty if disabled).
func (cfg Config) DeviceName() string { return cfg.deviceName }

// ShouldAdaptMTU returns whether we should adapt the MTU calculation to
// account for encapsulation.
func (cfg Config) ShouldAdaptMTU() bool { return cfg.shouldAdaptMTU }

func (cfg Config) datapathConfigProvider() (dpcfgdef.NodeOut, dpcfgdef.NodeFnOut) {
	defines := make(dpcfgdef.Map)
	definesFn := func() (dpcfgdef.Map, error) { return nil, nil }

	if cfg.EncapProtocol() != Disabled {
		defines[fmt.Sprintf("TUNNEL_PROTOCOL_%s", strings.ToUpper(VXLAN.String()))] = VXLAN.toDpID()
		defines[fmt.Sprintf("TUNNEL_PROTOCOL_%s", strings.ToUpper(Geneve.String()))] = Geneve.toDpID()
		defines["TUNNEL_PROTOCOL"] = cfg.EncapProtocol().toDpID()
		defines["TUNNEL_PORT"] = fmt.Sprintf("%d", cfg.Port())
		defines["TUNNEL_SRC_PORT_LOW"] = fmt.Sprintf("%d", cfg.SrcPortLow())
		defines["TUNNEL_SRC_PORT_HIGH"] = fmt.Sprintf("%d", cfg.SrcPortHigh())

		definesFn = func() (dpcfgdef.Map, error) {
			tunnelDev, err := safenetlink.LinkByName(cfg.DeviceName())
			if err != nil {
				return nil, fmt.Errorf("failed to retrieve device info for %q: %w", cfg.DeviceName(), err)
			}

			return dpcfgdef.Map{
				"ENCAP_IFINDEX": fmt.Sprintf("%d", tunnelDev.Attrs().Index),
			}, nil
		}
	}

	return dpcfgdef.NodeOut{NodeDefines: defines}, dpcfgdef.NewNodeFnOut(definesFn)
}

// EnablerOut allows requesting to enable tunneling functionalities.
type EnablerOut struct {
	cell.Out
	Enabler enabler `group:"request-enable-tunneling"`
}

// NewEnabler returns an object to be injected through hive to request to
// enable tunneling functionalities. Extra options are meaningful only when
// enable is set to true, and are ignored otherwise.
func NewEnabler(enable bool, opts ...enablerOpt) EnablerOut {
	enabler := enabler{enable: enable, needsMTUAdaptation: enable}

	for _, opt := range opts {
		opt(&enabler)
	}

	return EnablerOut{Enabler: enabler}
}

// WithValidator allows to register extra validation functions
// to assert that the configured tunnel protocol matches the one expected by
// the given feature.
func WithValidator(validator func(EncapProtocol) error) enablerOpt {
	return func(te *enabler) {
		te.validators = append(te.validators, validator)
	}
}

// WithoutMTUAdaptation conveys that the given feature request
// to enable tunneling, but the MTU adaptation is already handled externally.
func WithoutMTUAdaptation() enablerOpt {
	return func(te *enabler) {
		te.needsMTUAdaptation = false
	}
}

type enabler struct {
	enable             bool
	needsMTUAdaptation bool
	validators         []func(EncapProtocol) error
}

type enablerOpt func(*enabler)

// userCfg wraps the tunnel-related user configurations.
type userCfg struct {
	TunnelProtocol        string
	TunnelSourcePortRange string
	TunnelPort            uint16
	UnderlayProtocol      string
}

// Flags implements the cell.Flagger interface, to register the given flags.
func (def userCfg) Flags(flags *pflag.FlagSet) {
	flags.String("tunnel-protocol", def.TunnelProtocol, "Encapsulation protocol to use for the overlay (\"vxlan\" or \"geneve\")")
	flags.Uint16("tunnel-port", def.TunnelPort, fmt.Sprintf("Tunnel port (default %d for \"vxlan\" and %d for \"geneve\")", defaults.TunnelPortVXLAN, defaults.TunnelPortGeneve))
	flags.String("tunnel-source-port-range", def.TunnelSourcePortRange, fmt.Sprintf("Tunnel source port range hint (default %s)", defaults.TunnelSourcePortRange))
	flags.String("underlay-protocol", def.UnderlayProtocol, "IP family for the underlay (\"ipv4\" or \"ipv6\")")
}

var defaultConfig = userCfg{
	TunnelProtocol:        defaults.TunnelProtocol,
	TunnelSourcePortRange: defaults.TunnelSourcePortRange,
	TunnelPort:            0, // auto-detect based on the protocol.
	UnderlayProtocol:      defaults.UnderlayProtocol,
}
