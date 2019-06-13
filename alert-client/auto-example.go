package main

import (
	"time"

	"github.com/YaoZengzeng/practice/alert-client/alertapi"
)

func init() {
	// 用Alertmanager的地址以及一个可选的http.RoundTripper初始化alertapi
	alertapi.Init("http://127.0.0.1:9093", nil)
}

// 定义一系列Alert
var (
	diskRunningFull = alertapi.NewAlert(
		// 一组标签唯一地表示了一个alert，一般必须要有key为"alertname"的label
		// 表示alert的名字，其余的label表示alert的属性，可任意添加，为后续扩展设计考虑，
		// 建议当前仅配置"alertname"这唯一一个label
		alertapi.LabelSet{
			"alertname": "DiskRunningFull",
		},
		// annotations同样是一组键值对，它作为附加信息，可以在其中包含关于
		// alert的详细信息
		alertapi.AnnotationSet{
			"info": "The disk sda1 is running full",
			"summary": "please check the instance example1",
		},
	)
	memoryRunningFull = alertapi.NewAlert(
		alertapi.LabelSet{
			"alertname": "DiskRunningFull",
		},
		nil,		
	)
)

func main() {
	diskFull, memoryFull := true, true

	now := time.Now()

	// 当条件满足时触发报警
	if diskFull {
		// 可以指定0，1，2（超过部分自动忽略）个时间参数
		// 若存在，第一个参数总为报警的起始时间，若不指定，则alert manager会将接收到该报警的时间作为起始时间
		// 第二个参数表示报警的结束时间，一般要大于起始时间，若不指定，alert manager会将它设置为起始时间加上
		// 默认的resolved time
		diskRunningFull.Push(now, now.Add(time.Duration(5 * time.Minute)))
	}

	if memoryFull {
		// 起始时间，结束时间都不指定
		memoryRunningFull.Push()
	}
}
