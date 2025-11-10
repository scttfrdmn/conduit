# Protein Science Suite - Complete Exemplar Specification

**Version:** 1.0.0  
**Domain:** Structural Biology & Protein Engineering  
**Status:** Reference Implementation  
**Last Updated:** 2025-11-04

---

## Table of Contents

1. [Overview](#overview)
2. [Core Model Suite](#core-model-suite)
3. [Benchmark Framework](#benchmark-framework)
4. [Simple Workflows](#simple-workflows)
5. [Agentic Workflows](#agentic-workflows)
6. [UI Specifications](#ui-specifications)
7. [Deployment Patterns](#deployment-patterns)
8. [Success Metrics](#success-metrics)

---

## Overview

### Domain Description

The Protein Science suite provides standardized access to state-of-the-art machine learning models for:
- **Structure Prediction** - Predict 3D protein structures from sequence
- **Protein Design** - Design novel protein sequences for target functions
- **Binding Prediction** - Predict protein-protein and protein-ligand interactions
- **Function Prediction** - Predict protein properties and functions
- **Antibody Engineering** - Specialized tools for therapeutic antibody development

### Target Users

**Academic Researchers:**
- Structural biologists
- Protein engineers
- Computational biologists
- Graduate students

**Industry:**
- Pharmaceutical companies (drug discovery)
- Biotechnology companies (enzyme engineering, antibody development)
- Agricultural biotech (crop improvement)
- Industrial biotechnology (protein catalysts)

### Value Proposition

**vs. Traditional HPC:**
- ⚡ Instant compute (no queue waits of 2-3 days)
- 💰 Pay-per-use (no wasted allocations)
- 🔄 Reproducible (containerized environments)
- 📦 Easy deployment (no environment setup)

**vs. Running Locally:**
- 🚀 GPU access (no local GPU required)
- 📊 Scalability (batch thousands of predictions)
- 🔧 Managed infrastructure (no maintenance)
- 🤝 Collaboration (shared endpoints)

---

## Core Model Suite

### 1. Structure Prediction Models

#### AlphaFold2-Multimer

```yaml
name: "alphafold2-multimer"
version: "2.3.2"
domain: "protein-structure-prediction"
type: "single-model"

description: |
  Highly accurate protein structure prediction using deep learning.
  Supports monomers, multimers, and protein complexes.

citation:
  paper: "https://doi.org/10.1038/s41586-021-03819-2"
  authors: ["Jumper et al."]
  year: 2021
  bibtex: |
    @article{jumper2021alphafold,
      title={Highly accurate protein structure prediction with AlphaFold},
      author={Jumper, John and Evans, Richard and others},
      journal={Nature},
      volume={596},
      pages={583--589},
      year={2021}
    }

doi: "10.5281/zenodo.8234567"

authors:
  - name: "John Jumper"
    orcid: "0000-0001-9407-8948"
    institution: "DeepMind"
    
github_repo: "https://github.com/deepmind/alphafold"

model_artifacts:
  weights_uri: "s3://aws-open-data-scientific-models/alphafold2-multimer/v2.3.2/"
  weights_doi: "10.5281/zenodo.8234567"
  size_gb: 3.5
  format: "pickle"
  checksum_sha256: "abc123def456..."
  open_data_registry: true

runtime:
  framework: "jax"
  python_version: "3.10"
  dependencies: "requirements.txt"
  
inference:
  entrypoint: "inference.py"
  handler: "predict"
  input_schema: "schemas/input.json"
  output_schema: "schemas/output.json"

hardware:
  gpu_required: true
  min_gpu_memory_gb: 16
  recommended_instance: "ml.g5.2xlarge"
  supports_batch: true
  optimal_batch_size: 1

benchmarks:
  standard:
    - dataset: "casp15-subset"
      metric: "tm_score"
      result: 0.92
      instance: "ml.g5.2xlarge"
      cost_per_1k: "$2.45"
      latency_p50_ms: 850
      latency_p99_ms: 1200
      throughput_per_hour: 4235
      date: "2024-11-01"
      
    - dataset: "casp15-subset"
      metric: "tm_score"
      result: 0.92
      instance: "ml.g5.12xlarge"
      cost_per_1k: "$8.20"
      latency_p50_ms: 120
      latency_p99_ms: 180
      throughput_per_hour: 30000
      date: "2024-11-01"

datasets:
  training: "https://doi.org/10.1038/s41586-021-03819-2"
  validation: "CASP14"
  benchmark: "CASP15"

license: "Apache-2.0"

tags: 
  - "proteins"
  - "structure-prediction"
  - "multimer"
  - "deep-learning"

quality:
  technical_validation: "passed"
  peer_review: "certified"
  certification_date: "2024-10-15"
  reviewers:
    - name: "Dr. Jane Smith"
      orcid: "0000-0002-1234-5678"
      institution: "MIT"

ui:
  streamlit:
    enabled: true
    template: "protein-structure"
    
  inputs:
    - name: "sequence"
      type: "text_area"
      label: "Protein Sequence (FASTA format)"
      placeholder: ">protein1\nMKLLVGDDS...\n>protein2\nMAFGKL..."
      validation: "fasta_format"
      required: true
      
    - name: "num_recycles"
      type: "slider"
      label: "Number of Recycles"
      min: 1
      max: 20
      default: 3
      help: "More recycles = higher accuracy but slower. 3 is usually sufficient."
      
    - name: "model_preset"
      type: "select"
      label: "Model Preset"
      options: 
        - value: "monomer"
          label: "Monomer"
          help: "Single chain prediction"
        - value: "multimer"
          label: "Multimer"
          help: "Multiple chain complex prediction"
        - value: "monomer_ptm"
          label: "Monomer with pTM"
          help: "Single chain with predicted template modeling score"
      default: "monomer"
      
    - name: "use_templates"
      type: "checkbox"
      label: "Use structural templates"
      default: true
      help: "Search for and use template structures from PDB"
      
  outputs:
    - name: "structure"
      type: "molecule_viewer"
      format: "pdb"
      download: true
      viewer: "3dmol"
      
    - name: "confidence_plot"
      type: "line_chart"
      title: "Per-Residue Confidence (pLDDT)"
      x_label: "Residue"
      y_label: "Confidence"
      
    - name: "pae_heatmap"
      type: "heatmap"
      title: "Predicted Aligned Error"
      colormap: "viridis"
      
    - name: "summary_metrics"
      type: "metrics"
      fields:
        - name: "mean_plddt"
          label: "Mean pLDDT"
          format: ".2f"
        - name: "ptm_score"
          label: "pTM Score"
          format: ".3f"
          
  examples:
    - name: "T1083 (CASP15 target)"
      sequence: ">T1083\nMKLLVGDDSAFAMILKNYGEKV..."
      num_recycles: 3
      model_preset: "monomer"
      
    - name: "Antibody-Antigen Complex"
      sequence: ">Heavy_Chain\nEVQLVES...\n>Light_Chain\nDIQMTQ...\n>Antigen\nMAPRT..."
      model_preset: "multimer"
      num_recycles: 5
```

#### ESMFold

```yaml
name: "esmfold"
version: "1.0.0"
domain: "protein-structure-prediction"
type: "single-model"

description: |
  Fast protein structure prediction using evolutionary scale modeling.
  60x faster than AlphaFold2 with competitive accuracy for many targets.

citation:
  paper: "https://doi.org/10.1126/science.ade2574"
  authors: ["Lin et al."]
  year: 2023

doi: "10.5281/zenodo.8234568"

model_artifacts:
  weights_uri: "hf://facebook/esmfold_v1"
  huggingface_id: "facebook/esmfold_v1"
  size_gb: 2.8
  format: "pytorch"

runtime:
  framework: "pytorch"
  python_version: "3.10"
  
inference:
  entrypoint: "inference.py"
  handler: "predict"

hardware:
  gpu_required: true
  min_gpu_memory_gb: 8
  recommended_instance: "ml.g5.xlarge"
  supports_batch: true
  optimal_batch_size: 4

benchmarks:
  standard:
    - dataset: "casp15-subset"
      metric: "tm_score"
      result: 0.87
      instance: "ml.g5.xlarge"
      cost_per_1k: "$0.85"
      latency_p50_ms: 45
      latency_p99_ms: 80
      throughput_per_hour: 80000
      date: "2024-11-01"

license: "MIT"

tags: ["proteins", "structure-prediction", "fast", "language-model"]

ui:
  streamlit:
    enabled: true
    template: "protein-structure"
```

---

### 2. Protein Design Models

#### ProteinMPNN

```yaml
name: "proteinmpnn"
version: "1.0.1"
domain: "protein-design"
type: "single-model"

description: |
  Design protein sequences that fold into specified structures using
  message passing neural networks. Ideal for de novo protein design
  and stabilization.

citation:
  paper: "https://doi.org/10.1126/science.add2187"
  authors: ["Dauparas et al."]
  year: 2022

doi: "10.5281/zenodo.8234569"

github_repo: "https://github.com/dauparas/ProteinMPNN"

model_artifacts:
  weights_uri: "s3://aws-open-data-scientific-models/proteinmpnn/v1.0.1/"
  size_gb: 0.5
  format: "pytorch"
  open_data_registry: true

runtime:
  framework: "pytorch"
  python_version: "3.10"

inference:
  entrypoint: "inference.py"
  handler: "predict"

hardware:
  gpu_required: true
  min_gpu_memory_gb: 4
  recommended_instance: "ml.g5.xlarge"
  supports_batch: true
  optimal_batch_size: 8

benchmarks:
  standard:
    - dataset: "protein-design-benchmark"
      metric: "sequence_recovery"
      result: 0.52
      instance: "ml.g5.xlarge"
      cost_per_1k: "$0.45"
      latency_p50_ms: 25
      throughput_per_hour: 144000
      date: "2024-11-01"

license: "MIT"

tags: ["proteins", "design", "inverse-folding"]

ui:
  streamlit:
    enabled: true
    template: "protein-design"
    
  inputs:
    - name: "structure"
      type: "file_upload"
      label: "Backbone Structure (PDB)"
      accept: [".pdb"]
      required: true
      
    - name: "num_sequences"
      type: "number"
      label: "Number of Sequences to Design"
      min: 1
      max: 100
      default: 10
      
    - name: "temperature"
      type: "slider"
      label: "Sampling Temperature"
      min: 0.1
      max: 2.0
      default: 0.1
      step: 0.1
      help: "Higher = more diverse sequences"
      
    - name: "fixed_positions"
      type: "text_input"
      label: "Fixed Positions (comma-separated)"
      placeholder: "1,5,10-20,45"
      required: false
      help: "Specify residue positions to keep unchanged"
      
  outputs:
    - name: "designed_sequences"
      type: "table"
      columns: ["rank", "sequence", "score", "recovery"]
      download: true
      
    - name: "sequence_logos"
      type: "image"
      title: "Sequence Logo"
      
    - name: "score_distribution"
      type: "histogram"
```

---

### 3. Binding & Docking Models

#### DiffDock

```yaml
name: "diffdock"
version: "1.1.0"
domain: "molecular-docking"
type: "single-model"

description: |
  Diffusion-based molecular docking for predicting protein-ligand
  binding poses and affinities. State-of-the-art accuracy without
  requiring known binding sites.

citation:
  paper: "https://doi.org/10.48550/arXiv.2210.01776"
  authors: ["Corso et al."]
  year: 2023

doi: "10.5281/zenodo.8234570"

github_repo: "https://github.com/gcorso/DiffDock"

model_artifacts:
  weights_uri: "s3://aws-open-data-scientific-models/diffdock/v1.1.0/"
  size_gb: 1.2
  format: "pytorch"

runtime:
  framework: "pytorch"
  python_version: "3.10"
  dependencies: "requirements.txt"

hardware:
  gpu_required: true
  min_gpu_memory_gb: 8
  recommended_instance: "ml.g5.xlarge"

benchmarks:
  standard:
    - dataset: "pdbbind-core-set"
      metric: "rmsd_lt_2a"
      result: 0.38
      instance: "ml.g5.xlarge"
      cost_per_1k: "$1.20"
      latency_p50_ms: 150
      date: "2024-11-01"

license: "MIT"

tags: ["docking", "binding", "drug-discovery", "diffusion"]

ui:
  streamlit:
    enabled: true
    template: "molecular-docking"
    
  inputs:
    - name: "protein"
      type: "file_upload"
      label: "Protein Structure (PDB)"
      accept: [".pdb"]
      required: true
      
    - name: "ligand"
      type: "text_area"
      label: "Ligand SMILES"
      placeholder: "CC(C)Cc1ccc(cc1)C(C)C(O)=O"
      required: true
      
    - name: "num_poses"
      type: "number"
      label: "Number of Poses"
      min: 1
      max: 50
      default: 10
      
  outputs:
    - name: "poses"
      type: "molecule_viewer"
      format: "sdf"
      multi_pose: true
      
    - name: "confidence_scores"
      type: "table"
      columns: ["pose", "confidence", "rmsd"]
```

---

### 4. Antibody-Specific Models

#### IgFold

```yaml
name: "igfold"
version: "1.0.0"
domain: "antibody-structure"
type: "single-model"

description: |
  Fast and accurate antibody structure prediction from sequence.
  Optimized for VH-VL pairing and CDR loop conformations.

citation:
  paper: "https://doi.org/10.1038/s41467-023-38063-x"
  authors: ["Ruffolo et al."]
  year: 2023

doi: "10.5281/zenodo.8234571"

model_artifacts:
  weights_uri: "hf://Exscientia/IgFold"
  size_gb: 0.8
  format: "pytorch"

runtime:
  framework: "pytorch"
  python_version: "3.10"

hardware:
  gpu_required: true
  min_gpu_memory_gb: 6
  recommended_instance: "ml.g5.xlarge"

benchmarks:
  standard:
    - dataset: "sabdab-test-set"
      metric: "cdr_h3_rmsd"
      result: 2.1
      instance: "ml.g5.xlarge"
      cost_per_1k: "$0.60"
      latency_p50_ms: 35
      date: "2024-11-01"

license: "BSD-3-Clause"

tags: ["antibody", "structure-prediction", "therapeutic"]
```

---

## Benchmark Framework

### Standard Benchmark Datasets

#### CASP15 Subset

```yaml
benchmark_dataset:
  id: "casp15-subset"
  name: "CASP15 Structure Prediction Subset"
  version: "1.0.0"
  
  description: |
    100 representative targets from CASP15 (Critical Assessment of 
    protein Structure Prediction) spanning diverse protein families,
    sizes, and structural classes.
    
  storage:
    uri: "s3://aws-open-data-scientific-models/benchmarks/casp15-subset/"
    size_gb: 2.5
    format: "hdf5"
    
  contents:
    num_targets: 100
    target_types:
      - monomer: 60
      - homodimer: 20
      - heterodimer: 15
      - higher_order: 5
    size_range: "50-800 residues"
    
  ground_truth:
    structures: "experimental_pdbs/"
    format: "pdb"
    resolution_cutoff: "3.0A"
    
  metrics:
    primary:
      - name: "tm_score"
        description: "Template Modeling Score"
        range: [0, 1]
        threshold_good: 0.5
        threshold_excellent: 0.8
        
      - name: "gdt_ts"
        description: "Global Distance Test Total Score"
        range: [0, 100]
        
      - name: "rmsd"
        description: "Root Mean Square Deviation (Cα)"
        unit: "angstrom"
        lower_is_better: true
        
  usage:
    citation: "https://doi.org/10.1002/prot.26556"
    license: "CC-BY-4.0"
```

#### Protein Design Benchmark

```yaml
benchmark_dataset:
  id: "protein-design-benchmark"
  name: "Protein Inverse Folding Benchmark"
  version: "1.0.0"
  
  description: |
    Curated set of protein structures for testing sequence design
    algorithms. Includes native sequences for computing recovery rates.
    
  storage:
    uri: "s3://aws-open-data-scientific-models/benchmarks/protein-design/"
    size_gb: 1.8
    
  contents:
    num_structures: 500
    structure_types:
      - monomers: 350
      - dimers: 100
      - designed_proteins: 50
      
  ground_truth:
    native_sequences: "sequences.fasta"
    native_structures: "structures/"
    
  metrics:
    primary:
      - name: "sequence_recovery"
        description: "Fraction of recovered native residues"
        range: [0, 1]
        
      - name: "perplexity"
        description: "Sequence likelihood"
        lower_is_better: true
```

### Running Benchmarks

```bash
# Publisher workflow
cd my-model/
conduit benchmark run \
  --standard-suite protein-structure \
  --instances ml.g5.2xlarge,ml.g5.12xlarge \
  --output benchmarks/

# Generates:
# - benchmarks/casp15-results.json
# - benchmarks/plots/
# - Updates model.yaml with results
```

### Benchmark Visualization

Generated plots include:
- TM-score distribution histogram
- Cost vs accuracy scatter
- Latency vs throughput curves
- Comparison to baseline models

---

## Simple Workflows

### Workflow 1: Structure Prediction Pipeline

```yaml
name: "structure-prediction-pipeline"
version: "1.0.0"
type: "pipeline"
domain: "protein-structure"

description: |
  Complete pipeline for predicting protein structure with quality control
  and validation steps.

doi: "10.5281/zenodo.9234567"

authors:
  - name: "Research Lab PI"
    orcid: "0000-0001-2345-6789"
    institution: "Stanford University"

composition:
  type: "sequential"
  
  steps:
    - id: "structure_prediction"
      name: "Predict Structure"
      model: "alphafold2-multimer@v2.3.2"
      inputs:
        sequence: "${workflow.input.sequence}"
        num_recycles: "${workflow.input.num_recycles}"
      outputs:
        - structure
        - confidence
        - pae
        
    - id: "quality_check"
      name: "Quality Control"
      model: "structure-validator@v1.0"
      inputs:
        structure: "${steps.structure_prediction.outputs.structure}"
      outputs:
        - quality_metrics
        - pass_fail
      
    - id: "refinement"
      name: "Structure Refinement"
      model: "rosetta-relax@v2.0"
      condition: "${steps.quality_check.outputs.pass_fail} == 'pass'"
      inputs:
        structure: "${steps.structure_prediction.outputs.structure}"
      outputs:
        - refined_structure
        - energy_score

deployment:
  strategy: "distributed"
  instance_mapping:
    structure_prediction: "ml.g5.2xlarge"
    quality_check: "ml.t3.medium"  # CPU only
    refinement: "ml.c5.4xlarge"  # CPU intensive

benchmarks:
  cost_per_run: "$2.85"
  avg_latency_ms: 3500
  success_rate: 0.95

ui:
  streamlit:
    enabled: true
    template: "pipeline"
```

### Workflow 2: Protein Design + Validation

```yaml
name: "protein-design-validation"
version: "1.0.0"
type: "pipeline"

description: |
  Design protein sequences for a target backbone, then validate
  by predicting their structures.

composition:
  type: "sequential"
  
  steps:
    - id: "design_sequences"
      model: "proteinmpnn@v1.0.1"
      inputs:
        structure: "${workflow.input.backbone}"
        num_sequences: 10
        temperature: 0.1
        
    - id: "validate_designs"
      model: "esmfold@v1.0.0"
      inputs:
        sequences: "${steps.design_sequences.outputs.sequences}"
      outputs:
        - predicted_structures
        
    - id: "compare_structures"
      model: "structure-comparison@v1.0"
      inputs:
        reference: "${workflow.input.backbone}"
        predictions: "${steps.validate_designs.outputs.predicted_structures}"
      outputs:
        - tm_scores
        - best_design

deployment:
  strategy: "distributed"
  instance_mapping:
    design_sequences: "ml.g5.xlarge"
    validate_designs: "ml.g5.xlarge"
    compare_structures: "ml.t3.medium"

benchmarks:
  cost_per_run: "$1.50"
  avg_latency_ms: 2000
```

---

## Agentic Workflows

### Workflow 3: Therapeutic Antibody Designer (Full Spec)

```yaml
name: "therapeutic-antibody-designer"
version: "1.0.0"
type: "agentic"
domain: "antibody-engineering"

description: |
  AI agent that designs therapeutic antibodies against specified targets.
  Autonomously plans multi-step workflows combining structure prediction,
  binder design, affinity optimization, and developability assessment.

doi: "10.5281/zenodo.9234570"

authors:
  - name: "Computational Biology Lab"
    institution: "MIT"

coordinator:
  type: "foundation-model"
  provider: "bedrock"
  model_id: "anthropic.claude-3-5-sonnet-20241022-v2:0"
  
  mcp_servers:
    - name: "uniprot"
      uri: "mcp://uniprot-server"
      description: "Access UniProt protein database"
      
    - name: "pdb"
      uri: "mcp://pdb-server"
      description: "Access Protein Data Bank"
      
    - name: "sabdab"
      uri: "mcp://sabdab-server"
      description: "Structural Antibody Database"
      
    - name: "pubmed"
      uri: "mcp://pubmed-server"
      description: "Scientific literature search"
      
  aws_strands:
    enabled: true
    connections:
      - name: "internal_antibody_db"
        type: "postgres"
        endpoint: "${env.ANTIBODY_DB_ENDPOINT}"
        description: "Internal antibody screening data"

  system_prompt: |
    You are an expert therapeutic antibody engineer with deep knowledge of:
    - Antibody structure and function
    - Protein-protein interactions
    - Developability and manufacturability
    - Immunogenicity prediction
    - Clinical antibody design principles
    
    Your goal is to design novel therapeutic antibodies that:
    1. Bind target antigen with high affinity (KD < 1nM preferred)
    2. Have favorable developability profiles
    3. Minimize immunogenicity risk
    4. Are manufacturable at scale
    
    Plan and execute multi-step computational workflows using available tools.
    Provide scientific rationale for each decision.
    Iterate based on results to optimize designs.
    Be creative but scientifically rigorous.

available_tools:
  scientific_models:
    - name: "predict_target_structure"
      model: "alphafold3@v1.0.0"
      description: "Predict 3D structure of target antigen from sequence"
      inputs:
        - sequence: string
      outputs:
        - structure: pdb
        - confidence: float
      cost_per_call: "$2.50"
      
    - name: "design_antibody_binder"
      model: "rfdiffusion-antibody@v1.0"
      description: "Design antibody that binds to target epitope"
      inputs:
        - target_structure: pdb
        - epitope_residues: list[int]
      outputs:
        - antibody_structure: pdb
        - binding_score: float
      cost_per_call: "$3.50"
      
    - name: "optimize_sequence"
      model: "proteinmpnn@v1.0.1"
      description: "Optimize antibody sequence for given structure"
      inputs:
        - structure: pdb
        - num_designs: int
      outputs:
        - sequences: list[string]
      cost_per_call: "$0.50"
      
    - name: "predict_affinity"
      model: "equibind@v2.0"
      description: "Predict binding affinity"
      inputs:
        - antibody: pdb
        - antigen: pdb
      outputs:
        - affinity_kd_nm: float
        - confidence: float
      cost_per_call: "$1.20"
      
    - name: "assess_developability"
      model: "antibody-developability@v1.0"
      description: "Predict manufacturability and stability"
      inputs:
        - sequence: string
      outputs:
        - expression_score: float
        - stability_score: float
        - aggregation_risk: float
        - polyreactivity_risk: float
      cost_per_call: "$0.30"
      
    - name: "predict_immunogenicity"
      model: "immunogenicity-predictor@v2.0"
      description: "Predict immunogenic epitopes"
      inputs:
        - sequence: string
      outputs:
        - t_cell_epitopes: list
        - risk_score: float
      cost_per_call: "$0.40"
      
    - name: "humanize_sequence"
      model: "antibody-humanizer@v1.0"
      description: "Humanize non-human antibody sequences"
      inputs:
        - sequence: string
        - species: string
      outputs:
        - humanized_sequence: string
        - identity_to_human: float
      cost_per_call: "$0.50"

  data_tools:
    - name: "search_known_antibodies"
      type: "mcp"
      server: "sabdab"
      function: "search_by_antigen"
      description: "Find existing antibodies targeting similar antigens"
      
    - name: "get_target_info"
      type: "mcp"
      server: "uniprot"
      function: "get_protein"
      description: "Retrieve target protein information from UniProt"
      
    - name: "search_pdb_structures"
      type: "mcp"
      server: "pdb"
      function: "search"
      description: "Search PDB for related structures"
      
    - name: "literature_search"
      type: "mcp"
      server: "pubmed"
      function: "search"
      description: "Search scientific literature"
      
    - name: "query_internal_data"
      type: "aws_strand"
      connection: "internal_antibody_db"
      description: "Query proprietary antibody screening data"

workflow_config:
  max_iterations: 20
  reasoning_budget: 50  # Max tool calls
  early_stopping:
    condition: "affinity < 1.0 and developability > 0.8"
    
  constraints:
    max_cost: "$100"
    timeout_minutes: 60

benchmarks:
  validation_set: "therapeutic-antibodies-benchmark"
  success_rate: 0.75
  avg_cost_per_design: "$45"
  avg_time_minutes: 25

deployment:
  instance_type: "ml.g5.12xlarge"
  auto_scale: true
  min_instances: 0
  max_instances: 10

ui:
  streamlit:
    enabled: true
    template: "agentic-chat"
    
  interface:
    type: "chat"
    
    initial_prompts:
      - "Design an antibody against PD-1 for cancer immunotherapy"
      - "Create a SARS-CoV-2 neutralizing antibody"
      - "Design bispecific antibody for CD3 and tumor antigen"
      
    display_config:
      show_tool_calls: true
      show_reasoning: true
      show_cost: true
      stream_response: true
      
    sidebar:
      metrics:
        - label: "Tool Calls"
          value: "${session.tool_count}"
        - label: "Total Cost"
          value: "$${session.total_cost:.2f}"
        - label: "Time Elapsed"
          value: "${session.elapsed_time}s"
      
      controls:
        - type: "slider"
          name: "reasoning_budget"
          label: "Max Tool Calls"
          min: 10
          max: 100
          default: 50
          
        - type: "number"
          name: "max_cost"
          label: "Budget ($)"
          min: 10
          max: 500
          default: 100

  outputs:
    - name: "designed_antibodies"
      type: "table"
      columns: 
        - "rank"
        - "sequence"
        - "affinity_kd_nm"
        - "developability_score"
        - "immunogenicity_risk"
      download: true
      
    - name: "decision_trace"
      type: "tree_view"
      expandable: true
      
    - name: "final_structures"
      type: "molecule_viewer"
      multi_structure: true
```

### Usage Example: Antibody Designer

```python
from conduit import AgenticModel

# Load the agent
agent = AgenticModel.load("therapeutic-antibody-designer@v1.0.0")

# Deploy endpoint
endpoint = agent.deploy(
    instance_type="ml.g5.12xlarge",
    endpoint_name="antibody-designer-prod"
)

# Run design campaign
result = endpoint.run(
    objective="""
    Design a therapeutic antibody that:
    - Targets SARS-CoV-2 Spike protein receptor binding domain (RBD)
    - Has neutralizing activity (blocks ACE2 binding)
    - Affinity KD < 1 nM
    - Human framework
    - Favorable developability profile
    """,
    
    constraints={
        "species": "human",
        "min_affinity_kd_nm": 1.0,
        "min_developability_score": 0.7,
        "avoid_immunogenicity": True,
        "allow_glycosylation": False
    },
    
    context={
        "target_sequence": "NITNLCPFGEVFNATR...",  # RBD sequence
        "competing_antibodies": ["REGN10933", "REGN10987"],
        "preferred_epitope": "ACE2_binding_site"
    },
    
    budget={
        "max_cost_usd": 100,
        "max_tool_calls": 50,
        "timeout_minutes": 60
    }
)

# Result includes:
print(f"Designed {len(result.antibodies)} candidates")
print(f"Best affinity: {result.best_candidate.affinity_kd_nm:.2f} nM")
print(f"Total cost: ${result.total_cost:.2f}")
print(f"Tool calls used: {result.tool_calls_count}")

# Get full decision trace
trace = result.decision_trace
for step in trace.steps:
    print(f"{step.action}: {step.rationale}")
    print(f"  Result: {step.result}")
    
# Download results
result.save("antibody_design_results/")
```

### Agent Execution Trace Example

```
=== Antibody Design Session ===
Objective: Design antibody against SARS-CoV-2 RBD

[Step 1] Planning
Agent: I'll start by gathering information about the target and existing antibodies.
→ Calling: search_known_antibodies(antigen="SARS-CoV-2 RBD")
→ Calling: literature_search(query="SARS-CoV-2 neutralizing antibodies epitopes")
Result: Found 47 existing antibodies. Most target 3 main epitopes.

[Step 2] Target Structure
Agent: Need 3D structure of RBD. Checking PDB first.
→ Calling: search_pdb_structures(query="SARS-CoV-2 spike RBD")
Result: Found 7ANH with 1.95Å resolution. Using this.

[Step 3] Epitope Analysis
Agent: Based on literature, ACE2 binding site is key epitope. Analyzing residues.
Rationale: This epitope is conserved and functionally critical for viral entry.
Decision: Target residues 417, 446, 449, 453, 455, 456, 475-489, 493, 496, 498, 501

[Step 4] Initial Design
Agent: Designing binder targeting ACE2 binding site.
→ Calling: design_antibody_binder(
    target_structure="7ANH",
    epitope_residues=[417, 446, 449...]
)
Result: Generated Fv structure with binding score 0.82

[Step 5] Sequence Optimization
Agent: Optimizing sequence while maintaining structure.
→ Calling: optimize_sequence(structure=<fv>, num_designs=10)
Result: Generated 10 sequence variants

[Step 6] Affinity Screening
Agent: Predicting affinities for all 10 variants.
→ Calling: predict_affinity (10 times in parallel)
Results: KD range 0.8 - 15.3 nM. Top 3 < 2 nM.

[Step 7] Developability Check
Agent: Assessing top 3 candidates for developability issues.
→ Calling: assess_developability (3 times)
Results:
  - Candidate 1: expression=0.85, stability=0.78, aggregation=low
  - Candidate 2: expression=0.71, stability=0.92, aggregation=medium
  - Candidate 3: expression=0.88, stability=0.82, aggregation=low

[Step 8] Immunogenicity Analysis
Agent: Checking for immunogenic epitopes in top candidates.
→ Calling: predict_immunogenicity (3 times)
Results: All candidates have low immunogenicity risk (<0.2)

[Step 9] Humanization
Agent: Candidate 1 has non-human framework residues. Humanizing.
→ Calling: humanize_sequence(sequence=..., species="mouse")
Result: Humanized sequence with 94% identity to human germline

[Step 10] Final Validation
Agent: Re-predicting affinity for humanized candidate.
→ Calling: predict_affinity(humanized_candidate, target)
Result: KD = 1.2 nM (slightly reduced but still excellent)

=== Design Complete ===
Best Candidate:
  - Affinity: 1.2 nM
  - Developability: 0.85/1.0
  - Immunogenicity risk: Low (0.15)
  - Human germline identity: 94%
  - Sequence: EVQLVESGGGLVQPGG...

Tool calls used: 32/50
Total cost: $47.30
Time: 18 minutes

Recommendation: Proceed with experimental validation. 
Consider producing candidate 3 as backup (higher stability).
```

---

## UI Specifications

### Streamlit Templates

#### Template 1: Protein Structure Viewer

```python
# Auto-generated from ui: section in model.yaml
import streamlit as st
from conduit import Model
import py3Dmol

st.set_page_config(page_title="AlphaFold2", layout="wide")

# Load model
@st.cache_resource
def load_model():
    return Model.load("alphafold2-multimer@v2.3.2")

model = load_model()

# Header with citation
st.title("🧬 AlphaFold2 Multimer")
st.markdown(model.description)

with st.sidebar:
    st.markdown(f"**DOI:** [{model.doi}](https://doi.org/{model.doi})")
    st.markdown(f"**Paper:** [Link]({model.citation.paper})")
    st.markdown(f"**GitHub:** [Repo]({model.github_repo})")
    
    st.divider()
    
    st.caption(f"Model version: {model.version}")
    st.caption(f"Certified: {model.quality.certification_date}")

# Input section
st.header("Input")

col1, col2 = st.columns([2, 1])

with col1:
    sequence = st.text_area(
        "Protein Sequence (FASTA format)",
        height=150,
        placeholder=">protein1\nMKLLVGDDS...\n>protein2\nMAFGKL...",
        help="Enter protein sequence(s) in FASTA format"
    )

with col2:
    model_preset = st.selectbox(
        "Model Preset",
        options=["monomer", "multimer", "monomer_ptm"],
        help="Choose prediction mode"
    )
    
    num_recycles = st.slider(
        "Number of Recycles",
        min_value=1,
        max_value=20,
        value=3,
        help="More recycles = higher accuracy but slower"
    )
    
    use_templates = st.checkbox(
        "Use structural templates",
        value=True,
        help="Search PDB for template structures"
    )

# Examples
with st.expander("📋 Try an Example"):
    example_choice = st.selectbox(
        "Select example",
        ["Custom", "T1083 (CASP15)", "Antibody-Antigen Complex"]
    )
    
    if example_choice == "T1083 (CASP15)":
        if st.button("Load Example"):
            st.session_state.sequence = ">T1083\nMKLLVGDDS..."
            st.rerun()

# Predict button
if st.button("🚀 Predict Structure", type="primary"):
    if not sequence:
        st.error("Please provide a sequence")
    else:
        with st.spinner("Running AlphaFold2... This may take 1-2 minutes"):
            # Progress bar
            progress_bar = st.progress(0)
            status = st.empty()
            
            # Deploy endpoint if needed (cached)
            status.text("Deploying model endpoint...")
            progress_bar.progress(10)
            
            endpoint = model.deploy(
                instance_type="ml.g5.2xlarge",
                auto_scale=True
            )
            
            # Run prediction
            status.text("Running prediction...")
            progress_bar.progress(30)
            
            result = endpoint.predict({
                "sequence": sequence,
                "num_recycles": num_recycles,
                "model_preset": model_preset,
                "use_templates": use_templates
            })
            
            progress_bar.progress(100)
            status.text("Complete!")
            
            # Store in session state
            st.session_state.result = result

# Display results
if "result" in st.session_state:
    result = st.session_state.result
    
    st.header("Results")
    
    # Metrics
    col1, col2, col3, col4 = st.columns(4)
    with col1:
        st.metric("Mean pLDDT", f"{result.mean_plddt:.2f}")
    with col2:
        st.metric("pTM Score", f"{result.ptm_score:.3f}")
    with col3:
        st.metric("Time", f"{result.inference_time:.1f}s")
    with col4:
        st.metric("Cost", f"${result.cost:.4f}")
    
    # Structure viewer
    st.subheader("Predicted Structure")
    
    view = py3Dmol.view(width=800, height=600)
    view.addModel(result.structure_pdb, "pdb")
    view.setStyle({"cartoon": {"color": "spectrum"}})
    view.zoomTo()
    st.components.v1.html(view._make_html(), height=620)
    
    # Download button
    st.download_button(
        "📥 Download PDB",
        data=result.structure_pdb,
        file_name=f"prediction_{result.job_id}.pdb",
        mime="chemical/x-pdb"
    )
    
    # Confidence plot
    st.subheader("Per-Residue Confidence (pLDDT)")
    st.line_chart(result.plddt_per_residue)
    
    # PAE heatmap
    st.subheader("Predicted Aligned Error")
    st.pyplot(result.pae_heatmap_figure)
    
    # Additional info
    with st.expander("📊 Detailed Metrics"):
        st.json(result.detailed_metrics)
```

#### Template 2: Agentic Chat Interface

```python
# Auto-generated for agentic workflows
import streamlit as st
from conduit import AgenticModel

st.set_page_config(page_title="Antibody Designer", layout="wide")

# Load agent
@st.cache_resource
def load_agent():
    return AgenticModel.load("therapeutic-antibody-designer@v1.0.0")

agent = load_agent()

# Initialize session state
if "messages" not in st.session_state:
    st.session_state.messages = []
if "tool_count" not in st.session_state:
    st.session_state.tool_count = 0
if "total_cost" not in st.session_state:
    st.session_state.total_cost = 0.0

# Header
st.title("🔬 Therapeutic Antibody Designer")
st.markdown("AI agent for computational antibody design")

# Sidebar
with st.sidebar:
    st.header("Agent Configuration")
    
    reasoning_budget = st.slider(
        "Max Tool Calls",
        min_value=10,
        max_value=100,
        value=50
    )
    
    max_cost = st.number_input(
        "Budget ($)",
        min_value=10,
        max_value=500,
        value=100
    )
    
    st.divider()
    
    st.header("Session Metrics")
    st.metric("Tool Calls", st.session_state.tool_count)
    st.metric("Total Cost", f"${st.session_state.total_cost:.2f}")
    st.progress(st.session_state.tool_count / reasoning_budget)
    
    st.divider()
    
    if st.button("🗑️ Clear Session"):
        st.session_state.messages = []
        st.session_state.tool_count = 0
        st.session_state.total_cost = 0.0
        st.rerun()

# Example prompts
st.subheader("Quick Start")
col1, col2, col3 = st.columns(3)

with col1:
    if st.button("🎯 Design anti-PD-1"):
        prompt = "Design an antibody against PD-1 for cancer immunotherapy"
        
with col2:
    if st.button("🦠 COVID neutralizer"):
        prompt = "Create a SARS-CoV-2 neutralizing antibody"
        
with col3:
    if st.button("🔗 Bispecific"):
        prompt = "Design bispecific antibody for CD3 and tumor antigen"

# Chat interface
st.subheader("Design Session")

# Display chat history
for message in st.session_state.messages:
    with st.chat_message(message["role"]):
        st.markdown(message["content"])
        
        # Show tool calls if present
        if "tool_calls" in message and message["tool_calls"]:
            with st.expander(f"🔧 Tool Calls ({len(message['tool_calls'])})"):
                for i, tool_call in enumerate(message["tool_calls"]):
                    st.code(
                        f"{tool_call['name']}({tool_call['args']})",
                        language="python"
                    )
                    if "result" in tool_call:
                        st.json(tool_call["result"])
                    if "cost" in tool_call:
                        st.caption(f"Cost: ${tool_call['cost']:.4f}")

# User input
if prompt := st.chat_input("Describe the antibody you want to design..."):
    # Add user message
    st.session_state.messages.append({
        "role": "user",
        "content": prompt
    })
    
    # Display user message
    with st.chat_message("user"):
        st.markdown(prompt)
    
    # Agent response
    with st.chat_message("assistant"):
        response_placeholder = st.empty()
        tool_calls_placeholder = st.empty()
        
        # Stream response from agent
        full_response = ""
        tool_calls = []
        
        for chunk in agent.stream(
            prompt,
            context={
                "previous_messages": st.session_state.messages,
                "reasoning_budget": reasoning_budget,
                "max_cost": max_cost
            }
        ):
            if chunk.type == "text":
                full_response += chunk.content
                response_placeholder.markdown(full_response + "▌")
                
            elif chunk.type == "tool_call":
                tool_calls.append(chunk.tool_call)
                st.session_state.tool_count += 1
                st.session_state.total_cost += chunk.cost
                
                with tool_calls_placeholder.container():
                    st.status(f"🔧 Using {chunk.tool_call.name}...")
        
        response_placeholder.markdown(full_response)
        
        # Store message
        st.session_state.messages.append({
            "role": "assistant",
            "content": full_response,
            "tool_calls": tool_calls
        })
        
        # Update sidebar
        st.rerun()
```

---

## Deployment Patterns

### Pattern 1: Single Model Deployment

```python
from conduit import Model

# Load model
model = Model.load("alphafold2-multimer@v2.3.2")

# Deploy to Bedrock
endpoint = model.deploy_to_bedrock(
    endpoint_name="alphafold2-prod",
    instance_type="ml.g5.2xlarge",
    initial_instance_count=1,
    auto_scaling={
        "min_instances": 0,
        "max_instances": 10,
        "target_invocations_per_instance": 100
    },
    tags={
        "project": "protein-structure",
        "cost-center": "research"
    }
)

# Use endpoint
result = endpoint.predict({
    "sequence": "MKLLVGDDS...",
    "num_recycles": 3
})

print(f"pLDDT: {result['mean_plddt']:.2f}")
```

### Pattern 2: Batch Processing

```python
from conduit import Model

model = Model.load("alphafold2-multimer@v2.3.2")

# Batch predict (automatically handles large batches)
sequences = [
    "MKLLVGDDS...",
    "MAFGKLQPE...",
    # ... 10,000 more sequences
]

results = model.batch_predict(
    inputs=[{"sequence": seq} for seq in sequences],
    batch_size=100,
    output_s3_uri="s3://my-bucket/results/",
    instance_type="ml.g5.12xlarge",
    max_concurrent_transforms=10
)

# Results are written to S3 as they complete
# Can monitor progress
for result in results:
    print(f"Processed {result.sequence_id}: pLDDT={result.mean_plddt}")
```

### Pattern 3: Pipeline Deployment

```python
from conduit import Pipeline

# Load pipeline
pipeline = Pipeline.load("protein-design-validation@v1.0.0")

# Deploy (automatically deploys all components)
pipeline_endpoint = pipeline.deploy(
    mode="managed",  # Conduit handles orchestration
    instance_mapping={
        "design_sequences": "ml.g5.xlarge",
        "validate_designs": "ml.g5.xlarge",
        "compare_structures": "ml.t3.medium"
    }
)

# Use as single endpoint
result = pipeline_endpoint.run({
    "backbone": load_pdb("target_backbone.pdb"),
    "num_designs": 10
})

print(f"Best design: {result.best_design.sequence}")
print(f"TM-score: {result.best_design.tm_score:.3f}")
```

### Pattern 4: Agentic Deployment

```python
from conduit import AgenticModel

# Load agent
agent = AgenticModel.load("therapeutic-antibody-designer@v1.0.0")

# Deploy
endpoint = agent.deploy(
    instance_type="ml.g5.12xlarge",
    endpoint_name="antibody-designer",
    mcp_config={
        "servers": ["uniprot", "pdb", "sabdab"],
        "cache_ttl": 3600
    },
    strands_config={
        "connections": ["internal_antibody_db"]
    }
)

# Run design campaign
result = endpoint.run(
    objective="Design antibody against SARS-CoV-2 RBD",
    constraints={
        "min_affinity_kd_nm": 1.0,
        "min_developability_score": 0.7
    },
    budget={
        "max_cost_usd": 100,
        "max_tool_calls": 50
    }
)
```

---

## Success Metrics

### Publisher Metrics

**Model Quality:**
- Benchmark performance vs baselines
- Peer review pass rate
- Citation count
- Usage/download count

**Engagement:**
- Models published per month: Target 20+
- Active publishers: Target 50+ organizations
- Model updates/versions: Target 2+ per model/year

### Consumer Metrics

**Adoption:**
- Unique users per month: Target 1,000+
- Deployments per month: Target 5,000+
- Institutions using: Target 100+

**Usage:**
- Inference requests per month: Target 1M+
- Compute hours per month: Target 10,000+
- Cost per inference: Target $0.50-$5.00

**Satisfaction:**
- Time to first prediction: Target <5 minutes
- Success rate (predictions work): Target >95%
- User satisfaction score: Target >4.5/5

### Business Metrics (AWS)

**Revenue Impact:**
- Bedrock inference compute: Target $500k+/month
- Foundation model usage: Target $200k+/month
- Storage (S3): Target $50k+/month

**Strategic:**
- Enterprise customers: Target 20+
- Government/national lab partnerships: Target 5+
- Published case studies: Target 10+

### Scientific Impact

**Research Output:**
- Papers citing models: Target 100+/year
- Novel discoveries enabled: Track qualitatively
- Time saved vs HPC: Target 80% reduction

**Community Health:**
- GitHub stars/forks: Target 5,000+
- Community contributions: Target 50+ contributors
- Domain coverage: Target 10+ scientific domains

---

## Appendix: File Structure

### Example Model Repository

```
alphafold2-multimer/
├── model.yaml                      # Complete specification
├── README.md                       # Human-readable docs
├── LICENSE                         # Apache-2.0
│
├── src/
│   ├── inference.py               # Main inference code
│   ├── model.py                   # Model architecture
│   ├── data.py                    # Data processing
│   └── utils.py                   # Helper functions
│
├── requirements.txt               # Python dependencies
├── environment.yml                # Conda environment (optional)
│
├── schemas/
│   ├── input.json                 # Input schema (JSON Schema)
│   └── output.json                # Output schema
│
├── examples/
│   ├── example_input.json
│   ├── example_output.json
│   └── notebook.ipynb             # Usage examples
│
├── tests/
│   ├── test_inference.py
│   ├── test_model.py
│   └── fixtures/
│       ├── test_sequence.fasta
│       └── expected_output.pdb
│
├── benchmarks/
│   ├── casp15_results.json
│   ├── benchmark_script.py
│   └── plots/
│       ├── accuracy_vs_cost.png
│       └── latency_distribution.png
│
├── docs/
│   ├── usage.md
│   ├── architecture.md
│   └── api.md
│
└── .github/
    └── workflows/
        ├── validate.yml           # Auto-validation on PR
        └── publish.yml            # Auto-publish on tag
```

---

## Next Steps

**For Implementation:**

1. **Phase 1: Core Infrastructure (Weeks 1-4)**
   - Build CLI tool (`conduit` command)
   - Implement model.yaml parser
   - Create Bedrock deployment automation
   - Build basic catalog website

2. **Phase 2: Seed Models (Weeks 5-8)**
   - Partner with 3-5 research groups
   - Port AlphaFold2, ESMFold, ProteinMPNN
   - Create benchmark framework
   - Generate initial benchmarks

3. **Phase 3: Workflows (Weeks 9-12)**
   - Implement pipeline support
   - Build first agentic workflow
   - Create Streamlit UI generator
   - Deploy hosted UI service

4. **Phase 4: Launch (Week 13)**
   - Soft launch to beta users
   - Gather feedback
   - Iterate on UX
   - Public launch

**For Adoption:**

1. Conference presentations (NeurIPS, ICLR, ISMB)
2. Blog posts ("Stop waiting for GPU queues")
3. Academic partnerships
4. Enterprise pilots with pharma/biotech

---

**End of Specification**
