package cmd

import (
	"archive/zip"
	"fmt"
	"path"
	"strings"

	"howett.net/plist"
)

type ipaBundleInfo struct {
	Version     string
	BuildNumber string
}

func extractBundleInfoFromIPA(ipaPath string) (ipaBundleInfo, error) {
	reader, err := zip.OpenReader(ipaPath)
	if err != nil {
		return ipaBundleInfo{}, fmt.Errorf("open IPA: %w", err)
	}
	defer reader.Close()

	for _, file := range reader.File {
		if file.FileInfo().IsDir() {
			continue
		}
		if !isTopLevelAppInfoPlist(file.Name) {
			continue
		}
		return readBundleInfoFromInfoPlist(file)
	}

	return ipaBundleInfo{}, fmt.Errorf("Info.plist not found in IPA")
}

func isTopLevelAppInfoPlist(name string) bool {
	cleaned := path.Clean(name)
	if !strings.HasPrefix(cleaned, "Payload/") || !strings.HasSuffix(cleaned, "/Info.plist") {
		return false
	}
	dir := path.Dir(cleaned)
	if !strings.HasSuffix(dir, ".app") {
		return false
	}
	return path.Dir(dir) == "Payload"
}

func readBundleInfoFromInfoPlist(file *zip.File) (ipaBundleInfo, error) {
	reader, err := file.Open()
	if err != nil {
		return ipaBundleInfo{}, fmt.Errorf("open Info.plist: %w", err)
	}
	defer reader.Close()

	var info struct {
		ShortVersion string `plist:"CFBundleShortVersionString"`
		BuildVersion string `plist:"CFBundleVersion"`
	}
	decoder := plist.NewDecoder(reader)
	if err := decoder.Decode(&info); err != nil {
		return ipaBundleInfo{}, fmt.Errorf("decode Info.plist: %w", err)
	}

	return ipaBundleInfo{
		Version:     strings.TrimSpace(info.ShortVersion),
		BuildNumber: strings.TrimSpace(info.BuildVersion),
	}, nil
}
