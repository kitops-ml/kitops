# Packing RAG Systems with KitOps

## Technical Documentation & Integration Guide

### Using Jozu Hub Registry

**Version 1.0**

A comprehensive guide to packaging, distributing, and deploying RAG (Retrieval-Augmented Generation) systems using KitOps and Jozu Hub for reproducible AI model management.

---

## 1. Introduction

### 1.1 Overview

This documentation provides a comprehensive guide for packaging Retrieval-Augmented Generation (RAG) systems using KitOps and Jozu Hub. RAG combines large language models with external knowledge retrieval to provide contextually accurate and up-to-date responses. KitOps enables the packaging, versioning, and distribution of these complex AI systems as portable, reproducible ModelKits stored in Jozu Hub registry.

### 1.2 What is KitOps?

KitOps is an open-source tool for packaging AI/ML projects using OCI (Open Container Initiative) standards. It allows you to create ModelKits that bundle models, datasets, code, and configurations into a single, versioned artifact that can be stored in OCI-compliant registries like Jozu Hub.

### 1.3 What is Jozu Hub?

Jozu Hub (jozu.ml) is a specialized OCI registry designed specifically for AI/ML artifacts. It provides:

- **Free public hosting** for open-source AI projects and ModelKits
- **Optimized storage and transfer** for large model files and datasets
- **Web UI** for browsing, discovering, and managing ModelKits
- **Version control and metadata tracking** for AI artifacts
- **Team collaboration features** for private ModelKits
- **No Docker installation required** - KitOps works standalone

### 1.4 Why Use KitOps with Jozu Hub for RAG?

KitOps with Jozu Hub provides several key benefits for RAG deployment:

- **Reproducibility**: Ensures consistent RAG pipeline deployment across environments
- **Version Control**: Tracks changes to embeddings, models, and vector databases
- **Portability**: Package entire RAG systems for deployment anywhere
- **Collaboration**: Share complete RAG implementations via Jozu Hub
- **Discoverability**: Public ModelKits on Jozu Hub can be discovered by the community
- **No Container Dependencies**: Works without Docker or container runtimes
- **Standardization**: Use OCI standards for AI artifact management

---

## 2. Prerequisites

### 2.1 System Requirements

- **Operating System**: Linux, macOS, or Windows with WSL2
- **Memory**: Minimum 8GB RAM (16GB+ recommended for large models)
- **Storage**: 20GB+ free disk space for models and vector databases
- **Internet Connection**: Required for Jozu Hub access

### 2.2 Required Software

- Python 3.8 or higher
- Git for version control
- KitOps CLI (installation covered in next section)
- Jozu Hub account (free at jozu.ml)

> **📝 Note**: Docker is NOT required. KitOps works independently of Docker and communicates directly with Jozu Hub registry.

### 2.3 Knowledge Prerequisites

- Basic understanding of RAG architecture and components
- Familiarity with vector databases (ChromaDB, Pinecone, Weaviate, etc.)
- Basic Python programming and ML concepts
- Understanding of OCI registry concepts

---

## 3. Installation and Setup

### 3.1 Installing KitOps CLI

#### Linux/macOS Installation

Use the official installation script:

```bash
curl -s https://get.kitops.ml | bash
```

#### Windows Installation

Download the Windows binary from the releases page:

```powershell
# PowerShell
iwr https://github.com/jozu-ai/kitops/releases/latest/download/kitops-windows-x86_64.zip -OutFile kitops.zip

# Extract and add to PATH
Expand-Archive kitops.zip -DestinationPath C:\kitops
$env:Path += ";C:\kitops"
```

#### Verify Installation

```bash
kit version

# Expected output:
# KitOps version X.X.X
```

### 3.2 Creating a Jozu Hub Account

1. Visit <https://jozu.ml> and click **Sign Up**
2. Create your account with email or GitHub authentication
3. Verify your email address
4. Generate a Personal Access Token (PAT) from your account settings
5. Save your token securely - you will need it for authentication

### 3.3 Authenticate with Jozu Hub

Login to Jozu Hub using your credentials:

```bash
# Interactive login
kit login jozu.ml

# You will be prompted for:
# Username: your-username
# Password/Token: your-personal-access-token
```

Alternative authentication using environment variable:

```bash
# Set token as environment variable
export JOZU_TOKEN=your-personal-access-token

# Login using token
echo $JOZU_TOKEN | kit login jozu.ml -u your-username --password-stdin
```

#### Verify Authentication

```bash
# Verify you're logged in
kit login jozu.ml --test

# Expected output:
# Successfully authenticated with jozu.ml
```

---

## 4. RAG System Architecture

### 4.1 RAG Components

A typical RAG system consists of the following components that need to be packaged:

- **Embedding Model**: Converts text to vector representations (e.g., sentence-transformers, OpenAI embeddings)
- **Vector Database**: Stores and retrieves embeddings (ChromaDB, Pinecone, Weaviate, FAISS)
- **Large Language Model**: Generates responses (GPT, Claude, LLaMA, Mistral)
- **Document Loader**: Processes source documents (PDFs, text files, web pages)
- **Retrieval Pipeline**: Orchestrates query processing and context retrieval
- **Application Code**: API endpoints, business logic, and orchestration

### 4.2 Directory Structure

Recommended project structure for a KitOps-packaged RAG system:

```
rag-project/
├── Kitfile                    # KitOps manifest
├── models/
│   ├── embeddings/
│   │   └── all-MiniLM-L6-v2/  # Embedding model
│   └── llm/
│       └── mistral-7b/        # LLM model
├── vectordb/
│   └── chroma_db/             # Vector database files
├── datasets/
│   └── knowledge_base/        # Source documents
│       ├── docs/
│       └── metadata.json
├── code/
│   ├── ingest.py              # Document ingestion
│   ├── query.py               # Query pipeline
│   ├── app.py                 # Main application
│   └── requirements.txt       # Python dependencies
├── configs/
│   ├── embedding_config.yaml
│   ├── retrieval_config.yaml
│   └── llm_config.yaml
└── docs/
    └── README.md              # Documentation
```

---

## 5. Creating the Kitfile

### 5.1 Kitfile Fundamentals

The Kitfile is a YAML manifest that defines all components of your RAG system. It follows the ModelKit specification and uses OCI standards for packaging to Jozu Hub.

### 5.2 Basic Kitfile Structure

```yaml
manifestVersion: 1.0
package:
  name: rag-system
  version: 1.0.0
  description: RAG system with embeddings and LLM
  authors:
    - Your Name <email@example.com>
  
model:
  name: embedding-model
  path: ./models/embeddings/all-MiniLM-L6-v2
  framework: sentence-transformers
  description: Sentence embedding model for document vectorization
  
code:
  - path: ./code
    description: RAG application code
    
datasets:
  - name: knowledge-base
    path: ./datasets/knowledge_base
    description: Corpus for RAG retrieval

vectordb:
  - name: chroma-db
    path: ./vectordb/chroma_db
    description: Vector embeddings database
```

### 5.3 Complete RAG Kitfile Example

A production-ready Kitfile with multiple components for Jozu Hub:

```yaml
manifestVersion: 1.0

package:
  name: enterprise-rag-system
  version: 2.1.0
  description: Enterprise RAG with multi-model support
  authors:
    - AI Team <ai-team@company.com>
  license: Apache-2.0
  tags:
    - rag
    - nlp
    - retrieval
    - production

model:
  - name: embedding-model
    path: ./models/embeddings/all-mpnet-base-v2
    framework: sentence-transformers
    version: 2.2.2
    description: High-quality embedding model
    
  - name: reranker-model
    path: ./models/reranker/ms-marco-MiniLM
    framework: sentence-transformers
    description: Cross-encoder for reranking results
    
  - name: llm-model
    path: ./models/llm/mistral-7b-instruct-v0.2
    framework: transformers
    format: gguf
    description: Instruction-tuned LLM

code:
  - path: ./code
    description: Main RAG application
  - path: ./scripts
    description: Deployment and utility scripts

datasets:
  - name: knowledge-base-prod
    path: ./datasets/knowledge_base_v2
    description: Production knowledge corpus
    
  - name: evaluation-set
    path: ./datasets/eval_queries
    description: Evaluation queries and ground truth

configs:
  - path: ./configs/embedding_config.yaml
  - path: ./configs/retrieval_config.yaml
  - path: ./configs/llm_config.yaml
  - path: ./configs/deployment.yaml

vectordb:
  - name: primary-vectordb
    path: ./vectordb/chroma_db_prod
    backend: chromadb
    description: Main vector database
    
environment:
  EMBEDDING_BATCH_SIZE: "32"
  RETRIEVAL_TOP_K: "5"
  LLM_MAX_TOKENS: "2048"
  VECTOR_DIMENSION: "768"
  
docs:
  - path: ./docs/README.md
  - path: ./docs/API.md
  - path: ./docs/DEPLOYMENT.md
```

---

## 6. Best Practices for Jozu Hub

### 6.1 Model Management

- **Version Control**: Use semantic versioning (major.minor.patch) for ModelKits in Jozu Hub
- **Model Formats**: Prefer quantized formats (GGUF, ONNX) to reduce upload size and storage
- **Embedding Consistency**: Lock embedding model versions to prevent drift across deployments
- **Metadata Tags**: Use Jozu Hub tags for discoverability (rag, nlp, production, etc.)
- **Public vs Private**: Use private repositories for proprietary models, public for open-source

### 6.2 Vector Database Management

- **Persistence**: Always include vector DB files in the ModelKit for complete reproducibility
- **Indexing**: Pre-build indexes before packaging for faster deployment startup
- **Metadata**: Store document metadata alongside embeddings for filtering
- **Compression**: Use compression for large vector databases to optimize Jozu Hub storage

### 6.3 Dataset Management

- **Data Versioning**: Track dataset changes with clear version increments in Jozu Hub
- **Preprocessing**: Include preprocessed data to ensure consistency across environments
- **Documentation**: Add metadata files describing data sources and processing steps
- **Size Optimization**: Compress large datasets before pushing to Jozu Hub

### 6.4 Jozu Hub Naming Conventions

Follow these naming conventions for clarity in Jozu Hub:

- **Repository Format**: `jozu.ml/username/modelkit-name:version`
- **Use lowercase and hyphens**: `my-rag-system` (not `My_RAG_System`)
- **Include purpose in name**: `customer-support-rag`, `legal-doc-rag`
- **Version tags**: Use semantic versioning or environment tags (`v1.0.0`, `prod`, `staging`)

### 6.5 Security Best Practices

- **Secrets Management**: Never include API keys or credentials in ModelKits
- **Access Control**: Use Jozu Hub private repositories for proprietary models
- **Token Security**: Rotate Jozu Hub Personal Access Tokens regularly
- **Scanning**: Run security scans on dependencies before packaging
- **Data Privacy**: Ensure datasets comply with privacy regulations before uploading

---

## 7. Step-by-Step Integration Guide

### 7.1 Step 1: Prepare Your RAG Project

#### Initialize Project Structure

```bash
# Create project directory
mkdir rag-project && cd rag-project

# Create subdirectories
mkdir -p models/{embeddings,llm} vectordb datasets code configs docs

# Initialize git repository
git init

# Create .gitignore
cat > .gitignore << EOF
*.pyc
__pycache__/
.env
.venv/
models/*.bin
vectordb/
EOF
```

### 7.2 Step 2: Download and Prepare Models

#### Download Embedding Model

```python
# Using Python
python3 << EOF
from sentence_transformers import SentenceTransformer

# Download model
model = SentenceTransformer('sentence-transformers/all-MiniLM-L6-v2')

# Save locally
model.save('models/embeddings/all-MiniLM-L6-v2')
print("Embedding model downloaded successfully")
EOF
```

#### Download LLM Model

```bash
# For local LLMs (using huggingface-cli)
pip install huggingface-hub

# Download quantized model
huggingface-cli download TheBloke/Mistral-7B-Instruct-v0.2-GGUF \
  mistral-7b-instruct-v0.2.Q4_K_M.gguf \
  --local-dir models/llm/mistral-7b \
  --local-dir-use-symlinks False
```

### 7.3 Step 3: Create Vector Database

#### Setup ChromaDB

```bash
# Create ingest script: code/ingest.py
cat > code/ingest.py << 'EOF'
import chromadb
from sentence_transformers import SentenceTransformer
import os
from pathlib import Path

# Initialize embedding model
model = SentenceTransformer('../models/embeddings/all-MiniLM-L6-v2')

# Create ChromaDB client
client = chromadb.PersistentClient(path="../vectordb/chroma_db")

# Create or get collection
collection = client.get_or_create_collection(
    name="knowledge_base",
    metadata={"description": "RAG knowledge base"}
)

# Load and process documents
docs_path = Path("../datasets/knowledge_base/docs")
documents = []
metadatas = []
ids = []

for idx, file_path in enumerate(docs_path.glob("*.txt")):
    with open(file_path, 'r') as f:
        content = f.read()
        documents.append(content)
        metadatas.append({
            "source": file_path.name,
            "type": "text"
        })
        ids.append(f"doc_{idx}")

# Generate embeddings and add to collection
embeddings = model.encode(documents).tolist()

collection.add(
    embeddings=embeddings,
    documents=documents,
    metadatas=metadatas,
    ids=ids
)

print(f"Indexed {len(documents)} documents")
EOF
```

#### Run Ingestion

```bash
# Install dependencies
pip install chromadb sentence-transformers

# Add sample documents
mkdir -p datasets/knowledge_base/docs
echo "Sample document about AI and machine learning" > datasets/knowledge_base/docs/doc1.txt
echo "Information about RAG systems and retrieval" > datasets/knowledge_base/docs/doc2.txt

# Run ingestion
python code/ingest.py
```

### 7.4 Step 4: Implement RAG Query Pipeline

```bash
# Create query script: code/query.py
cat > code/query.py << 'EOF'
import chromadb
from sentence_transformers import SentenceTransformer
from llama_cpp import Llama

class RAGPipeline:
    def __init__(self):
        # Load embedding model
        self.embedding_model = SentenceTransformer(
            '../models/embeddings/all-MiniLM-L6-v2'
        )
        
        # Connect to vector DB
        self.client = chromadb.PersistentClient(
            path="../vectordb/chroma_db"
        )
        self.collection = self.client.get_collection(
            name="knowledge_base"
        )
        
        # Load LLM
        self.llm = Llama(
            model_path="../models/llm/mistral-7b/mistral-7b-instruct-v0.2.Q4_K_M.gguf",
            n_ctx=4096,
            n_threads=8
        )
    
    def retrieve(self, query, top_k=5):
        # Generate query embedding
        query_embedding = self.embedding_model.encode([query]).tolist()
        
        # Search vector DB
        results = self.collection.query(
            query_embeddings=query_embedding,
            n_results=top_k
        )
        
        return results
    
    def generate(self, query, context):
        # Build prompt with context
        prompt = f"""<s>[INST] Use the following context to answer the question.

Context:
{context}

Question: {query}

Answer: [/INST]"""
        
        # Generate response
        response = self.llm(
            prompt,
            max_tokens=512,
            temperature=0.7,
            top_p=0.9
        )
        
        return response['choices'][0]['text'].strip()
    
    def query(self, question):
        # Retrieve relevant documents
        results = self.retrieve(question)
        
        # Build context from results
        context = "\n\n".join(results['documents'][0])
        
        # Generate answer
        answer = self.generate(question, context)
        
        return {
            'answer': answer,
            'sources': results['metadatas'][0]
        }

if __name__ == "__main__":
    rag = RAGPipeline()
    response = rag.query("What is artificial intelligence?")
    print(f"Answer: {response['answer']}")
    print(f"Sources: {response['sources']}")
EOF
```

### 7.5 Step 5: Create Application Entry Point

```bash
# Create main app: code/app.py
cat > code/app.py << 'EOF'
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
from query import RAGPipeline
import uvicorn

app = FastAPI(title="RAG API")
rag_pipeline = RAGPipeline()

class Query(BaseModel):
    question: str
    top_k: int = 5

class Response(BaseModel):
    answer: str
    sources: list

@app.post("/query", response_model=Response)
async def query_rag(query: Query):
    try:
        result = rag_pipeline.query(query.question)
        return Response(**result)
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

@app.get("/health")
async def health_check():
    return {"status": "healthy"}

if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=8000)
EOF
```

#### Create Requirements File

```bash
# code/requirements.txt
cat > code/requirements.txt << EOF
chromadb==0.4.22
sentence-transformers==2.3.1
llama-cpp-python==0.2.52
fastapi==0.109.0
uvicorn==0.27.0
pydantic==2.5.3
torch==2.1.2
transformers==4.37.2
EOF
```

### 7.6 Step 6: Create Configuration Files

```bash
# Embedding configuration
cat > configs/embedding_config.yaml << EOF
model:
  name: all-MiniLM-L6-v2
  path: ../models/embeddings/all-MiniLM-L6-v2
  dimension: 384
  max_seq_length: 256
  batch_size: 32
  normalize_embeddings: true
EOF

# Retrieval configuration
cat > configs/retrieval_config.yaml << EOF
vectordb:
  backend: chromadb
  path: ../vectordb/chroma_db
  collection: knowledge_base
  
retrieval:
  top_k: 5
  similarity_threshold: 0.7
  rerank: false
  
chunking:
  chunk_size: 512
  chunk_overlap: 50
  separator: "\n\n"
EOF

# LLM configuration
cat > configs/llm_config.yaml << EOF
model:
  name: mistral-7b-instruct-v0.2
  path: ../models/llm/mistral-7b/mistral-7b-instruct-v0.2.Q4_K_M.gguf
  format: gguf
  
generation:
  max_tokens: 512
  temperature: 0.7
  top_p: 0.9
  top_k: 40
  repeat_penalty: 1.1
  
context:
  context_length: 4096
  threads: 8
EOF
```

### 7.7 Step 7: Create the Kitfile

```bash
# Create Kitfile in project root
cat > Kitfile << EOF
manifestVersion: 1.0

package:
  name: rag-system-complete
  version: 1.0.0
  description: Complete RAG system with embeddings, vector DB, and LLM
  authors:
    - Your Name <email@example.com>
  tags:
    - rag
    - nlp
    - production
    - chromadb
    - mistral

model:
  - name: embedding-model
    path: ./models/embeddings/all-MiniLM-L6-v2
    framework: sentence-transformers
    description: Sentence embedding model
    
  - name: llm-model
    path: ./models/llm/mistral-7b
    framework: llama-cpp
    format: gguf
    description: Mistral 7B instruction model

code:
  - path: ./code
    description: RAG application code

datasets:
  - name: knowledge-base
    path: ./datasets/knowledge_base
    description: Source documents corpus

vectordb:
  - name: chroma-vectordb
    path: ./vectordb/chroma_db
    description: Indexed embeddings database

configs:
  - path: ./configs/embedding_config.yaml
  - path: ./configs/retrieval_config.yaml  
  - path: ./configs/llm_config.yaml

environment:
  EMBEDDING_BATCH_SIZE: "32"
  RETRIEVAL_TOP_K: "5"
  LLM_MAX_TOKENS: "512"
EOF
```

### 7.8 Step 8: Pack the ModelKit

```bash
# Build the ModelKit for Jozu Hub
kit pack . -t jozu.ml/your-username/rag-system:1.0.0

# Verify the build
kit list

# Inspect the ModelKit
kit inspect jozu.ml/your-username/rag-system:1.0.0
```

> **📝 Note**: Replace 'your-username' with your actual Jozu Hub username

### 7.9 Step 9: Push to Jozu Hub

```bash
# Push to Jozu Hub registry
kit push jozu.ml/your-username/rag-system:1.0.0

# Tag for different environments
kit tag jozu.ml/your-username/rag-system:1.0.0 jozu.ml/your-username/rag-system:latest
kit tag jozu.ml/your-username/rag-system:1.0.0 jozu.ml/your-username/rag-system:production

# Push tags
kit push jozu.ml/your-username/rag-system:latest
kit push jozu.ml/your-username/rag-system:production
```

After pushing, your ModelKit will be visible on Jozu Hub at: `https://jozu.ml/your-username/rag-system`

### 7.10 Step 10: Deploy and Test

#### Pull from Jozu Hub and Unpack

```bash
# On deployment server
# Authenticate with Jozu Hub first
kit login jozu.ml

# Pull from Jozu Hub
kit pull jozu.ml/your-username/rag-system:1.0.0

# Unpack to local directory
kit unpack jozu.ml/your-username/rag-system:1.0.0 -d ./rag-deployment

# Navigate to deployment
cd rag-deployment
```

#### Setup Environment

```bash
# Create virtual environment
python -m venv venv
source venv/bin/activate  # On Windows: venv\Scripts\activate

# Install dependencies
pip install -r code/requirements.txt

# Set environment variables
export EMBEDDING_BATCH_SIZE=32
export RETRIEVAL_TOP_K=5
export LLM_MAX_TOKENS=512
```

#### Run Application

```bash
# Start the RAG API
cd code
python app.py

# Test in another terminal
curl -X POST http://localhost:8000/query \
  -H "Content-Type: application/json" \
  -d '{"question": "What is AI?", "top_k": 5}'
```

---

## 8. Advanced Topics

### 8.1 Using Jozu Hub Web Interface

Jozu Hub provides a web interface for managing your ModelKits:

- **Browse ModelKits**: Navigate to <https://jozu.ml/explore> to discover public RAG systems
- **View Details**: Click on any ModelKit to see metadata, tags, and version history
- **Manage Access**: Configure public/private visibility and team permissions
- **Download Instructions**: Copy pull commands directly from the web UI
- **Analytics**: View pull counts and usage statistics for your ModelKits

### 8.2 Multi-Stage RAG Pipelines

For complex RAG systems with multiple stages, organize components into separate ModelKits on Jozu Hub:

```bash
# Base retrieval kit
kit pack ./retrieval-stage -t jozu.ml/your-username/rag-retrieval:1.0.0
kit push jozu.ml/your-username/rag-retrieval:1.0.0

# Reranking kit
kit pack ./reranking-stage -t jozu.ml/your-username/rag-reranker:1.0.0
kit push jozu.ml/your-username/rag-reranker:1.0.0

# Generation kit
kit pack ./generation-stage -t jozu.ml/your-username/rag-generator:1.0.0
kit push jozu.ml/your-username/rag-generator:1.0.0

# Composite kit referencing others
cat > Kitfile << EOF
manifestVersion: 1.0
package:
  name: rag-complete-pipeline
  version: 1.0.0
  
dependencies:
  - jozu.ml/your-username/rag-retrieval:1.0.0
  - jozu.ml/your-username/rag-reranker:1.0.0
  - jozu.ml/your-username/rag-generator:1.0.0
EOF
```

### 8.3 Continuous Updates and Versioning

Implement versioning strategy for incremental updates on Jozu Hub:

- **Major version (2.0.0)**: Breaking changes, new architecture, incompatible model changes
- **Minor version (1.1.0)**: New features, model updates, dataset additions (backward compatible)
- **Patch version (1.0.1)**: Bug fixes, config updates, documentation improvements

### 8.4 CI/CD Integration with Jozu Hub

Example GitHub Actions workflow for automated ModelKit building and pushing to Jozu Hub:

```yaml
# .github/workflows/build-modelkit.yml
name: Build and Push ModelKit to Jozu Hub

on:
  push:
    tags:
      - 'v*'

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Install KitOps
        run: |
          curl -s https://get.kitops.ml | bash
          
      - name: Login to Jozu Hub
        run: |
          echo ${{ secrets.JOZU_TOKEN }} | \
          kit login jozu.ml -u ${{ secrets.JOZU_USERNAME }} --password-stdin
          
      - name: Build ModelKit
        run: |
          kit pack . -t jozu.ml/${{ secrets.JOZU_USERNAME }}/rag-system:${{ github.ref_name }}
          
      - name: Push to Jozu Hub
        run: |
          kit push jozu.ml/${{ secrets.JOZU_USERNAME }}/rag-system:${{ github.ref_name }}
          
      - name: Tag as latest
        if: startsWith(github.ref, 'refs/tags/v')
        run: |
          kit tag jozu.ml/${{ secrets.JOZU_USERNAME }}/rag-system:${{ github.ref_name }} \
                  jozu.ml/${{ secrets.JOZU_USERNAME }}/rag-system:latest
          kit push jozu.ml/${{ secrets.JOZU_USERNAME }}/rag-system:latest
```

Configure the following secrets in your GitHub repository:

- `JOZU_USERNAME`: Your Jozu Hub username
- `JOZU_TOKEN`: Your Jozu Hub Personal Access Token

### 8.5 Performance Optimization

- **Model Quantization**: Use 4-bit or 8-bit quantized models to reduce size and speed up uploads
- **Layer Caching**: Jozu Hub caches identical layers - reuse model versions when possible
- **Lazy Loading**: Load models on-demand rather than at startup to reduce memory
- **Compression**: Compress large files before adding to ModelKit
- **Batch Processing**: Process multiple queries in batches for efficiency

---

## 9. Troubleshooting

### 9.1 Common Issues with Jozu Hub

| Issue | Cause | Solution |
|-------|-------|----------|
| Authentication failed | Invalid credentials or expired token | Regenerate Personal Access Token from Jozu Hub settings and login again |
| Push timeout | Large files or slow network | Use `--timeout` flag or compress large model files before packing |
| ModelKit not found | Incorrect repository name or not pushed | Verify repository name with `kit list` and ensure push completed |
| Permission denied | Attempting to push to others repository | Check repository ownership or request access from owner |
| Layer already exists error | Trying to overwrite existing tag | Use a new version tag or delete old tag first with `kit remove` |
| Quota exceeded | Free tier storage limit reached | Upgrade to paid plan or remove old ModelKits |

### 9.2 Debug Commands

```bash
# Validate Kitfile
kit validate Kitfile

# Inspect ModelKit contents
kit inspect jozu.ml/username/rag-system:1.0.0 --format json

# List all layers
kit info jozu.ml/username/rag-system:1.0.0

# Check Jozu Hub connectivity
kit login jozu.ml --test

# Verbose mode for debugging
kit pack . -t jozu.ml/username/test:latest --verbose

# Remove corrupted local cache
kit remove jozu.ml/username/rag-system:1.0.0 --force

# Check Jozu Hub status
curl -I https://jozu.ml/health
```

### 9.3 Performance Diagnostics

```bash
# Monitor pack process
time kit pack . -t jozu.ml/username/test:latest

# Check ModelKit size
kit info jozu.ml/username/rag-system:1.0.0 | grep Size

# Profile unpacking
time kit unpack jozu.ml/username/rag-system:1.0.0 -d test-dir

# Verify file integrity
kit verify jozu.ml/username/rag-system:1.0.0

# Test upload speed to Jozu Hub
kit push jozu.ml/username/test:latest --progress
```

---

## 10. References and Resources

### 10.1 Official Documentation

- **KitOps Official Docs**: <https://kitops.ml/docs>
- **Jozu Hub Website**: <https://jozu.ml>
- **KitOps GitHub**: <https://github.com/jozu-ai/kitops>
- **ModelKit Specification**: <https://kitops.ml/docs/modelkit-spec>
- **OCI Distribution Spec**: <https://github.com/opencontainers/distribution-spec>

### 10.2 RAG Frameworks and Tools

- **LangChain**: <https://python.langchain.com>
- **LlamaIndex**: <https://www.llamaindex.ai>
- **ChromaDB**: <https://www.trychroma.com>
- **Sentence Transformers**: <https://www.sbert.net>

### 10.3 Community and Support

- **KitOps Slack Community**: <https://kitops.ml/slack>
- **Jozu Hub Support**: <support@jozu.ml>
- **Stack Overflow Tag**: kitops
- **GitHub Discussions**: <https://github.com/jozu-ai/kitops/discussions>

### 10.4 Example Projects on Jozu Hub

- **RAG with LangChain**: jozu.ml/examples/rag-langchain
- **Enterprise RAG Pipeline**: jozu.ml/examples/enterprise-rag
- **Multi-Modal RAG**: jozu.ml/examples/multimodal-rag
- **Browse all examples**: <https://jozu.ml/explore?tags=rag>

---

## Appendix A: Kitfile Schema Reference

```yaml
manifestVersion: string (required)
  - Current version: "1.0"

package: object (required)
  name: string (required)
  version: string (required, semver format)
  description: string (optional)
  authors: array of strings (optional)
  license: string (optional)
  tags: array of strings (optional)

model: object or array (optional)
  name: string (required)
  path: string (required)
  framework: string (optional)
  version: string (optional)
  format: string (optional)
  description: string (optional)

code: object or array (optional)
  path: string (required)
  description: string (optional)

datasets: object or array (optional)
  name: string (required)
  path: string (required)
  description: string (optional)

vectordb: object or array (optional)
  name: string (required)
  path: string (required)
  backend: string (optional)
  description: string (optional)

configs: array (optional)
  path: string (required)

environment: object (optional)
  KEY: VALUE (string pairs)

docs: array (optional)
  path: string (required)

dependencies: array (optional)
  - jozu.ml/username/modelkit:version
```

---
