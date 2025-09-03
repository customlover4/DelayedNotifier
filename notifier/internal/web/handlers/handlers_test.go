package handlers

import (
	"delayednotifier/internal/entities/notification"
	"delayednotifier/internal/service"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

type ServiceMock struct {
	createF func(n notification.Notification) (int64, error)
	getF    func(id int64) (notification.Notification, error)
	deleteF func(id int64) error
	updateF func(status string, id int64) error
}

func (sm *ServiceMock) CreateNotification(n notification.Notification) (int64, error) {
	return sm.createF(n)
}

func (sm *ServiceMock) Notification(id int64) (notification.Notification, error) {
	return sm.getF(id)
}

func (sm *ServiceMock) DeleteNotification(id int64) error {
	return sm.deleteF(id)
}

func (sm *ServiceMock) UpdateNotificationStatus(status string, id int64) error {
	return sm.updateF(status, id)
}

func TestMain(t *testing.T) {
	type args struct {
		s notifyer
	}
	tests := []struct {
		name string
		args
		want int
	}{
		{
			name: "default",
			args: args{
				s: &ServiceMock{},
			},
			want: http.StatusOK,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)

			tmpDir := t.TempDir()

			templateContent := `<html><head></head></html>`

			templatePath := filepath.Join(tmpDir, "test.html")
			err := os.WriteFile(templatePath, []byte(templateContent), 0644)
			if err != nil {
				t.Errorf("can't create tmp dir with templates")
				return
			}

			rr := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/main", nil)
			h := Main(tt.s)

			g := gin.Default()
			g.LoadHTMLFiles(templatePath)
			g.GET("/main", h)
			g.ServeHTTP(rr, req)
			if tt.want != rr.Result().StatusCode {
				t.Errorf(
					"Main() status code get=%d, want %d",
					rr.Result().StatusCode, tt.want,
				)
			}
		})
	}
}

func TestCreating(t *testing.T) {
	type args struct {
		s notifyer
	}
	tests := []struct {
		name string
		args
		body string
		code int
	}{
		{
			name: "good",
			args: args{
				s: &ServiceMock{
					createF: func(n notification.Notification) (int64, error) {
						return 0, nil
					},
				},
			},
			body: `{"message": "hi", "telegram_id": "123", "date": "2000-12-22T15:00:00.000Z"}`,
			code: http.StatusOK,
		},
		{
			name: "empty body",
			args: args{
				s: &ServiceMock{},
			},
			body: "",
			code: http.StatusBadRequest,
		},
		{
			name: "bad json",
			args: args{
				s: &ServiceMock{},
			},
			body: "{",
			code: http.StatusBadRequest,
		},
		{
			name: "wrong data in json",
			args: args{
				s: &ServiceMock{
					createF: func(n notification.Notification) (int64, error) {
						return 0, service.ErrNotValidData
					},
				},
			},
			body: `{"message": "hi", "telegram_id": "123", "date": "not valid"}`,
			code: http.StatusBadRequest,
		},
		{
			name: "wrong telegram_id in json",
			args: args{
				s: &ServiceMock{
					createF: func(n notification.Notification) (int64, error) {
						return 0, service.ErrNotValidData
					},
				},
			},
			body: `{"message": "hi", "telegram_id": "-100", "date": "2000-12-22T15:00:00.000Z"}`,
			code: http.StatusBadRequest,
		},
		{
			name: "wrong data in json",
			args: args{
				s: &ServiceMock{
					createF: func(n notification.Notification) (int64, error) {
						return 0, service.ErrNotValidData
					},
				},
			},
			body: `{"message": "hi", "telegram_id": "10", "date": "2000-12-20T15:00:00.000Z"}`,
			code: http.StatusServiceUnavailable,
		},
		{
			name: "internal error",
			args: args{
				s: &ServiceMock{
					createF: func(n notification.Notification) (int64, error) {
						return 0, errors.New("unknown error")
					},
				},
			},
			body: `{"message": "hi", "telegram_id": "123", "date": "2000-12-20T15:00:00.000Z"}`,
			code: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			rr := httptest.NewRecorder()
			req := httptest.NewRequest(
				http.MethodPost, "/main", strings.NewReader(tt.body),
			)
			h := CreateNotify(tt.s)

			g := gin.Default()
			g.POST("/main", h)
			g.ServeHTTP(rr, req)
			if tt.code != rr.Result().StatusCode {
				t.Errorf(
					"Main() status code get=%d, want %d, body %s",
					rr.Result().StatusCode, tt.code, rr.Body.String(),
				)
			}
		})
	}
}

func TestGetter(t *testing.T) {
	type args struct {
		s notifyer
	}
	tests := []struct {
		name string
		args
		code int
		id   string
	}{
		{
			name: "good",
			args: args{
				s: &ServiceMock{
					getF: func(id int64) (notification.Notification, error) {
						return notification.Notification{}, nil
					},
				},
			},
			code: http.StatusOK,
			id:   "1",
		},
		{
			name: "not numeric id",
			args: args{
				s: &ServiceMock{
					getF: func(id int64) (notification.Notification, error) {
						return notification.Notification{}, nil
					},
				},
			},
			code: http.StatusBadRequest,
			id:   "haha",
		},
		{
			name: "not valid id",
			args: args{
				s: &ServiceMock{
					getF: func(id int64) (notification.Notification, error) {
						return notification.Notification{}, service.ErrNotValidData
					},
				},
			},
			code: http.StatusBadRequest,
			id:   "-1",
		},
		{
			name: "not found notification",
			args: args{
				s: &ServiceMock{
					getF: func(id int64) (notification.Notification, error) {
						return notification.Notification{}, service.ErrNotFound
					},
				},
			},
			code: http.StatusNotFound,
			id:   "10000000",
		},
		{
			name: "unknown err",
			args: args{
				s: &ServiceMock{
					getF: func(id int64) (notification.Notification, error) {
						return notification.Notification{}, errors.New("unknown")
					},
				},
			},
			code: http.StatusInternalServerError,
			id:   "10000000",
		},
		{
			name: "business err",
			args: args{
				s: &ServiceMock{
					getF: func(id int64) (notification.Notification, error) {
						return notification.Notification{}, service.ErrNotValidData
					},
				},
			},
			code: http.StatusServiceUnavailable,
			id:   "10000000",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			url := "/endpoint"
			rr := httptest.NewRecorder()
			rUrl := fmt.Sprintf("%s/%s", url, tt.id)
			req := httptest.NewRequest(http.MethodGet, rUrl, nil)

			g := gin.Default()
			h := GetNotify(tt.s)
			g.GET(url+"/:id", h)
			g.ServeHTTP(rr, req)
			if tt.code != rr.Result().StatusCode {
				t.Errorf(
					"Main() status code get=%d, want %d, body %s",
					rr.Result().StatusCode, tt.code, rr.Body.String(),
				)
			}
		})
	}
}

func TestDeleter(t *testing.T) {
	type args struct {
		s notifyer
	}
	tests := []struct {
		name string
		args
		code int
		id   string
	}{
		{
			name: "good",
			args: args{
				s: &ServiceMock{
					deleteF: func(id int64) error {
						return nil
					},
				},
			},
			code: http.StatusOK,
			id:   "1",
		},
		{
			name: "not numeric id",
			args: args{
				s: &ServiceMock{
					deleteF: func(id int64) error {
						return nil
					},
				},
			},
			code: http.StatusBadRequest,
			id:   "haha",
		},
		{
			name: "not valid id",
			args: args{
				s: &ServiceMock{
					deleteF: func(id int64) error {
						return service.ErrNotValidData
					},
				},
			},
			code: http.StatusBadRequest,
			id:   "-1",
		},
		{
			name: "not found notification",
			args: args{
				s: &ServiceMock{
					deleteF: func(id int64) error {
						return service.ErrNotAffected
					},
				},
			},
			code: http.StatusNotFound,
			id:   "10000000",
		},
		{
			name: "unknown err",
			args: args{
				s: &ServiceMock{
					deleteF: func(id int64) error {
						return errors.New("unknown")
					},
				},
			},
			code: http.StatusInternalServerError,
			id:   "10000000",
		},
		{
			name: "business err",
			args: args{
				s: &ServiceMock{
					deleteF: func(id int64) error {
						return service.ErrNotValidData
					},
				},
			},
			code: http.StatusServiceUnavailable,
			id:   "10000000",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			url := "/endpoint"
			rr := httptest.NewRecorder()
			rUrl := fmt.Sprintf("%s/%s", url, tt.id)
			req := httptest.NewRequest(http.MethodDelete, rUrl, nil)

			g := gin.Default()
			h := DeleteNotify(tt.s)
			g.DELETE(url+"/:id", h)
			g.ServeHTTP(rr, req)
			if tt.code != rr.Result().StatusCode {
				t.Errorf(
					"Main() status code get=%d, want %d, body %s",
					rr.Result().StatusCode, tt.code, rr.Body.String(),
				)
			}
		})
	}
}

func TestUpdater(t *testing.T) {
	type args struct {
		s notifyer
	}
	tests := []struct {
		name string
		args
		code int
		id   string
		body string
	}{
		{
			name: "good",
			args: args{
				s: &ServiceMock{
					updateF: func(status string, id int64) error {
						return nil
					},
				},
			},
			code: http.StatusOK,
			body: `{"status": "complete"}`,
			id:   "1",
		},
		{
			name: "not numeric id",
			args: args{
				s: &ServiceMock{
					updateF: func(status string, id int64) error {
						return nil
					},
				},
			},
			code: http.StatusBadRequest,
			body: `{"status": "complete"}`,
			id:   "haha",
		},
		{
			name: "not valid id",
			args: args{
				s: &ServiceMock{
					updateF: func(status string, id int64) error {
						return service.ErrNotValidData
					},
				},
			},
			code: http.StatusBadRequest,
			body: `{"status": "complete"}`,
			id:   "-1",
		},
		{
			name: "not found notification",
			args: args{
				s: &ServiceMock{
					updateF: func(status string, id int64) error {
						return service.ErrNotAffected
					},
				},
			},
			code: http.StatusNotFound,
			body: `{"status": "complete"}`,
			id:   "10000000",
		},
		{
			name: "unknown err",
			args: args{
				s: &ServiceMock{
					updateF: func(status string, id int64) error {
						return errors.New("unknown")
					},
				},
			},
			code: http.StatusInternalServerError,
			body: `{"status": "complete"}`,
			id:   "10000000",
		},
		{
			name: "unknown status",
			args: args{
				s: &ServiceMock{
					updateF: func(status string, id int64) error {
						return nil
					},
				},
			},
			code: http.StatusBadRequest,
			body: `{"status": "test"}`,
			id:   "10000000",
		},
		{
			name: "wrong json",
			args: args{
				s: &ServiceMock{
					updateF: func(status string, id int64) error {
						return nil
					},
				},
			},
			code: http.StatusBadRequest,
			body: `{"status": "test"`,
			id:   "10000000",
		},
		{
			name: "wrong json",
			args: args{
				s: &ServiceMock{
					updateF: func(status string, id int64) error {
						return service.ErrNotValidData
					},
				},
			},
			code: http.StatusServiceUnavailable,
			body: `{"status": "pending"}`,
			id:   "10000000",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			url := "/endpoint"
			rr := httptest.NewRecorder()
			rUrl := fmt.Sprintf("%s/%s", url, tt.id)
			req := httptest.NewRequest(http.MethodPatch, rUrl, strings.NewReader(tt.body))

			g := gin.Default()
			h := UpdateNotify(tt.s)
			g.PATCH(url+"/:id", h)
			g.ServeHTTP(rr, req)
			if tt.code != rr.Result().StatusCode {
				t.Errorf(
					"Main() status code get=%d, want %d, body %s",
					rr.Result().StatusCode, tt.code, rr.Body.String(),
				)
			}
		})
	}
}
