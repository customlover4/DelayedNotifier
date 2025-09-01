package notification

import (
	"bytes"
	"encoding/binary"
	"time"
)

// Notify модель отложенного уведомления
type Notification struct {
	ID               int64     `db:"id"`
	TelegramID       int64     `db:"telegram_id"`
	Message          string    `db:"message"`
	Email            string    `db:"email"`
	Status           string    `db:"status"`
	Date             time.Time `db:"dt"`
}

const (
	DateLayout = time.RFC3339

	StatusPending  = "pending"
	StatusComplete = "complete"
)

func (n Notification) DT() string {
	return n.Date.Format(time.DateTime)
}

func (n Notification) MarshalBinary() ([]byte, error) {
	base := make([]byte, 0, 128)
	b := bytes.NewBuffer(base)

	if err := binary.Write(b, binary.LittleEndian, n.ID); err != nil {
		return nil, err
	}
	if err := binary.Write(b, binary.LittleEndian, n.TelegramID); err != nil {
		return nil, err
	}

	strFields := []string{n.Message, n.Email, n.Status}
	for _, f := range strFields {
		if err := binary.Write(b, binary.LittleEndian, int32(len(f))); err != nil {
			return nil, err
		}
		if err := binary.Write(b, binary.LittleEndian, []byte(f)); err != nil {
			return nil, err
		}
	}
	d, err := n.Date.MarshalBinary()
	if err != nil {
		return nil, err
	}
	if err := binary.Write(b, binary.LittleEndian, int32(len(d))); err != nil {
		return nil, err
	}
	if err := binary.Write(b, binary.LittleEndian, d); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func (n *Notification) UnmarshalBinary(data []byte) error {
	if data == nil {
		return nil
	}

	b := bytes.NewReader(data)

	var ID int64
	if err := binary.Read(b, binary.LittleEndian, &ID); err != nil {
		return err
	}
	n.ID = ID
	var TelegramID int64
	if err := binary.Read(b, binary.LittleEndian, &TelegramID); err != nil {
		return err
	}
	n.TelegramID = TelegramID

	strFields := []*string{&n.Message, &n.Email, &n.Status}
	for _, f := range strFields {
		var l int32
		if err := binary.Read(b, binary.LittleEndian, &l); err != nil {
			return err
		}
		bf := make([]byte, l)
		if err := binary.Read(b, binary.LittleEndian, bf); err != nil {
			return err
		}
		*f = string(bf)
	}
	t := time.Time{}
	var l int32
	if err := binary.Read(b, binary.LittleEndian, &l); err != nil {
		return err
	}
	bf := make([]byte, l)
	if err := binary.Read(b, binary.LittleEndian, bf); err != nil {
		return err
	}
	t.UnmarshalBinary(bf)
	n.Date = t

	return nil
}
