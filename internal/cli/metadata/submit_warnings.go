package metadata

import (
	"github.com/rudrankriyam/App-Store-Connect-CLI/internal/cli/shared"
)

func effectiveVersionCreateLocalization(localPatch versionLocalPatch) VersionLocalization {
	createLoc := localPatch.localization
	if hasVersionContent(localPatch.createLocalization) {
		createLoc = localPatch.createLocalization
	}
	return NormalizeVersionLocalization(createLoc)
}

func versionCreateWarningForPatch(locale string, localPatch versionLocalPatch, remoteExists bool, mode string) (shared.SubmitReadinessCreateWarning, bool) {
	if remoteExists {
		return shared.SubmitReadinessCreateWarning{}, false
	}
	return shared.SubmitReadinessCreateWarningForLocale(
		locale,
		versionAttributes(locale, effectiveVersionCreateLocalization(localPatch), true),
		mode,
	)
}

func versionCreateWarningsForPatches(local map[string]versionLocalPatch, remote map[string]VersionLocalization, mode string) []shared.SubmitReadinessCreateWarning {
	warnings := make([]shared.SubmitReadinessCreateWarning, 0, len(local))
	for _, locale := range sortedKeys(local) {
		warning, ok := versionCreateWarningForPatch(locale, local[locale], hasRemoteVersionLocalization(remote, locale), mode)
		if !ok {
			continue
		}
		warnings = append(warnings, warning)
	}
	return shared.NormalizeSubmitReadinessCreateWarnings(warnings)
}

func hasRemoteVersionLocalization(remote map[string]VersionLocalization, locale string) bool {
	if len(remote) == 0 {
		return false
	}
	_, ok := remote[locale]
	return ok
}
