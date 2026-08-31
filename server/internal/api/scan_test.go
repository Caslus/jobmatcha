package api

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/caslus/jobmatcha/internal/model"
	"github.com/gin-gonic/gin"
)

type scanFake struct {
	startID  uint
	startErr error
	job      *model.ScanJobResponse
	getErr   error
}

func (f scanFake) StartScan() (uint, error)                      { return f.startID, f.startErr }
func (f scanFake) GetJob(uint) (*model.ScanJobResponse, error)   { return f.job, f.getErr }
func (f scanFake) GetLatestJob() (*model.ScanJobResponse, error) { return f.job, f.getErr }
func (scanFake) SupportsAdapter(string) bool                     { return false }

func TestScanHandlerMapsServiceOutcomesToHTTP(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tc := range []struct {
		name, method, path string
		fake               scanFake
		want               int
	}{
		{"start", http.MethodPost, "/", scanFake{startID: 7}, http.StatusAccepted},
		{"start failure", http.MethodPost, "/", scanFake{startErr: errors.New("db")}, http.StatusInternalServerError},
		{"bad id", http.MethodGet, "/bad", scanFake{}, http.StatusBadRequest},
		{"missing job", http.MethodGet, "/1", scanFake{}, http.StatusNotFound},
		{"job", http.MethodGet, "/1", scanFake{job: &model.ScanJobResponse{ID: 1, Status: "completed"}}, http.StatusOK},
		{"latest failure", http.MethodGet, "/latest", scanFake{getErr: errors.New("db")}, http.StatusInternalServerError},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := gin.New()
			h := NewScanHandler(tc.fake)
			if tc.path == "/latest" {
				r.GET("/latest", h.GetLatest)
			} else if tc.method == http.MethodPost {
				r.POST("/", h.Start)
			} else {
				r.GET("/:id", h.GetByID)
			}
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(tc.method, tc.path, nil))
			if w.Code != tc.want {
				t.Fatalf("status = %d, want %d: %s", w.Code, tc.want, w.Body.String())
			}
		})
	}
}
