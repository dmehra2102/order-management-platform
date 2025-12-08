package metrics

import "github.com/prometheus/client_golang/prometheus"

type Metrics struct {
	OrdersCreated    prometheus.Counter
	OrdersConfirmed  prometheus.Counter
	OrdersFailed     prometheus.Counter
	OrderProcessTime prometheus.Histogram
	KafkaErrors      prometheus.Counter
	DBErrors         prometheus.Counter
}

func New() *Metrics {
	return &Metrics{
		OrdersCreated: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "orders_created_total",
			Help: "Total number of orders created",
		}),
		OrdersConfirmed: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "orders_confirmed_total",
			Help: "Total number of orders confirmed",
		}),
		OrdersFailed: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "orders_failed_total",
			Help: "Total number of orders failed",
		}),
		OrderProcessTime: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "order_process_duration_seconds",
			Help:    "Time taken to process orders",
			Buckets: []float64{0.1, 0.5, 1, 2, 5},
		}),
		KafkaErrors: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "kafka_errors_total",
			Help: "Total Kafka errors",
		}),
		DBErrors: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "db_errors_total",
			Help: "Total database errors",
		}),
	}
}

func (m *Metrics) Register() error {
	if err := prometheus.Register(m.OrdersCreated); err != nil {
		return err
	}
	if err := prometheus.Register(m.OrdersConfirmed); err != nil {
		return err
	}
	if err := prometheus.Register(m.OrdersFailed); err != nil {
		return err
	}
	if err := prometheus.Register(m.OrderProcessTime); err != nil {
		return err
	}
	if err := prometheus.Register(m.KafkaErrors); err != nil {
		return err
	}
	if err := prometheus.Register(m.DBErrors); err != nil {
		return err
	}
	return nil
}
