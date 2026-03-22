# libdns Oracle Cloud

`libdns` provider for Oracle Cloud Infrastructure DNS.

This package implements:

- `GetRecords`
- `AppendRecords`
- `SetRecords`
- `DeleteRecords`
- `ListZones`

## Authentication

The provider currently keeps OCI auth intentionally simple:

- explicit API key fields on `oraclecloud.Provider`
- OCI config file credentials
- Oracle CLI environment variables

Supported `Auth` values:

- `""` or `auto`
- `api_key`
- `config_file`
- `environment`

`instance_principal` is also wired through, but the package is primarily aimed at API-key based usage for now.

If you use `Auth: "config_file"` with `ConfigFile: "~/.oci/config"`, you do not also need to provide
`TenancyOCID`, `UserOCID`, `Fingerprint`, `Region`, or `PrivateKey*` fields. `ConfigProfile` is optional
and defaults to `DEFAULT` (or `OCI_CLI_PROFILE` if set).

## Provider Fields

```go
provider := oraclecloud.Provider{
    Auth:               "auto",
    ConfigFile:         "~/.oci/config",
    ConfigProfile:      "DEFAULT",
    TenancyOCID:        "ocid1.tenancy.oc1..example",
    UserOCID:           "ocid1.user.oc1..example",
    Fingerprint:        "12:34:56:78:90:ab:cd:ef:12:34:56:78:90:ab:cd:ef",
    PrivateKeyPath:     "~/.oci/oci_api_key.pem",
    PrivateKeyPassphrase: "",
    Region:             "us-phoenix-1",
    Scope:              "GLOBAL",
    ViewID:             "",
    CompartmentID:      "ocid1.compartment.oc1..example",
}
```

## Environment Variables

The provider accepts Oracle's documented OCI CLI environment variables:

- `OCI_CLI_KEY_CONTENT` or `OCI_CLI_KEY_FILE`
- `OCI_CLI_PASSPHRASE`
- `OCI_CLI_TENANCY`
- `OCI_CLI_USER`
- `OCI_CLI_FINGERPRINT`
- `OCI_CLI_REGION`
- `OCI_CLI_CONFIG_FILE`
- `OCI_CLI_PROFILE`

## Versioning

This module follows Semantic Versioning, with Git tags like `v1.2.3`.

For Go modules, that means:

- `fix:` conventional commits map to patch releases
- `feat:` conventional commits map to minor releases
- `feat!:` or any commit with a `BREAKING CHANGE:` footer maps to a major release
- if this module ever releases `v2+`, the Go module path must also change to include the major suffix
  such as `github.com/Djelibeybi/libdns-oraclecloud/v2`

GitHub Actions uses `release-please` to turn conventional commits merged to `main` into release PRs,
SemVer tags, changelog updates, and GitHub Releases.

## Notes

- In `Auth: "auto"` mode, explicit API-key fields take precedence over config-file auth. If you want to
  force use of `~/.oci/config`, set `Auth: "config_file"`.
- For private zones accessed by name, OCI requires `ViewID`.
- `SetRecords` is atomic per RRSet because OCI exposes RRSet replacement as a single operation, but it is not atomic across multiple distinct RRSets.
- The module path is `github.com/Djelibeybi/libdns-oraclecloud`.
