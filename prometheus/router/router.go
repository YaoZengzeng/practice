package routers

import (
	"github.com/YaoZengzeng/practice/prometheus/controller"

	"github.com/astaxie/beego"
)

func init() {
	controller := controller.NewPrometheusController()
	beego.Router("/backend/prometheus/clusters/:cluster/namespaces/:namespace/pods/:pod", controller, "*:MonitorPod")
	beego.Router("/backend/prometheus/clusters/:cluster/namespaces/:namespace/pods/:pod/metrics", controller, "*:PodMetrics")
	beego.Router("/backend/prometheus/clusters/:cluster/namespaces/:namespace/pods/:pod/metrics-records", controller, "*:PodMetricsRecords")
	beego.Router("/backend/prometheus/clusters/:cluster/namespaces/:namespace/pods/:pod/series", controller, "*:PodSeries")
	beego.Router("/backend/prometheus/clusters/:cluster/nodes/:node", controller, "*:MonitorNode")
}
