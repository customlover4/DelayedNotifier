package postgres

import (
	"context"
	"delayednotifier/internal/entities/notification"
	"fmt"
)

func (p *Postgres) CreateNotification(n notification.Notification) (int64, error) {
	const op = "internal.storage.postgres.CreateNotification"

	var id int64

	q := fmt.Sprintf(
		`insert into %s 
		(telegram_id, message, email, dt)
		values ($1, $2, $3, $4) returning id;`,
		NotificationTable,
	)

	row := p.db.Master.QueryRowContext(
		context.Background(),
		q, n.TelegramID, n.Message, n.Email, n.DT(),
	)
	if row.Err() != nil {
		return id, row.Err()
	}
	err := row.Scan(&id)
	if err != nil {
		return id, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (p *Postgres) Notification(id int64) (notification.Notification, error) {
	const op = "internal.storage.postgres.Notification"

	var r notification.Notification

	q := fmt.Sprintf("select * from %s where id = $1;", NotificationTable)

	row := p.db.Master.QueryRowContext(
		context.Background(), q, id,
	)
	if row.Err() != nil {
		return r, fmt.Errorf("%s: %w", op, row.Err())
	}
	err := row.Scan(
		&r.ID, &r.TelegramID, &r.Message, &r.Email, &r.Status, &r.Date,
	)
	if err != nil {
		return r, err
	}

	return r, nil
}

func (p *Postgres) UpdateNotificationStatus(status string, id int64) (int64, error) {
	const op = "internal.storage.postgres.UpdateNotificationStatus"

	q := fmt.Sprintf(
		"update %s set status = $1 where id = $2;", NotificationTable,
	)

	r, err := p.db.ExecContext(context.Background(), q, status, id)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	affected, err := r.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return affected, nil
}

func (p *Postgres) DeleteNotification(id int64) (int64, error) {
	const op = "internal.storage.postgres.UpdateNotificationStatus"

	q := fmt.Sprintf(
		"delete from %s where id = $1;", NotificationTable,
	)

	r, err := p.db.ExecContext(context.Background(), q, id)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	affected, err := r.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return affected, nil
}
