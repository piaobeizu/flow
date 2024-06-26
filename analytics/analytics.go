package analytics

import (
	"runtime"

	segment "github.com/segmentio/analytics-go"
	"github.com/sirupsen/logrus"
)

const segmentWriteKey string = "oU2iC4shRUBfEboaO0FDuDIUk49Ime92"

type publisher interface {
	Publish(string, map[string]interface{})
	Close()
}

// Client is an analytics client that implements the publisher interface
var Client publisher

// NullClient is a drop in non-functional analytics publisher
type NullClient struct{}

// Initialize does nothing
func (c *NullClient) Initialize() error {
	return nil
}

// Publish would send a tracking event
func (c *NullClient) Publish(event string, props map[string]interface{}) {
	if len(props) == 0 {
		logrus.Infof("analytics event %s - properties: nil", event)
	} else {
		logrus.Infof("analytics event %s - properties: %+v", event, props)
	}
}

// Close the analytics connection
func (c *NullClient) Close() {}

func InitAnalytics(segment bool) {
	Client = &NullClient{}
	if segment {
		mid, err := MachineID()
		if err != nil {
			panic(err)
		}
		client, err := NewClient(segmentWriteKey, mid)
		if err != nil {
			panic(err)
		}
		Client = client
	}
}

var ctx = &segment.Context{
	App: segment.AppInfo{
		Name:      "flowctl",
		Namespace: "flow",
	},
	OS: segment.OSInfo{
		Name: runtime.GOOS + " " + runtime.GOARCH,
	},
	Extra: map[string]interface{}{"direct": true},
}

// Client for the Segment.io analytics service
type SegmentClient struct {
	client    segment.Client
	machineID string
}

// NewClient returns a new segment analytics client
func NewClient(writeKey, machineID string) (*SegmentClient, error) {
	client, err := segment.NewWithConfig(writeKey, segment.Config{Verbose: true})
	if err != nil {
		return nil, err
	}
	return &SegmentClient{
		client:    client,
		machineID: machineID,
	}, nil
}

// Publish enqueues the sending of a tracking event
func (c *SegmentClient) Publish(event string, props map[string]interface{}) {
	logrus.Tracef("segment event %s - properties: %+v", event, props)
	err := c.client.Enqueue(segment.Track{
		Context:     ctx,
		AnonymousId: c.machineID,
		Event:       event,
		Properties:  props,
	})
	if err != nil {
		logrus.Debugf("failed to submit telemetry: %s", err)
	}
}

// Close the analytics connection
func (c *SegmentClient) Close() {
	c.client.Close()
}
