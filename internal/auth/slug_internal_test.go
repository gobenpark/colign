package auth

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateOrgSlug_Internal(t *testing.T) {
	t.Run("returns lowercase slug with random suffix", func(t *testing.T) {
		slug := generateOrgSlug("My Organization")

		assert.True(t, strings.HasPrefix(slug, "my-organization-"), "slug should start with lowercased, dash-separated name")
		// The suffix is 8 hex characters (4 bytes = 8 hex chars)
		parts := strings.SplitN(slug, "my-organization-", 2)
		require.Len(t, parts, 2)
		suffix := parts[1]
		assert.Len(t, suffix, 8, "random suffix should be 8 hex characters")
	})

	t.Run("replaces spaces with dashes", func(t *testing.T) {
		slug := generateOrgSlug("Hello World Org")

		assert.True(t, strings.HasPrefix(slug, "hello-world-org-"))
		// No spaces should remain in the slug
		assert.False(t, strings.Contains(slug, " "))
	})

	t.Run("lowercases the name", func(t *testing.T) {
		slug := generateOrgSlug("UPPERCASE NAME")

		assert.True(t, strings.HasPrefix(slug, "uppercase-name-"))
		// The prefix should be all lowercase
		prefix := slug[:len(slug)-9] // remove the "-" + 8 hex chars
		assert.Equal(t, strings.ToLower(prefix), prefix)
	})

	t.Run("trims whitespace", func(t *testing.T) {
		slug := generateOrgSlug("  Padded Name  ")

		assert.True(t, strings.HasPrefix(slug, "padded-name-"))
	})

	t.Run("successive calls produce different slugs", func(t *testing.T) {
		slug1 := generateOrgSlug("Same Name")
		slug2 := generateOrgSlug("Same Name")

		assert.NotEqual(t, slug1, slug2, "random suffix should make slugs unique")
	})
}
