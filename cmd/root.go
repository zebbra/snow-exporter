package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	"github.com/zebbra/snow-exporter/internal/lib/collector"
	"github.com/zebbra/snow-exporter/internal/lib/snow"
	"github.com/zebbra/snow-exporter/internal/lib/version"
	"go.uber.org/zap"
)

const userEnv = "SNOW_USER"
const passwordEnv = "SNOW_PASSWORD"

var rootCmd = &cobra.Command{
	Use:           "snow-exporter --snow.endpoint <url>",
	SilenceErrors: true,
	Version:       fmt.Sprintf("%s-%s", version.Version, version.Commit),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if os.Getenv(userEnv) == "" || os.Getenv(passwordEnv) == "" {
			return fmt.Errorf(
				"Please provide SNOW credentials in environment variables %s and %s",
				userEnv,
				passwordEnv,
			)
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		endpoint, err := cmd.Flags().GetString("snow.endpoint")

		if err != nil {
			return err
		}

		addr, err := cmd.Flags().GetString("web.listen-address")

		if err != nil {
			return err
		}

		metricsPath, err := cmd.Flags().GetString("web.metrics-path")

		if err != nil {
			return err
		}

		scrapeInterval, err := cmd.Flags().GetDuration("scrape.interval")

		if err != nil {
			return err
		}

		logger, _ := zap.NewProduction()
		defer logger.Sync()
		sugar := logger.Sugar()

		snowClient := snow.NewClient(
			endpoint,
			os.Getenv(userEnv),
			os.Getenv(passwordEnv),
		)

		reg := prometheus.NewPedanticRegistry()
		_ = reg.Register(collectors.NewBuildInfoCollector())
		_ = reg.Register(collectors.NewGoCollector())
		_ = reg.Register(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))

		mainCache := cache.New(5*scrapeInterval, 10*scrapeInterval)

		sugar.Infof("Start initial data collection")

		ctx := context.Background()

		errorCounter := collector.Counter(0)

		sc := &collector.StatisticsCollector{
			Logger:       sugar,
			Cache:        mainCache,
			ErrorCounter: &errorCounter,
		}

		_ = sc.Run(ctx)
		_ = reg.Register(sc)

		snc := &collector.SNOWCollector{
			Logger:       sugar,
			Client:       snowClient,
			Cache:        mainCache,
			ErrorCounter: &errorCounter,
		}

		_ = snc.Run(ctx)
		_ = reg.Register(snc)

		sugar.Infof("Start collector threads")
		scraper := time.NewTicker(scrapeInterval + scrapeInterval/2)
		go func() {
			for {
				select {
				case <-scraper.C:
					go func() {
						ctx, cancel := context.WithTimeout(ctx, scrapeInterval)
						defer cancel()
						_ = snc.Run(ctx)
					}()
				}
			}
		}()

		http.Handle(metricsPath, promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))

		http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			max, _ := cmd.Flags().GetInt("scrape.max-errors")

			if errorCounter.Get() > int64(max) {
				w.WriteHeader(500)
				_, _ = w.Write([]byte("Unhealthy"))
				return
			}

			_, _ = w.Write([]byte("OK"))
		})

		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte(`
            <html>
            <head><title>SNOW Exporter Metrics</title></head>
            <body>
            <h1>Metrics</h1>
            <p><a href='` + metricsPath + `'>Metrics</a></p>
            </body>
            </html>
        `))
		})

		sugar.Infof("Start listening for connections on %s", addr)
		return http.ListenAndServe(addr, nil)
	},
}

// Execute runs root command
func Execute() {
	rootCmd.Flags().String("snow.endpoint", "", "URL of SNOW API")
	_ = rootCmd.MarkFlagRequired("snow.endpoint")

	rootCmd.Flags().String("web.listen-address", ":9911", "Address on which to expose metrics and web interface.")
	rootCmd.Flags().String("web.metrics-path", "/metrics", "Path under which to expose metrics.")

	rootCmd.Flags().Duration("scrape.interval", 15*time.Second, "Polling interval")
	rootCmd.Flags().Int("scrape.max-errors", 25, "Max scrape errors before reporting exporter as unhealthy")

	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
