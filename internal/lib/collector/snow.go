package collector

import (
	"context"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/zebbra/snow-exporter/internal/lib/snow"
	"go.uber.org/zap"
)

type SNOWCollector struct {
	Cache         *cache.Cache
	Client        *snow.Client
	Logger        *zap.SugaredLogger
	ErrorCounter  *Counter
	ScrapeCounter *Counter
}

func (c *SNOWCollector) Run(ctx context.Context) error {
	c.Logger.Infow("Refresh incident list")
	startTime := time.Now()

	incidents, err := c.Client.Incident(ctx)

	if err != nil {
		c.Logger.Errorw(
			"Error fetching incident list",
			"error", err,
		)

		c.ErrorCounter.Inc()
		return err
	}

	c.Cache.Set("incidents", incidents, cache.DefaultExpiration)
	c.Logger.Infow(
		"Refresh done",
		"duration",
		time.Now().Sub(startTime),
	)

	return nil
}

func (c *SNOWCollector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(c, ch)
}

func (c *SNOWCollector) Collect(ch chan<- prometheus.Metric) {
	incidents := []snow.Incident{}

	if d, found := c.Cache.Get("incidents"); found {
		incidents = d.([]snow.Incident)
	} else {
		return
	}

	// general stats
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			"snow_incident_count",
			"Number of incidents in SNOW",
			[]string{},
			nil,
		),
		prometheus.GaugeValue,
		float64(len(incidents)),
	)

	for _, i := range incidents {
		// incident info
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc(
				"snow_incident_info",
				"Info about a SNOW incident",
				[]string{"ShortDescription", "ClosedAt", "Description", "Priority", "ChildIncidents", "SysId", "Number", "OpenedAt", "ResolvedBy", "CallerId", "Location", "State", "AssignedTo"},
				nil,
			),
			prometheus.GaugeValue,
			1.0,
			i.ShortDescription,
			i.ClosedAt,
			i.Description,
			i.Priority,
			i.ChildIncidents,
			i.SysId,
			i.Number,
			i.OpenedAt,
			i.ResolvedBy,
			i.CallerId,
			i.Location,
			i.State,
			i.AssignedTo,
		)
	}

}
