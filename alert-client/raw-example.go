package main

import (
	"fmt"
	"time"
	"context"

	"github.com/YaoZengzeng/practice/alert-client/alertapi"
)

func main() {
	// 构建alertmanager的HTTP client
	c, err := alertapi.NewClient(alertapi.Config{
		// alertmanager的地址
		Address: "http://127.0.0.1:9093",
	})
	if err != nil {
		fmt.Printf("Construct alertmanager client failed: %v", err)
		return
	}


	client := alertapi.NewAlertAPI(c)

	now := time.Now()
	// 构建一系列的alert
	alerts := []alertapi.Alert{
		{
			Labels:	alertapi.LabelSet{
				// 一组标签唯一地表示了一个alert，一般必须要有key为"alertname"的label
				// 表示alert的名字，其余的label表示alert的属性，可任意添加
				"alertname": "DiskRunningFull",
				"dev": "sda1",
				"instance": "example1",
				// 一般可以指定key为"severity"的label表示告警级别
				"severity": "warning",
			},
			Annotations: alertapi.AnnotationSet{
				// annotations同样是一组键值对，它作为附加信息，可以在其中包含关于
				// alert的详细信息
				"info": "The disk sda1 is running full",
				"summary": "please check the instance example1",
			},
			// StartsAt表示alert的起始时间，若不指定，则alert manager会将接收到该
			// alert的时间作为起始时间
			StartsAt: now,
			// EndsAt表示alert的结束时间，一般要大于StartsAt，若不指定，alert manager
			// 会将它设置为StartsAt加上默认的resolved time
			EndsAt: now.Add(time.Duration(5 * time.Minute)),
		},
		{
			Labels:	alertapi.LabelSet{
				"alertname": "MemoryRunningFull",
				"severity": "critical",
			},
		},
	}

	err = client.Push(context.Background(), alerts...)
	if err != nil {
		fmt.Printf("Push alerts to alert manager failed: %v\n", err)
	}
}
