package notification_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/smykla-labs/claude-hooks/internal/validators/notification"
	"github.com/smykla-labs/claude-hooks/pkg/hook"
	"github.com/smykla-labs/claude-hooks/pkg/logger"
)

var _ = Describe("BellValidator", func() {
	var (
		validator *notification.BellValidator
		ctx       *hook.Context
	)

	BeforeEach(func() {
		validator = notification.NewBellValidator(logger.NewNoOpLogger())
		ctx = &hook.Context{
			EventType: hook.Notification,
		}
	})

	Describe("Validate", func() {
		Context("when notification type is bell", func() {
			BeforeEach(func() {
				ctx.NotificationType = "bell"
			})

			It("should pass", func() {
				result := validator.Validate(ctx)
				Expect(result.Passed).To(BeTrue())
			})

			It("should not block", func() {
				result := validator.Validate(ctx)
				Expect(result.ShouldBlock).To(BeFalse())
			})
		})

		Context("when notification type is not bell", func() {
			BeforeEach(func() {
				ctx.NotificationType = "other"
			})

			It("should pass", func() {
				result := validator.Validate(ctx)
				Expect(result.Passed).To(BeTrue())
			})
		})

		Context("when notification type is empty", func() {
			It("should pass", func() {
				result := validator.Validate(ctx)
				Expect(result.Passed).To(BeTrue())
			})
		})
	})
})
