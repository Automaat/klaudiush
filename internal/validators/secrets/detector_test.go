package secrets_test

import (
	"regexp"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/smykla-labs/klaudiush/internal/validator"
	"github.com/smykla-labs/klaudiush/internal/validators/secrets"
)

var _ = Describe("PatternDetector", func() {
	var detector *secrets.PatternDetector

	BeforeEach(func() {
		detector = secrets.NewDefaultPatternDetector()
	})

	Describe("Detect", func() {
		It("should return empty for empty content", func() {
			findings := detector.Detect("")
			Expect(findings).To(BeEmpty())
		})

		It("should return empty for safe content", func() {
			findings := detector.Detect("This is safe content without any secrets.")
			Expect(findings).To(BeEmpty())
		})

		It("should detect multiple secrets in content", func() {
			content := `
AWS_KEY=AKIAIOSFODNN7EXAMPLE
GITHUB_TOKEN=ghp_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
`
			findings := detector.Detect(content)
			Expect(findings).To(HaveLen(2))
		})

		It("should report correct line numbers", func() {
			content := `line 1
line 2
AKIAIOSFODNN7EXAMPLE on line 3
line 4`
			findings := detector.Detect(content)
			Expect(findings).To(HaveLen(1))
			Expect(findings[0].Line).To(Equal(3))
		})

		It("should report correct column numbers", func() {
			content := `prefix AKIAIOSFODNN7EXAMPLE suffix`
			findings := detector.Detect(content)
			Expect(findings).To(HaveLen(1))
			Expect(findings[0].Column).To(Equal(8)) // "prefix " is 7 chars, so column 8
		})

		It("should capture the matched text", func() {
			content := `key=AKIAIOSFODNN7EXAMPLE`
			findings := detector.Detect(content)
			Expect(findings).To(HaveLen(1))
			Expect(findings[0].Match).To(Equal("AKIAIOSFODNN7EXAMPLE"))
		})

		It("should include pattern metadata in findings", func() {
			content := `ghp_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx`
			findings := detector.Detect(content)
			Expect(findings).To(HaveLen(1))
			Expect(findings[0].Pattern).NotTo(BeNil())
			Expect(findings[0].Pattern.Name).To(Equal("github-pat"))
			Expect(findings[0].Pattern.Description).To(Equal("GitHub Personal Access Token"))
			Expect(findings[0].Pattern.ErrorCode).To(Equal(validator.ErrSecretsToken))
		})
	})

	Describe("AddPatterns", func() {
		It("should allow adding custom patterns", func() {
			customPattern := secrets.Pattern{
				Name:        "custom-key",
				Description: "Custom API Key",
				Regex:       regexp.MustCompile(`CUSTOM_[A-Z0-9]{16}`),
				ErrorCode:   validator.ErrSecretsAPIKey,
			}
			detector.AddPatterns(customPattern)

			content := `my_key=CUSTOM_ABCDEFGH12345678`
			findings := detector.Detect(content)
			Expect(findings).To(HaveLen(1))
			Expect(findings[0].Pattern.Name).To(Equal("custom-key"))
		})
	})

	Describe("DefaultPatterns", func() {
		It("should return all default patterns", func() {
			patterns := secrets.DefaultPatterns()
			Expect(patterns).NotTo(BeEmpty())
			Expect(len(patterns)).To(BeNumerically(">=", 20))
		})

		It("should return a copy, not the original", func() {
			patterns1 := secrets.DefaultPatterns()
			patterns2 := secrets.DefaultPatterns()
			patterns1[0].Name = "modified"
			Expect(patterns2[0].Name).NotTo(Equal("modified"))
		})
	})
})

var _ = Describe("Pattern detection coverage", func() {
	var detector *secrets.PatternDetector

	BeforeEach(func() {
		detector = secrets.NewDefaultPatternDetector()
	})

	DescribeTable(
		"should detect various secret types",
		func(content, expectedPatternName string) {
			findings := detector.Detect(content)
			Expect(findings).NotTo(BeEmpty(), "Expected to detect secret in: %s", content)
			// Find the expected pattern in findings
			var found bool
			for _, f := range findings {
				if f.Pattern.Name == expectedPatternName {
					found = true
					break
				}
			}
			Expect(
				found,
			).To(BeTrue(), "Expected pattern %s in findings for: %s", expectedPatternName, content)
		},
		Entry("AWS Access Key ID", "AKIAIOSFODNN7EXAMPLE", "aws-access-key-id"),
		Entry("GitHub PAT", "ghp_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx", "github-pat"),
		Entry("GitHub OAuth", "gho_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx", "github-oauth"),
		Entry("GitHub App (ghs)", "ghs_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx", "github-app"),
		Entry("GitHub App (ghu)", "ghu_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx", "github-app"),
		Entry("GitHub Refresh", "ghr_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx", "github-refresh"),
		Entry("GitLab PAT", "glpat-abcdefghijklmnopqrst", "gitlab-pat"),
		Entry("Google API Key", "AIzaSyD-abcdefghijklmnopqrstuvwxyz12345", "google-api-key"),
		Entry("RSA Private Key", "-----BEGIN RSA PRIVATE KEY-----", "private-key-rsa"),
		Entry("DSA Private Key", "-----BEGIN DSA PRIVATE KEY-----", "private-key-dsa"),
		Entry("EC Private Key", "-----BEGIN EC PRIVATE KEY-----", "private-key-ec"),
		Entry("OpenSSH Private Key", "-----BEGIN OPENSSH PRIVATE KEY-----", "private-key-openssh"),
		Entry("PGP Private Key", "-----BEGIN PGP PRIVATE KEY BLOCK-----", "private-key-pgp"),
		Entry("MongoDB Connection", "mongodb://user:pass@host:27017/db", "mongodb-conn"),
		Entry("PostgreSQL Connection", "postgresql://user:pass@host:5432/db", "postgres-conn"),
		Entry("MySQL Connection", "mysql://user:pass@host:3306/db", "mysql-conn"),
		Entry("Redis Connection", "redis://user:pass@host:6379", "redis-conn"),
		Entry("NPM Token", "npm_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx", "npm-token"),
		Entry("Stripe Live Key", "sk_live_abcdefghijklmnopqrstuvwx", "stripe-api-key"),
		Entry("Stripe Test Key", "pk_test_abcdefghijklmnopqrstuvwx", "stripe-api-key"),
		Entry(
			"SendGrid Key",
			"SG.abcdefghijklmnopqrstuv.wxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789abc",
			"sendgrid-api-key",
		),
		Entry("Mailgun Key", "key-01234567890123456789012345678901", "mailgun-api-key"),
		//nolint:lll // JWT test data is intentionally long
		Entry(
			"JWT",
			"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgNryP4J3jVmNHl0w5N_XgL0n3I9PlFUP0THsR8U",
			"jwt-token",
		),
	)
})
