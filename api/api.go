package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"
)

// API is the handler for the API
type API struct {
	log *logrus.Logger
}

// NewApi creates a new API
func NewApi(logger *logrus.Logger) *API {
	if logger == nil {
		logger = logrus.New()
		logger.SetLevel(logrus.TraceLevel)
		logger.SetReportCaller(true)
		logger.SetFormatter(&logrus.TextFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
			FullTimestamp:   true,
			PadLevelText:    true,
			CallerPrettyfier: func(frame *runtime.Frame) (function string, file string) {
				return "", fmt.Sprintf(" %s:%d", frame.File, frame.Line)
			},
		})
	}

	return &API{logger}
}

// Router returns the router for the API
func (a *API) Router() func(r chi.Router) {
	return func(r chi.Router) {
		r.Use(a.headersMiddleware)

		r.Get("/online", a.online)
		r.Get("/admins", a.admins)
		r.Get("/user", a.users)
		r.Get("/user/{user}", a.user)
	}
}

func (a *API) headersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

type Response struct {
	Message string `json:",omitempty"`
	Error   string `json:",omitempty"`
}

// write writes the response
func write(w http.ResponseWriter, statusCode int, data interface{}) {
	if statusCode != http.StatusOK {
		w.WriteHeader(statusCode)
	}

	if data == nil {
		if statusCode == http.StatusOK {
			data = Response{Message: http.StatusText(statusCode)}
		} else {
			data = Response{Error: http.StatusText(statusCode)}
		}
	}

	_ = json.NewEncoder(w).Encode(data)
}

const apiHref = "https://training-server.com/api"

func (c *API) online(w http.ResponseWriter, r *http.Request) {
	b, status, _ := c.trainingApiGet(fmt.Sprintf("%s/online", apiHref), r.Context())
	write(w, status, b)
}

func (c *API) users(w http.ResponseWriter, r *http.Request) {
	b, status, _ := c.trainingApiGet(fmt.Sprintf("%s/user", apiHref), r.Context())
	write(w, status, b)
}

func (c *API) admins(w http.ResponseWriter, r *http.Request) {
	b, status, _ := c.trainingApiGet(fmt.Sprintf("%s/admin", apiHref), r.Context())
	write(w, status, b)
}

func (c *API) user(w http.ResponseWriter, r *http.Request) {
	b, status, _ := c.trainingApiGet(fmt.Sprintf("%s/user/%s", apiHref, chi.URLParam(r, "user")), r.Context())
	write(w, status, b)
}

func (c *API) trainingApiGet(href string, requestCtx context.Context) (interface{}, int, error) {
	ctx, cancel := context.WithTimeout(requestCtx, 10*time.Second)
	defer cancel()
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, href, nil)
	if err != nil {
		return Response{Error: err.Error()}, http.StatusInternalServerError, err
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return Response{Error: err.Error()}, http.StatusInternalServerError, err
	}

	defer func() {
		_ = response.Body.Close()
	}()

	if response.StatusCode != http.StatusOK {
		return Response{Error: response.Status}, response.StatusCode, nil
	}

	b, err := io.ReadAll(response.Body)
	if err != nil {
		c.log.Error(err)
		return Response{Error: err.Error()}, http.StatusInternalServerError, err
	}

	var a interface{}
	err = json.Unmarshal(b, &a)
	if err != nil {
		return Response{Error: err.Error()}, http.StatusInternalServerError, err
	}

	return a, response.StatusCode, nil
}
