package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"ocm/internal/model"
	"ocm/internal/storage"
)

type Dependencies struct {
	Store *storage.Store
}

func New(dep Dependencies) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		_ = r
		writeJSON(w, http.StatusOK, map[string]string{"ok": "true"})
	})

	mux.HandleFunc("GET /services", func(w http.ResponseWriter, r *http.Request) {
		svcs, err := dep.Store.ListServices(r.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusOK, svcs)
	})

	mux.HandleFunc("GET /overview", func(w http.ResponseWriter, r *http.Request) {
		svcs, err := dep.Store.ListServices(r.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		csaSum := 0.0
		csaCount := 0
		ocmSum := 0.0
		ocmCount := 0
		for _, s := range svcs {
			series, err := dep.Store.GetMetricSeries(r.Context(), s.ID, model.MetricCSA)
			if err != nil {
				writeError(w, http.StatusInternalServerError, err)
				return
			}
			if len(series) == 0 {
				// no CSA
			} else {
				v := series[len(series)-1].Value
				csaSum += v
				csaCount++
			}

			scores, err := dep.Store.GetScoreSeries(r.Context(), s.ID)
			if err != nil {
				writeError(w, http.StatusInternalServerError, err)
				return
			}
			if len(scores) == 0 {
				continue
			}
			ocmSum += scores[len(scores)-1].Score
			ocmCount++
		}
		var csaAvg *float64
		if csaCount > 0 {
			v := csaSum / float64(csaCount)
			csaAvg = &v
		}
		var ocmAvg *float64
		if ocmCount > 0 {
			v := ocmSum / float64(ocmCount)
			ocmAvg = &v
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"serviceCount":    len(svcs),
			"servicesWithCSA": csaCount,
			"csaSum":          csaSum,
			"csaAvg":          csaAvg,
			"servicesWithOCM": ocmCount,
			"ocmSum":          ocmSum,
			"ocmAvg":          ocmAvg,
		})
	})

	mux.HandleFunc("GET /services/", func(w http.ResponseWriter, r *http.Request) {
		// /services/{id}/metrics/{metricType}
		// /services/{id}/metrics/{metricType}/evidence
		// /services/{id}/scores
		path := strings.TrimPrefix(r.URL.Path, "/services/")
		parts := strings.Split(strings.Trim(path, "/"), "/")
		if len(parts) < 2 {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
			return
		}
		id, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid service id"})
			return
		}
		switch parts[1] {
		case "scores":
			series, err := dep.Store.GetScoreSeries(r.Context(), id)
			if err != nil {
				writeError(w, http.StatusInternalServerError, err)
				return
			}
			writeJSON(w, http.StatusOK, series)
			return
		case "metrics":
			if len(parts) == 3 {
				mt := model.MetricType(parts[2])
				series, err := dep.Store.GetMetricSeries(r.Context(), id, mt)
				if err != nil {
					writeError(w, http.StatusInternalServerError, err)
					return
				}
				writeJSON(w, http.StatusOK, series)
				return
			}
			if len(parts) == 4 && parts[3] == "evidence" {
				mt := model.MetricType(parts[2])
				evs, err := dep.Store.ListEvidenceLatest(r.Context(), id, mt)
				if err != nil {
					writeError(w, http.StatusInternalServerError, err)
					return
				}
				writeJSON(w, http.StatusOK, evs)
				return
			}
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "expected /services/{id}/metrics/{type} or /services/{id}/metrics/{type}/evidence"})
			return
		default:
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
			return
		}
	})

	return mux
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{"error": err.Error()})
}
