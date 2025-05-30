// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Cilium

package cni

import (
	"context"
	"fmt"
	"log/slog"
	"path"

	"github.com/cilium/hive/cell"
	"github.com/spf13/pflag"

	"github.com/cilium/cilium/api/v1/models"
	"github.com/cilium/cilium/pkg/controller"
	"github.com/cilium/cilium/pkg/option"
	cnitypes "github.com/cilium/cilium/plugins/cilium-cni/types"
)

var Cell = cell.Module(
	"cni-config",
	"CNI configuration manager",

	cell.Config(defaultConfig),
	cell.Provide(enableConfigManager),
)

type Config struct {
	WriteCNIConfWhenReady string
	ReadCNIConf           string
	CNIChainingMode       string
	CNILogFile            string
	CNIExclusive          bool
	CNIChainingTarget     string
	CNIExternalRouting    bool
}

type CNIConfigManager interface {
	// GetMTU provides the MTU from the provided CNI configuration file.
	// This is only useful if ReadCNIConfiguration is set *and* the file specifies an MTU.
	GetMTU() int

	// GetChainingMode returns the configured CNI chaining mode
	GetChainingMode() string

	// Status returns the status of the CNI manager.
	// Cannot return nil.
	Status() *models.Status

	GetCustomNetConf() *cnitypes.NetConf

	// ExternalRoutingEnabled returns true if the chained plugin implements
	// routing for Endpoints (Pods).
	ExternalRoutingEnabled() bool
}

var defaultConfig = Config{
	WriteCNIConfWhenReady: "",
	ReadCNIConf:           "",
	CNIChainingMode:       "none",
	CNILogFile:            "/var/run/cilium/cilium-cni.log",
	CNIExclusive:          false,
	CNIChainingTarget:     "",
	CNIExternalRouting:    false,
}

func (cfg Config) Flags(flags *pflag.FlagSet) {
	flags.String(option.WriteCNIConfigurationWhenReady, defaultConfig.WriteCNIConfWhenReady, "Write the CNI configuration to the specified path when agent is ready")
	flags.String(option.ReadCNIConfiguration, defaultConfig.ReadCNIConf, fmt.Sprintf("CNI configuration file to use as a source for --%s. If not supplied, a suitable one will be generated.", option.WriteCNIConfigurationWhenReady))
	flags.String(option.CNIChainingMode, defaultConfig.CNIChainingMode, "Enable CNI chaining with the specified plugin")
	flags.String(option.CNILogFile, defaultConfig.CNILogFile, "Path where the CNI plugin should write logs")
	flags.String(option.CNIChainingTarget, defaultConfig.CNIChainingTarget, "CNI network name into which to insert the Cilium chained configuration. Use '*' to select any network.")
	flags.Bool(option.CNIExclusive, defaultConfig.CNIExclusive, "Whether to remove other CNI configurations")
	flags.Bool(option.CNIExternalRouting, defaultConfig.CNIExternalRouting, "Whether the chained CNI plugin handles routing on the node")
}

func enableConfigManager(lc cell.Lifecycle, logger *slog.Logger, cfg Config, dcfg *option.DaemonConfig /*only for .Debug*/) CNIConfigManager {
	c := newConfigManager(logger, cfg, dcfg.Debug)
	lc.Append(c)
	return c
}

func newConfigManager(logger *slog.Logger, cfg Config, debug bool) *cniConfigManager {
	if cfg.CNIChainingMode == "aws-cni" && cfg.CNIChainingTarget == "" {
		cfg.CNIChainingTarget = "aws-cni"
		cfg.CNIExternalRouting = true
	}

	if cfg.CNIChainingTarget != "" && cfg.CNIChainingMode == "" {
		cfg.CNIChainingMode = "generic-veth"
	}

	if cfg.CNIChainingMode == "" {
		cfg.CNIChainingMode = "none"
	}

	s := models.Status{
		Msg:   "CNI controller not started",
		State: models.StatusStateFailure,
	}

	c := &cniConfigManager{
		config:     cfg,
		debug:      debug,
		logger:     logger,
		controller: controller.NewManager(),
	}

	c.status.Store(&s)

	c.cniConfDir, c.cniConfFile = path.Split(cfg.WriteCNIConfWhenReady)
	c.ctx, c.doneFunc = context.WithCancel(context.Background())

	return c
}
