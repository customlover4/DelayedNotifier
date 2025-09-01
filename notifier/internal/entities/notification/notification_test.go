package notification

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"
)

func BenchmarkBinary(b *testing.B) {
	tm, _ := time.Parse(DateLayout, "2000-12-22 15:00")

	for i := 0; i < b.N; i++ {
		m := Notification{
			ID: 1, TelegramID: 123, Message: "hihi",
			Email: "asd@asad.com", Status: "pending", Date: tm,
		}
		_, _ = m.MarshalBinary()
		// t := Notification{}
		// t.UnmarshalBinary(b)
	}
}

func BenchmarkJSON(b *testing.B) {
	tm, _ := time.Parse(DateLayout, "2000-12-22 15:00")
	type tmp struct {
		R string `json:"r"`
	}
	for i := 0; i < b.N; i++ {
		m := Notification{
			ID: 1, TelegramID: 123, Message: "hihi",
			Email: "asd@asad.com", Status: "pending", Date: tm,
		}
		v, _ := json.Marshal(m)
		t := Notification{}
		json.Unmarshal(v, &t)
	}
}

func BenchmarkJSONBinary(b *testing.B) {
	tm, _ := time.Parse(DateLayout, "2000-12-22 15:00")

	for i := 0; i < b.N; i++ {

		m := Notification{
			ID: 1, TelegramID: 123, Message: "hihi",
			Email: "asd@asad.com", Status: "pending", Date: tm,
		}
		b, _ := m.MarshalBinary()
		v, _ := json.Marshal(b)
		var r []byte
		json.Unmarshal(v, &r)
		t := Notification{}
		t.UnmarshalBinary(r)
	}
}

func TestMarshalUnmarshal(t *testing.T) {
	tm, _ := time.Parse(DateLayout, "2000-12-22 15:00")
	tests := []struct {
		name string
		data Notification
		want Notification
	}{
		{
			name: "good",
			data: Notification{
				ID: 1, TelegramID: 123, Message: "hihi",
				Email: "asd@asad.com", Status: "pending", Date: tm,
			},
			want: Notification{
				ID: 1, TelegramID: 123, Message: "hihi",
				Email: "asd@asad.com", Status: "pending", Date: tm,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, _ := tt.data.MarshalBinary()
			tmp := Notification{}
			tmp.UnmarshalBinary(b)
			if !reflect.DeepEqual(tt.want, tmp) {
				t.Errorf("Marhsal/Unmarshal() get=%v, want %v", tmp, tt.want)
			}
		})
	}
}
