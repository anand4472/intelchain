package rpc

import (
	"fmt"
	"net"
	"strings"

	"github.com/intelchain-itc/intelchain/eth/rpc"
	bootnodeConfigs "github.com/intelchain-itc/intelchain/internal/configs/bootnode"
	"github.com/intelchain-itc/intelchain/internal/configs/intelchain"
	"github.com/intelchain-itc/intelchain/internal/utils"
	itcboot "github.com/intelchain-itc/intelchain/itc_boot"
)

// Version enum
const (
	V1 Version = iota
	V2
	Eth
	Debug
	Trace
)

const (
	// APIVersion used for DApp's, bumped after RPC refactor (7/2020)
	APIVersion = "1.0"
	// LogTag is the tag found in the log for all RPC logs
	LogTag = "[BOOT_RPC]"
	// HTTPPortOffset ..
	HTTPPortOffset = 500
	// WSPortOffset ..
	WSPortOffset = 800
)

var (
	// HTTPModules ..
	HTTPModules = []string{"itcboot", "itcbootv2", "eth", "debug", "trace"}
	// WSModules ..
	WSModules = []string{"itcboot", "itcbootv2", "eth", "debug", "trace"}

	httpListener     net.Listener
	httpHandler      *rpc.Server
	wsListener       net.Listener
	wsHandler        *rpc.Server
	httpEndpoint     = ""
	wsEndpoint       = ""
	httpVirtualHosts = []string{"*"}
	httpOrigins      = []string{"*"}
	wsOrigins        = []string{"*"}
)

// Version of the RPC
type Version int

// Namespace of the RPC version
func (n Version) Namespace() string {
	return HTTPModules[n]
}

// StartServers starts the http & ws servers
func StartServers(itcboot *itcboot.BootService, apis []rpc.API, config bootnodeConfigs.RPCServerConfig, rpcOpt intelchain.RpcOptConfig) error {
	apis = append(apis, getBootAPIs(itcboot, config)...)
	// load method filter from file (if exist)
	var rmf rpc.RpcMethodFilter
	rpcFilterFilePath := strings.TrimSpace(rpcOpt.RpcFilterFile)
	if len(rpcFilterFilePath) > 0 {
		if err := rmf.LoadRpcMethodFiltersFromFile(rpcFilterFilePath); err != nil {
			return err
		}
	} else {
		rmf.ExposeAll()
	}

	if config.HTTPEnabled {
		timeouts := rpc.HTTPTimeouts{
			ReadTimeout:  config.HTTPTimeoutRead,
			WriteTimeout: config.HTTPTimeoutWrite,
			IdleTimeout:  config.HTTPTimeoutIdle,
		}
		httpEndpoint = fmt.Sprintf("%v:%v", config.HTTPIp, config.HTTPPort)
		if err := startBootServiceHTTP(apis, &rmf, timeouts); err != nil {
			return err
		}
	}

	if config.WSEnabled {
		wsEndpoint = fmt.Sprintf("%v:%v", config.WSIp, config.WSPort)
		if err := startBootServiceWS(apis, &rmf); err != nil {
			return err
		}
	}

	return nil
}

// StopServers stops the http & ws servers
func StopServers() error {
	if httpListener != nil {
		if err := httpListener.Close(); err != nil {
			return err
		}
		httpListener = nil
		utils.Logger().Info().
			Str("url", fmt.Sprintf("http://%s", httpEndpoint)).
			Msg("HTTP endpoint closed")
	}
	if httpHandler != nil {
		httpHandler.Stop()
		httpHandler = nil
	}
	if wsListener != nil {
		if err := wsListener.Close(); err != nil {
			return err
		}
		wsListener = nil
		utils.Logger().Info().
			Str("url", fmt.Sprintf("http://%s", wsEndpoint)).
			Msg("WS endpoint closed")
	}
	if wsHandler != nil {
		wsHandler.Stop()
		wsHandler = nil
	}
	return nil
}

// getBootAPIs returns all the API methods for the RPC interface
func getBootAPIs(itcboot *itcboot.BootService, config bootnodeConfigs.RPCServerConfig) []rpc.API {
	publicAPIs := []rpc.API{
		// Public methods
		NewPublicBootAPI(itcboot, V1),
		NewPublicBootAPI(itcboot, V2),
	}

	return publicAPIs
}

func startBootServiceHTTP(apis []rpc.API, rmf *rpc.RpcMethodFilter, httpTimeouts rpc.HTTPTimeouts) (err error) {
	httpListener, httpHandler, err = rpc.StartHTTPEndpoint(
		httpEndpoint, apis, HTTPModules, rmf, httpOrigins, httpVirtualHosts, httpTimeouts,
	)
	if err != nil {
		return err
	}

	utils.Logger().Info().
		Str("url", fmt.Sprintf("http://%s", httpEndpoint)).
		Str("cors", strings.Join(httpOrigins, ",")).
		Str("vhosts", strings.Join(httpVirtualHosts, ",")).
		Msg("HTTP endpoint opened")

	fmt.Printf("Started Boot Node RPC server at: %v\n", httpEndpoint)
	return nil
}

func startBootServiceWS(apis []rpc.API, rmf *rpc.RpcMethodFilter) (err error) {
	wsListener, wsHandler, err = rpc.StartWSEndpoint(wsEndpoint, apis, WSModules, rmf, wsOrigins, true)
	if err != nil {
		return err
	}

	utils.Logger().Info().
		Str("url", fmt.Sprintf("ws://%s", wsEndpoint)).
		Msg("WebSocket endpoint opened")
	fmt.Printf("Started Boot Node WS server at: %v\n", wsEndpoint)
	return nil
}
