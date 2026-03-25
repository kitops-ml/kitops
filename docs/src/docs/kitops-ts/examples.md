# Examples

The [`examples/`](https://github.com/kitops-ml/kitops-ts/tree/main/examples) directory contains runnable scripts for common workflows. Each example is a standalone Node.js script you can copy and adapt.

---

## Build a Kitfile programmatically

**File:** [`examples/build-kitfile.js`](https://github.com/kitops-ml/kitops-ts/blob/main/examples/build-kitfile.js)

Useful when you want to generate the Kitfile from existing metadata (e.g. a training config, a model card, or `package.json`) rather than maintaining it as a static file.

```javascript
import { writeFile } from 'fs/promises';
import { stringify as toYaml } from 'yaml';
import { pack } from '@kitops/kitops-ts';

function buildKitfile({ manifestVersion = '1.0.0', ...rest }) {
  return toYaml({ manifestVersion, ...rest }, { lineWidth: 0 });
}

const tag = `${process.env.REGISTRY ?? 'registry.example.com'}/org/sentiment-model:v1.0.0`;

const yaml = buildKitfile({
  package: {
    name: 'sentiment-model',
    version: '1.0.0',
    description: 'Fine-tuned sentiment classification model',
    authors: ['ML Team'],
    license: 'Apache-2.0',
  },
  model: {
    name: 'sentiment-classifier',
    path: './model/weights.pt',
    format: 'pytorch',
  },
  datasets: [
    { name: 'training-set', path: './data/train.csv', license: 'CC-BY-4.0' },
    { name: 'validation-set', path: './data/val.csv', license: 'CC-BY-4.0' },
  ],
  docs: [
    { path: './README.md', description: 'Model card' },
  ],
});

await writeFile('./Kitfile', yaml);
await pack('.', { tag });
```

**Run:**
```bash
node examples/build-kitfile.js
```

---

## CI/CD pipeline

**File:** [`examples/ci-cd.js`](https://github.com/kitops-ml/kitops-ts/blob/main/examples/ci-cd.js)

Pack and publish a ModelKit from a CI environment. Reads credentials and version metadata from environment variables — nothing sensitive ends up in the script or the process list.

```javascript
import { login, pack, push, logout } from '@kitops/kitops-ts';

function requireEnv(name) {
  const value = process.env[name];
  if (!value) throw new Error(`Missing required environment variable: ${name}`);
  return value;
}

const registry = requireEnv('REGISTRY');
const user = requireEnv('REGISTRY_USER');
const pass = requireEnv('REGISTRY_PASS');
const version = process.env.MODEL_VERSION ?? 'latest';

const ref = `${registry}/org/my-model:${version}`;

await login(registry, user, pass);

try {
  await pack('.', { tag: ref });
  await push(ref);
} finally {
  // Always log out so credentials aren't cached on shared runners.
  await logout(registry);
}
```

**GitHub Actions usage:**

```yaml
- name: Publish ModelKit
  run: node examples/ci-cd.js
  env:
    REGISTRY: ghcr.io
    REGISTRY_USER: ${{ github.actor }}
    REGISTRY_PASS: ${{ secrets.GITHUB_TOKEN }}
    MODEL_VERSION: ${{ github.ref_name }}
```

---

## Tag and promote a release

**File:** [`examples/tag-and-promote.js`](https://github.com/kitops-ml/kitops-ts/blob/main/examples/tag-and-promote.js)

A common release promotion workflow:

1. CI publishes a release candidate (`rc1`) to a staging registry.
2. After QA sign-off, the same digest is promoted to production by re-tagging — no re-packing, no risk of bit-rot between environments.

```javascript
import { login, pull, tag, push, inspect, logout } from '@kitops/kitops-ts';

const stagingRegistry = process.env.STAGING_REGISTRY ?? 'staging.example.com';
const prodRegistry = process.env.PROD_REGISTRY ?? 'registry.example.com';

const rcRef = `${stagingRegistry}/org/my-model:rc1`;
const v1Ref = `${prodRegistry}/org/my-model:v1.0.0`;
const latestRef = `${prodRegistry}/org/my-model:latest`;

// Inspect the candidate before promoting
const metadata = await inspect('org/my-model', 'rc1', { remote: true });
console.log(`Promoting digest: ${metadata.digest}`);

// Pull the RC so we can tag it locally
await login(stagingRegistry, user, pass);
await pull(rcRef);
await logout(stagingRegistry);

// Create two local aliases
await tag(rcRef, v1Ref);
await tag(rcRef, latestRef);

// Push both tags to production
await login(prodRegistry, user, pass);
await push(v1Ref);
await push(latestRef);
await logout(prodRegistry);
```

---

## LLM fine-tuning dataset packaging

**File:** [`examples/llm-prompts-ci.js`](https://github.com/kitops-ml/kitops-ts/blob/main/examples/llm-prompts-ci.js)

Package a prompt dataset for LLM fine-tuning. The `prompts` layer is designed for versioning prompt templates, instruction sets, and RLHF preference data alongside model weights and training code — keeping the entire fine-tuning run fully reproducible.

```javascript
import { writeFile, access, mkdir } from 'fs/promises';
import { stringify as toYaml } from 'yaml';
import { login, pack, push, logout } from '@kitops/kitops-ts';

function requireEnv(name) {
  const value = process.env[name];
  if (!value) throw new Error(`Missing required environment variable: ${name}`);
  return value;
}

const registry = process.env.REGISTRY ?? 'registry.example.com';
const user = requireEnv('REGISTRY_USER');
const pass = requireEnv('REGISTRY_PASS');
const version = process.env.MODEL_VERSION ?? 'latest';

const ref = `${registry}/org/llm-finetune:v${version}`;
const workdir = './fine-tune';

// Ensure the working directory exists before writing into it.
await mkdir(workdir, { recursive: true });

// Generate a Kitfile if one isn't already committed.
const kitfilePath = `${workdir}/Kitfile`;
const kitfileExists = await access(kitfilePath).then(() => true).catch(() => false);

if (!kitfileExists) {
  const kitfile = {
    manifestVersion: '1.0.0',
    package: { name: 'llm-finetune', version },
    model: { name: 'base-checkpoint', path: './checkpoints' },
    prompts: [
      { path: './prompts/system.txt', description: 'System prompt template' },
      { path: './prompts/instructions.jsonl', description: 'Instruction-following examples' },
      { path: './prompts/preferences.jsonl', description: 'RLHF preference pairs' },
    ],
    datasets: [
      { name: 'training-corpus', path: './data/train.jsonl' },
    ],
    code: [{ path: './train.py', description: 'Fine-tuning entry point' }],
  };

  await writeFile(kitfilePath, toYaml(kitfile, { lineWidth: 0 }));
  console.log('Kitfile generated.');
}

await login(registry, user, pass);

try {
  await pack(workdir, { tag: ref });
  await push(ref);
} finally {
  await logout(registry);
}
```

**Run:**
```bash
MODEL_VERSION=0.2.0 REGISTRY_USER=... REGISTRY_PASS=... node examples/llm-prompts-ci.js
```
