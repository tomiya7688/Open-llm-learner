package permission

import (
	"fmt"
	"sort"

	"github.com/tomiya7688/Open-llm-learner/internal/modpkg"
)

type DeniedError struct {
	ModID   string
	Missing []string
}

func (e *DeniedError) Error() string {
	return fmt.Sprintf("MOD %s is missing required permission grants: %v", e.ModID, e.Missing)
}

func Missing(requested, granted []string) []string {
	grantSet := make(map[string]struct{}, len(granted))
	for _, value := range granted {
		grantSet[value] = struct{}{}
	}
	var missing []string
	for _, value := range requested {
		if _, ok := grantSet[value]; !ok {
			missing = append(missing, value)
		}
	}
	sort.Strings(missing)
	return missing
}

func CheckManifest(manifest *modpkg.Manifest, granted []string) error {
	if manifest == nil {
		return fmt.Errorf("MOD manifest is nil")
	}
	missing := Missing(manifest.Permissions, granted)
	if len(missing) > 0 {
		return &DeniedError{ModID: manifest.ID, Missing: missing}
	}
	return nil
}
