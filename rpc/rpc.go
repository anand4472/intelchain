package rpc

import (
	"fmt"
	"net"
	"strings"

	"github.com/intelchain-itc/intelchain/eth/rpc"
	"github.com/intelchain-itc/intelchain/internal/configs/intelchain"
	nodeconfig "github.com/intelchain-itc/intelchain/internal/configs/node"
	"github.com/intelchain-itc/intelchain/internal/utils"
	"github.com/intelchain-itc/intelchain/itc"
	eth "github.com/intelchain-itc/intelchain/rpc/eth"
	v1 "github.com/intelchain-itc/intelchain/rpc/v1"
	v2 "github.com/intelchain-itc/intelchain/rpc/v2"
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
	APIVersion = "1.1"
	// LogTag is the tag found in the log for all RPC logs
	LogTag = "[RPC]"
	// HTTPPortOffset ..
	HTTPPortOffset = 500
	// WSPortOffset ..
	WSPortOffset = 800

	netNamespace   = "net"
	netV1Namespace = "netv1"
	netV2Namespace = "netv2"
	web3Namespace  = "web3"
)

var (
	// HTTPModules ..
	HTTPModules = []string{"itc", "itcv2", "eth", "debug", "trace", netNamespace, netV1Namespace, netV2Namespace, web3Namespace, "explorer", "preimages"}
	// WSModules ..
	WSModules = []string{"itc", "itcv2", "eth", "debug", "trace", netNamespace, netV1Namespace, netV2Namespace, web3Namespace, "web3"}

	httpListener     net.Listener
	httpHandler      *rpc.Server
	wsListener       net.Listener
	wsHandler        *rpc.Server
	httpEndpoint     = ""
	httpAuthEndpoint = ""
	wsEndpoint       = ""
	wsAuthEndpoint   = ""
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
func StartServers(itc *itc.Intelchain, apis []rpc.API, config nodeconfig.RPCServerConfig, rpcOpt intelchain.RpcOptConfig) error {
	apis = append(apis, getAPIs(itc, config)...)
	authApis := append(apis, getAuthAPIs(itc, config.DebugEnabled, config.RateLimiterEnabled, config.RequestsPerSecond)...)
	if rpcOpt.PreimagesEnabled {
		authApis = append(authApis, NewPreimagesAPI(itc, "preimages"))
	}
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
		if err := startHTTP(apis, &rmf, timeouts); err != nil {
			return err
		}

		httpAuthEndpoint = fmt.Sprintf("%v:%v", config.HTTPIp, config.HTTPAuthPort)
		if err := startAuthHTTP(authApis, &rmf, timeouts); err != nil {
			return err
		}
	}

	if config.WSEnabled {
		wsEndpoint = fmt.Sprintf("%v:%v", config.WSIp, config.WSPort)
		if err := startWS(apis, &rmf); err != nil {
			return err
		}

		wsAuthEndpoint = fmt.Sprintf("%v:%v", config.WSIp, config.WSAuthPort)
		if err := startAuthWS(authApis, &rmf); err != nil {
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

func getAuthAPIs(itc *itc.Intelchain, debugEnable bool, rateLimiterEnable bool, ratelimit int) []rpc.API {
	return []rpc.API{
		NewPublicTraceAPI(itc, Debug), // Debug version means geth trace rpc
		NewPublicTraceAPI(itc, Trace), // Trace version means parity trace rpc
	}
}

// getAPIs returns all the API methods for the RPC interface
func getAPIs(itc *itc.Intelchain, config nodeconfig.RPCServerConfig) []rpc.API {
	publicAPIs := []rpc.API{
		// Public methods
		NewPublicIntelchainAPI(itc, V1),
		NewPublicIntelchainAPI(itc, V2),
		NewPublicBlockchainAPI(itc, V1, config.RateLimiterEnabled, config.RequestsPerSecond),
		NewPublicBlockchainAPI(itc, V2, config.RateLimiterEnabled, config.RequestsPerSecond),
		NewPublicContractAPI(itc, V1, config.RateLimiterEnabled, config.RequestsPerSecond, config.EvmCallTimeout),
		NewPublicContractAPI(itc, V2, config.RateLimiterEnabled, config.RequestsPerSecond, config.EvmCallTimeout),
		NewPublicTransactionAPI(itc, V1),
		NewPublicTransactionAPI(itc, V2),
		NewPublicPoolAPI(itc, V1, config.RateLimiterEnabled, config.RequestsPerSecond),
		NewPublicPoolAPI(itc, V2, config.RateLimiterEnabled, config.RequestsPerSecond),
	}

	// Legacy methods (subject to removal)
	if config.LegacyRPCsEnabled {
		publicAPIs = append(publicAPIs,
			v1.NewPublicLegacyAPI(itc, "itc"),
			v2.NewPublicLegacyAPI(itc, "itcv2"),
		)
	}

	if config.StakingRPCsEnabled {
		publicAPIs = append(publicAPIs,
			NewPublicStakingAPI(itc, V1),
			NewPublicStakingAPI(itc, V2),
		)
	}

	if config.EthRPCsEnabled {
		publicAPIs = append(publicAPIs,
			NewPublicIntelchainAPI(itc, Eth),
			NewPublicBlockchainAPI(itc, Eth, config.RateLimiterEnabled, config.RequestsPerSecond),
			NewPublicContractAPI(itc, Eth, config.RateLimiterEnabled, config.RequestsPerSecond, config.EvmCallTimeout),
			NewPublicTransactionAPI(itc, Eth),
			NewPublicPoolAPI(itc, Eth, config.RateLimiterEnabled, config.RequestsPerSecond),
			eth.NewPublicEthService(itc, "eth"),
		)
	}

	publicDebugAPIs := []rpc.API{
		//Public debug API
		NewPublicDebugAPI(itc, V1),
		NewPublicDebugAPI(itc, V2),
	}

	privateAPIs := []rpc.API{
		NewPrivateDebugAPI(itc, V1),
		NewPrivateDebugAPI(itc, V2),
	}

	if config.DebugEnabled {
		apis := append(publicAPIs, publicDebugAPIs...)
		return append(apis, privateAPIs...)
	}
	return publicAPIs
}

func startHTTP(apis []rpc.API, rmf *rpc.RpcMethodFilter, httpTimeouts rpc.HTTPTimeouts) (err error) {
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
	fmt.Printf("Started RPC server at: %v\n", httpEndpoint)
	return nil
}

func startAuthHTTP(apis []rpc.API, rmf *rpc.RpcMethodFilter, httpTimeouts rpc.HTTPTimeouts) (err error) {
	httpListener, httpHandler, err = rpc.StartHTTPEndpoint(
		httpAuthEndpoint, apis, HTTPModules, rmf, httpOrigins, httpVirtualHosts, httpTimeouts,
	)
	if err != nil {
		return err
	}

	utils.Logger().Info().
		Str("url", fmt.Sprintf("http://%s", httpAuthEndpoint)).
		Str("cors", strings.Join(httpOrigins, ",")).
		Str("vhosts", strings.Join(httpVirtualHosts, ",")).
		Msg("HTTP endpoint opened")
	fmt.Printf("Started Auth-RPC server at: %v\n", httpAuthEndpoint)
	return nil
}

func startWS(apis []rpc.API, rmf *rpc.RpcMethodFilter) (err error) {
	wsListener, wsHandler, err = rpc.StartWSEndpoint(wsEndpoint, apis, WSModules, rmf, wsOrigins, true)
	if err != nil {
		return err
	}

	utils.Logger().Info().
		Str("url", fmt.Sprintf("ws://%s", wsListener.Addr())).
		Msg("WebSocket endpoint opened")
	fmt.Printf("Started WS server at: %v\n", wsEndpoint)
	return nil
}

func startAuthWS(apis []rpc.API, rmf *rpc.RpcMethodFilter) (err error) {
	wsListener, wsHandler, err = rpc.StartWSEndpoint(wsAuthEndpoint, apis, WSModules, rmf, wsOrigins, true)
	if err != nil {
		return err
	}

	utils.Logger().Info().
		Str("url", fmt.Sprintf("ws://%s", wsListener.Addr())).
		Msg("WebSocket endpoint opened")
	fmt.Printf("Started Auth-WS server at: %v\n", wsAuthEndpoint)
	return nil
}
