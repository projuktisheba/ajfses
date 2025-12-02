package utils

import (
	"log"
	"os"
	"path/filepath"
	"strings"
)

// GetWordpressProjectDirectory returns the full path for a WordPress project
// based on the provided domain name.
// It also empty string when error occurs
func GetWordpressProjectDirectory() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(homeDir, "projuktisheba", "bin", "wordpress")
}

// GetWordpressProjectName returns a unique name for a WordPress project
// based on the provided domain name.
// It also empty string when error occurs
func GetWordpressProjectName(domainName string) string {
	// trim leading and trailing dots
	domainName = strings.Trim(domainName, ".")
	// replace the dots with underscore (example: subdomain.domain.com will be subdomain_domain_com)
	return strings.ReplaceAll(domainName, ".", "_")
}

// GetPHPProjectBaseDirectory returns the full path for the base directory of PHP project
// It also empty string when error occurs
func GetPHPProjectBaseDirectory() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(homeDir, "projuktisheba", "bin", "PHP")
}
// GetPHPProjectDirectory returns the full path for a PHP project
// based on the provided domain name.
// It also empty string when error occurs
func GetPHPProjectDirectory(domainName string) string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Println(err)
		return ""
	}
	return filepath.Join(homeDir, "projuktisheba", "bin", "PHP", domainName)
}


// GetPHPProjectName returns a unique name for a PHP project
// based on the provided domain name.
// It also empty string when error occurs
func GetPHPProjectName(domainName string) string {
	// trim leading and trailing dots
	domainName = strings.Trim(domainName, ".")
	// replace the dots with underscore (example: subdomain.domain.com will be subdomain_domain_com)
	return strings.ReplaceAll(domainName, ".", "_")
}

// GetTempDirectory returns the full path for temporary directory
// It also empty string when error occurs
func GetTempDirectory() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(homeDir, "projuktisheba", "temp")
}
