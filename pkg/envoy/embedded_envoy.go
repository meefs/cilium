// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Cilium

package envoy

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/cilium/lumberjack/v2"
	cilium "github.com/cilium/proxy/go/cilium/api"
	envoy_config_bootstrap "github.com/envoyproxy/go-control-plane/envoy/config/bootstrap/v3"
	envoy_config_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_config_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_config_endpoint "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	envoy_config_overload "github.com/envoyproxy/go-control-plane/envoy/config/overload/v3"
	envoy_extensions_bootstrap_internal_listener_v3 "github.com/envoyproxy/go-control-plane/envoy/extensions/bootstrap/internal_listener/v3"
	envoy_extensions_resource_monitors_downstream_connections "github.com/envoyproxy/go-control-plane/envoy/extensions/resource_monitors/downstream_connections/v3"
	envoy_config_upstream "github.com/envoyproxy/go-control-plane/envoy/extensions/upstreams/http/v3"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/cilium/cilium/pkg/flowdebug"
	"github.com/cilium/cilium/pkg/logging"
	"github.com/cilium/cilium/pkg/logging/logfields"
	"github.com/cilium/cilium/pkg/metrics"
	"github.com/cilium/cilium/pkg/time"
)

const (
	envoyLogLevelOff      = "off"
	envoyLogLevelCritical = "critical"
	envoyLogLevelError    = "error"
	envoyLogLevelWarning  = "warning"
	envoyLogLevelInfo     = "info"
	envoyLogLevelDebug    = "debug"
	envoyLogLevelTrace    = "trace"
)

var (
	// envoyLevelMap maps slog.Level values to Envoy (spdlog) log levels.
	envoyLevelMap = map[slog.Level]string{
		logging.LevelPanic: envoyLogLevelOff,
		logging.LevelFatal: envoyLogLevelCritical,
		slog.LevelError:    envoyLogLevelError,
		slog.LevelWarn:     envoyLogLevelWarning,
		slog.LevelInfo:     envoyLogLevelInfo,
		slog.LevelDebug:    envoyLogLevelDebug,
		// spdlog "trace" not mapped
	}

	tracing = false
)

const (
	ciliumEnvoyStarter = "cilium-envoy-starter"
	ciliumEnvoy        = "cilium-envoy"

	maxActiveDownstreamConnections = 50000
)

// EnableTracing changes Envoy log level to "trace", producing the most logs.
func EnableTracing() {
	tracing = true
}

func mapLogLevel(agentLogLevel slog.Level, defaultEnvoyLogLevel string) string {
	// Set Envoy loglevel to trace if debug AND verbose Engoy logging is enabled
	if agentLogLevel == slog.LevelDebug && tracing {
		return envoyLogLevelTrace
	}

	// Suppress the debug level if not debugging at flow level.
	if agentLogLevel == slog.LevelDebug && !flowdebug.Enabled() {
		return envoyLogLevelInfo
	}

	// If defined, use explicit default log level for Envoy
	if defaultEnvoyLogLevel != "" {
		return defaultEnvoyLogLevel
	}

	// Fall back to current log level of the agent
	return envoyLevelMap[agentLogLevel]
}

// Envoy manages a running Envoy proxy instance via the
// ListenerDiscoveryService and RouteDiscoveryService gRPC APIs.
type EmbeddedEnvoy struct {
	stopCh chan struct{}
	errCh  chan error
	admin  *EnvoyAdminClient
}

type embeddedEnvoyConfig struct {
	runDir                   string
	logPath                  string
	defaultLogLevel          string
	baseID                   uint64
	keepCapNetBindService    bool
	connectTimeout           int64
	maxRequestsPerConnection uint32
	maxConnectionDuration    time.Duration
	idleTimeout              time.Duration
	maxConcurrentRetries     uint32
}

// startEmbeddedEnvoyInternal starts an Envoy proxy instance.
func (o *onDemandXdsStarter) startEmbeddedEnvoyInternal(config embeddedEnvoyConfig) (*EmbeddedEnvoy, error) {
	envoy := &EmbeddedEnvoy{
		stopCh: make(chan struct{}),
		errCh:  make(chan error, 1),
		admin:  NewEnvoyAdminClientForSocket(o.logger, GetSocketDir(config.runDir), config.defaultLogLevel),
	}

	bootstrapDir := filepath.Join(config.runDir, "envoy")

	// make sure envoy dir exists
	os.Mkdir(bootstrapDir, 0777)

	// Make sure sockets dir exists
	os.Mkdir(GetSocketDir(config.runDir), 0777)

	bootstrapFilePath := filepath.Join(bootstrapDir, "bootstrap.pb")

	o.writeBootstrapConfigFile(bootstrapConfig{
		filePath:                 bootstrapFilePath,
		nodeId:                   "host~127.0.0.1~no-id~localdomain", // node id format inherited from Istio
		cluster:                  ingressClusterName,
		adminPath:                getAdminSocketPath(GetSocketDir(config.runDir)),
		xdsSock:                  getXDSSocketPath(GetSocketDir(config.runDir)),
		egressClusterName:        egressClusterName,
		ingressClusterName:       ingressClusterName,
		connectTimeout:           config.connectTimeout,
		maxRequestsPerConnection: config.maxRequestsPerConnection,
		maxConnectionDuration:    config.maxConnectionDuration,
		idleTimeout:              config.idleTimeout,
		maxConcurrentRetries:     config.maxConcurrentRetries,
	})

	o.logger.Debug("Envoy: Starting embedded Envoy")

	// make it a buffered channel, so we can not only
	// read the written value but also skip it in
	// case no one reader reads it.
	started := make(chan bool, 1)
	go func() {
		var logWriter io.WriteCloser
		var logFormat string
		if config.logPath != "" {
			// Use the Envoy default log format when logging to a separate file
			logFormat = "[%Y-%m-%d %T.%e][%t][%l][%n] %v"
			logger := &lumberjack.Logger{
				Filename:   config.logPath,
				MaxSize:    100, // megabytes
				MaxBackups: 3,
				MaxAge:     28,   // days
				Compress:   true, // disabled by default
			}
			logWriter = logger
		} else {
			// Use log format that looks like Cilium logs when integrating logs
			// The logs will be reported as coming from the cilium-agent, so
			// we add the thread id to be able to differentiate between Envoy's
			// main and worker threads.
			logFormat = "%t|%l|%n|%v"

			// Create a piper that parses and writes into logrus the log
			// messages from Envoy.
			logWriter = o.newEnvoyLogPiper()
		}
		defer logWriter.Close()

		envoyArgs := []string{"-l", mapLogLevel(logging.GetSlogLevel(o.logger), config.defaultLogLevel), "-c", bootstrapFilePath, "--base-id", strconv.FormatUint(config.baseID, 10), "--log-format", logFormat}
		envoyStarterArgs := []string{}
		if config.keepCapNetBindService {
			envoyStarterArgs = append(envoyStarterArgs, "--keep-cap-net-bind-service", "--")
		}
		envoyStarterArgs = append(envoyStarterArgs, envoyArgs...)

		for {
			cmd := exec.Command(ciliumEnvoyStarter, envoyStarterArgs...)
			cmd.Stderr = logWriter
			cmd.Stdout = logWriter

			if err := cmd.Start(); err != nil {
				o.logger.Warn("Envoy: Failed to start proxy",
					logfields.Error, err,
				)
				select {
				case started <- false:
				default:
				}
				return
			}

			o.logger.Info("Envoy: Proxy started",
				logfields.PID, cmd.Process.Pid,
			)
			metrics.SubprocessStart.WithLabelValues(ciliumEnvoyStarter).Inc()
			select {
			case started <- true:
			default:
			}

			// We do not return after a successful start, but watch the Envoy process
			// and restart it if it crashes.
			// Waiting for the process execution is done in the goroutime.
			// The purpose of the "crash channel" is to inform the loop about their
			// Envoy process crash - after closing that channel by the goroutime,
			// the loop continues, the channel is recreated and the new process
			// is watched again.
			crashCh := make(chan struct{})
			go func() {
				if err := cmd.Wait(); err != nil {
					o.logger.Warn("Envoy: Proxy crashed",
						logfields.PID, cmd.Process.Pid,
						logfields.Error, err,
					)
					// Avoid busy loop & hogging CPU resources by waiting before restarting envoy.
					time.Sleep(100 * time.Millisecond)
				} else {
					o.logger.Info("Envoy: Proxy terminated",
						logfields.PID, cmd.Process.Pid,
					)
				}
				close(crashCh)
			}()

			select {
			case <-crashCh:
				// Start Envoy again
				continue
			case <-envoy.stopCh:
				o.logger.Info("Envoy: Stopping embedded Envoy proxy",
					logfields.PID, cmd.Process.Pid,
				)
				if err := envoy.admin.quit(); err != nil {
					o.logger.Error("Envoy: Envoy admin quit failed, killing process",
						logfields.PID, cmd.Process.Pid,
						logfields.Error, err,
					)
					if err := cmd.Process.Kill(); err != nil {
						o.logger.Error("Envoy: Stopping Envoy failed",
							logfields.Error, err,
						)
						envoy.errCh <- err
					}
				}
				close(envoy.errCh)
				return
			}
		}
	}()

	if <-started {
		return envoy, nil
	}

	return nil, errors.New("failed to start embedded Envoy server")
}

// newEnvoyLogPiper creates a writer that parses and logs log messages written by Envoy.
func (o *onDemandXdsStarter) newEnvoyLogPiper() io.WriteCloser {
	reader, writer := io.Pipe()
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(nil, 1024*1024)
	go func() {
		for scanner.Scan() {
			line := scanner.Text()

			logThreadID := "unknown"
			logLevel := "debug"
			logSubsys := "unknown"
			logMsg := ""

			parts := strings.SplitN(line, "|", 4)
			// Parse the line as a log message written by Envoy, assuming it
			// uses the configured format: "%t|%l|%n|%v".
			if len(parts) == 4 {
				logThreadID = parts[0]
				logLevel = parts[1]
				logSubsys = fmt.Sprintf("envoy-%s", parts[2])
				// TODO: Parse msg to extract the source filename, line number, etc.
				logMsg = fmt.Sprintf("[%s", parts[3])
			} else {
				// If this line can't be parsed, it continues a multi-line log
				// message. In this case, log it at the same level and with the
				// same fields as the previous line.
				logMsg = line
			}

			scopedLog := o.logger.With(
				logfields.LogSubsys, logSubsys,
				logfields.ThreadID, logThreadID,
			)

			if len(logMsg) == 0 {
				continue
			}

			// Map the Envoy log level to a logrus level.
			switch logLevel {
			case envoyLogLevelOff, envoyLogLevelCritical, envoyLogLevelError:
				scopedLog.Error(logMsg)
			case envoyLogLevelWarning:
				// Demote expected warnings to info level
				if strings.Contains(logMsg, "gRPC config: initial fetch timed out for") {
					scopedLog.Info(logMsg)
					continue
				}
				scopedLog.Warn(logMsg)
			case envoyLogLevelInfo:
				scopedLog.Info(logMsg)
			case envoyLogLevelDebug, envoyLogLevelTrace:
				scopedLog.Debug(logMsg)
			default:
				scopedLog.Debug(logMsg)
			}
		}
		if err := scanner.Err(); err != nil {
			o.logger.Error("Error while parsing Envoy logs",
				logfields.Error, err,
			)
		}
		reader.Close()
	}()
	return writer
}

// Stop kills the Envoy process started with startEmbeddedEnvoy. The gRPC API streams are terminated
// first.
func (e *EmbeddedEnvoy) Stop() error {
	close(e.stopCh)
	err, ok := <-e.errCh
	if ok {
		return err
	}
	return nil
}

func (e *EmbeddedEnvoy) GetAdminClient() *EnvoyAdminClient {
	return e.admin
}

type bootstrapConfig struct {
	filePath                 string
	nodeId                   string
	cluster                  string
	adminPath                string
	xdsSock                  string
	egressClusterName        string
	ingressClusterName       string
	connectTimeout           int64
	maxRequestsPerConnection uint32
	maxConnectionDuration    time.Duration
	idleTimeout              time.Duration
	maxConcurrentRetries     uint32
}

func (o *onDemandXdsStarter) writeBootstrapConfigFile(config bootstrapConfig) {
	useDownstreamProtocol := map[string]*anypb.Any{
		"envoy.extensions.upstreams.http.v3.HttpProtocolOptions": toAny(&envoy_config_upstream.HttpProtocolOptions{
			CommonHttpProtocolOptions: &envoy_config_core.HttpProtocolOptions{
				IdleTimeout:              durationpb.New(config.idleTimeout),
				MaxRequestsPerConnection: wrapperspb.UInt32(config.maxRequestsPerConnection),
				MaxConnectionDuration:    durationpb.New(config.maxConnectionDuration),
			},
			UpstreamProtocolOptions: &envoy_config_upstream.HttpProtocolOptions_UseDownstreamProtocolConfig{
				UseDownstreamProtocolConfig: &envoy_config_upstream.HttpProtocolOptions_UseDownstreamHttpConfig{},
			},
		}),
	}

	useDownstreamProtocolAutoSNI := map[string]*anypb.Any{
		"envoy.extensions.upstreams.http.v3.HttpProtocolOptions": toAny(&envoy_config_upstream.HttpProtocolOptions{
			UpstreamHttpProtocolOptions: &envoy_config_core.UpstreamHttpProtocolOptions{
				//	Setting AutoSni or AutoSanValidation options here may crash
				//	Envoy, when Cilium Network filter already passes these from
				//	downstream to upstream.
			},
			CommonHttpProtocolOptions: &envoy_config_core.HttpProtocolOptions{
				IdleTimeout:              durationpb.New(config.idleTimeout),
				MaxRequestsPerConnection: wrapperspb.UInt32(config.maxRequestsPerConnection),
				MaxConnectionDuration:    durationpb.New(config.maxConnectionDuration),
			},
			UpstreamProtocolOptions: &envoy_config_upstream.HttpProtocolOptions_UseDownstreamProtocolConfig{
				UseDownstreamProtocolConfig: &envoy_config_upstream.HttpProtocolOptions_UseDownstreamHttpConfig{},
			},
		}),
	}

	http2ProtocolOptions := map[string]*anypb.Any{
		"envoy.extensions.upstreams.http.v3.HttpProtocolOptions": toAny(&envoy_config_upstream.HttpProtocolOptions{
			UpstreamProtocolOptions: &envoy_config_upstream.HttpProtocolOptions_ExplicitHttpConfig_{
				ExplicitHttpConfig: &envoy_config_upstream.HttpProtocolOptions_ExplicitHttpConfig{
					ProtocolConfig: &envoy_config_upstream.HttpProtocolOptions_ExplicitHttpConfig_Http2ProtocolOptions{},
				},
			},
		}),
	}

	clusterRetryLimits := &envoy_config_cluster.CircuitBreakers{
		Thresholds: []*envoy_config_cluster.CircuitBreakers_Thresholds{{
			MaxRetries: &wrapperspb.UInt32Value{Value: config.maxConcurrentRetries},
		}},
	}

	bs := &envoy_config_bootstrap.Bootstrap{
		Node: &envoy_config_core.Node{Id: config.nodeId, Cluster: config.cluster},
		StaticResources: &envoy_config_bootstrap.Bootstrap_StaticResources{
			Clusters: []*envoy_config_cluster.Cluster{
				{
					Name:                          egressClusterName,
					ClusterDiscoveryType:          &envoy_config_cluster.Cluster_Type{Type: envoy_config_cluster.Cluster_ORIGINAL_DST},
					ConnectTimeout:                &durationpb.Duration{Seconds: config.connectTimeout, Nanos: 0},
					CleanupInterval:               &durationpb.Duration{Seconds: config.connectTimeout, Nanos: 500000000},
					LbPolicy:                      envoy_config_cluster.Cluster_CLUSTER_PROVIDED,
					TypedExtensionProtocolOptions: useDownstreamProtocol,
					CircuitBreakers:               clusterRetryLimits,
				},
				{
					Name:                          egressTLSClusterName,
					ClusterDiscoveryType:          &envoy_config_cluster.Cluster_Type{Type: envoy_config_cluster.Cluster_ORIGINAL_DST},
					ConnectTimeout:                &durationpb.Duration{Seconds: config.connectTimeout, Nanos: 0},
					CleanupInterval:               &durationpb.Duration{Seconds: config.connectTimeout, Nanos: 500000000},
					LbPolicy:                      envoy_config_cluster.Cluster_CLUSTER_PROVIDED,
					TypedExtensionProtocolOptions: useDownstreamProtocolAutoSNI,
					TransportSocket: &envoy_config_core.TransportSocket{
						Name: "cilium.tls_wrapper",
						ConfigType: &envoy_config_core.TransportSocket_TypedConfig{
							TypedConfig: toAny(&cilium.UpstreamTlsWrapperContext{}),
						},
					},
				},
				{
					Name:                          ingressClusterName,
					ClusterDiscoveryType:          &envoy_config_cluster.Cluster_Type{Type: envoy_config_cluster.Cluster_ORIGINAL_DST},
					ConnectTimeout:                &durationpb.Duration{Seconds: config.connectTimeout, Nanos: 0},
					CleanupInterval:               &durationpb.Duration{Seconds: config.connectTimeout, Nanos: 500000000},
					LbPolicy:                      envoy_config_cluster.Cluster_CLUSTER_PROVIDED,
					TypedExtensionProtocolOptions: useDownstreamProtocol,
				},
				{
					Name:                          ingressTLSClusterName,
					ClusterDiscoveryType:          &envoy_config_cluster.Cluster_Type{Type: envoy_config_cluster.Cluster_ORIGINAL_DST},
					ConnectTimeout:                &durationpb.Duration{Seconds: config.connectTimeout, Nanos: 0},
					CleanupInterval:               &durationpb.Duration{Seconds: config.connectTimeout, Nanos: 500000000},
					LbPolicy:                      envoy_config_cluster.Cluster_CLUSTER_PROVIDED,
					TypedExtensionProtocolOptions: useDownstreamProtocolAutoSNI,
					TransportSocket: &envoy_config_core.TransportSocket{
						Name: "cilium.tls_wrapper",
						ConfigType: &envoy_config_core.TransportSocket_TypedConfig{
							TypedConfig: toAny(&cilium.UpstreamTlsWrapperContext{}),
						},
					},
				},
				{
					Name:                 CiliumXDSClusterName,
					ClusterDiscoveryType: &envoy_config_cluster.Cluster_Type{Type: envoy_config_cluster.Cluster_STATIC},
					ConnectTimeout:       &durationpb.Duration{Seconds: config.connectTimeout, Nanos: 0},
					LbPolicy:             envoy_config_cluster.Cluster_ROUND_ROBIN,
					LoadAssignment: &envoy_config_endpoint.ClusterLoadAssignment{
						ClusterName: CiliumXDSClusterName,
						Endpoints: []*envoy_config_endpoint.LocalityLbEndpoints{{
							LbEndpoints: []*envoy_config_endpoint.LbEndpoint{{
								HostIdentifier: &envoy_config_endpoint.LbEndpoint_Endpoint{
									Endpoint: &envoy_config_endpoint.Endpoint{
										Address: &envoy_config_core.Address{
											Address: &envoy_config_core.Address_Pipe{
												Pipe: &envoy_config_core.Pipe{Path: config.xdsSock},
											},
										},
									},
								},
							}},
						}},
					},
					TypedExtensionProtocolOptions: http2ProtocolOptions,
				},
				{
					Name:                 adminClusterName,
					ClusterDiscoveryType: &envoy_config_cluster.Cluster_Type{Type: envoy_config_cluster.Cluster_STATIC},
					ConnectTimeout:       &durationpb.Duration{Seconds: config.connectTimeout, Nanos: 0},
					LbPolicy:             envoy_config_cluster.Cluster_ROUND_ROBIN,
					LoadAssignment: &envoy_config_endpoint.ClusterLoadAssignment{
						ClusterName: adminClusterName,
						Endpoints: []*envoy_config_endpoint.LocalityLbEndpoints{{
							LbEndpoints: []*envoy_config_endpoint.LbEndpoint{{
								HostIdentifier: &envoy_config_endpoint.LbEndpoint_Endpoint{
									Endpoint: &envoy_config_endpoint.Endpoint{
										Address: &envoy_config_core.Address{
											Address: &envoy_config_core.Address_Pipe{
												Pipe: &envoy_config_core.Pipe{Path: config.adminPath},
											},
										},
									},
								},
							}},
						}},
					},
				},
			},
		},
		DynamicResources: &envoy_config_bootstrap.Bootstrap_DynamicResources{
			LdsConfig: CiliumXDSConfigSource,
			CdsConfig: CiliumXDSConfigSource,
		},
		Admin: &envoy_config_bootstrap.Admin{
			Address: &envoy_config_core.Address{
				Address: &envoy_config_core.Address_Pipe{
					Pipe: &envoy_config_core.Pipe{Path: config.adminPath},
				},
			},
		},
		BootstrapExtensions: []*envoy_config_core.TypedExtensionConfig{
			{
				Name:        "envoy.bootstrap.internal_listener",
				TypedConfig: toAny(&envoy_extensions_bootstrap_internal_listener_v3.InternalListener{}),
			},
		},
		OverloadManager: &envoy_config_overload.OverloadManager{
			ResourceMonitors: []*envoy_config_overload.ResourceMonitor{{
				Name: "envoy.resource_monitors.global_downstream_max_connections",
				ConfigType: &envoy_config_overload.ResourceMonitor_TypedConfig{
					TypedConfig: toAny(&envoy_extensions_resource_monitors_downstream_connections.DownstreamConnectionsConfig{
						MaxActiveDownstreamConnections: maxActiveDownstreamConnections,
					}),
				},
			}},
		},
	}

	o.logger.Debug("Envoy: Writing Bootstrap config",
		logfields.Resource, bs,
	)
	data, err := proto.Marshal(bs)
	if err != nil {
		o.logger.Error("Envoy: Error marshaling Envoy bootstrap",
			logfields.Error, err,
		)
		return
	}
	if err := os.WriteFile(config.filePath, data, 0644); err != nil {
		o.logger.Error("Envoy: Error writing Envoy bootstrap file",
			logfields.Error, err,
		)
	}
}

// getEmbeddedEnvoyVersion returns the envoy binary version string
func getEmbeddedEnvoyVersion() (string, error) {
	out, err := exec.Command(ciliumEnvoy, "--version").Output()
	if err != nil {
		return "", fmt.Errorf("failed to execute '%s --version': %w", ciliumEnvoy, err)
	}
	envoyVersionString := strings.TrimSpace(string(out))

	envoyVersionArray := strings.Fields(envoyVersionString)
	if len(envoyVersionArray) < 3 {
		return "", fmt.Errorf("failed to extract version from truncated Envoy version string")
	}

	return envoyVersionArray[2], nil
}
