package server

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// ОТДЕЛЬНЫЙ registry БЕЗ default collectors (NewGoCollector/NewProcessCollector НЕ регистрируем):
// при federation router-proxy конкатенирует /metrics promptlint + footer-proxy простой склейкой,
// а дефолтные process_*/go_* имена совпали бы у обоих процессов -> дубли -> scrape падает. Здесь
// экспонируем ТОЛЬКО бизнес-метрики promptlint -> concat безопасен.
var (
	metricsReg = prometheus.NewRegistry()

	analyzeTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "promptlint_analyze_total",
		Help: "Total /analyze requests processed.",
	})
	domainTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "promptlint_domain_total",
		Help: "Prompt classifications by dominant domain.",
	}, []string{"domain"})
	complexityTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "promptlint_complexity_total",
		Help: "Prompt classifications by complexity level.",
	}, []string{"level"})
)

func init() {
	metricsReg.MustRegister(analyzeTotal, domainTotal, complexityTotal)
}

// metricsHandler — /metrics из отдельного registry (только бизнес-метрики, без runtime-дефолтов).
func metricsHandler() http.Handler {
	return promhttp.HandlerFor(metricsReg, promhttp.HandlerOpts{})
}

// recordAnalyze инкрементит счётчики классификации. domain — ДОМИНИРУЮЩИЙ домен (argmax вектора).
func recordAnalyze(domain, complexity string) {
	analyzeTotal.Inc()
	if domain != "" {
		domainTotal.WithLabelValues(domain).Inc()
	}
	if complexity != "" {
		complexityTotal.WithLabelValues(complexity).Inc()
	}
}

// topDomain — домен с максимальным score (argmax). Тай-брейк по имени -> детерминизм.
func topDomain(scores map[string]float64) string {
	best, bestScore := "", -1.0
	for d, s := range scores {
		if s > bestScore || (s == bestScore && (best == "" || d < best)) {
			bestScore = s
			best = d
		}
	}
	return best
}
