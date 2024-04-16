package main

import (
	"DZ_ITOG/cmd"
	configs "DZ_ITOG/config"
	"DZ_ITOG/handlers"
	"DZ_ITOG/repo"
	"database/sql"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

func initMetrics() {
	promauto.NewCounter(prometheus.CounterOpts{
		Name: "myapp_processed_ops_total",
		Help: "The total number of processed events",
	})

	promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "myapp_response_latency_seconds",
		Help:    "Histogram of response latency (seconds)",
		Buckets: prometheus.DefBuckets,
	})
}
func DatabaseMiddleware(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if db == nil {
			logrus.Error("Database is nil")
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Database instance not available"})
			return
		}
		c.Set("db", db)
		c.Next()
	}
}

func main() {
	initMetrics()

	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.InfoLevel)

	//http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	config, err := configs.LoadConfig("./config/config.yaml")
	if err != nil {
		logrus.Fatalf("Cannot load config: %v", err)
	}

	db, err := repo.InitDB(config)
	if err != nil {
		logrus.Fatalf("Database initialization failed: %v", err)
	}

	router := gin.Default()
	router.Use(DatabaseMiddleware(db))
	router.POST("/transactions", handlers.CreateTransaction)
	router.GET("/transactions", handlers.GetAllTransactions)
	router.GET("/transactions/:id", handlers.GetTransactionByID)
	router.PUT("/transactions/:id", handlers.UpdateTransaction)
	router.DELETE("/transactions/:id", handlers.DeleteTransaction)
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))
	router.Run(":8080")

	cmd.Run(db)
}
