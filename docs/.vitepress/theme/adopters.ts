export type Adopter = {
  /** Organization name, as it should be displayed. */
  name: string
  /** Optional logo in `src/public/images/adopters/`. Falls back to the name when absent. */
  logo?: string
  /** Rendered width/height of the logo, matching its intrinsic aspect ratio. */
  logoWidth?: number
  logoHeight?: number
  /** Link to the organization's site. */
  url?: string
}

export type CaseStudyAdopter = Adopter & {
  /** How the organization uses KitOps — a sentence or two. */
  usage: string
  /** Link to the published case study. Required — that is what puts them in this section. */
  caseStudyUrl: string
  /** Optional pull quote from someone at the organization. */
  quote?: {
    text: string
    /** A person's name, or a role when the source only attributes by role. */
    author: string
    title?: string
  }
}

/**
 * Adopters with a published case study. Rendered as cards with a
 * "Read the full case study" link.
 *
 * Logos go in `docs/src/public/images/adopters/`.
 */
export const caseStudyAdopters: CaseStudyAdopter[] = [
  {
    name: 'Arlequin AI',
    url: 'https://arlequin.ai',
    caseStudyUrl: 'https://jozu.com/arlequin-kitops-case-study/',
    usage: 'A topological deep learning research lab that standardized on ModelKits as its packaging and versioning standard — bundling the model card, MLflow experiment pointers, LakeFS dataset references, and configuration into a single OCI artifact it promotes from dev to staging to production.',
    quote: {
      text: 'It felt like discovering containerization. It feels like Docker for models.',
      author: 'Aymeric Alixe',
      title: 'MLOps/DevOps Engineer'
    }
  },
  {
    name: 'BDR',
    caseStudyUrl: 'https://jozu.com/case-study/',
    usage: 'An IT security provider delivering AI/ML projects for government agencies and security-conscious organizations, using ModelKits to package datasets and model milestones across its development lifecycle and to automate Kubernetes deployments.',
    quote: {
      text: 'We have MLflow, but executives need a tamper proof and auditable solution. We were considering customizing MLflow to add secure storage but now we don’t have to!',
      author: 'MLOps Tech Lead'
    }
  }
]

/**
 * Adopters without a case study. Rendered as a simple list.
 *
 * Only add organizations that have given permission to be listed.
 *
 * Example entry:
 *
 *   { name: 'Example Corp', url: 'https://example.com' }
 */
export const listedAdopters: Adopter[] = [
  { name: 'Oak Ridge National Laboratory', url: 'https://www.ornl.gov' },
  { name: 'Arlequin AI', url: 'https://arlequin.ai' },
  { name: 'Pacific Northwest National Laboratory', url: 'https://www.pnnl.gov' },
  { name: 'De Sammensluttede Vognmænd (DSV)', url: 'https://www.dsv.com' },
  { name: 'Jozu', url: 'https://jozu.com' }
]
