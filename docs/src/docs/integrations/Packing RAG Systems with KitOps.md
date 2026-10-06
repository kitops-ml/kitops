# Packaging RAG Systems

## Overview

RAG (Retrieval-Augmented Generation) systems combine language models with knowledge retrieval. This guide shows how to package RAG components—embedding models, vector databases, LLMs, and code—into a single ModelKit.

**What you'll package:**
- Embedding models for vectorization
- Vector databases (ChromaDB, Pinecone, FAISS, etc.)
- Language models
- Application code and configs

---

## Prerequisites

- Python 3.8+
- KitOps CLI ([installation](https://kitops.ml/docs/cli/installation.html))
- OCI registry access (Docker Hub, Jozu Hub, GHCR, etc.)

---

## Project Structure

```
rag-project/
├── Kitfile
├── models/
│   ├── embeddings/
│   └── llm/
├── vectordb/
├── datasets/
├── code/
│   ├── ingest.py
│   ├── query.py
│   └── requirements.txt
└── configs/
```

---

## Instructions

### 1. Download Models

```bash
# Embedding model
python3 -c "
from sentence_transformers import SentenceTransformer
model = SentenceTransformer('all-MiniLM-L6-v2')
model.save('models/embeddings/all-MiniLM-L6-v2')
"

# LLM (quantized)
huggingface-cli download TheBloke/Mistral-7B-Instruct-v0.2-GGUF \
  mistral-7b-instruct-v0.2.Q4_K_M.gguf \
  --local-dir models/llm/mistral-7b
```

### 2. Build Vector Database

```python
# code/ingest.py
import chromadb
from sentence_transformers import SentenceTransformer

model = SentenceTransformer('../models/embeddings/all-MiniLM-L6-v2')
client = chromadb.PersistentClient(path="../vectordb/chroma_db")
collection = client.get_or_create_collection("knowledge_base")

# Load documents, generate embeddings, store
# ... your logic here ...
```

Run this before packing to populate the vector database.

### 3. Create Kitfile

```yaml
manifestVersion: 1.0

package:
  name: my-rag-system
  version: 1.0.0

model:
  - name: embedding-model
    path: ./models/embeddings/all-MiniLM-L6-v2
    framework: sentence-transformers
    
  - name: llm-model
    path: ./models/llm/mistral-7b
    framework: llama-cpp

code:
  - path: ./code

datasets:
  - name: knowledge-base
    path: ./datasets/docs

vectordb:
  - name: chroma-db
    path: ./vectordb/chroma_db

configs:
  - path: ./configs/retrieval_config.yaml
```

### 4. Pack and Push

```bash
# Pack
kit pack . -t myregistry.com/username/my-rag:1.0.0

# Push to any OCI registry
kit push myregistry.com/username/my-rag:1.0.0          # Docker Hub
kit push jozu.ml/username/my-rag:1.0.0                 # Jozu Hub
kit push ghcr.io/username/my-rag:1.0.0                 # GitHub
```

### 5. Deploy

```bash
# Pull
kit pull myregistry.com/username/my-rag:1.0.0

# Unpack
kit unpack myregistry.com/username/my-rag:1.0.0 -d ./deploy

# Run
cd deploy/code
pip install -r requirements.txt
python app.py
```

---

## Example Query Pipeline

```python
# code/query.py
import chromadb
from sentence_transformers import SentenceTransformer
from llama_cpp import Llama

class RAGPipeline:
    def __init__(self):
        self.embedding_model = SentenceTransformer('../models/embeddings/all-MiniLM-L6-v2')
        client = chromadb.PersistentClient(path="../vectordb/chroma_db")
        self.collection = client.get_collection("knowledge_base")
        self.llm = Llama(model_path="../models/llm/mistral-7b/mistral-7b-instruct-v0.2.Q4_K_M.gguf")
    
    def query(self, question, top_k=5):
        # Get embeddings and search
        query_embedding = self.embedding_model.encode([question])
        results = self.collection.query(query_embeddings=query_embedding.tolist(), n_results=top_k)
        
        # Generate with context
        context = "\n\n".join(results['documents'][0])
        prompt = f"Context:\n{context}\n\nQuestion: {question}\n\nAnswer:"
        response = self.llm(prompt, max_tokens=512)
        
        return response['choices'][0]['text']
```

---

## Tips

- Use quantized models (GGUF, ONNX) to reduce size
- Build vector indexes before packing
- Version your datasets with your models
- Tag by environment: `prod`, `staging`, `dev`

- Troubleshooting
IssueFixSlow uploadsUse quantized models; compress databasesAuth failuresRun kit login <registry>Missing filesCheck paths in Kitfile match your directories

---

