# Conduit: AWS-Native Scientific Model Platform

## Executive Summary

**Conduit** (working name) is an open-source, AWS-native platform that makes publishing and deploying scientific ML models as simple as installing software packages. Think "Homebrew for scientific models" with cloud deployment built-in.

### Core Problem

- **HPC Queue Hell**: Researchers wait days for GPU allocations while cloud resources sit idle
- **Model Reproducibility**: Hard to share/reproduce scientific ML models
- **Accessibility**: Complex infrastructure prevents smaller institutions from using cutting-edge models
- **Cost Opacity**: Researchers don't know actual compute costs vs benefits

### Solution

Standardized model packaging + Automated cloud deployment + Cost transparency + Scientific rigor (DOIs, benchmarks, signatures)

---

## Key Design Decisions

### 1. No Containers Required (Huge Simplification)

Unlike Garden, no Docker knowledge needed:

```yaml
# model.yaml
runtime:
  framework: "pytorch"
  python_version: "3.11"
  dependencies: "requirements.txt"

inference:
  entrypoint: "inference.py"
  handler: "predict"
```

Bedrock handles the serving infrastructure.

### 2. Flexible Weight Storage

```yaml
weights_uri: "s3://bucket/weights/"           # S3
# OR: "hf://org/model"                        # HuggingFace
# OR: "https://zenodo.org/record/123/..."    # Zenodo
# OR: "s3://aws-open-data-scientific-models" # AWS Open Data
```

Publishers choose what works. GitHub repo is source of truth.

### 3. SageMaker Studio Lab Priority

Not Colab first - **Studio Lab** first:

- Free, no AWS account initially needed
- 15GB persistent storage (Colab doesn't have this)
- Direct Bedrock access
- Natural upgrade path to SageMaker Studio

**Strategic Funnel**: Studio Lab (free) → SageMaker Studio (paid) → Full AWS ecosystem

### 4. Agentic Workflows (Killer Feature)

Foundation model (Claude) + specialist scientific models + MCP servers + AWS Strands:

```yaml
coordinator:
  model: "claude-sonnet-4-5"
  mcp_servers: ["uniprot", "pdb", "pubmed"]
  aws_strands: ["internal_db"]

available_tools:
  - predict_structure (AlphaFold)
  - design_binder (RFdiffusion)
  - predict_affinity (EquiBind)
```

**Example**: "Design an antibody against SARS-CoV-2 RBD with KD < 1nM" → Agent autonomously plans workflow, calls models, iterates, produces candidates with full scientific rationale.

### 5. Sigstore Integration (Security Differentiator)

Cryptographic signing via Sigstore (keyless, OIDC-based):

- Sign model weights, code, benchmarks
- Immutable transparency log (Rekor)
- Perfect for pharma/regulated industries
- **Nobody else is doing this in scientific ML**

---

## Architecture

### Technology Stack

- **Backend**: Go (fast compilation, single binary, excellent AWS SDK)
- **Database**: PostgreSQL + OpenSearch (catalog)
- **Frontend**: Next.js
- **CLI**: Cobra framework
- **Cloud**: AWS Bedrock (primary), SageMaker (secondary)

### Key Components

1. **CLI Tool** (`conduit`)
   - `init` - Create model from template
   - `publish` - Publish to catalog (+ optional signing)
   - `deploy` - Deploy to Bedrock
   - `benchmark` - Run standardized benchmarks
   - `verify` - Verify Sigstore signatures

2. **Model Registry** (GitHub-based, decentralized)
   - Each model = GitHub repo with `model.yaml`
   - Catalog crawls repos, indexes metadata
   - DOI minting via Zenodo

3. **Auto-Generated UIs**
   - Streamlit apps (full-featured)
   - Gradio interfaces (simple demos)
   - Jupyter notebooks (Studio Lab optimized)

4. **Deployment Engine**
   - Bedrock Custom Model Import
   - Automatic instance selection
   - Cost/performance benchmarking
   - Spot instance support

---

## Launch Strategy

### Phase 1: Protein Science (Exemplar Domain)

20+ models: AlphaFold2, ESMFold, ProteinMPNN, DiffDock, IgFold, etc.

**Why proteins first:**

- ✅ Massive commercial value (pharma, biotech)
- ✅ Active academic community
- ✅ Clear HPC pain point (AlphaFold queue = days)
- ✅ Computational intensity = $$$ AWS revenue
- ✅ Regulatory compliance story (Sigstore)

### Phase 2-4: Expand to Other Domains

- Materials Science (batteries, semiconductors, catalysts)
- Drug Discovery (docking, ADMET, retrosynthesis)
- Climate/Earth Science (weather, climate projections)

---

## Competitive Positioning

|                     | **Conduit**        | **Garden**     | **HuggingFace** | **Traditional HPC** |
| ------------------- | ------------------ | -------------- | --------------- | ------------------- |
| **Speed**           | Instant            | Instant        | Instant         | Days (queues)       |
| **Cost**            | Pay-per-use        | Pay-per-use    | Pay-per-use     | Wasted allocations  |
| **Focus**           | Scientific models  | Physics/mat'ls | LLMs mainly     | Everything          |
| **Cloud**           | AWS-native         | Modal/Globus   | Agnostic        | On-prem             |
| **Signing**         | Sigstore ✓         | No             | No              | No                  |
| **Agentic**         | Claude + tools ✓   | No             | Limited         | No                  |
| **DOIs**            | Zenodo ✓           | Limited        | No              | No                  |
| **No-code UI**      | Auto-gen ✓         | No             | Spaces          | No                  |

**Key Differentiator**: Only platform combining scientific rigor (DOIs, benchmarks), cloud deployment (Bedrock), AI reasoning (agentic workflows), and supply chain security (Sigstore).

---

## ISC Conference Opportunity

**Paper Title**: "Democratizing Scientific AI: A Cloud-Native Framework for Model Publishing and Execution"

**Angle**: HPC researcher at AWS shows how cloud eliminates queue waits while maintaining scientific rigor.

**BoF Session**: "Beyond the Queue: Cloud-Native Scientific Computing"

- Panel with HPC directors, cloud researchers, funding agencies, industry users
- Live demo of publishing + deploying in minutes vs. days on HPC
- Timeline: Submission Feb/Mar for June conference

---

## What Makes This Compelling

1. **Solves Real Pain**: Built Top 500 systems, know the queue problem is intolerable

2. **Perfect Timing**: Garden moving at academic pace, we can ship at commercial speed

3. **AWS Advantage**:
   - Deep Bedrock knowledge
   - Credibility in cloud space
   - Ability to evangelize internally
   - But project lives outside AWS (credible to academics)

4. **Unique Combo**: Nobody else has:
   - Agentic workflows (Claude reasoning + specialist models)
   - Supply chain security (Sigstore)
   - AWS-native integration (Studio Lab → Bedrock)
   - Academic legitimacy (DOIs, FAIR principles)
   - Cost transparency

5. **Network Effects**: More publishers → more models → more users → more publishers

6. **Business Model**: Open source tools (free), AWS pays via compute consumption (Bedrock inference, SageMaker, storage)

---

## Implementation Roadmap

### MVP (Weeks 1-4)

- Core CLI commands (init, validate, publish)
- Model.yaml parser and validator
- Basic Bedrock deployment
- PostgreSQL catalog backend
- Simple web catalog viewer

### Phase 1 (Weeks 5-8)

- Complete protein science model suite (20+ models)
- Studio Lab notebook generation
- Sigstore signing integration
- Streamlit/Gradio UI generation
- Benchmark framework

### Phase 2 (Weeks 9-12)

- Agentic workflow engine
- MCP server integration
- AWS Strands support
- Advanced deployment (spot instances, multi-region)
- Enterprise features

### Phase 3 (Weeks 13-16)

- Additional domains (materials, drug discovery)
- Model Registry federated search
- Advanced analytics
- Community building
- ISC paper preparation

---

## Success Metrics

### Technical

- \>100 models published across 3+ domains
- 95% reduction in time-to-first-result vs HPC queues
- 40% cost savings using spot instances vs on-demand
- Go Report Card: A+
- API response time: <100ms (p95)

### Adoption

- 1,000+ CLI downloads (month 3)
- 50+ active publishers
- 500+ model deployments
- 10+ institutional adopters

### Business

- Bedrock inference hours attributed to Conduit
- SageMaker Studio upgrades from Studio Lab
- AWS Open Data Registry models published

### Academic

- ISC paper accepted
- 3+ research groups using in production
- Cited in 10+ papers
- Working group / SIG formation

---

## Risk Mitigation

### Technical Risks

- **Bedrock Custom Model limitations**: Maintain SageMaker fallback
- **Cost unpredictability**: Spot instances + cost guardrails
- **Model compatibility**: Comprehensive testing framework

### Adoption Risks

- **Garden captures market**: Move faster, differentiate (Sigstore, agentic)
- **AWS perception as vendor lock-in**: Open source, multi-cloud capable
- **Critical mass of models**: Seed with 100+ models at launch

### Business Risks

- **AWS doesn't support**: Internal evangelism, show usage metrics
- **Regulatory concerns**: Sigstore addresses, compliance documentation
- **Sustainability**: Foundation model (Linux Foundation style)

---

## Next Steps

1. ✅ Initialize professional Go project structure
2. ✅ Set up GitHub Projects/Milestones/Issues
3. Implement core model.yaml parser
4. Build basic CLI (init, validate)
5. Set up PostgreSQL schema
6. Implement Bedrock deployment
7. Create first example model (AlphaFold2)
8. Alpha testing with 3-5 research groups

---

## References

- Design specifications: See DESIGN_CONVO.md
- Go implementation plan: See go-implementation-plan.md
- Protein science suite: See protein-science-suite-spec.md
- SageMaker integration: See sagemaker-integration-spec.md
- Sigstore integration: See sigstore-integration-spec.md

**Status**: Ready to build. Let's ship this.
