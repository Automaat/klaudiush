package git_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/smykla-labs/claude-hooks/internal/validators/git"
	"github.com/smykla-labs/claude-hooks/pkg/hook"
	"github.com/smykla-labs/claude-hooks/pkg/logger"
)

var _ = Describe("BranchValidator", func() {
	var (
		v   *git.BranchValidator
		ctx *hook.Context
	)

	BeforeEach(func() {
		v = git.NewBranchValidator(logger.NewNoOpLogger())
		ctx = &hook.Context{
			EventType: hook.PreToolUse,
			ToolName:  hook.Bash,
			ToolInput: hook.ToolInput{},
		}
	})

	Describe("git checkout -b", func() {
		Context("with valid branch names", func() {
			It("should pass for feat/add-feature", func() {
				ctx.ToolInput.Command = "git checkout -b feat/add-feature"
				result := v.Validate(context.Background(), ctx)
				Expect(result.Passed).To(BeTrue())
			})

			It("should pass for fix/bug-123", func() {
				ctx.ToolInput.Command = "git checkout -b fix/bug-123"
				result := v.Validate(context.Background(), ctx)
				Expect(result.Passed).To(BeTrue())
			})

			It("should pass for docs/update-readme", func() {
				ctx.ToolInput.Command = "git checkout -b docs/update-readme"
				result := v.Validate(context.Background(), ctx)
				Expect(result.Passed).To(BeTrue())
			})

			It("should pass for chore/cleanup-code", func() {
				ctx.ToolInput.Command = "git checkout -b chore/cleanup-code"
				result := v.Validate(context.Background(), ctx)
				Expect(result.Passed).To(BeTrue())
			})

			It("should pass for ci/update-workflow", func() {
				ctx.ToolInput.Command = "git checkout -b ci/update-workflow"
				result := v.Validate(context.Background(), ctx)
				Expect(result.Passed).To(BeTrue())
			})
		})

		Context("with protected branches", func() {
			It("should skip validation for main", func() {
				ctx.ToolInput.Command = "git checkout -b main"
				result := v.Validate(context.Background(), ctx)
				Expect(result.Passed).To(BeTrue())
			})

			It("should skip validation for master", func() {
				ctx.ToolInput.Command = "git checkout -b master"
				result := v.Validate(context.Background(), ctx)
				Expect(result.Passed).To(BeTrue())
			})
		})

		Context("with invalid branch names", func() {
			It("should fail for uppercase characters", func() {
				ctx.ToolInput.Command = "git checkout -b Feat/add-feature"
				result := v.Validate(context.Background(), ctx)
				Expect(result.Passed).To(BeFalse())
				Expect(result.Message).To(ContainSubstring("lowercase"))
				Expect(result.Message).To(ContainSubstring("feat/add-feature"))
			})

			It("should fail for missing type", func() {
				ctx.ToolInput.Command = "git checkout -b add-feature"
				result := v.Validate(context.Background(), ctx)
				Expect(result.Passed).To(BeFalse())
				Expect(result.Message).To(ContainSubstring("type/description"))
			})

			It("should fail for invalid type", func() {
				ctx.ToolInput.Command = "git checkout -b invalid/add-feature"
				result := v.Validate(context.Background(), ctx)
				Expect(result.Passed).To(BeFalse())
				Expect(result.Message).To(ContainSubstring("Invalid branch type"))
			})

			It("should fail for spaces in branch name", func() {
				ctx.ToolInput.Command = "git checkout -b feat/add feature"
				result := v.Validate(context.Background(), ctx)
				Expect(result.Passed).To(BeFalse())
			})

			It("should fail for uppercase in description", func() {
				ctx.ToolInput.Command = "git checkout -b feat/Add-Feature"
				result := v.Validate(context.Background(), ctx)
				Expect(result.Passed).To(BeFalse())
				Expect(result.Message).To(ContainSubstring("lowercase"))
			})

			It("should fail for underscore separator", func() {
				ctx.ToolInput.Command = "git checkout -b feat_add_feature"
				result := v.Validate(context.Background(), ctx)
				Expect(result.Passed).To(BeFalse())
			})
		})
	})

	Describe("git branch", func() {
		Context("with valid branch names", func() {
			It("should pass for feat/add-feature", func() {
				ctx.ToolInput.Command = "git branch feat/add-feature"
				result := v.Validate(context.Background(), ctx)
				Expect(result.Passed).To(BeTrue())
			})

			It("should pass for fix/bug-456", func() {
				ctx.ToolInput.Command = "git branch fix/bug-456"
				result := v.Validate(context.Background(), ctx)
				Expect(result.Passed).To(BeTrue())
			})
		})

		Context("with delete operations", func() {
			It("should skip validation for -d flag", func() {
				ctx.ToolInput.Command = "git branch -d Invalid-Branch"
				result := v.Validate(context.Background(), ctx)
				Expect(result.Passed).To(BeTrue())
			})

			It("should skip validation for -D flag", func() {
				ctx.ToolInput.Command = "git branch -D Invalid-Branch"
				result := v.Validate(context.Background(), ctx)
				Expect(result.Passed).To(BeTrue())
			})

			It("should skip validation for --delete flag", func() {
				ctx.ToolInput.Command = "git branch --delete Invalid-Branch"
				result := v.Validate(context.Background(), ctx)
				Expect(result.Passed).To(BeTrue())
			})
		})

		Context("with invalid branch names", func() {
			It("should fail for uppercase characters", func() {
				ctx.ToolInput.Command = "git branch Fix/bug-456"
				result := v.Validate(context.Background(), ctx)
				Expect(result.Passed).To(BeFalse())
				Expect(result.Message).To(ContainSubstring("lowercase"))
			})

			It("should fail for missing type", func() {
				ctx.ToolInput.Command = "git branch bug-fix"
				result := v.Validate(context.Background(), ctx)
				Expect(result.Passed).To(BeFalse())
			})
		})
	})

	Describe("chained commands", func() {
		It("should validate branch in chained command", func() {
			ctx.ToolInput.Command = "git fetch upstream && git checkout -b feat/new-feature"
			result := v.Validate(context.Background(), ctx)
			Expect(result.Passed).To(BeTrue())
		})

		It("should fail for invalid branch in chained command", func() {
			ctx.ToolInput.Command = "git fetch upstream && git checkout -b Invalid-Branch"
			result := v.Validate(context.Background(), ctx)
			Expect(result.Passed).To(BeFalse())
		})
	})

	Describe("non-branch commands", func() {
		It("should pass for git checkout without -b", func() {
			ctx.ToolInput.Command = "git checkout main"
			result := v.Validate(context.Background(), ctx)
			Expect(result.Passed).To(BeTrue())
		})

		It("should pass for git status", func() {
			ctx.ToolInput.Command = "git status"
			result := v.Validate(context.Background(), ctx)
			Expect(result.Passed).To(BeTrue())
		})

		It("should pass for git commit", func() {
			ctx.ToolInput.Command = "git commit -sS -m 'test'"
			result := v.Validate(context.Background(), ctx)
			Expect(result.Passed).To(BeTrue())
		})
	})
})
