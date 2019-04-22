package routers

import (
	"github.com/YaoZengzeng/practice/prometheus/controller"

	"github.com/astaxie/beego"
)

func init() {
	beego.Router("/backend/prometheus/clusters/:cluster/namespaces/:namespace/pods/:pod", &controller.PrometheusController{}, "*:MonitorPod")
	beego.Router("/backend/prometheus/clusters/:cluster/nodes/:node", &controller.PrometheusController{}, "*:MonitorNode")
}
