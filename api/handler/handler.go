// Copyright (C) 2014-2018 Goodrain Co., Ltd.
// RAINBOND, Application Management Platform

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version. For any non-GPL usage of Rainbond,
// one or multiple Commercial Licenses authorized by Goodrain Co., Ltd.
// must be obtained first.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package handler

import (
	"github.com/coreos/etcd/clientv3"
	"github.com/goodrain/rainbond/api/client/prometheus"
	api_db "github.com/goodrain/rainbond/api/db"
	"github.com/goodrain/rainbond/api/handler/group"
	"github.com/goodrain/rainbond/api/handler/share"
	"github.com/goodrain/rainbond/builder/sources/registry"
	"github.com/goodrain/rainbond/cmd/api/option"
	"github.com/goodrain/rainbond/db"
	"github.com/goodrain/rainbond/pkg/generated/clientset/versioned"
	etcdutil "github.com/goodrain/rainbond/util/etcd"
	"github.com/goodrain/rainbond/worker/client"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"kubevirt.io/client-go/kubecli"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
	gateway "sigs.k8s.io/gateway-api/pkg/client/clientset/versioned/typed/apis/v1beta1"
)

// InitHandle 初始化handle
func InitHandle(conf option.Config,
	etcdClientArgs *etcdutil.ClientArgs,
	statusCli *client.AppRuntimeSyncClient,
	etcdcli *clientv3.Client,
	kubeClient *kubernetes.Clientset,
	rainbondClient versioned.Interface,
	k8sClient k8sclient.Client,
	config *rest.Config,
	mapper meta.RESTMapper,
	dynamicClient dynamic.Interface,
	gatewayClient *gateway.GatewayV1beta1Client,
	kubevirtCli kubecli.KubevirtClient,
	registryCli *registry.Registry,
) error {
	mq := api_db.MQManager{
		EtcdClientArgs: etcdClientArgs,
		DefaultServer:  conf.MQAPI,
	}
	mqClient, errMQ := mq.NewMQManager()
	if errMQ != nil {
		logrus.Errorf("new MQ manager failed, %v", errMQ)
		return errMQ
	}
	prometheusCli, err := prometheus.NewPrometheus(&prometheus.Options{
		Endpoint: conf.PrometheusEndpoint,
	})
	if err != nil {
		logrus.Errorf("new prometheus client failure, %v", err)
		return err
	}
	dbmanager := db.GetManager()
	defaultServieHandler = CreateManager(conf, mqClient, etcdcli, statusCli, prometheusCli, rainbondClient, kubeClient, kubevirtCli, dbmanager, registryCli)
	defaultPluginHandler = CreatePluginManager(mqClient)
	defaultAppHandler = CreateAppManager(mqClient)
	defaultTenantHandler = CreateTenManager(mqClient, statusCli, &conf, kubeClient, prometheusCli, k8sClient)
	defaultHelmHandler = CreateHelmManager(kubeClient, rainbondClient, config, mapper)
	defaultNetRulesHandler = CreateNetRulesManager(etcdcli)
	defaultCloudHandler = CreateCloudManager(conf)
	defaultAPPBackupHandler = group.CreateBackupHandle(mqClient, statusCli, etcdcli)
	defaultEventHandler = CreateLogManager(conf, etcdcli)
	shareHandler = &share.ServiceShareHandle{MQClient: mqClient, EtcdCli: etcdcli}
	pluginShareHandler = &share.PluginShareHandle{MQClient: mqClient, EtcdCli: etcdcli}
	if err := CreateTokenIdenHandler(conf); err != nil {
		logrus.Errorf("create token identification mannager error, %v", err)
		return err
	}
	defaultGatewayHandler = CreateGatewayManager(dbmanager, mqClient, etcdcli, gatewayClient, kubeClient)
	def3rdPartySvcHandler = Create3rdPartySvcHandler(dbmanager, statusCli)
	operationHandler = CreateOperationHandler(mqClient)
	batchOperationHandler = CreateBatchOperationHandler(mqClient, statusCli, operationHandler)
	defaultAppRestoreHandler = NewAppRestoreHandler()
	defPodHandler = NewPodHandler(statusCli)
	defClusterHandler = NewClusterHandler(kubeClient, conf.RbdNamespace, conf.GrctlImage, config, mapper, prometheusCli, rainbondClient, statusCli, dynamicClient, gatewayClient, mqClient)
	defaultVolumeTypeHandler = CreateVolumeTypeManger(statusCli)
	defaultEtcdHandler = NewEtcdHandler(etcdcli)
	defaultmonitorHandler = NewMonitorHandler(prometheusCli)
	defServiceEventHandler = NewServiceEventHandler()
	defApplicationHandler = NewApplicationHandler(statusCli, prometheusCli, rainbondClient, kubeClient, dynamicClient)
	defRegistryAuthSecretHandler = CreateRegistryAuthSecretManager(dbmanager, mqClient, etcdcli)
	defNodesHandler = NewNodesHandler(kubeClient, conf.RbdNamespace, config, mapper, prometheusCli)
	return nil
}

var defaultServieHandler ServiceHandler
var shareHandler *share.ServiceShareHandle
var pluginShareHandler *share.PluginShareHandle
var defaultmonitorHandler MonitorHandler

// GetMonitorHandle get monitor handler
func GetMonitorHandle() MonitorHandler {
	return defaultmonitorHandler
}

// GetShareHandle get share handle
func GetShareHandle() *share.ServiceShareHandle {
	return shareHandler
}

// GetPluginShareHandle get plugin share handle
func GetPluginShareHandle() *share.PluginShareHandle {
	return pluginShareHandler
}

// GetServiceManager get manager
func GetServiceManager() ServiceHandler {
	return defaultServieHandler
}

var defaultPluginHandler PluginHandler

// GetPluginManager get manager
func GetPluginManager() PluginHandler {
	return defaultPluginHandler
}

var defaultTenantHandler TenantHandler

// GetTenantManager get manager
func GetTenantManager() TenantHandler {
	return defaultTenantHandler
}

var defaultHelmHandler HelmHandler

// GetHelmManager get manager
func GetHelmManager() HelmHandler {
	return defaultHelmHandler
}

var defaultNetRulesHandler NetRulesHandler

// GetRulesManager get manager
func GetRulesManager() NetRulesHandler {
	return defaultNetRulesHandler
}

var defaultCloudHandler CloudHandler

// GetCloudManager get manager
func GetCloudManager() CloudHandler {
	return defaultCloudHandler
}

var defaultEventHandler EventHandler

// GetEventHandler get event handler
func GetEventHandler() EventHandler {
	return defaultEventHandler
}

var defaultAppHandler *AppAction

// GetAppHandler GetAppHandler
func GetAppHandler() *AppAction {
	return defaultAppHandler
}

var defaultAPPBackupHandler *group.BackupHandle

// GetAPPBackupHandler GetAPPBackupHandler
func GetAPPBackupHandler() *group.BackupHandle {
	return defaultAPPBackupHandler
}

var defaultGatewayHandler GatewayHandler

// GetGatewayHandler returns a default GatewayHandler
func GetGatewayHandler() GatewayHandler {
	return defaultGatewayHandler
}

var def3rdPartySvcHandler *ThirdPartyServiceHanlder

// Get3rdPartySvcHandler returns the defalut ThirdParthServiceHanlder
func Get3rdPartySvcHandler() *ThirdPartyServiceHanlder {
	return def3rdPartySvcHandler
}

var batchOperationHandler *BatchOperationHandler

// GetBatchOperationHandler get handler
func GetBatchOperationHandler() *BatchOperationHandler {
	return batchOperationHandler
}

var operationHandler *OperationHandler

// GetOperationHandler get handler
func GetOperationHandler() *OperationHandler {
	return operationHandler
}

var defaultAppRestoreHandler AppRestoreHandler

// GetAppRestoreHandler returns a default AppRestoreHandler
func GetAppRestoreHandler() AppRestoreHandler {
	return defaultAppRestoreHandler
}

var defPodHandler PodHandler

// GetPodHandler returns the defalut PodHandler
func GetPodHandler() PodHandler {
	return defPodHandler
}

var defaultEtcdHandler *EtcdHandler

// GetEtcdHandler returns the default etcd handler.
func GetEtcdHandler() *EtcdHandler {
	return defaultEtcdHandler
}

var defClusterHandler ClusterHandler

// GetClusterHandler returns the default cluster handler.
func GetClusterHandler() ClusterHandler {
	return defClusterHandler
}

var defNodesHandler NodesHandler

// GetNodesHandler returns the default cluster handler.
func GetNodesHandler() NodesHandler {
	return defNodesHandler
}

var defApplicationHandler ApplicationHandler

// GetApplicationHandler  returns the default tenant application handler.
func GetApplicationHandler() ApplicationHandler {
	return defApplicationHandler
}

var defServiceEventHandler *ServiceEventHandler

// GetServiceEventHandler -
func GetServiceEventHandler() *ServiceEventHandler {
	return defServiceEventHandler
}

var defRegistryAuthSecretHandler RegistryAuthSecretHandler

// GetRegistryAuthSecretHandler -
func GetRegistryAuthSecretHandler() RegistryAuthSecretHandler {
	return defRegistryAuthSecretHandler
}
