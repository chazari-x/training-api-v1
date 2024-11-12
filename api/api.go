package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"strconv"
	"time"

	"github.com/chazari-x/training-api-v1/model"
	"github.com/chazari-x/training-api-v1/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-pg/pg/v10"
	"github.com/sirupsen/logrus"
)

// API is the handler for the API
type API struct {
	log *logrus.Logger
	s   *storage.Storage
}

// NewApi creates a new API
func NewApi(logger *logrus.Logger, storage *storage.Storage) *API {
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

	return &API{logger, storage}
}

// Router returns the router for the API
func (c *API) Router() func(r chi.Router) {
	return func(r chi.Router) {
		r.Use(c.headersMiddleware)

		r.Get("/online", c.online)
		r.Get("/admins", c.admins)
		r.Get("/user", c.users)
		r.Get("/user/{user}", c.user)
		r.Get("/v2/user", c.v2user)
	}
}

func (c *API) headersMiddleware(next http.Handler) http.Handler {
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

func (c *API) v2user(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Get("nickname") != "" {
		response, err := http.Get(fmt.Sprintf("%s/user/%s", apiHref, r.URL.Query().Get("nickname")))
		if err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		defer func() {
			_ = response.Body.Close()
		}()

		if response.StatusCode != http.StatusOK {
			http.Error(w, response.Status, response.StatusCode)
			return
		}

		body, err := io.ReadAll(response.Body)
		if err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		var user struct {
			Data model.ApiUser `json:"data"`
		}
		if err = json.Unmarshal(body, &user); err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		if u, err := c.s.SelectById(user.Data.ID); err == nil {
			var longUser = model.LongUser{
				ID:                      user.Data.ID,
				Login:                   user.Data.Login,
				Access:                  user.Data.Access,
				Moder:                   user.Data.Moder,
				Verify:                  user.Data.Verify,
				VerifyTxt:               user.Data.VerifyTxt,
				Mute:                    user.Data.Mute,
				Online:                  user.Data.Online,
				PlayerID:                user.Data.PlayerID,
				RegDate:                 user.Data.RegDate,
				LastLogin:               user.Data.LastLogin,
				Warn:                    user.Data.Warn,
				Avatar:                  u.Avatar,
				Background:              u.Background,
				VIP:                     u.VIP,
				SocialCredits:           u.SocialCredits,
				Kills:                   u.Kills,
				Deaths:                  u.Deaths,
				CopChaseRating:          u.CopChaseRating,
				Punishments:             u.Punishments,
				Achievement:             u.Achievement,
				Telegram:                u.Telegram,
				Prefix:                  u.Prefix,
				Star:                    u.Star,
				ApplicationVerification: u.ApplicationVerification,
			}

			marshal, err := json.Marshal(longUser)
			if err != nil {
				http.Error(w, "Bad Request", http.StatusBadRequest)
				return
			}

			if _, err = w.Write(marshal); err != nil {
				http.Error(w, "Bad Request", http.StatusBadRequest)
				return
			}
		} else {
			if !errors.Is(err, pg.ErrNoRows) {
				c.log.Error(err)
			}

			marshal, err := json.Marshal(user.Data)
			if err != nil {
				http.Error(w, "Bad Request", http.StatusBadRequest)
				return
			}

			if _, err = w.Write(marshal); err != nil {
				http.Error(w, "Bad Request", http.StatusBadRequest)
				return
			}
		}
	} else {
		limit := r.URL.Query().Get("limit")
		if limit == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		limitInt, err := strconv.Atoi(limit)
		if err != nil || limitInt > 1000 || limitInt < 1 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		page := r.URL.Query().Get("page")
		if page == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		pageInt, err := strconv.Atoi(page)
		if err != nil || pageInt < 1 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		orderBy := r.URL.Query().Get("orderBy")
		if orderBy == "" {
			orderBy = "account_id"
		}

		users, err := c.s.SearchByPageAndLimit(r.URL.Query().Get("search"), limitInt, 0+(pageInt-1)*limitInt, orderBy)
		if err != nil {
			if errors.Is(err, pg.ErrNoRows) {
				w.WriteHeader(http.StatusNotFound)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}
			return
		}

		marshal, err := json.Marshal(users)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		_, err = w.Write(marshal)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
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
