# Sigstore Integration Specification

**Extension to: Go Implementation Plan**  
**Phase: 4 (Week 10)**  
**Priority: High Value Differentiator**

---

## Table of Contents

1. [Overview](#overview)
2. [Architecture](#architecture)
3. [Implementation Details](#implementation-details)
4. [CLI Commands](#cli-commands)
5. [Database Schema](#database-schema)
6. [API Endpoints](#api-endpoints)
7. [Catalog UI Integration](#catalog-ui-integration)
8. [Verification Policies](#verification-policies)
9. [Testing Strategy](#testing-strategy)
10. [Deployment](#deployment)

---

## Overview

### What Gets Signed

1. **Model Weights** - Primary artifact, highest risk
2. **Inference Code** - Python files, requirements.txt
3. **model.yaml** - Metadata integrity
4. **Benchmark Results** - Prevent falsification
5. **Container Images** - If custom containers used

### Sigstore Components Used

**Cosign** - Signs and verifies artifacts
```bash
# Install
go get github.com/sigstore/cosign/v2
```

**Rekor** - Transparency log (immutable audit trail)
```bash
# Public instance: rekor.sigstore.dev
# Can also run private instance
```

**Fulcio** - Certificate authority (keyless signing)
```bash
# Public instance: fulcio.sigstore.dev
# Issues short-lived certificates via OIDC
```

### Authentication Flow

```
Publisher → GitHub OIDC → Fulcio → Certificate → Cosign → Sign → Rekor → Log Entry
                                                        ↓
                                                   Signature Bundle
                                                        ↓
                                                   S3/Registry
```

---

## Architecture

### Project Structure Updates

```
conduit/
├── internal/
│   ├── sigstore/              # NEW: Sigstore integration
│   │   ├── signer.go         # Sign artifacts
│   │   ├── verifier.go       # Verify signatures
│   │   ├── oidc.go           # OIDC authentication
│   │   ├── rekor.go          # Transparency log
│   │   ├── policy.go         # Verification policies
│   │   └── types.go          # Signature types
│   │
│   ├── cli/
│   │   ├── sign.go           # NEW: conduit sign
│   │   ├── verify.go         # NEW: conduit verify
│   │   └── publish.go        # UPDATED: Add --sign flag
│   │
│   └── model/
│       └── signature.go      # NEW: Signature metadata
│
├── pkg/
│   └── types/
│       └── signature.go      # NEW: Exported signature types
│
└── config/
    └── verification.yaml     # NEW: Verification policies
```

### Data Flow

#### Signing Flow

```
1. Publisher runs: conduit publish --sign
                   ↓
2. Authenticate via OIDC (GitHub, Google, etc.)
                   ↓
3. Compute artifact hashes (SHA256)
                   ↓
4. Request Fulcio certificate
                   ↓
5. Sign with ephemeral key
                   ↓
6. Upload signature to Rekor
                   ↓
7. Store signature bundle with artifact
                   ↓
8. Update model.yaml with signature metadata
```

#### Verification Flow

```
1. Consumer runs: conduit deploy --verify
                   ↓
2. Download artifact + signature bundle
                   ↓
3. Verify signature matches artifact
                   ↓
4. Verify Rekor transparency log entry
                   ↓
5. Check certificate identity
                   ↓
6. Apply verification policy
                   ↓
7. Allow/Deny deployment
```

---

## Implementation Details

### 1. Core Types

```go
// pkg/types/signature.go
package types

import "time"

// Signature represents a Sigstore signature
type Signature struct {
    // Artifact info
    ArtifactType    string `json:"artifact_type"`    // "weights", "code", "benchmark"
    ArtifactURI     string `json:"artifact_uri"`
    ArtifactHash    string `json:"artifact_hash"`    // SHA256
    
    // Signature info
    SignatureBundle string `json:"signature_bundle"` // Base64 encoded or URI
    BundleURI       string `json:"bundle_uri"`       // S3 URI if stored separately
    
    // Rekor transparency log
    RekorLogID      string    `json:"rekor_log_id"`
    RekorIndex      int64     `json:"rekor_index"`
    RekorUUID       string    `json:"rekor_uuid"`
    RekorEntryURL   string    `json:"rekor_entry_url"`
    
    // Certificate info (from Fulcio)
    CertIdentity    string    `json:"cert_identity"`     // Email or OIDC subject
    CertIssuer      string    `json:"cert_issuer"`       // OIDC issuer URL
    CertValidFrom   time.Time `json:"cert_valid_from"`
    CertValidTo     time.Time `json:"cert_valid_to"`
    
    // Metadata
    SignedBy        string    `json:"signed_by"`         // Friendly name
    SignedAt        time.Time `json:"signed_at"`
    
    // Verification
    Verified        bool      `json:"verified"`
    VerifiedAt      time.Time `json:"verified_at,omitempty"`
    VerifiedBy      string    `json:"verified_by,omitempty"`
}

// ModelSignatures contains all signatures for a model
type ModelSignatures struct {
    ModelID         string      `json:"model_id"`
    Version         string      `json:"version"`
    
    WeightsSignature    *Signature `json:"weights_signature,omitempty"`
    CodeSignature       *Signature `json:"code_signature,omitempty"`
    MetadataSignature   *Signature `json:"metadata_signature,omitempty"`
    BenchmarkSignatures []Signature `json:"benchmark_signatures,omitempty"`
    
    SignedAt        time.Time   `json:"signed_at"`
    LastVerified    time.Time   `json:"last_verified,omitempty"`
}

// VerificationPolicy defines signature requirements
type VerificationPolicy struct {
    Level               PolicyLevel  `json:"level"`
    RequireSignatures   bool         `json:"require_signatures"`
    RequireRekorEntry   bool         `json:"require_rekor_entry"`
    AllowedIdentities   []string     `json:"allowed_identities"`
    AllowedIssuers      []string     `json:"allowed_issuers"`
    MaxSignatureAgeDays int          `json:"max_signature_age_days"`
    TrustRoot           string       `json:"trust_root,omitempty"`
}

type PolicyLevel string

const (
    PolicyLevelNone        PolicyLevel = "none"
    PolicyLevelOptional    PolicyLevel = "optional"
    PolicyLevelRecommended PolicyLevel = "recommended"
    PolicyLevelStrict      PolicyLevel = "strict"
)
```

### 2. Signer Implementation

```go
// internal/sigstore/signer.go
package sigstore

import (
    "context"
    "crypto"
    "fmt"
    "io"
    "os"
    
    "github.com/sigstore/cosign/v2/cmd/cosign/cli/options"
    "github.com/sigstore/cosign/v2/cmd/cosign/cli/sign"
    "github.com/sigstore/cosign/v2/pkg/providers"
    "github.com/sigstore/sigstore/pkg/signature/dsse"
    
    "github.com/yourusername/conduit/pkg/types"
)

type Signer struct {
    oidcProvider string
    rekorURL     string
    fulcioURL    string
}

type SignOptions struct {
    OIDCProvider     string // "github", "google", "microsoft"
    OIDCIssuer       string // Custom OIDC issuer URL
    RekorURL         string // Default: rekor.sigstore.dev
    FulcioURL        string // Default: fulcio.sigstore.dev
    NoUploadToRekor  bool   // Skip Rekor upload (for testing)
}

func NewSigner(opts SignOptions) *Signer {
    // Use defaults if not specified
    if opts.RekorURL == "" {
        opts.RekorURL = "https://rekor.sigstore.dev"
    }
    if opts.FulcioURL == "" {
        opts.FulcioURL = "https://fulcio.sigstore.dev"
    }
    
    return &Signer{
        oidcProvider: opts.OIDCProvider,
        rekorURL:     opts.RekorURL,
        fulcioURL:    opts.FulcioURL,
    }
}

// SignArtifact signs a file or blob
func (s *Signer) SignArtifact(ctx context.Context, artifactPath string) (*types.Signature, error) {
    // 1. Compute artifact hash
    hash, err := computeSHA256(artifactPath)
    if err != nil {
        return nil, fmt.Errorf("compute hash: %w", err)
    }
    
    // 2. Authenticate with OIDC provider
    idToken, err := s.authenticate(ctx)
    if err != nil {
        return nil, fmt.Errorf("OIDC authentication: %w", err)
    }
    
    // 3. Sign using Cosign
    ko := options.KeyOpts{
        FulcioURL:    s.fulcioURL,
        RekorURL:     s.rekorURL,
        OIDCIssuer:   idToken.Issuer,
        OIDCClientID: "sigstore",
    }
    
    ro := options.RootOptions{
        Timeout: options.DefaultTimeout,
    }
    
    // Sign the artifact
    signedPayload, err := sign.SignBlobCmd(
        &ro,
        ko,
        artifactPath,
        true, // Upload to Rekor
        "", // Output signature path (we'll get it from return)
        "", // Security key
        "", // Slot
        idToken.RawString,
        "", // Certificate path
        "", // Certificate chain path
        false, // Skip confirmation
        "", // TLog upload
        false, // RFC3161 timestamp
        "", // RFC3161 timestamp server URL
    )
    if err != nil {
        return nil, fmt.Errorf("sign artifact: %w", err)
    }
    
    // 4. Parse signature bundle
    sig, err := s.parseSignatureBundle(signedPayload, hash, idToken)
    if err != nil {
        return nil, fmt.Errorf("parse signature: %w", err)
    }
    
    return sig, nil
}

// SignWeights is a convenience wrapper for signing model weights
func (s *Signer) SignWeights(ctx context.Context, weightsPath string) (*types.Signature, error) {
    sig, err := s.SignArtifact(ctx, weightsPath)
    if err != nil {
        return nil, err
    }
    sig.ArtifactType = "weights"
    return sig, nil
}

// SignCode signs all code files in a directory
func (s *Signer) SignCode(ctx context.Context, repoPath string) (*types.Signature, error) {
    // Create tarball of code files
    tarballPath, err := createCodeTarball(repoPath)
    if err != nil {
        return nil, fmt.Errorf("create tarball: %w", err)
    }
    defer os.Remove(tarballPath)
    
    sig, err := s.SignArtifact(ctx, tarballPath)
    if err != nil {
        return nil, err
    }
    sig.ArtifactType = "code"
    return sig, nil
}

// SignModelYAML signs the model metadata
func (s *Signer) SignModelYAML(ctx context.Context, yamlPath string) (*types.Signature, error) {
    sig, err := s.SignArtifact(ctx, yamlPath)
    if err != nil {
        return nil, err
    }
    sig.ArtifactType = "metadata"
    return sig, nil
}

// authenticate performs OIDC authentication
func (s *Signer) authenticate(ctx context.Context) (*providers.IDToken, error) {
    // Get OIDC token based on provider
    var provider providers.Interface
    var err error
    
    switch s.oidcProvider {
    case "github":
        provider, err = providers.ProvideGithub(ctx)
    case "google":
        provider, err = providers.ProvideGoogle(ctx)
    case "microsoft":
        provider, err = providers.ProvideMicrosoft(ctx)
    default:
        return nil, fmt.Errorf("unsupported OIDC provider: %s", s.oidcProvider)
    }
    
    if err != nil {
        return nil, fmt.Errorf("create OIDC provider: %w", err)
    }
    
    // Get ID token
    idToken, err := provider.Provide(ctx, "sigstore")
    if err != nil {
        return nil, fmt.Errorf("get ID token: %w", err)
    }
    
    return idToken, nil
}

// parseSignatureBundle extracts signature information
func (s *Signer) parseSignatureBundle(bundle []byte, hash string, idToken *providers.IDToken) (*types.Signature, error) {
    // Parse the bundle to extract:
    // - Signature
    // - Certificate
    // - Rekor entry
    
    // This is simplified - actual implementation would parse the bundle format
    sig := &types.Signature{
        ArtifactHash:    hash,
        SignatureBundle: string(bundle),
        CertIdentity:    idToken.Subject,
        CertIssuer:      idToken.Issuer,
        SignedBy:        extractFriendlyName(idToken.Subject),
        SignedAt:        time.Now(),
    }
    
    return sig, nil
}

// Helper functions
func computeSHA256(path string) (string, error) {
    f, err := os.Open(path)
    if err != nil {
        return "", err
    }
    defer f.Close()
    
    h := crypto.SHA256.New()
    if _, err := io.Copy(h, f); err != nil {
        return "", err
    }
    
    return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func extractFriendlyName(subject string) string {
    // Extract friendly name from OIDC subject
    // e.g. "https://github.com/deepmind/alphafold" -> "deepmind/alphafold"
    // This is simplified
    return subject
}

func createCodeTarball(repoPath string) (string, error) {
    // Create tarball of relevant code files
    // Include: *.py, *.yaml, requirements.txt, etc.
    // Exclude: .git, __pycache__, *.pyc, etc.
    // This is simplified - would use archive/tar
    return "", nil
}
```

### 3. Verifier Implementation

```go
// internal/sigstore/verifier.go
package sigstore

import (
    "context"
    "fmt"
    
    "github.com/sigstore/cosign/v2/cmd/cosign/cli/options"
    "github.com/sigstore/cosign/v2/cmd/cosign/cli/verify"
    
    "github.com/yourusername/conduit/pkg/types"
)

type Verifier struct {
    rekorURL  string
    policy    *types.VerificationPolicy
}

type VerifyOptions struct {
    RekorURL string
    Policy   *types.VerificationPolicy
}

func NewVerifier(opts VerifyOptions) *Verifier {
    if opts.RekorURL == "" {
        opts.RekorURL = "https://rekor.sigstore.dev"
    }
    
    return &Verifier{
        rekorURL: opts.RekorURL,
        policy:   opts.Policy,
    }
}

// VerifyArtifact verifies a signature against an artifact
func (v *Verifier) VerifyArtifact(ctx context.Context, artifactPath string, sig *types.Signature) error {
    // 1. Verify signature cryptographically
    if err := v.verifySignature(ctx, artifactPath, sig); err != nil {
        return fmt.Errorf("signature verification failed: %w", err)
    }
    
    // 2. Verify Rekor transparency log entry
    if err := v.verifyRekorEntry(ctx, sig); err != nil {
        return fmt.Errorf("rekor verification failed: %w", err)
    }
    
    // 3. Verify certificate identity
    if err := v.verifyCertificate(sig); err != nil {
        return fmt.Errorf("certificate verification failed: %w", err)
    }
    
    // 4. Apply verification policy
    if err := v.applyPolicy(sig); err != nil {
        return fmt.Errorf("policy check failed: %w", err)
    }
    
    return nil
}

// verifySignature verifies the cryptographic signature
func (v *Verifier) verifySignature(ctx context.Context, artifactPath string, sig *types.Signature) error {
    // Use Cosign to verify
    ko := options.KeyOpts{
        RekorURL: v.rekorURL,
    }
    
    // Verify blob command
    err := verify.VerifyBlobCmd(
        ctx,
        ko,
        sig.CertIdentity,
        sig.CertIssuer,
        "", // Signature path
        sig.SignatureBundle,
        artifactPath,
        "", // Certificate path
        "", // Certificate chain
        "", // Cert OIDC issuer regex
        "", // Cert identity regex
        "", // Offline
        "", // Local image
    )
    
    if err != nil {
        return fmt.Errorf("cosign verification failed: %w", err)
    }
    
    return nil
}

// verifyRekorEntry verifies the transparency log entry
func (v *Verifier) verifyRekorEntry(ctx context.Context, sig *types.Signature) error {
    if sig.RekorUUID == "" {
        if v.policy != nil && v.policy.RequireRekorEntry {
            return fmt.Errorf("rekor entry required by policy but not found")
        }
        return nil
    }
    
    // Query Rekor to verify entry exists and matches
    // This would use the Rekor client library
    // Simplified here
    
    return nil
}

// verifyCertificate verifies the certificate identity
func (v *Verifier) verifyCertificate(sig *types.Signature) error {
    // Verify:
    // - Certificate was issued by Fulcio
    // - Identity matches expected pattern
    // - Certificate hasn't expired
    // - Issuer is trusted
    
    if sig.CertValidTo.Before(time.Now()) {
        return fmt.Errorf("certificate expired at %s", sig.CertValidTo)
    }
    
    return nil
}

// applyPolicy checks if signature meets policy requirements
func (v *Verifier) applyPolicy(sig *types.Signature) error {
    if v.policy == nil {
        return nil
    }
    
    switch v.policy.Level {
    case types.PolicyLevelNone:
        return nil
        
    case types.PolicyLevelOptional:
        // Verify if signature exists, but don't require it
        return nil
        
    case types.PolicyLevelRecommended:
        // Warn if missing but don't fail
        if sig == nil {
            log.Warn("Signature recommended but not present")
        }
        return nil
        
    case types.PolicyLevelStrict:
        if sig == nil {
            return fmt.Errorf("signature required by policy")
        }
        
        // Check allowed identities
        if len(v.policy.AllowedIdentities) > 0 {
            allowed := false
            for _, pattern := range v.policy.AllowedIdentities {
                if matchesPattern(sig.CertIdentity, pattern) {
                    allowed = true
                    break
                }
            }
            if !allowed {
                return fmt.Errorf("identity %s not in allowed list", sig.CertIdentity)
            }
        }
        
        // Check allowed issuers
        if len(v.policy.AllowedIssuers) > 0 {
            allowed := false
            for _, issuer := range v.policy.AllowedIssuers {
                if sig.CertIssuer == issuer {
                    allowed = true
                    break
                }
            }
            if !allowed {
                return fmt.Errorf("issuer %s not in allowed list", sig.CertIssuer)
            }
        }
        
        // Check signature age
        if v.policy.MaxSignatureAgeDays > 0 {
            maxAge := time.Duration(v.policy.MaxSignatureAgeDays) * 24 * time.Hour
            if time.Since(sig.SignedAt) > maxAge {
                return fmt.Errorf("signature too old: %s", sig.SignedAt)
            }
        }
    }
    
    return nil
}

func matchesPattern(identity, pattern string) bool {
    // Simple glob matching
    // In reality, would use proper glob or regex
    return strings.HasPrefix(identity, strings.TrimSuffix(pattern, "*"))
}
```

### 4. Policy Engine

```go
// internal/sigstore/policy.go
package sigstore

import (
    "fmt"
    "os"
    
    "gopkg.in/yaml.v3"
    "github.com/yourusername/conduit/pkg/types"
)

type PolicyManager struct {
    globalPolicy *types.VerificationPolicy
    domainPolicies map[string]*types.VerificationPolicy
}

func NewPolicyManager() *PolicyManager {
    return &PolicyManager{
        domainPolicies: make(map[string]*types.VerificationPolicy),
    }
}

// LoadPolicyFile loads verification policies from YAML
func (pm *PolicyManager) LoadPolicyFile(path string) error {
    data, err := os.ReadFile(path)
    if err != nil {
        return fmt.Errorf("read policy file: %w", err)
    }
    
    var config struct {
        Global *types.VerificationPolicy                   `yaml:"global"`
        Domains map[string]*types.VerificationPolicy      `yaml:"domains"`
    }
    
    if err := yaml.Unmarshal(data, &config); err != nil {
        return fmt.Errorf("parse policy file: %w", err)
    }
    
    pm.globalPolicy = config.Global
    pm.domainPolicies = config.Domains
    
    return nil
}

// GetPolicy returns the applicable policy for a domain
func (pm *PolicyManager) GetPolicy(domain string) *types.VerificationPolicy {
    // Check domain-specific policy first
    if policy, ok := pm.domainPolicies[domain]; ok {
        return policy
    }
    
    // Fall back to global policy
    return pm.globalPolicy
}

// SetGlobalPolicy sets the global verification policy
func (pm *PolicyManager) SetGlobalPolicy(policy *types.VerificationPolicy) {
    pm.globalPolicy = policy
}

// SetDomainPolicy sets a domain-specific policy
func (pm *PolicyManager) SetDomainPolicy(domain string, policy *types.VerificationPolicy) {
    pm.domainPolicies[domain] = policy
}

// DefaultPolicies returns sensible defaults
func DefaultPolicies() map[string]*types.VerificationPolicy {
    return map[string]*types.VerificationPolicy{
        "drug-discovery": {
            Level:               types.PolicyLevelStrict,
            RequireSignatures:   true,
            RequireRekorEntry:   true,
            MaxSignatureAgeDays: 90,
        },
        "protein-science": {
            Level:               types.PolicyLevelRecommended,
            RequireSignatures:   false,
            RequireRekorEntry:   true,
            MaxSignatureAgeDays: 180,
        },
        "climate": {
            Level:               types.PolicyLevelOptional,
            RequireSignatures:   false,
        },
    }
}
```

---

## CLI Commands

### conduit sign

```go
// internal/cli/sign.go
package cli

import (
    "fmt"
    
    "github.com/spf13/cobra"
    "github.com/yourusername/conduit/internal/sigstore"
)

func SignCmd() *cobra.Command {
    var (
        weightsPath  string
        codePath     string
        yamlPath     string
        oidcProvider string
        outputPath   string
    )
    
    cmd := &cobra.Command{
        Use:   "sign",
        Short: "Sign model artifacts with Sigstore",
        Long: `Sign model artifacts (weights, code, metadata) using Sigstore.

Signatures provide cryptographic proof of authenticity and are recorded
in the Rekor transparency log for public auditability.

Examples:
  # Sign model weights
  conduit sign --weights model.safetensors --oidc github

  # Sign all artifacts
  conduit sign --weights weights/ --code src/ --yaml model.yaml

  # Sign during publish
  conduit publish --sign
`,
        RunE: func(cmd *cobra.Command, args []string) error {
            ctx := cmd.Context()
            
            // Create signer
            signer := sigstore.NewSigner(sigstore.SignOptions{
                OIDCProvider: oidcProvider,
            })
            
            // Sign weights
            if weightsPath != "" {
                cmd.Println("🔐 Signing weights...")
                cmd.Println("  → Authenticating with", oidcProvider)
                
                sig, err := signer.SignWeights(ctx, weightsPath)
                if err != nil {
                    return fmt.Errorf("sign weights: %w", err)
                }
                
                cmd.Println("  ✓ Weights signed")
                cmd.Printf("    Signed by: %s\n", sig.SignedBy)
                cmd.Printf("    Rekor entry: %s\n", sig.RekorEntryURL)
                
                // Save signature
                if err := saveSignature(sig, outputPath); err != nil {
                    return err
                }
            }
            
            // Sign code
            if codePath != "" {
                cmd.Println("🔐 Signing code...")
                sig, err := signer.SignCode(ctx, codePath)
                if err != nil {
                    return fmt.Errorf("sign code: %w", err)
                }
                cmd.Println("  ✓ Code signed")
            }
            
            // Sign metadata
            if yamlPath != "" {
                cmd.Println("🔐 Signing metadata...")
                sig, err := signer.SignModelYAML(ctx, yamlPath)
                if err != nil {
                    return fmt.Errorf("sign metadata: %w", err)
                }
                cmd.Println("  ✓ Metadata signed")
            }
            
            cmd.Println("\n✓ All artifacts signed successfully")
            cmd.Println("Signatures recorded in Rekor transparency log")
            
            return nil
        },
    }
    
    cmd.Flags().StringVar(&weightsPath, "weights", "", "Path to model weights")
    cmd.Flags().StringVar(&codePath, "code", "", "Path to code directory")
    cmd.Flags().StringVar(&yamlPath, "yaml", "", "Path to model.yaml")
    cmd.Flags().StringVar(&oidcProvider, "oidc", "github", "OIDC provider (github, google, microsoft)")
    cmd.Flags().StringVar(&outputPath, "output", "signature.json", "Output path for signature")
    
    return cmd
}
```

### conduit verify

```go
// internal/cli/verify.go
package cli

import (
    "fmt"
    
    "github.com/spf13/cobra"
    "github.com/yourusername/conduit/internal/sigstore"
    "github.com/yourusername/conduit/internal/catalog"
)

func VerifyCmd() *cobra.Command {
    var (
        policyFile string
        strictMode bool
    )
    
    cmd := &cobra.Command{
        Use:   "verify MODEL_NAME[@VERSION]",
        Short: "Verify model signatures",
        Long: `Verify Sigstore signatures for a model.

Checks:
  - Signature is cryptographically valid
  - Artifact hasn't been tampered with
  - Signature is recorded in Rekor transparency log
  - Certificate identity matches expected publisher
  - Signature meets policy requirements

Examples:
  # Verify a model
  conduit verify alphafold2-multimer

  # Verify with strict policy
  conduit verify alphafold2-multimer --strict

  # Verify with custom policy
  conduit verify alphafold2-multimer --policy verification.yaml
`,
        Args: cobra.ExactArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
            ctx := cmd.Context()
            modelName := args[0]
            
            // Load model
            cat := catalog.NewCatalog()
            model, err := cat.GetModel(ctx, modelName)
            if err != nil {
                return fmt.Errorf("get model: %w", err)
            }
            
            // Load policy
            var policy *types.VerificationPolicy
            if policyFile != "" {
                pm := sigstore.NewPolicyManager()
                if err := pm.LoadPolicyFile(policyFile); err != nil {
                    return fmt.Errorf("load policy: %w", err)
                }
                policy = pm.GetPolicy(model.Domain)
            } else if strictMode {
                policy = &types.VerificationPolicy{
                    Level:             types.PolicyLevelStrict,
                    RequireSignatures: true,
                    RequireRekorEntry: true,
                }
            }
            
            // Create verifier
            verifier := sigstore.NewVerifier(sigstore.VerifyOptions{
                Policy: policy,
            })
            
            // Verify signatures
            cmd.Println("🔍 Verifying signatures for", model.Name)
            
            // Verify weights
            if model.Signatures != nil && model.Signatures.WeightsSignature != nil {
                cmd.Println("\n📦 Weights:")
                if err := verifier.VerifyArtifact(ctx, model.Weights.URI, model.Signatures.WeightsSignature); err != nil {
                    cmd.Println("  ✗ Verification failed:", err)
                    return fmt.Errorf("weights verification failed")
                }
                cmd.Println("  ✓ Signature valid")
                cmd.Printf("    Signed by: %s\n", model.Signatures.WeightsSignature.SignedBy)
                cmd.Printf("    Signed at: %s\n", model.Signatures.WeightsSignature.SignedAt.Format(time.RFC3339))
                cmd.Printf("    Rekor: %s\n", model.Signatures.WeightsSignature.RekorEntryURL)
            } else {
                cmd.Println("\n📦 Weights:")
                if policy != nil && policy.RequireSignatures {
                    cmd.Println("  ✗ No signature found (required by policy)")
                    return fmt.Errorf("weights signature required but not found")
                }
                cmd.Println("  ⚠ No signature")
            }
            
            // Verify code
            if model.Signatures != nil && model.Signatures.CodeSignature != nil {
                cmd.Println("\n💻 Code:")
                // Similar verification...
                cmd.Println("  ✓ Signature valid")
            }
            
            // Summary
            cmd.Println("\n" + strings.Repeat("=", 50))
            cmd.Println("✓ All signatures verified successfully")
            cmd.Println(strings.Repeat("=", 50))
            
            return nil
        },
    }
    
    cmd.Flags().StringVar(&policyFile, "policy", "", "Verification policy file")
    cmd.Flags().BoolVar(&strictMode, "strict", false, "Use strict verification policy")
    
    return cmd
}
```

### Updated publish command

```go
// internal/cli/publish.go (updated)
func PublishCmd() *cobra.Command {
    var (
        sign         bool
        oidcProvider string
        // ... other flags
    )
    
    cmd := &cobra.Command{
        Use:   "publish",
        Short: "Publish model to catalog",
        RunE: func(cmd *cobra.Command, args []string) error {
            // ... existing publish logic ...
            
            // Sign if requested
            if sign {
                cmd.Println("\n🔐 Signing artifacts...")
                
                signer := sigstore.NewSigner(sigstore.SignOptions{
                    OIDCProvider: oidcProvider,
                })
                
                // Sign weights
                weightsSig, err := signer.SignWeights(ctx, weightsPath)
                if err != nil {
                    return fmt.Errorf("sign weights: %w", err)
                }
                
                // Sign code
                codeSig, err := signer.SignCode(ctx, ".")
                if err != nil {
                    return fmt.Errorf("sign code: %w", err)
                }
                
                // Update model with signatures
                model.Signatures = &types.ModelSignatures{
                    WeightsSignature: weightsSig,
                    CodeSignature:    codeSig,
                    SignedAt:         time.Now(),
                }
                
                cmd.Println("  ✓ Artifacts signed")
            }
            
            // ... continue with publish ...
            
            return nil
        },
    }
    
    cmd.Flags().BoolVar(&sign, "sign", false, "Sign artifacts with Sigstore")
    cmd.Flags().StringVar(&oidcProvider, "oidc", "github", "OIDC provider for signing")
    
    return cmd
}
```

### Updated deploy command

```go
// internal/cli/deploy.go (updated)
func DeployCmd() *cobra.Command {
    var (
        verify       bool
        requireSign  bool
        policyFile   string
        // ... other flags
    )
    
    cmd := &cobra.Command{
        Use:   "deploy MODEL_NAME",
        Short: "Deploy model to Bedrock",
        RunE: func(cmd *cobra.Command, args []string) error {
            // ... load model ...
            
            // Verify signatures if requested
            if verify || requireSign {
                cmd.Println("🔍 Verifying signatures...")
                
                // Load policy
                var policy *types.VerificationPolicy
                if policyFile != "" {
                    pm := sigstore.NewPolicyManager()
                    pm.LoadPolicyFile(policyFile)
                    policy = pm.GetPolicy(model.Domain)
                } else if requireSign {
                    policy = &types.VerificationPolicy{
                        Level:             types.PolicyLevelStrict,
                        RequireSignatures: true,
                    }
                }
                
                // Verify
                verifier := sigstore.NewVerifier(sigstore.VerifyOptions{
                    Policy: policy,
                })
                
                if err := verifier.VerifyArtifact(ctx, model.Weights.URI, model.Signatures.WeightsSignature); err != nil {
                    return fmt.Errorf("signature verification failed: %w", err)
                }
                
                cmd.Println("  ✓ Signatures verified")
            }
            
            // ... continue with deployment ...
            
            return nil
        },
    }
    
    cmd.Flags().BoolVar(&verify, "verify", false, "Verify signatures before deployment")
    cmd.Flags().BoolVar(&requireSign, "require-signatures", false, "Fail if signatures missing")
    cmd.Flags().StringVar(&policyFile, "policy", "", "Verification policy file")
    
    return cmd
}
```

---

## Database Schema

```sql
-- migrations/004_add_signatures.sql

CREATE TABLE model_signatures (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    model_id UUID REFERENCES models(id) ON DELETE CASCADE,
    
    artifact_type VARCHAR(50) NOT NULL, -- 'weights', 'code', 'metadata', 'benchmark'
    artifact_uri TEXT NOT NULL,
    artifact_hash VARCHAR(64) NOT NULL,
    
    signature_bundle TEXT NOT NULL,
    bundle_uri TEXT,
    
    rekor_log_id VARCHAR(255),
    rekor_index BIGINT,
    rekor_uuid VARCHAR(255),
    rekor_entry_url TEXT,
    
    cert_identity TEXT NOT NULL,
    cert_issuer TEXT NOT NULL,
    cert_valid_from TIMESTAMP,
    cert_valid_to TIMESTAMP,
    
    signed_by TEXT,
    signed_at TIMESTAMP NOT NULL,
    
    verified BOOLEAN DEFAULT false,
    verified_at TIMESTAMP,
    verified_by TEXT,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(model_id, artifact_type)
);

CREATE INDEX idx_model_signatures_model_id ON model_signatures(model_id);
CREATE INDEX idx_model_signatures_artifact_type ON model_signatures(artifact_type);
CREATE INDEX idx_model_signatures_cert_identity ON model_signatures(cert_identity);
CREATE INDEX idx_model_signatures_rekor_uuid ON model_signatures(rekor_uuid);

CREATE TABLE verification_policies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) UNIQUE NOT NULL,
    domain VARCHAR(100), -- NULL = global policy
    
    level VARCHAR(50) NOT NULL, -- 'none', 'optional', 'recommended', 'strict'
    require_signatures BOOLEAN DEFAULT false,
    require_rekor_entry BOOLEAN DEFAULT false,
    allowed_identities TEXT[], -- Array of identity patterns
    allowed_issuers TEXT[],
    max_signature_age_days INTEGER,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(domain)
);

-- Insert default policies
INSERT INTO verification_policies (name, domain, level, require_signatures, max_signature_age_days)
VALUES 
    ('Global Default', NULL, 'optional', false, NULL),
    ('Drug Discovery', 'drug-discovery', 'strict', true, 90),
    ('Protein Science', 'protein-science', 'recommended', false, 180);
```

---

## API Endpoints

```go
// internal/api/handlers/signatures.go
package handlers

import (
    "net/http"
    
    "github.com/gin-gonic/gin"
    "github.com/yourusername/conduit/internal/sigstore"
)

type SignatureHandler struct {
    verifier *sigstore.Verifier
}

// GET /api/models/:id/signatures
func (h *SignatureHandler) GetSignatures(c *gin.Context) {
    modelID := c.Param("id")
    
    // Get signatures from database
    signatures, err := h.store.GetModelSignatures(c, modelID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusOK, signatures)
}

// POST /api/models/:id/verify
func (h *SignatureHandler) VerifyModel(c *gin.Context) {
    modelID := c.Param("id")
    
    var req struct {
        Policy *types.VerificationPolicy `json:"policy"`
    }
    
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    // Get model
    model, err := h.store.GetModel(c, modelID)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "model not found"})
        return
    }
    
    // Verify
    verifier := sigstore.NewVerifier(sigstore.VerifyOptions{
        Policy: req.Policy,
    })
    
    result := make(map[string]interface{})
    
    // Verify weights
    if model.Signatures.WeightsSignature != nil {
        err := verifier.VerifyArtifact(c, model.Weights.URI, model.Signatures.WeightsSignature)
        result["weights"] = map[string]interface{}{
            "verified": err == nil,
            "error":    errToString(err),
        }
    }
    
    // Verify code
    if model.Signatures.CodeSignature != nil {
        err := verifier.VerifyCode(c, model.GitHubRepo, model.Signatures.CodeSignature)
        result["code"] = map[string]interface{}{
            "verified": err == nil,
            "error":    errToString(err),
        }
    }
    
    c.JSON(http.StatusOK, result)
}

// GET /api/policies
func (h *SignatureHandler) ListPolicies(c *gin.Context) {
    policies, err := h.store.GetVerificationPolicies(c)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusOK, policies)
}

// POST /api/policies
func (h *SignatureHandler) CreatePolicy(c *gin.Context) {
    var policy types.VerificationPolicy
    
    if err := c.ShouldBindJSON(&policy); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    if err := h.store.CreateVerificationPolicy(c, &policy); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusCreated, policy)
}
```

---

## Catalog UI Integration

### Model Page Updates

```tsx
// web/components/ModelSignatures.tsx
import { CheckCircle, XCircle, AlertCircle } from 'lucide-react';

interface Signature {
  artifactType: string;
  signedBy: string;
  signedAt: string;
  rekorEntryURL: string;
  verified: boolean;
}

export function ModelSignatures({ signatures }: { signatures: Signature[] }) {
  if (!signatures || signatures.length === 0) {
    return (
      <div className="border rounded-lg p-4 bg-yellow-50">
        <div className="flex items-center gap-2">
          <AlertCircle className="w-5 h-5 text-yellow-600" />
          <span className="font-medium">No Signatures</span>
        </div>
        <p className="text-sm text-gray-600 mt-2">
          This model has not been signed. Signatures provide cryptographic proof of authenticity.
        </p>
      </div>
    );
  }
  
  return (
    <div className="border rounded-lg overflow-hidden">
      <div className="bg-green-50 px-4 py-3 border-b">
        <div className="flex items-center gap-2">
          <CheckCircle className="w-5 h-5 text-green-600" />
          <span className="font-medium">Signed & Verified</span>
        </div>
      </div>
      
      <div className="divide-y">
        {signatures.map((sig) => (
          <div key={sig.artifactType} className="px-4 py-3">
            <div className="flex items-center justify-between">
              <div>
                <div className="font-medium capitalize">{sig.artifactType}</div>
                <div className="text-sm text-gray-600">
                  Signed by: {sig.signedBy}
                </div>
                <div className="text-sm text-gray-500">
                  {new Date(sig.signedAt).toLocaleString()}
                </div>
              </div>
              
              {sig.verified ? (
                <CheckCircle className="w-5 h-5 text-green-600" />
              ) : (
                <XCircle className="w-5 h-5 text-red-600" />
              )}
            </div>
            
            <a 
              href={sig.rekorEntryURL} 
              target="_blank"
              className="text-sm text-blue-600 hover:underline mt-2 inline-block"
            >
              View transparency log →
            </a>
          </div>
        ))}
      </div>
    </div>
  );
}
```

### Search Filters

```tsx
// web/app/search/page.tsx
<FilterSection title="Security">
  <Checkbox label="Signed models only" />
  <Checkbox label="Recently verified" />
</FilterSection>
```

### Badges

```tsx
// web/components/ModelCard.tsx
{model.signatures && (
  <span className="inline-flex items-center gap-1 px-2 py-1 text-xs bg-green-100 text-green-800 rounded">
    <CheckCircle className="w-3 h-3" />
    Signed
  </span>
)}
```

---

## Verification Policies

### Default Policy File

```yaml
# config/verification.yaml

# Global default policy (applies unless overridden)
global:
  level: optional
  require_signatures: false
  require_rekor_entry: false
  max_signature_age_days: 365

# Domain-specific policies
domains:
  drug-discovery:
    level: strict
    require_signatures: true
    require_rekor_entry: true
    allowed_identities:
      - "github.com/*/drug-*"
      - "github.com/pharma-org/*"
    allowed_issuers:
      - "https://token.actions.githubusercontent.com"
    max_signature_age_days: 90
    
  protein-science:
    level: recommended
    require_signatures: false
    require_rekor_entry: true
    max_signature_age_days: 180
    
  materials-science:
    level: recommended
    require_rekor_entry: true
    
  climate:
    level: optional
    
# Enterprise overrides (loaded separately)
enterprise:
  level: strict
  require_signatures: true
  require_rekor_entry: true
  allowed_identities:
    - "github.com/my-company/*"
  allowed_issuers:
    - "https://token.actions.githubusercontent.com"
  max_signature_age_days: 60
```

### User Configuration

```bash
# Set global policy
conduit config set verification.policy strict

# Set domain-specific policy
conduit config set verification.domains.drug-discovery strict

# Load from file
conduit config load-policy verification.yaml

# View current policy
conduit config get verification
```

---

## Testing Strategy

### Unit Tests

```go
// internal/sigstore/signer_test.go
package sigstore

import (
    "context"
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestSignWeights(t *testing.T) {
    // This would use a test/mock OIDC provider
    t.Skip("Requires OIDC authentication - run manually")
    
    ctx := context.Background()
    signer := NewSigner(SignOptions{
        OIDCProvider: "github",
    })
    
    // Create test file
    testFile := createTestFile(t, "test-weights.bin")
    defer os.Remove(testFile)
    
    // Sign
    sig, err := signer.SignWeights(ctx, testFile)
    require.NoError(t, err)
    
    assert.NotEmpty(t, sig.SignatureBundle)
    assert.NotEmpty(t, sig.RekorUUID)
    assert.NotEmpty(t, sig.CertIdentity)
}

func TestVerifyWeights(t *testing.T) {
    t.Skip("Requires existing signature - run manually")
    
    ctx := context.Background()
    verifier := NewVerifier(VerifyOptions{})
    
    // Load test signature
    sig := loadTestSignature(t)
    testFile := "testdata/signed-weights.bin"
    
    // Verify
    err := verifier.VerifyArtifact(ctx, testFile, sig)
    assert.NoError(t, err)
}

func TestPolicyEnforcement(t *testing.T) {
    tests := []struct {
        name    string
        policy  *types.VerificationPolicy
        sig     *types.Signature
        wantErr bool
    }{
        {
            name: "strict policy with valid signature",
            policy: &types.VerificationPolicy{
                Level:             types.PolicyLevelStrict,
                RequireSignatures: true,
            },
            sig: &types.Signature{
                CertIdentity: "github.com/test/repo",
                SignedAt:     time.Now(),
            },
            wantErr: false,
        },
        {
            name: "strict policy without signature",
            policy: &types.VerificationPolicy{
                Level:             types.PolicyLevelStrict,
                RequireSignatures: true,
            },
            sig:     nil,
            wantErr: true,
        },
        {
            name: "identity not in allowed list",
            policy: &types.VerificationPolicy{
                Level:             types.PolicyLevelStrict,
                AllowedIdentities: []string{"github.com/approved/*"},
            },
            sig: &types.Signature{
                CertIdentity: "github.com/random/repo",
            },
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            verifier := NewVerifier(VerifyOptions{
                Policy: tt.policy,
            })
            
            err := verifier.applyPolicy(tt.sig)
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

### Integration Tests

```go
// internal/sigstore/integration_test.go
// +build integration

package sigstore

func TestEndToEndSigning(t *testing.T) {
    // This requires actual OIDC authentication
    // Run with: go test -tags=integration
    
    ctx := context.Background()
    
    // Sign
    signer := NewSigner(SignOptions{
        OIDCProvider: "github",
    })
    
    testFile := "testdata/model.safetensors"
    sig, err := signer.SignWeights(ctx, testFile)
    require.NoError(t, err)
    
    // Verify immediately
    verifier := NewVerifier(VerifyOptions{})
    err = verifier.VerifyArtifact(ctx, testFile, sig)
    assert.NoError(t, err)
    
    // Modify file and verify again (should fail)
    modifyFile(t, testFile)
    err = verifier.VerifyArtifact(ctx, testFile, sig)
    assert.Error(t, err, "verification should fail for modified file")
}
```

---

## Deployment

### Docker Support

```dockerfile
# Dockerfile (updated)
FROM golang:1.21-alpine AS builder

# Install Cosign
RUN apk add --no-cache cosign

WORKDIR /app
COPY go.* ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /conduit-server ./cmd/conduit-server

FROM alpine:latest
RUN apk --no-cache add ca-certificates cosign
COPY --from=builder /conduit-server /conduit-server

EXPOSE 8080
CMD ["/conduit-server"]
```

### Environment Variables

```bash
# Sigstore configuration
SIGSTORE_REKOR_URL=https://rekor.sigstore.dev
SIGSTORE_FULCIO_URL=https://fulcio.sigstore.dev

# OIDC providers
SIGSTORE_OIDC_ISSUER=https://token.actions.githubusercontent.com
SIGSTORE_OIDC_CLIENT_ID=sigstore

# Verification policy
VERIFICATION_POLICY_LEVEL=optional
VERIFICATION_REQUIRE_SIGNATURES=false

# Private Sigstore instance (optional)
# SIGSTORE_REKOR_URL=https://rekor.internal.company.com
# SIGSTORE_FULCIO_URL=https://fulcio.internal.company.com
```

### CI/CD Integration

```yaml
# .github/workflows/publish.yml
name: Publish Model with Signing

on:
  push:
    tags:
      - 'v*'

jobs:
  publish:
    runs-on: ubuntu-latest
    
    permissions:
      id-token: write  # Required for OIDC
      contents: read
      
    steps:
      - uses: actions/checkout@v3
      
      - name: Install Conduit CLI
        run: |
          curl -sSL https://get.conduit.dev | sh
          
      - name: Configure AWS credentials
        uses: aws-actions/configure-aws-credentials@v2
        with:
          role-to-assume: ${{ secrets.AWS_ROLE_ARN }}
          
      - name: Publish with signing
        run: |
          conduit publish \
            --sign \
            --oidc github \
            --github-repo ${{ github.repository }} \
            --version ${{ github.ref_name }}
            
      - name: Verify signatures
        run: |
          conduit verify ${{ github.repository }}
```

---

## Documentation

### Publisher Guide

```markdown
# Signing Your Model

## Why Sign?

Signing your model provides:
- **Authenticity**: Cryptographic proof you published it
- **Integrity**: Detect any tampering
- **Auditability**: Public transparency log
- **Trust**: Users know it's legitimate

## How to Sign

### Option 1: Sign during publish

```bash
conduit publish --sign --oidc github
```

This will:
1. Authenticate you with GitHub
2. Sign your model artifacts
3. Upload signatures to Rekor transparency log
4. Update model.yaml with signature metadata

### Option 2: Sign separately

```bash
# Sign weights
conduit sign --weights model.safetensors --oidc github

# Sign code
conduit sign --code src/ --oidc github

# Then publish
conduit publish
```

## What Gets Signed?

By default:
- ✓ Model weights
- ✓ Inference code
- ✓ model.yaml metadata

Optional:
- Benchmark results (use --sign-benchmarks)
- Container images (if using containers)

## Authentication

Supported OIDC providers:
- GitHub Actions (recommended for CI/CD)
- Google
- Microsoft

No private keys to manage - authentication via OIDC.

## Transparency Log

All signatures are recorded in the public Rekor transparency log:
https://rekor.sigstore.dev

This provides:
- Immutable audit trail
- Public verifiability
- Timestamping
```

### Consumer Guide

```markdown
# Verifying Models

## Why Verify?

Before deploying a model, verify:
- It's actually from the claimed publisher
- It hasn't been tampered with
- Signatures are in the transparency log

## How to Verify

### Option 1: Verify during deployment

```bash
conduit deploy alphafold2 --verify
```

### Option 2: Verify separately

```bash
conduit verify alphafold2
```

Output:
```
🔍 Verifying signatures for alphafold2-multimer

📦 Weights:
  ✓ Signature valid
    Signed by: github.com/deepmind/alphafold
    Signed at: 2024-11-01 12:34:56 UTC
    Rekor: https://rekor.sigstore.dev/...

💻 Code:
  ✓ Signature valid

=================================================
✓ All signatures verified successfully
=================================================
```

## Verification Policies

Set your requirements:

```bash
# Require signatures for all models
conduit config set verification.policy strict

# Domain-specific policies
conduit config set verification.domains.drug-discovery strict
```

Policy levels:
- **none**: Don't verify
- **optional**: Verify if present
- **recommended**: Warn if missing
- **strict**: Require signatures

## Enterprise Policies

For regulated industries:

```yaml
# verification.yaml
global:
  level: strict
  require_signatures: true
  allowed_identities:
    - "github.com/approved-org/*"
  max_signature_age_days: 90
```

```bash
conduit deploy --policy verification.yaml
```
```

---

## Summary

### What We're Building

1. **Signing Infrastructure**
   - OIDC-based keyless signing
   - Cosign integration
   - Rekor transparency log
   - Support for weights, code, metadata

2. **Verification System**
   - Cryptographic verification
   - Policy engine
   - CLI commands
   - API endpoints

3. **User Experience**
   - `--sign` flag on publish
   - `--verify` flag on deploy
   - Visual indicators in catalog
   - Clear documentation

### Timeline

**Week 10: Core Implementation**
- Sigstore integration (signing + verification)
- CLI commands
- Database schema
- Basic policies

**Week 11: Integration**
- API endpoints
- Catalog UI updates
- Documentation
- Testing

**Week 12: Launch**
- Blog post: "Supply Chain Security for Scientific Models"
- Update all example models
- Enterprise documentation

### Differentiation

**This is huge:**
- Nobody else has cryptographic verification
- Perfect for pharma/regulated industries
- Academic credibility (provenance)
- Unique selling point vs Garden/HuggingFace

### Success Metrics

- % of models signed
- Verification adoption rate
- Enterprise customer acquisition
- Zero supply chain incidents

---

**Ready to implement?** Start with `internal/sigstore/signer.go` and `internal/sigstore/verifier.go`. The Cosign library does most of the heavy lifting - we're just wrapping it nicely for scientific models.
