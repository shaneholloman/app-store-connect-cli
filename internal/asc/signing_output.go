package asc

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
)

// BundleIDDeleteResult represents CLI output for bundle ID deletions.
type BundleIDDeleteResult struct {
	ID      string `json:"id"`
	Deleted bool   `json:"deleted"`
}

// BundleIDCapabilityDeleteResult represents CLI output for capability deletions.
type BundleIDCapabilityDeleteResult struct {
	ID      string `json:"id"`
	Deleted bool   `json:"deleted"`
}

// CertificateRevokeResult represents CLI output for certificate revocations.
type CertificateRevokeResult struct {
	ID      string `json:"id"`
	Revoked bool   `json:"revoked"`
}

// ProfileDeleteResult represents CLI output for profile deletions.
type ProfileDeleteResult struct {
	ID      string `json:"id"`
	Deleted bool   `json:"deleted"`
}

// ProfileDownloadResult represents CLI output for profile downloads.
type ProfileDownloadResult struct {
	ID         string `json:"id"`
	Name       string `json:"name,omitempty"`
	OutputPath string `json:"outputPath"`
}

func printBundleIDsTable(resp *BundleIDsResponse) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tName\tIdentifier\tPlatform\tSeed ID")
	for _, item := range resp.Data {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			item.ID,
			compactWhitespace(item.Attributes.Name),
			item.Attributes.Identifier,
			item.Attributes.Platform,
			item.Attributes.SeedID,
		)
	}
	return w.Flush()
}

func printBundleIDsMarkdown(resp *BundleIDsResponse) error {
	fmt.Fprintln(os.Stdout, "| ID | Name | Identifier | Platform | Seed ID |")
	fmt.Fprintln(os.Stdout, "| --- | --- | --- | --- | --- |")
	for _, item := range resp.Data {
		fmt.Fprintf(os.Stdout, "| %s | %s | %s | %s | %s |\n",
			item.ID,
			escapeMarkdown(item.Attributes.Name),
			escapeMarkdown(item.Attributes.Identifier),
			escapeMarkdown(string(item.Attributes.Platform)),
			escapeMarkdown(item.Attributes.SeedID),
		)
	}
	return nil
}

func printBundleIDCapabilitiesTable(resp *BundleIDCapabilitiesResponse) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tCapability\tSettings")
	for _, item := range resp.Data {
		fmt.Fprintf(w, "%s\t%s\t%s\n",
			item.ID,
			item.Attributes.CapabilityType,
			formatCapabilitySettings(item.Attributes.Settings),
		)
	}
	return w.Flush()
}

func printBundleIDCapabilitiesMarkdown(resp *BundleIDCapabilitiesResponse) error {
	fmt.Fprintln(os.Stdout, "| ID | Capability | Settings |")
	fmt.Fprintln(os.Stdout, "| --- | --- | --- |")
	for _, item := range resp.Data {
		fmt.Fprintf(os.Stdout, "| %s | %s | %s |\n",
			item.ID,
			escapeMarkdown(item.Attributes.CapabilityType),
			escapeMarkdown(formatCapabilitySettings(item.Attributes.Settings)),
		)
	}
	return nil
}

func printBundleIDDeleteResultTable(result *BundleIDDeleteResult) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tDeleted")
	fmt.Fprintf(w, "%s\t%t\n", result.ID, result.Deleted)
	return w.Flush()
}

func printBundleIDDeleteResultMarkdown(result *BundleIDDeleteResult) error {
	fmt.Fprintln(os.Stdout, "| ID | Deleted |")
	fmt.Fprintln(os.Stdout, "| --- | --- |")
	fmt.Fprintf(os.Stdout, "| %s | %t |\n",
		escapeMarkdown(result.ID),
		result.Deleted,
	)
	return nil
}

func printBundleIDCapabilityDeleteResultTable(result *BundleIDCapabilityDeleteResult) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tDeleted")
	fmt.Fprintf(w, "%s\t%t\n", result.ID, result.Deleted)
	return w.Flush()
}

func printBundleIDCapabilityDeleteResultMarkdown(result *BundleIDCapabilityDeleteResult) error {
	fmt.Fprintln(os.Stdout, "| ID | Deleted |")
	fmt.Fprintln(os.Stdout, "| --- | --- |")
	fmt.Fprintf(os.Stdout, "| %s | %t |\n",
		escapeMarkdown(result.ID),
		result.Deleted,
	)
	return nil
}

func printCertificatesTable(resp *CertificatesResponse) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tName\tType\tExpiration\tSerial")
	for _, item := range resp.Data {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			item.ID,
			compactWhitespace(certificateDisplayName(item.Attributes)),
			item.Attributes.CertificateType,
			item.Attributes.ExpirationDate,
			item.Attributes.SerialNumber,
		)
	}
	return w.Flush()
}

func printCertificatesMarkdown(resp *CertificatesResponse) error {
	fmt.Fprintln(os.Stdout, "| ID | Name | Type | Expiration | Serial |")
	fmt.Fprintln(os.Stdout, "| --- | --- | --- | --- | --- |")
	for _, item := range resp.Data {
		fmt.Fprintf(os.Stdout, "| %s | %s | %s | %s | %s |\n",
			item.ID,
			escapeMarkdown(certificateDisplayName(item.Attributes)),
			escapeMarkdown(item.Attributes.CertificateType),
			escapeMarkdown(item.Attributes.ExpirationDate),
			escapeMarkdown(item.Attributes.SerialNumber),
		)
	}
	return nil
}

func printCertificateRevokeResultTable(result *CertificateRevokeResult) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tRevoked")
	fmt.Fprintf(w, "%s\t%t\n", result.ID, result.Revoked)
	return w.Flush()
}

func printCertificateRevokeResultMarkdown(result *CertificateRevokeResult) error {
	fmt.Fprintln(os.Stdout, "| ID | Revoked |")
	fmt.Fprintln(os.Stdout, "| --- | --- |")
	fmt.Fprintf(os.Stdout, "| %s | %t |\n",
		escapeMarkdown(result.ID),
		result.Revoked,
	)
	return nil
}

func printProfilesTable(resp *ProfilesResponse) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tName\tType\tState\tExpiration")
	for _, item := range resp.Data {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			item.ID,
			compactWhitespace(item.Attributes.Name),
			item.Attributes.ProfileType,
			item.Attributes.ProfileState,
			item.Attributes.ExpirationDate,
		)
	}
	return w.Flush()
}

func printProfilesMarkdown(resp *ProfilesResponse) error {
	fmt.Fprintln(os.Stdout, "| ID | Name | Type | State | Expiration |")
	fmt.Fprintln(os.Stdout, "| --- | --- | --- | --- | --- |")
	for _, item := range resp.Data {
		fmt.Fprintf(os.Stdout, "| %s | %s | %s | %s | %s |\n",
			item.ID,
			escapeMarkdown(item.Attributes.Name),
			escapeMarkdown(item.Attributes.ProfileType),
			escapeMarkdown(string(item.Attributes.ProfileState)),
			escapeMarkdown(item.Attributes.ExpirationDate),
		)
	}
	return nil
}

func printProfileDeleteResultTable(result *ProfileDeleteResult) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tDeleted")
	fmt.Fprintf(w, "%s\t%t\n", result.ID, result.Deleted)
	return w.Flush()
}

func printProfileDeleteResultMarkdown(result *ProfileDeleteResult) error {
	fmt.Fprintln(os.Stdout, "| ID | Deleted |")
	fmt.Fprintln(os.Stdout, "| --- | --- |")
	fmt.Fprintf(os.Stdout, "| %s | %t |\n",
		escapeMarkdown(result.ID),
		result.Deleted,
	)
	return nil
}

func printProfileDownloadResultTable(result *ProfileDownloadResult) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tName\tOutput Path")
	fmt.Fprintf(w, "%s\t%s\t%s\n",
		result.ID,
		compactWhitespace(result.Name),
		result.OutputPath,
	)
	return w.Flush()
}

func printProfileDownloadResultMarkdown(result *ProfileDownloadResult) error {
	fmt.Fprintln(os.Stdout, "| ID | Name | Output Path |")
	fmt.Fprintln(os.Stdout, "| --- | --- | --- |")
	fmt.Fprintf(os.Stdout, "| %s | %s | %s |\n",
		escapeMarkdown(result.ID),
		escapeMarkdown(result.Name),
		escapeMarkdown(result.OutputPath),
	)
	return nil
}

func joinSigningList(values []string) string {
	if len(values) == 0 {
		return ""
	}
	return strings.Join(values, ", ")
}

func printSigningFetchResultTable(result *SigningFetchResult) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Bundle ID\tBundle ID Resource\tProfile Type\tProfile ID\tProfile File\tCertificate IDs\tCertificate Files\tCreated")
	fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%t\n",
		result.BundleID,
		result.BundleIDResource,
		result.ProfileType,
		result.ProfileID,
		result.ProfileFile,
		joinSigningList(result.CertificateIDs),
		joinSigningList(result.CertificateFiles),
		result.Created,
	)
	return w.Flush()
}

func printSigningFetchResultMarkdown(result *SigningFetchResult) error {
	fmt.Fprintln(os.Stdout, "| Bundle ID | Bundle ID Resource | Profile Type | Profile ID | Profile File | Certificate IDs | Certificate Files | Created |")
	fmt.Fprintln(os.Stdout, "| --- | --- | --- | --- | --- | --- | --- | --- |")
	fmt.Fprintf(os.Stdout, "| %s | %s | %s | %s | %s | %s | %s | %t |\n",
		escapeMarkdown(result.BundleID),
		escapeMarkdown(result.BundleIDResource),
		escapeMarkdown(result.ProfileType),
		escapeMarkdown(result.ProfileID),
		escapeMarkdown(result.ProfileFile),
		escapeMarkdown(joinSigningList(result.CertificateIDs)),
		escapeMarkdown(joinSigningList(result.CertificateFiles)),
		result.Created,
	)
	return nil
}

func formatCapabilitySettings(settings []CapabilitySetting) string {
	if len(settings) == 0 {
		return ""
	}
	payload, err := json.Marshal(settings)
	if err != nil {
		return ""
	}
	return sanitizeTerminal(string(payload))
}

func certificateDisplayName(attrs CertificateAttributes) string {
	if strings.TrimSpace(attrs.DisplayName) != "" {
		return attrs.DisplayName
	}
	return attrs.Name
}
