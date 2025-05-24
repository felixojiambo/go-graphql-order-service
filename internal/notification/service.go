package notification

import "context"

type NotificationService interface {
	SendOrderSMS(ctx context.Context, phone, msg string) error
	SendOrderEmail(ctx context.Context, email, subject, body string) error
}
