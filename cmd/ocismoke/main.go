package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	oraclecloud "github.com/Djelibeybi/libdns-oraclecloud"
	"github.com/libdns/libdns"
)

func main() {
	log.SetFlags(0)

	var (
		zone                 = flag.String("zone", "", "DNS zone name or OCI zone OCID (required)")
		name                 = flag.String("name", "", "Relative TXT record name to create; defaults to a random _libdns-smoke-* label")
		value                = flag.String("value", "", "TXT value to write; defaults to a generated smoke-test value")
		ttl                  = flag.Duration("ttl", 30*time.Second, "TXT record TTL")
		timeout              = flag.Duration("timeout", 2*time.Minute, "Overall timeout for the smoke test")
		auth                 = flag.String("auth", "auto", "Provider auth mode: auto, config_file, environment, api_key")
		configFile           = flag.String("config-file", "", "OCI config file path")
		configProfile        = flag.String("config-profile", "", "OCI config profile name")
		privateKey           = flag.String("private-key", "", "Inline OCI private key PEM")
		privateKeyPath       = flag.String("private-key-path", "", "Path to OCI private key PEM")
		privateKeyPassphrase = flag.String("private-key-passphrase", "", "OCI private key passphrase")
		tenancyOCID          = flag.String("tenancy-ocid", "", "OCI tenancy OCID")
		userOCID             = flag.String("user-ocid", "", "OCI user OCID")
		fingerprint          = flag.String("fingerprint", "", "OCI API key fingerprint")
		region               = flag.String("region", "", "OCI region")
		scope                = flag.String("scope", "", "OCI DNS scope: GLOBAL or PRIVATE")
		viewID               = flag.String("view-id", "", "OCI DNS view OCID; required for private zones by name")
	)
	flag.Parse()

	if strings.TrimSpace(*zone) == "" {
		log.Fatal("missing required -zone")
	}

	suffix := randomSuffix()
	recordName := strings.TrimSpace(*name)
	if recordName == "" {
		recordName = "_libdns-smoke-" + suffix
	}

	txtValue := strings.TrimSpace(*value)
	if txtValue == "" {
		txtValue = "libdns-oraclecloud smoke test " + suffix
	}

	provider := &oraclecloud.Provider{
		Auth:                 strings.TrimSpace(*auth),
		ConfigFile:           strings.TrimSpace(*configFile),
		ConfigProfile:        strings.TrimSpace(*configProfile),
		PrivateKey:           strings.TrimSpace(*privateKey),
		PrivateKeyPath:       strings.TrimSpace(*privateKeyPath),
		PrivateKeyPassphrase: strings.TrimSpace(*privateKeyPassphrase),
		TenancyOCID:          strings.TrimSpace(*tenancyOCID),
		UserOCID:             strings.TrimSpace(*userOCID),
		Fingerprint:          strings.TrimSpace(*fingerprint),
		Region:               strings.TrimSpace(*region),
		Scope:                strings.TrimSpace(*scope),
		ViewID:               strings.TrimSpace(*viewID),
	}

	record := libdns.TXT{
		Name: recordName,
		TTL:  *ttl,
		Text: txtValue,
	}

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	log.Printf("Zone: %s", *zone)
	log.Printf("Auth mode: %s", provider.Auth)
	log.Printf("Creating TXT record %q with TTL %s", record.Name, record.TTL)

	created, err := provider.AppendRecords(ctx, *zone, []libdns.Record{record})
	if err != nil {
		log.Fatalf("append TXT record: %v", err)
	}
	if len(created) == 0 {
		log.Fatalf("append TXT record: provider returned no created records")
	}

	printRecords("Created", created)

	log.Printf("Fetching records to confirm access and visibility")
	records, err := provider.GetRecords(ctx, *zone)
	if err != nil {
		log.Fatalf("get records after append: %v", err)
	}

	if !containsTXT(records, record.Name, txtValue) {
		log.Fatalf("TXT record %q with expected value was not found after creation", record.Name)
	}
	log.Printf("Confirmed TXT record is visible via GetRecords")

	log.Printf("Deleting TXT record %q", record.Name)
	deleted, err := provider.DeleteRecords(ctx, *zone, created)
	if err != nil {
		log.Fatalf("delete TXT record: %v", err)
	}
	if len(deleted) == 0 {
		log.Fatalf("delete TXT record: provider returned no deleted records")
	}

	printRecords("Deleted", deleted)

	records, err = provider.GetRecords(ctx, *zone)
	if err != nil {
		log.Fatalf("get records after delete: %v", err)
	}
	if containsTXT(records, record.Name, txtValue) {
		log.Fatalf("TXT record %q is still present after delete", record.Name)
	}

	log.Printf("Smoke test passed")
}

func printRecords(label string, records []libdns.Record) {
	for _, record := range records {
		rr := record.RR()
		log.Printf("%s: %s %s %s %s", label, rr.Name, rr.Type, rr.TTL, rr.Data)
	}
}

func containsTXT(records []libdns.Record, name, value string) bool {
	for _, record := range records {
		txt, ok := record.(libdns.TXT)
		if !ok {
			continue
		}
		if txt.Name == name && txtMatchesValue(txt.Text, value) {
			return true
		}
	}
	return false
}

func txtMatchesValue(actual, expected string) bool {
	if actual == expected {
		return true
	}

	normalized, ok := parseTXTChunks(actual)
	return ok && normalized == expected
}

func parseTXTChunks(input string) (string, bool) {
	input = strings.TrimSpace(input)
	if input == "" || !strings.Contains(input, "\"") {
		return "", false
	}

	var out strings.Builder
	for len(input) > 0 {
		input = strings.TrimSpace(input)
		if input == "" {
			break
		}
		if input[0] != '"' {
			return "", false
		}

		end := findQuotedChunkEnd(input)
		if end <= 0 {
			return "", false
		}

		part, err := strconv.Unquote(input[:end])
		if err != nil {
			return "", false
		}
		out.WriteString(part)
		input = input[end:]
	}

	return out.String(), true
}

func findQuotedChunkEnd(input string) int {
	escaped := false
	for i := 1; i < len(input); i++ {
		switch {
		case escaped:
			escaped = false
		case input[i] == '\\':
			escaped = true
		case input[i] == '"':
			return i + 1
		}
	}
	return -1
}

func randomSuffix() string {
	var buf [4]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(buf[:])
}
