package service

import (
	"context"

	paymentcommon "github.com/NexaCard/API/internal/payment/common"
)

// detachOutboundRequestContext 将出站请求从上游 HTTP 连接生命周期中解耦，
// 避免浏览器断开、页面跳转或第三方回调连接中断直接取消下游外部请求。
func detachOutboundRequestContext(parent context.Context) (context.Context, context.CancelFunc) {
	if parent == nil {
		return paymentcommon.WithDefaultTimeout(context.Background())
	}
	return paymentcommon.WithDefaultTimeout(context.WithoutCancel(parent))
}
