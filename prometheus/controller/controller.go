package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	api "github.com/prometheus/client_golang/api"
	apiv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

type PrometheusController struct {
	beego.Controller
}

type DataSource struct {
	Url   string
	Token string
}

type status string

const (
	statusSuccess status = "success"
	statusError   status = "error"
)

// queryData is just a wrapper to be compatible with the Prometheus API.
type queryData struct {
	ResultType string      `json:"resultType"`
	Result     model.Value `json:"result"`
}

type Response struct {
	Status status      `json:"status"`
	Data   interface{} `json:"data,omitempty"`
	Error  string      `json:"error,omitempty"`
}

func (p *PrometheusController) getClient(dsInfo *DataSource) (apiv1.API, error) {
	cfg := api.Config{
		Address: dsInfo.Url,
	}

	client, err := api.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	return apiv1.NewAPI(client), nil
}

func parseTime(s string) (time.Time, error) {
	if t, err := strconv.ParseFloat(s, 64); err == nil {
		s, ns := math.Modf(t)
		ns = math.Round(ns*1000) / 1000
		return time.Unix(int64(s), int64(ns*float64(time.Second))), nil
	}
	if t, err := time.Parse(time.RFC3339Nano, s); err == nil {
		return t, nil
	}
	return time.Time{}, fmt.Errorf("cannot parse %q to a valid timestamp", s)
}

func parseDuration(s string) (time.Duration, error) {
	if d, err := strconv.ParseFloat(s, 64); err == nil {
		ts := d * float64(time.Second)
		if ts > float64(math.MaxInt64) || ts < float64(math.MinInt64) {
			return 0, fmt.Errorf("cannot parse %q to a valid duration. It overflows int64", s)
		}
		return time.Duration(ts), nil
	}
	if d, err := model.ParseDuration(s); err == nil {
		return time.Duration(d), nil
	}
	return 0, fmt.Errorf("cannot parse %q to a valid duration", s)
}

func (p *PrometheusController) MonitorPod() {
	cluster := p.GetString(":cluster")
	namespace := p.GetString(":namespace")
	pod := p.GetString(":pod")
	logs.Info("cluster: %s, namespace: %s, pod: %s", cluster, namespace, pod)

	client, err := p.getClient(&DataSource{Url: "http://localhost:9090"})
	if err != nil {
		logs.Error("Get Prometheus client failed: %v", err)
	}

	r := p.Ctx.Request

	start, err := parseTime(r.FormValue("start"))
	if err != nil {
		logs.Error("Parse start time failed: %v", err)
	}

	end, err := parseTime(r.FormValue("end"))
	if err != nil {
		logs.Error("Parse end time failed: %v", err)
	}

	if end.Before(start) {
		logs.Error("End before start")
	}

	step, err := parseDuration(r.FormValue("step"))
	if err != nil {
		logs.Error("Parse step failed: %v", err)
	}

	if step <= 0 {
		logs.Error("Zero or negative query resolution step width are not accepted")
	}

	timeRange := apiv1.Range{
		Start: start,
		End:   end,
		Step:  step,
	}
	data := model.Matrix{}

	queries := []string{"up", "process_start_time_seconds"}

	for _, query := range queries {
		value, err := client.QueryRange(context.Background(), query, timeRange)
		if err != nil {
			logs.Error("Query Prometheus failed: %v", err)
		}
		matrix, ok := value.(model.Matrix)
		if !ok {
			logs.Error("The type of QueryRange value is unexpected")
		}
		// Actually the maxtrix length should always be 1.
		if len(matrix) == 0 {
			logs.Error("The length of QueryRange value is 0")
		}
		// Add name label to let frontend know the meaning of corresponding samples.
		matrix[0].Metric["name"] = "cpu_usage"
		data = append(data, matrix[0])
	}

	b, err := json.Marshal(&Response{
		Status: statusSuccess,
		Data: &queryData{
			ResultType: "matrix",
			Result:     data,
		},
	})
	if err != nil {
		logs.Error("Marshal response failed: %v", err)
	}

	w := p.Ctx.ResponseWriter
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if n, err := w.Write(b); err != nil {
		logs.Error("Write response body failed: %v, bytesWritten: %v", err, n)
	}
}

func (p *PrometheusController) MonitorNode() {

}
