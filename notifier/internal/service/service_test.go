package service

import (
	"delayednotifier/internal/entities/notification"
	"delayednotifier/internal/storage"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type StorageMock struct {
	addNF   func(n notification.Notification) (int64, error)
	getF    func(id int64) (notification.Notification, error)
	deleteF func(id int64) error
	updateF func(status string, id int64) error
}

func (sm *StorageMock) CreateNotification(n notification.Notification) (int64, error) {
	return sm.addNF(n)
}

func (sm *StorageMock) GetNotification(id int64) (notification.Notification, error) {
	return sm.getF(id)
}

func (sm *StorageMock) DeleteNotification(id int64) error {
	return sm.deleteF(id)
}
func (sm *StorageMock) UpdateNotificationStatus(status string, id int64) error {
	return sm.updateF(status, id)
}

func TestService_CreateNotification(t *testing.T) {
	goodTM, err := time.Parse(notification.DateLayout, "3000-12-22T15:20:00.000Z")
	require.NoError(t, err)
	badTM, _ := time.Parse(notification.DateLayout, "1969-12-22T15:20:00.000Z")
	require.NoError(t, err)
	type fields struct {
		s storager
	}
	type args struct {
		n notification.Notification
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   error
	}{
		{
			name: "good",
			fields: fields{
				s: &StorageMock{
					addNF: func(n notification.Notification) (int64, error) {
						return 0, nil
					},
				},
			},
			args: args{
				n: notification.Notification{
					Message:    "test",
					TelegramID: 123,
					Email:      "asd@asd.com",
					Date:       goodTM,
				},
			},
			want: nil,
		},
		{
			name: "good 2",
			fields: fields{
				s: &StorageMock{
					addNF: func(n notification.Notification) (int64, error) {
						return 0, nil
					},
				},
			},
			args: args{
				n: notification.Notification{
					Message: "test",
					Email:   "asd@asd.com",
					Date:    goodTM,
				},
			},
			want: nil,
		},
		{
			name: "good 3",
			fields: fields{
				s: &StorageMock{
					addNF: func(n notification.Notification) (int64, error) {
						return 0, nil
					},
				},
			},
			args: args{
				n: notification.Notification{
					Message:    "test",
					TelegramID: 123,
					Date:       goodTM,
				},
			},
			want: nil,
		},
		{
			name: "good 4",
			fields: fields{
				s: &StorageMock{
					addNF: func(n notification.Notification) (int64, error) {
						return 0, nil
					},
				},
			},
			args: args{
				n: notification.Notification{
					Message:    "test",
					TelegramID: 123,
					Date:       goodTM,
				},
			},
			want: nil,
		},
		{
			name: "date in past",
			fields: fields{
				s: &StorageMock{
					addNF: func(n notification.Notification) (int64, error) {
						return 0, nil
					},
				},
			},
			args: args{
				n: notification.Notification{
					Message:    "test",
					TelegramID: 123,
					Email:      "asd@asd.com",
					Date:       badTM,
				},
			},
			want: ErrNotValidData,
		},
		{
			name: "bad telegram_id",
			fields: fields{
				s: &StorageMock{
					addNF: func(n notification.Notification) (int64, error) {
						return 0, nil
					},
				},
			},
			args: args{
				n: notification.Notification{
					Message:    "test",
					TelegramID: -1,
					Email:      "asd@asd.com",
					Date:       goodTM,
				},
			},
			want: ErrNotValidData,
		},
		{
			name: "error on telegram notification",
			fields: fields{
				s: &StorageMock{
					addNF: func(n notification.Notification) (int64, error) {
						return 0, ErrStorageInternal
					},
				},
			},
			args: args{
				n: notification.Notification{
					Message:    "test",
					TelegramID: 123,
					Email:      "asd@asd.com",
					Date:       goodTM,
				},
			},
			want: ErrStorageInternal,
		},
		{
			name: "error on email notification",
			fields: fields{
				s: &StorageMock{
					addNF: func(n notification.Notification) (int64, error) {
						return 0, errors.New("unknown")
					},
				},
			},
			args: args{
				n: notification.Notification{
					Message:    "test",
					TelegramID: 123,
					Email:      "asd@asd.com",
					Date:       goodTM,
				},
			},
			want: ErrStorageInternal,
		},
		{
			name: "wrong email",
			fields: fields{
				s: &StorageMock{
					addNF: func(n notification.Notification) (int64, error) {
						return 0, nil
					},
				},
			},
			args: args{
				n: notification.Notification{
					Message:    "test",
					TelegramID: 123,
					Email:      "asdasd.com",
					Date:       goodTM,
				},
			},
			want: ErrNotValidData,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := New(
				tt.fields.s,
			)
			if _, err := s.CreateNotification(tt.args.n); !errors.Is(err, tt.want) {
				t.Errorf("Service.CreateNotification() error = %v, wantErr %v", err, tt.want)
			}
		})
	}
}

func TestService_Notification(t *testing.T) {
	goodTM, _ := time.Parse(notification.DateLayout, "3000-12-22 15:20")
	type fields struct {
		s storager
	}
	type args struct {
		id int64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    notification.Notification
		wantErr error
	}{
		{
			name: "good",
			fields: fields{
				s: &StorageMock{
					getF: func(id int64) (notification.Notification, error) {
						return notification.Notification{
							ID:         1,
							Message:    "test",
							TelegramID: 123,
							Email:      "asd@asd.com",
							Date:       goodTM,
						}, nil
					},
				},
			},
			args: args{
				id: 1,
			},
			want: notification.Notification{
				ID:         1,
				Message:    "test",
				TelegramID: 123,
				Email:      "asd@asd.com",
				Date:       goodTM,
			},
			wantErr: nil,
		},
		{
			name: "not found",
			fields: fields{
				s: &StorageMock{
					getF: func(id int64) (notification.Notification, error) {
						return notification.Notification{}, storage.ErrNotFound
					},
				},
			},
			args: args{
				id: 100000,
			},
			want:    notification.Notification{},
			wantErr: ErrNotFound,
		},
		{
			name: "unknown error",
			fields: fields{
				s: &StorageMock{
					getF: func(id int64) (notification.Notification, error) {
						return notification.Notification{}, errors.New("unknown")
					},
				},
			},
			args: args{
				id: 1,
			},
			want:    notification.Notification{},
			wantErr: ErrStorageInternal,
		},
		{
			name: "bad id",
			fields: fields{
				s: &StorageMock{
					getF: func(id int64) (notification.Notification, error) {
						return notification.Notification{}, nil
					},
				},
			},
			args: args{
				id: -1,
			},
			want:    notification.Notification{},
			wantErr: ErrNotValidData,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := New(
				tt.fields.s,
			)
			got, err := s.Notification(tt.args.id)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Service.Notification() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Service.Notification() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestService_DeleteNotification(t *testing.T) {
	type fields struct {
		s storager
	}
	type args struct {
		id int64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
	}{
		{
			name: "good",
			fields: fields{
				s: &StorageMock{
					deleteF: func(id int64) error {
						return nil
					},
				},
			},
			args: args{
				id: 1,
			},
			wantErr: nil,
		},
		{
			name: "not affected",
			fields: fields{
				s: &StorageMock{
					deleteF: func(id int64) error {
						return storage.ErrNotAffected
					},
				},
			},
			args: args{
				id: 100000,
			},
			wantErr: ErrNotAffected,
		},
		{
			name: "unknown error",
			fields: fields{
				s: &StorageMock{
					deleteF: func(id int64) error {
						return errors.New("unknown")
					},
				},
			},
			args: args{
				id: 1,
			},
			wantErr: ErrStorageInternal,
		},
		{
			name: "bad id",
			fields: fields{
				s: &StorageMock{
					deleteF: func(id int64) error {
						return nil
					},
				},
			},
			args: args{
				id: -1,
			},
			wantErr: ErrNotValidData,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := New(
				tt.fields.s,
			)
			err := s.DeleteNotification(tt.args.id)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Service.DeleteNotification() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestService_UpdateNotificationStatus(t *testing.T) {
	type fields struct {
		str storager
	}
	type args struct {
		status string
		id     int64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   error
	}{
		{
			name: "good",
			fields: fields{
				str: &StorageMock{
					updateF: func(status string, id int64) error {
						return nil
					},
				},
			},
			args: args{
				status: "pending",
				id:     100,
			},
		},
		{
			name: "negative id",
			fields: fields{
				str: &StorageMock{
					updateF: func(status string, id int64) error {
						return nil
					},
				},
			},
			args: args{
				status: "pending",
				id:     -100,
			},
			want: ErrNotValidData,
		},
		{
			name: "unknown status",
			fields: fields{
				str: &StorageMock{
					updateF: func(status string, id int64) error {
						return nil
					},
				},
			},
			args: args{
				status: "test",
				id:     1,
			},
			want: ErrNotValidData,
		},
		{
			name: "good",
			fields: fields{
				str: &StorageMock{
					updateF: func(status string, id int64) error {
						return storage.ErrNotAffected
					},
				},
			},
			args: args{
				status: "pending",
				id:     100,
			},
			want: ErrNotAffected,
		},
		{
			name: "good",
			fields: fields{
				str: &StorageMock{
					updateF: func(status string, id int64) error {
						return errors.New("test")
					},
				},
			},
			args: args{
				status: "pending",
				id:     100,
			},
			want: ErrStorageInternal,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Service{
				str: tt.fields.str,
			}
			err := s.UpdateNotificationStatus(tt.args.status, tt.args.id)
			if !errors.Is(err, tt.want) {
				t.Errorf("Service.UpdateNotificationStatus() error = %v, wantErr %v", err, tt.want)
			}
		})
	}
}
