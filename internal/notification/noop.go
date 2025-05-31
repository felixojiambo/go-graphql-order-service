package notification

import "context"

// NoopNotificationService implements NotificationService but does nothing.
type NoopNotificationService struct{}

// SendOrderSMS is a no-op.
func (n *NoopNotificationService) SendOrderSMS(ctx context.Context, phone, msg string) error {
	return nil
}

// SendOrderEmail is a no-op.
func (n *NoopNotificationService) SendOrderEmail(ctx context.Context, email, subject, body string) error {
	return nil
}
