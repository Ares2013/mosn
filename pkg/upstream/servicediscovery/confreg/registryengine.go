package registry

import (
    "gitlab.alipay-inc.com/afe/mosn/pkg/log"
    "gitlab.alipay-inc.com/afe/mosn/pkg/upstream/servicediscovery/confreg/config"
    "gitlab.alipay-inc.com/afe/mosn/pkg/upstream/servicediscovery/confreg/servermanager"
    "sync"
    "gitlab.alipay-inc.com/afe/mosn/pkg/upstream/cluster"
)

var confregServerManager *servermanager.RegistryServerManager
var registryClient Client

var lock = new(sync.Mutex)

var ModuleStarted = false

//Setup registry module.
func init() {
    log.InitDefaultLogger("", log.INFO)

    rpcServerManager := servermanager.NewRPCServerManager()
    cf := &confregAdaptor{
        ca: &cluster.ClusterAdap,
    }
    rpcServerManager.RegisterRPCServerChangeListener(cf)

    go func() {
        re := &Endpoint{
            registryConfig: config.DefaultRegistryConfig,
        }
        re.StartListener()
    }()
}

func StartupRegistryModule(sysConfig *config.SystemConfig, registryConfig *config.RegistryConfig) Client {
    lock.Lock()

    defer func() {
        lock.Unlock()
    }()

    if ModuleStarted {
        return registryClient
    }
    confregServerManager = servermanager.NewRegistryServerManager(sysConfig, registryConfig)

    ModuleStarted = true

    return NewConfregClient(sysConfig, registryConfig, confregServerManager)
}