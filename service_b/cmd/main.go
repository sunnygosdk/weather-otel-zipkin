package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"

	"github.com/gorilla/mux"
	"github.com/sunnygosdk/weather-otel-zipkin/serviceb/internal/domain/application/dto"
	"github.com/sunnygosdk/weather-otel-zipkin/serviceb/internal/domain/entities"
	"github.com/sunnygosdk/weather-otel-zipkin/serviceb/internal/infrastructure/config"
	"github.com/sunnygosdk/weather-otel-zipkin/serviceb/internal/infrastructure/monitor"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"
	"go.opentelemetry.io/otel/trace"
)

var (
	cfg    *config.Conf
	tracer trace.Tracer
)

func main() {
	cfg = config.LoadConfig()

	setupOpenTelemetry()
	startServer()
}

func setupOpenTelemetry() {
	ot := monitor.NewOpenTelemetry()
	ot.ServiceName = "Service B"
	ot.ServiceVersion = "1"
	ot.ExporterEndpoint = fmt.Sprintf("%s/api/v2/spans", cfg.URLZipKin)
	tracer = ot.GetTracer()
}

func startServer() {
	r := mux.NewRouter()
	r.Use(otelmux.Middleware("Service B"))
	r.HandleFunc("/weather", getWeather).Methods("POST")

	log.Println("Listening on port :8081")

	if err := http.ListenAndServe(":8081", r); err != nil {
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
		httpError(w, http.StatusBadRequest, "Failed to unmarshal request body", err)
		return
	}

	if err = validateCEP(cep); err != nil {
		httpError(w, http.StatusUnprocessableEntity, "Invalid ZIP code", err)
		return
	}

	responseViaCEP, err := fetchViaCEP(ctx, cep)
	if responseViaCEP == nil {
		httpError(w, http.StatusNotFound, "Can not find ZIP code", err)
		return
	}

	if err != nil {
		httpError(w, http.StatusInternalServerError, "Failed to fetch ViaCEP", err)
		return
	}

	responseWeather, err := fetchWeather(ctx, responseViaCEP.Localidade)
	if err != nil {
		httpError(w, http.StatusInternalServerError, "Failed to fetch weather", err)
		return
	}

	tempF := responseWeather.Current.TempC*1.8 + 32
	tempK := responseWeather.Current.TempC + 273.15
	response := dto.ResponseDTO{
		City:  responseWeather.Location.Name,
		TempC: responseWeather.Current.TempC,
		TempF: tempF,
		TempK: tempK,
	}
	handleResponse(w, http.StatusOK, &response)
}

func validateCEP(cep entities.CEP) error {
	var validCEP = regexp.MustCompile(`^\d{5}-?\d{3}$`)
	if !validCEP.MatchString(cep.CEP) {
		return config.ErrZipCodeInvalido
	}
	return nil
}

func fetchViaCEP(ctx context.Context, cep entities.CEP) (*dto.ResponseViaCEPDTO, error) {
	ctx, span := tracer.Start(ctx, "request-via-cep")
	defer span.End()

	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("http://viacep.com.br/ws/%s/json/", cep.CEP), nil)
	if err != nil {
		return nil, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var response dto.ResponseViaCEPDTO
	if err = json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, err
	}

	if response.Cep == "" {
		return nil, config.ErrZipCodeNaoEncontrado
	}

	return &response, nil
}

func fetchWeather(ctx context.Context, city string) (*dto.ResponseWeatherDTO, error) {
	ctx, span := tracer.Start(ctx, "request-weather-api")
	defer span.End()

	weatherUrl := fmt.Sprintf("http://api.weatherapi.com/v1/current.json?key=%s&q=%s", "3841b81037a5427eb51191826241702", url.QueryEscape(city))
	req, err := http.NewRequestWithContext(ctx, "GET", weatherUrl, nil)
	if err != nil {
		return nil, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var response dto.ResponseWeatherDTO
	if err = json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, err
	}

	return &response, nil
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
		w.Write([]byte("Cannot find ZIP code"))
	default:
		w.Write([]byte("Internal Server Error"))
	}
}

func httpError(w http.ResponseWriter, statusCode int, message string, err error) {
	log.Printf("%s: %v", message, err)
	w.WriteHeader(statusCode)
	w.Write([]byte(message))
}
