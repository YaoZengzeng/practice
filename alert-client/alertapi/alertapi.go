package alertapi

import (
	"net/http"
	"bytes"
	"fmt"
	"time"
	"context"
	"encoding/json"
)

const (
	apiPrefix = "/api/v1"

	epAlerts   = apiPrefix + "/alerts"
)

// Alert represents an alert as expected by the AlertManager's push alert API.
type Alert struct {
	Labels       LabelSet  `json:"labels"`
	Annotations  AnnotationSet  `json:"annotations"`
	StartsAt     time.Time `json:"startsAt,omitempty"`
	EndsAt       time.Time `json:"endsAt,omitempty"`
	GeneratorURL string    `json:"generatorURL"`
}

// LabelSet represents a collection of label names and values as a map.
type LabelSet map[LabelName]LabelValue

// LabelName represents the name of a label.
type LabelName string

// LabelValue represents the value of a label.
type LabelValue string

// AnnotationSet represents a collection of annotation names and values as a map.
type AnnotationSet map[AnnotationName]AnnotationValue

// AnnotationName represents the name of a annotation.
type AnnotationName string

// AnnotationValue represents the value of a annotation.
type AnnotationValue string

func NewAlert(labels LabelSet, annotations AnnotationSet) *Alert {
	return &Alert{
		Labels:			labels,
		Annotations:	annotations,
	}
}

func (a *Alert) Push(times ...time.Time) error {
	alert := *a

	switch len(times) {
	case 1:
		alert.StartsAt = times[0]
	case 2:
		alert.StartsAt = times[0]
		alert.EndsAt = times[1]
	}

	return defaultAlertAPI.Push(context.Background(), alert)
}

// AlertAPI provides bindings for the Alertmanager's alert API.
type AlertAPI interface {
	// Push sends a list of alerts to the Alertmanager.
	Push(ctx context.Context, alerts ...Alert) error
}

type httpAlertAPI struct {
	client Client
}

// NewAlertAPI returns a new AlertAPI for the client.
func NewAlertAPI(c Client) AlertAPI {
	return newAlertAPI(c)
}

func newAlertAPI(c Client) AlertAPI {
	return &httpAlertAPI{client: c}
}

func (h *httpAlertAPI) Push(ctx context.Context, alerts ...Alert) error {
	u := h.client.URL(epAlerts, nil)

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(&alerts); err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, u.String(), &buf)
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}

	_, _, err = h.client.Do(ctx, req)
	return err
}

var defaultAlertAPI AlertAPI

func Init(address string, roundTripper http.RoundTripper) {
	client, _ := NewClient(Config{
		Address:		address,
		RoundTripper:	roundTripper,
	})

	defaultAlertAPI = newAlertAPI(client)
}
