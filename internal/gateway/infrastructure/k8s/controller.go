package k8s

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/gateway/application"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/tools/cache"
)

// RouteController 监听 K8s CRD 并将其状态同步到网关应用层
type RouteController struct {
	dynamicClient dynamic.Interface
	appService    *application.GatewayService
	logger        *slog.Logger
	resource      schema.GroupVersionResource
}

func NewRouteController(client dynamic.Interface, appService *application.GatewayService, logger *slog.Logger) *RouteController {
	return &RouteController{
		dynamicClient: client,
		appService:    appService,
		logger:        logger.With("module", "k8s_controller"),
		resource: schema.GroupVersionResource{
			Group:    "gateway.wyf.com",
			Version:  "v1",
			Resource: "gatewayroutes",
		},
	}
}

// Start 启动监听循环
func (c *RouteController) Start(ctx context.Context) error {
	factory := dynamicinformer.NewDynamicSharedInformerFactory(c.dynamicClient, time.Minute*30)
	informer := factory.ForResource(c.resource).Informer()

	if _, err := informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			c.reconcile(obj.(*unstructured.Unstructured))
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			c.reconcile(newObj.(*unstructured.Unstructured))
		},
		DeleteFunc: func(obj interface{}) {
			c.remove(obj.(*unstructured.Unstructured))
		},
	}); err != nil {
		return fmt.Errorf("failed to add event handler: %w", err)
	}

	go informer.Run(ctx.Done())

	if !cache.WaitForCacheSync(ctx.Done(), informer.HasSynced) {
		return fmt.Errorf("failed to sync cache")
	}

	c.logger.Info("K8s GatewayRoute Controller started successfully")
	return nil
}

// reconcile 核心调和逻辑：将 K8s 的 YAML 状态映射为业务路由
func (c *RouteController) reconcile(u *unstructured.Unstructured) {
	spec := u.Object["spec"].(map[string]interface{})

	path := spec["path"].(string)
	method := spec["method"].(string)
	backend := spec["backend"].(string)

	c.logger.Info("reconciling route from K8s CRD", "name", u.GetName(), "path", path)

	// 调用应用层更新路由表（如果存在则更新，不存在则创建）
	// 这里体现了声明式：最终状态由 K8s 决定
	_, err := c.appService.SyncRoute(context.Background(), application.SyncRouteRequest{
		ExternalID: string(u.GetUID()),
		Name:       u.GetName(),
		Path:       path,
		Method:     method,
		Backend:    backend,
		Source:     "K8S_CRD",
	})

	if err != nil {
		c.logger.Error("failed to sync route from K8s", "error", err)
	}
}

func (c *RouteController) remove(u *unstructured.Unstructured) {
	c.logger.Info("removing route deleted from K8s", "name", u.GetName())
	if err := c.appService.DeleteRouteByExternalID(context.Background(), string(u.GetUID())); err != nil {
		c.logger.Error("failed to delete route from K8s", "error", err, "uid", u.GetUID())
	}
}
