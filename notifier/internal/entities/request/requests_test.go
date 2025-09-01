package request

import (
	"testing"
)

func TestCreateNotify_Validate(t *testing.T) {
	type fields struct {
		Message    string
		TelegramID string
		Email      string
		Date       string
	}
	tests := []struct {
		name    string
		fields  fields
		wantMsg bool
	}{
		{
			name: "good data 1",
			fields: fields{
				Message:    "haha",
				TelegramID: "123",
				Email:      "asd@asd.com",
				Date:       "2000-12-22T15:06:00.000Z",
			},
			wantMsg: false,
		},
		{
			name: "good data 2",
			fields: fields{
				Message:    "haha",
				TelegramID: "12312",
				Email:      "asd@asd.com",
				Date:       "2000-12-22T15:06:00.000Z",
			},
			wantMsg: false,
		},
		{
			name: "good data 3",
			fields: fields{
				Message:    "haha",
				TelegramID: "12312",
				Date:       "2000-12-22T15:06:00.000Z",
			},
			wantMsg: false,
		},
		{
			name: "good data 4",
			fields: fields{
				Message: "haha",
				Email:   "asd@asd.com",
				Date:    "2000-12-22T15:06:00.000Z",
			},
			wantMsg: false,
		},
		{
			name: "empty text",
			fields: fields{
				Message:    "",
				TelegramID: "123",
				Email:      "asd@asd.com",
				Date:       "2000-12-22T15:06:00.000Z",
			},
			wantMsg: true,
		},
		{
			name: "tg & email empty",
			fields: fields{
				Message:    "asd",
				TelegramID: "",
				Email:      "",
				Date:       "2000-12-22T15:06:00.000Z",
			},
			wantMsg: true,
		},
		{
			name: "wrong tg id",
			fields: fields{
				Message:    "haha",
				TelegramID: "-1",
				Email:      "asd@asd.com",
				Date:       "2000-12-22T15:06:00.000Z",
			},
			wantMsg: true,
		},
		{
			name: "wrong email 1",
			fields: fields{
				Message:    "haha",
				TelegramID: "123",
				Email:      "asdasd.com",
				Date:       "2000-12-22T15:06:00.000Z",
			},
			wantMsg: true,
		},
		{
			name: "wrong email 2",
			fields: fields{
				Message:    "haha",
				TelegramID: "123",
				Email:      "asd@asdcom",
				Date:       "2000-12-22T15:06:00.000Z",
			},
			wantMsg: true,
		},
		{
			name: "wrong date",
			fields: fields{
				Message:    "haha",
				TelegramID: "123",
				Email:      "asd@asd.com",
				Date:       "2000-i12-22T15:06",
			},
			wantMsg: true,
		},
		{
			name: "empty date",
			fields: fields{
				Message:    "haha",
				TelegramID: "123",
				Email:      "asd@asd.com",
			},
			wantMsg: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &CreateNotification{
				Message:    tt.fields.Message,
				TelegramID: tt.fields.TelegramID,
				Email:      tt.fields.Email,
				Date:       tt.fields.Date,
			}
			_, got := c.Validate()
			if (tt.wantMsg && got == "") || (!tt.wantMsg && got != "") {
				t.Errorf("CreateNotify.Validate() got1 = %v, want %t", got, tt.wantMsg)
			}
		})
	}
}
