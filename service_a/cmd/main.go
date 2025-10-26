package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"

	"github.com/gorilla/mux"
	"github.com/sunnygosdk/weather-otel-zipkin/servicea/internal/application/dto"
	"github.com/sunnygosdk/weather-otel-zipkin/servicea/internal/domain/entities"
	"github.com/sunnygosdk/weather-otel-zipkin/servicea/internal/infrastructure/config"
	"github.com/sunnygosdk/weather-otel-zipkin/servicea/internal/infrastructure/monitor"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/trace"
)

var (
	cfg    *config.Config
	tracer trace.Tracer
)

func main() {
	cfg = config.SetConfig()

	setupOpenTelemetry()
	startServer()
}

func setupOpenTelemetry() {
	ot := monitor.NewOpenTelemetry()
	ot.ServiceName = "Service A"
	ot.ServiceVersion = "1"
	ot.ExporterEndpoint = fmt.Sprintf("%s/api/v2/spans", cfg.URLZipKin)

	tracer = ot.GetTracer()
}

func startServer() {
	r := mux.NewRouter()
	r.Use(otelmux.Middleware("Service A"))
	r.HandleFunc("/weather", getWeather).Methods("POST")

	log.Printf("Service B URL: %s", cfg.URLServiceB)
	log.Println("Listening on port :8080")

	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func getWeather(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "validate-zipcode")
	defer span.End()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		httpError(w, http.StatusBadRequest, "Failed to read request body", err)
		return
	}

	var cep entities.CEP
	if err = json.Unmarshal(body, &cep); err != nil {
		httpError(w, http.StatusUnprocessableEntity, "Invalid ZIP code", err)
		return
	}

	if err = validateCEP(cep); err != nil {
		httpError(w, http.StatusUnprocessableEntity, "Invalid ZIP code", err)
		return
	}

	response, statusCode, err := requestServiceB(ctx, body)
	if err != nil {
		httpError(w, http.StatusUnprocessableEntity, config.ErrZipCodeInvalido.Error(), err)
		return
	}

	handleResponse(w, statusCode, response)
}

func validateCEP(cep entities.CEP) error {
	var validCEP = regexp.MustCompile(`^\d{5}-?\d{3}$`)
	if !validCEP.MatchString(cep.Cep) {
		return config.ErrZipCodeInvalido
	}

	return nil
}

func requestServiceB(ctx context.Context, body []byte) (*dto.ResponseDTO, int, error) {
	ctx, span := tracer.Start(ctx, "request-service-b")
	defer span.End()

	client := http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)}
	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/weather", cfg.URLServiceB), bytes.NewBuffer(body))
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Printf("Failed to close response body: %v", err)
		}
	}(res.Body)

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	var response dto.ResponseDTO
	if err = json.Unmarshal(resBody, &response); err != nil {
		return nil, http.StatusInternalServerError, err
	}

	return &response, res.StatusCode, nil
}

func handleResponse(w http.ResponseWriter, statusCode int, response *dto.ResponseDTO) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if statusCode == http.StatusOK {
		if err := json.NewEncoder(w).Encode(response); err != nil {
			httpError(w, http.StatusInternalServerError, "Failed to marshal response", err)
		}
		return
	}

	switch statusCode {
	case http.StatusUnprocessableEntity:
		w.Write([]byte("Invalid ZIP code"))
	case http.StatusNotFound:
		w.Write([]byte("ZIP code not found"))
	default:
		w.Write([]byte("Internal Server Error"))
	}
}

func httpError(w http.ResponseWriter, statusCode int, message string, err error) {
	log.Printf("%s: %v", message, err)
	w.WriteHeader(statusCode)
	w.Write([]byte(message))
}
