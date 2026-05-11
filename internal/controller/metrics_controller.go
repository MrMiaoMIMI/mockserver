package controller

import (
	"github.com/MrMiaoMIMI/goshared/util/serverresp"
	"github.com/gin-gonic/gin"

	"github.com/MrMiaoMIMI/mockserver/internal/observability"
)

type MetricsController struct {
	runtimeMetrics *observability.RuntimeMetrics
}

func NewMetricsController(runtimeMetrics *observability.RuntimeMetrics) *MetricsController {
	return &MetricsController{runtimeMetrics: runtimeMetrics}
}

func (c *MetricsController) RuntimeMetrics(ctx *gin.Context) {
	serverresp.Success(ctx, c.runtimeMetrics.Snapshot())
}
