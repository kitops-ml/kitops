<script setup lang="ts">
import { isClient } from '@vueuse/core'
import { ref, computed, onMounted } from 'vue'
import { Vue3Marquee } from 'vue3-marquee'
import Accordion from './Accordion.vue'
import vGaTrack from '../directives/ga'
import axios from 'axios'
import VueTurnstile from 'vue-turnstile'

const error = ref('')
const email = ref('')
const token = ref('')
const favoriteDevOpsTool = ref('')
const isBusy = ref(false)
const isSubscribed = ref(false)
const isSuccess = ref(false)
const isProd = import.meta.env.PROD

const activeQuote = ref(0)
const quotes = [
  {
    text: `When you start producing many models, you need to have a way to store them, track where they are deployed,
      on what data they were trained on, which version of the runtime (if any) they are compatible with, and so on.
      It's inevitable for any company dealing with data science projects.`,
    authorName: 'Marco',
    authorTitle: 'Data Scientist',
    authorCompany: 'Orobix',
  },
  {
    text: `The whole idea [of KitOps] is awesome... the development process used to be a struggle.
      By creating a middle ground between Devs and Data Scientists, you brought both sides together.`,
    authorName: 'Paulo',
    authorTitle: 'Data Engineer',
    authorCompany: 'Zwift',
  },
  {
    text: `Currently we have a system that resembles half-MLOps but lacks data versioning tied to model versioning and
      configurations, and makes it difficult to deploy and keep track [of AI projects]`,
    authorName: 'Majid',
    authorTitle: 'Data Scientist',
    authorCompany: 'Siemens',
  },
  {
    text: `As we have an open source first policy we are forced to go with tools like MLflow.
      But I really don’t like their 'opinionated' approach to how practitioners should log artifacts.
      I like the idea of using existing best practices in the MLOps space.`,
    authorName: 'Niklas',
    authorTitle: 'MLOps Engineer',
    authorCompany: 'Bundesdruckerei',
  },
]

const quotesOffsetMobile = computed(() => {
  if (isClient) {
    return `translateX(${(activeQuote.value * window.innerWidth - (activeQuote.value * 6)) * -1}px)`
  }

  return `translateX(0)`
})

// current quote * card width + margin + half card)
const quotesOffsetDesktop = computed(() => `translateX(${((activeQuote.value * 664 + 16) + 332) * -1}px)`)

const subscribeToNewsletter = async () => {
  isBusy.value = true

  // Validate the captcha token with the server
  try {
    await axios.post('https://newsprxy.gorkem.workers.dev/', {
      email: email.value,
      userGroup: 'KitOps',
      formName: 'KitOps-Community',
      favoriteDevOpsTool: favoriteDevOpsTool.value
    }, {
      headers: {
        'cf-turnstile-response': token.value,
        'Content-Type': 'application/x-www-form-urlencoded',
        'Expect': '',
      }
    })

    isSuccess.value = true
    localStorage.setItem('subscribed', 'true')
  }
  catch(err: any) {
    error.value = err.response?.data?.errors?.flatMap((e: Error) => e.message)[0] || 'An unknown error occurred'
  }
  finally {
    isBusy.value = false
  }
}

onMounted(() => {
  isSubscribed.value = localStorage.getItem('subscribed') === 'true'

  setTimeout(() => {
    // check if there's an anchor link in the url and if so, scroll to that element id
    if (location.hash) {
      const el = document.querySelector(location.hash)
      if (el && el instanceof HTMLElement) {
        el.scrollIntoView({
          behavior:'smooth',
          block: 'center',
        })
      }
    }
  }, 100)
})
</script>

<template>
<div class="mt-32 md:mt-40  px-6 md:px-12 text-center content-container">
  <p class="h4 !font-normal !text-off-white">Simple, secure, and reproducible packaging for AI/ML projects</p>
  <h1 class="mt-4">The missing link in your AI pipeline</h1>

  <div class="flex flex-col lg:flex-row justify-center items-center gap-10 mt-10 md:mt-14 xl:mt-22">
    <div class="relative inline-block w-full max-w-48 lg:w-auto text-center lg:text-left group">
      <a href="/docs/cli/installation/#%F0%9F%AA%9F-windows-install" class="kit-button flex justify-center lg:justify-start items-center gap-2">
        <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5 fill-white" viewBox="0 0 24 24" fill="currentColor">
          <path d="M3 19H21V21H3V19ZM13 13.1716L19.0711 7.1005L20.4853 8.51472L12 17L3.51472 8.51472L4.92893 7.1005L11 13.1716V2H13V13.1716Z"></path>
        </svg>
        Windows
      </a>
      <div class="absolute right-0 pt-2 whitespace-nowrap min-w-52 hidden lg:group-hover:block lg:group-focus-within:block">
        <div class="py-1 bg-night border border-off-white mt-2" role="menu" aria-orientation="vertical" aria-labelledby="options-menu">
          <a href="https://github.com/kitops-ml/kitops/releases/latest/download/kitops-windows-x86_64.zip"
            target="_blank" class="block px-4 py-2 text-white hocus:text-night font-bold hocus:bg-gold" role="menuitem">
            Intel / AMD, 64-bit
          </a>
          <a href="https://github.com/kitops-ml/kitops/releases/latest/download/kitops-windows-arm64.zip"
            target="_blank" class="block px-4 py-2 text-white hocus:text-night font-bold hocus:bg-gold" role="menuitem">
            ARM 64-bit
          </a>
          <a href="https://github.com/kitops-ml/kitops/releases/latest/download/kitops-windows-i386.zip"
            target="_blank" class="block px-4 py-2 text-white hocus:text-night font-bold hocus:bg-gold" role="menuitem">
            Intel / AMD, 32-bit
          </a>
        </div>
      </div>
    </div>

    <div class="relative inline-block w-full max-w-48 lg:w-auto text-center lg:text-left group">
      <a href="/docs/cli/installation/#%F0%9F%8D%8E-macos-install" class="kit-button flex justify-center lg:justify-start items-center gap-2">
        <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5 fill-white" viewBox="0 0 24 24" fill="currentColor">
          <path d="M3 19H21V21H3V19ZM13 13.1716L19.0711 7.1005L20.4853 8.51472L12 17L3.51472 8.51472L4.92893 7.1005L11 13.1716V2H13V13.1716Z"></path>
        </svg>
        Mac
      </a>
      <div class="absolute right-0 pt-2 whitespace-nowrap min-w-52 hidden lg:group-hover:block lg:group-focus-within:block">
        <div class="py-1 bg-night border border-off-white mt-2" role="menu" aria-orientation="vertical" aria-labelledby="options-menu">
          <a href="https://github.com/kitops-ml/kitops/releases/latest/download/kitops-darwin-arm64.zip"
            target="_blank" class="block px-4 py-2 text-white hocus:text-night font-bold hocus:bg-gold" role="menuitem">
            Apple Silicon / ARM64
          </a>
          <a href="https://github.com/kitops-ml/kitops/releases/latest/download/kitops-darwin-x86_64.zip"
            target="_blank" class="block px-4 py-2 text-white hocus:text-night font-bold hocus:bg-gold" role="menuitem">
            Intel / x86_64
          </a>
        </div>
      </div>
    </div>

    <div class="relative inline-block w-full max-w-48 lg:w-auto text-center lg:text-left group">
      <a href="/docs/cli/installation/#%F0%9F%90%A7-linux-install" class="kit-button flex justify-center lg:justify-start items-center gap-2">
        <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5 fill-white" viewBox="0 0 24 24" fill="currentColor">
          <path d="M3 19H21V21H3V19ZM13 13.1716L19.0711 7.1005L20.4853 8.51472L12 17L3.51472 8.51472L4.92893 7.1005L11 13.1716V2H13V13.1716Z"></path>
        </svg>
        Linux
      </a>
      <div class="absolute right-0 pt-2 whitespace-nowrap min-w-52 hidden lg:group-hover:block lg:group-focus-within:block">
        <div class="py-1 bg-night border border-off-white mt-2" role="menu" aria-orientation="vertical" aria-labelledby="options-menu">
          <a href="https://github.com/kitops-ml/kitops/releases/latest/download/kitops-linux-x86_64.tar.gz"
            target="_blank" class="block px-4 py-2 text-white hocus:text-night font-bold hocus:bg-gold" role="menuitem">
            Intel / AMD, 64-bit
          </a>
          <a href="https://github.com/kitops-ml/kitops/releases/latest/download/kitops-linux-arm64.tar.gz"
            target="_blank" class="block px-4 py-2 text-white hocus:text-night font-bold hocus:bg-gold" role="menuitem">
            ARM 64-bit
          </a>
          <a href="https://github.com/kitops-ml/kitops/releases/latest/download/kitops-linux-i386.tar.gz"
            target="_blank" class="block px-4 py-2 text-white hocus:text-night font-bold hocus:bg-gold" role="menuitem">
            Intel / AMD, 32-bit
          </a>
        </div>
      </div>
    </div>
  </div>
</div>

<div id="howdoesitwork" class="mt-32 md:mt-40 xl:mt-60 px-6 md:px-12 text-center max-w-[1152px] content-container">
  <h2>What is Kitops<span class="font-heading font-extralight">?</span></h2>
  <div class="p1 mt-8 mx-8">
    KitOps is an open source DevOps tool that packages and versions your AI/ML model, datasets, code, and configuration into a reproducible artifact called a ModelKit. ModelKits are built on existing standards, ensuring compatibility with the tools your data scientists and developers already use.
  </div>

  <iframe width="1050" class="border-2 border-black max-w-full aspect-video mt-22 mx-auto" src="https://www.youtube.com/embed/iK9mnU0prRU?si=ommsLD32Kjj4RUMu&amp;autoplay=0" title="YouTube video player" frameborder="0" allow="accelerometer; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" referrerpolicy="strict-origin-when-cross-origin" allowfullscreen></iframe>
</div>

<div v-if="!isSubscribed" class="mt-32 md:mt-40 xl:mt-60 px-6 md:px-12 content-container" id="join">
  <h2 class="text-center">stay informed About Kitops</h2>

  <div class="text-center max-w-[600px] mx-auto mt-12">
    <template v-if="!isSuccess">
      <form @submit.prevent="subscribeToNewsletter" class="flex flex-col md:flex-row gap-10 lg:gap-4">
        <input required
          :disabled="isBusy"
          id="email"
          type="email"
          pattern="^[a-zA-Z0-9]+([._+\-][a-zA-Z0-9]+)*@[a-zA-Z0-9\-]+\.[a-zA-Z]{2,}$"
          name="email"
          placeholder="you@example.com"
          class="input"
          v-model="email"
          style="border: 1px solid var(--color-off-white)" />

        <input
          type="text"
          id="favoriteDevOpsTool"
          placeholder="What's your favorite devops tool?"
          name="favoriteDevOpsTool"
          v-model="favoriteDevOpsTool"
          class="hidden" />

        <button type="submit" :disabled="isBusy" class="kit-button kit-button-gold text-center mx-auto">
          JOIN THE LIST
        </button>
      </form>

      <div v-if="isProd" class="mt-10">
        <vue-turnstile site-key="0x4AAAAAAA1WT4LYaVrBtAD7" v-model="token" />
      </div>

      <p v-if="error" class="text-red-500 mt-6">{{ error }}</p>
    </template>

    <template v-else>
      <p class="mt-12">You are now subscribed to the newsletter.</p>
    </template>
  </div>
</div>

<div id="whykitops" class="mt-32 md:mt-40 xl:mt-60 px-6 md:px-12 content-container">
  <h2 class="text-center">Why Kit<span class="font-heading font-extralight">?</span></h2>

  <div class="mt-10 md:mt-14 xl:mt-22 grid grid-cols-1 md:grid-cols-2 gap-4 md:gap-x-4 md:gap-y-[4.5rem] xl:gap-22 max-w-[47.5rem] mx-auto">
    <div class="h4 !text-gold">Model handoffs are hard.</div>
    <div class="p2 space-y-4">
      <p>Moving a model from a Jupyter notebook to an ML tool or development server, then to a production server like Kubernetes is difficult because each tool uses its own packaging mechanism, and requires engineers to repackage the model multiple times. This slows down development and introduces risk.</p>
      <p>KitOps is an open source DevOps project built to standardize packaging, reproduction, deployment, and tracking of AI / ML models, so it can be run anywhere, just like application code</p>
      <p>KitOps solves multiple problems:</p>
    </div>

    <div class="h4 !text-gold xs:mt-12">Model traceability and reproducibility–</div>
    <div class="p2 space-y-4">
      <p>Unlike Dockerfiles, Kitfiles are a modular package - pull just a part of the ModelKit, like the model or dataset, or pull the whole package with one simple command.</p>
      <p>Storing ModelKits in your organization’s container registry provides a history of meaningful state changes for auditing. ModelKits are immutable so are perfect for a secure bill-of-materials (SBOM) initiative.</p>
    </div>

    <div class="h4 !text-gold xs:mt-12">Collaboration–</div>
    <div class="p2 space-y-4">
      <p>By building ModelKits on industry standards, anyone (not just data scientists) can participate in the model development lifecycle whether they’re integrating models with their application, experimenting with them locally, or deploying them to production.</p>
      <p>ModelKits can be stored in your existing container registry and work with the tools your team is already using, so you can use the same deployment pipelines and endpoints you’ve hardened with your application development process.</p>
    </div>

  </div>
</div>

<div class="mt-32 md:mt-40 xl:mt-60 px-6 md:px-12 content-container">
  <h2 class="text-center">Get stArted</h2>

  <div class="kit-cards mt-22 min-h-[32.5rem]">
    <div class="kit-card flex flex-col">
      <div class="h4 font-bold !text-salmon">1</div>
      <div class="mt-8 flex flex-col flex-1 justify-between">
        <p class="p2">Download and install Kit CLI.</p>
        <a href="/docs/cli/installation/" v-ga-track="{ category: 'button', label: 'install', location: 'get started' }" class="kit-button kit-button-salmon md:w-fit mt-6">Install the CLI</a>
      </div>
    </div>

    <div class="kit-card flex flex-col">
      <div class="h4 font-bold !text-cornflower">2</div>
      <div class="mt-8 flex flex-col flex-1 justify-between">
        <p class="p2">Create a simple manifest file called a Kitfile with your model, dataset and code. Then build and push the ModelKit to a registry for sharing.</p>
        <a href="/docs/kitfile/kf-overview/" class="kit-button kit-button-cornflower md:w-fit mt-6">LEARN MORE</a>
      </div>
    </div>

    <div class="kit-card flex flex-col">
      <div class="h4 font-bold !text-gold">3</div>
      <div class="mt-8 flex flex-col flex-1 justify-between">
        <p class="p2">Pull the ModelKit into your pipeline, or use <span class="text-gold">kit dev</span> to start working with the model locally.</p>
        <a href="/docs/use-cases/" class="kit-button md:w-fit mt-6">USE CASES</a>
      </div>
    </div>
  </div>
</div>

<div class="mt-32 md:mt-40 xl:mt-60 px-6 md:px-12 content-container">
  <h2 class="text-center">Key feAtuRes</h2>
  <p class="p2 text-center mt-16">Visit our <a href="https://github.com/kitops-ml/kitops" target="_blank" class="underline">GitHub repo</a> for a list of all features and our roadmap.</p>

  <ol class="grid grid-cols-1 md:grid-cols-2 gap-x-22 gap-y-16 mt-16 max-w-[960px] mx-auto p1">
    <li>
      <div class="text-off-white">🎁 Standards-based package</div>
      <p class="p2 mb-4 text-gray-06">A ModelKit package includes models, datasets, configurations, and code in an OCI artifact. Add as much or as little as your project needs.</p>
      <a class="text-off-white font-bold flex items-center gap-2 text-base" href="/docs/modelkit/compatibility/">
        LEARN MORE
        <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none">
          <path d="M16.1716 10.9999L10.8076 5.63589L12.2218 4.22168L20 11.9999L12.2218 19.778L10.8076 18.3638L16.1716 12.9999H4V10.9999H16.1716Z" fill="#ECECEC"/>
        </svg>
      </a>
    </li>
    <li>
      <div class="text-off-white">🔒 Tamper-proof</div>
      <p class="p2 mb-4 text-gray-06">Each ModelKit package is immutable and includes a SHA digest for itself, and every artifact it holds.</p>
      <a class="text-off-white font-bold flex items-center gap-2 text-base" href="/docs/modelkit/spec/">
        LEARN MORE
        <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none">
          <path d="M16.1716 10.9999L10.8076 5.63589L12.2218 4.22168L20 11.9999L12.2218 19.778L10.8076 18.3638L16.1716 12.9999H4V10.9999H16.1716Z" fill="#ECECEC"/>
        </svg>
      </a>
    </li>
    <li>
      <div class="text-off-white">🏭 Tags and versions</div>
      <p class="p2 mb-4 text-gray-06">Each ModelKit is tagged and versioned so everyone knows which dataset and model work together.</p>
      <a class="text-off-white font-bold flex items-center gap-2 text-base" href="/docs/why-kitops/">
        LEARN MORE
        <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none">
          <path d="M16.1716 10.9999L10.8076 5.63589L12.2218 4.22168L20 11.9999L12.2218 19.778L10.8076 18.3638L16.1716 12.9999H4V10.9999H16.1716Z" fill="#ECECEC"/>
        </svg>
      </a>
    </li>
    <li>
      <div class="text-off-white">🤗 Use with LLM, ML, or AI projects</div>
      <p class="p2 mb-4 text-gray-06">ModelKits can be used with any AI, ML, or LLM project - even multi-modal models.</p>
      <a class="text-off-white font-bold flex items-center gap-2 text-base" href="/docs/cli/cli-reference/#kit-tag">
        LEARN MORE
        <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none">
          <path d="M16.1716 10.9999L10.8076 5.63589L12.2218 4.22168L20 11.9999L12.2218 19.778L10.8076 18.3638L16.1716 12.9999H4V10.9999H16.1716Z" fill="#ECECEC"/>
        </svg>
      </a>
    </li>
    <li>
      <div class="text-off-white">🤖 Automation for CI/CD</div>
      <p class="p2 mb-4 text-gray-06">Pack or unpack a ModelKit locally or as part of your CI/CD workflow for testing, integration, or deployment.</p>
      <a class="text-off-white font-bold flex items-center gap-2 text-base" href="https://github.com/marketplace/actions/setup-kit-cli" target="_blank">
        LEARN MORE
        <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none">
          <path d="M16.1716 10.9999L10.8076 5.63589L12.2218 4.22168L20 11.9999L12.2218 19.778L10.8076 18.3638L16.1716 12.9999H4V10.9999H16.1716Z" fill="#ECECEC"/>
        </svg>
      </a>
    </li>
    <li>
      <div class="text-off-white">🏃‍♂️‍ Local dev mode</div>
      <p class="p2 mb-4 text-gray-06">Kit's Dev Mode lets your run an LLM locally, configure it, and prompt/chat with it instantly</p>
      <a class="text-off-white font-bold flex items-center gap-2 text-base" href="/docs/why-kitops/">
        LEARN MORE
        <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none">
          <path d="M16.1716 10.9999L10.8076 5.63589L12.2218 4.22168L20 11.9999L12.2218 19.778L10.8076 18.3638L16.1716 12.9999H4V10.9999H16.1716Z" fill="#ECECEC"/>
        </svg>
      </a>
    </li>
  </ol>

  <div class="mt-36 p2 text-center">
    <a href="/docs/use-cases/" class="kit-button mt-22">SEE THE KIT WORKFLOW</a>
  </div>
</div>

<div class="mt-32 md:mt-40 xl:mt-60 px-6 md:px-10 lg:px-0 w-full max-w-[960px] mx-auto">
  <h2 class="text-center">tRy running youR fiRst Model with Kitops</h2>
  <p class="p1 text-center mt-16">AI projects are more than just a model, you need a codebase, dataset, documentation too.</p>
  <p class="p1 text-center mt-4">Our quickstart ModelKits have everything you need in one easy to find place.</p>

  <ul class="grid grid-cols-1 md:grid-cols-2 gap-6 my-22">
    <li class="border border-gray-02 hover:border-gold transition-colors">
      <a href="https://jozu.ml/repository/jozu/llama3.1-8b" class="flex items-center justify-between px-14 py-6">
        Meta Llama3.1
        <svg xmlns="http://www.w3.org/2000/svg" xml:space="preserve" width="32" height="32" viewBox="0 0 209 297">
          <path fill="#fff" d="M22 297c-1-10-3-19-1-29 2-8 8-15 8-22-1-13-13-23-12-37 1-12 11-24 11-35-1-6-6-11-9-16-4-11-4-23 0-33 6-12 17-22 29-25 5-2 12 1 16-2 7-3 10-14 17-18 16-12 36-10 50 3 5 4 8 12 13 15 4 2 10 0 14 1 12 2 23 11 29 22s7 24 2 37c-2 5-8 10-8 16 0 11 10 23 10 35 1 14-10 24-11 37 0 7 6 14 8 22 2 10-1 19-1 29 4 0 10 1 13-1 5-2 4-9 4-13 0-8 0-16-2-23-1-5-5-10-3-15 5-16 9-28 7-44 0-9-6-18-6-26 0-5 4-11 6-16a64 64 0 0 0-10-55c-4-4-9-7-11-12-1-6 1-14 1-19 0-14-1-28-9-41-7-11-22-14-31-3-8 10-9 23-12 35-21-9-38-10-59 0-3-12-4-25-12-35-9-11-25-8-31 3-8 13-9 27-9 41 0 5 2 13 0 19-1 5-7 8-10 12-5 6-8 14-10 21-3 12-3 23 0 35 2 5 6 10 6 15 0 8-6 17-6 26-2 16 2 28 7 44 2 5-2 10-3 15-3 7-2 15-2 23 0 4-1 11 3 13s10 1 14 1M39 86c-1-13-1-25 3-37 1-3 2-11 6-11s6 7 7 10c3 7 8 24 4 32-2 5-15 4-20 6m131 0c-6-2-18-1-20-6-4-8 1-25 4-32 1-3 3-10 6-10 5 0 6 8 7 11 4 12 4 24 3 37m-73 51c-25 4-48 33-27 55 12 12 32 13 48 11 22-4 37-28 24-48a45 45 0 0 0-45-18m-50 3c-11 6-4 25 8 19s4-25-8-19m105 0c-11 6 0 25 11 19 10-6 1-25-11-19m-53 8c13-2 27 2 34 13 9 14-3 29-18 31-11 2-26 2-35-6-18-15 2-36 19-38m-1 14c-6 2 2 21 8 19 3-1 3-6 4-8 3-9-1-14-12-11z"/>
        </svg>
      </a>
    </li>
    <li class="border border-gray-02 hover:border-gold transition-colors">
      <a href="https://jozu.ml/repository/jozu/gemma-7b" class="flex items-center justify-between px-14 py-6">
        Google Gemma
        <svg xmlns="http://www.w3.org/2000/svg" width="32" height="32" viewBox="0 0 32 32" fill="none">
          <path d="M21.5623 14.6665L14.4103 7.51452L16.2959 5.62891L26.6668 15.9999L16.2959 26.3707L14.4103 24.4851L21.5623 17.3332H5.3335V14.6665H21.5623Z" fill="#ECECEC"/>
        </svg>
      </a>
    </li>
    <li class="border border-gray-02 hover:border-gold transition-colors">
      <a href="https://jozu.ml/repository/jozu/phi3" class="flex items-center justify-between px-14 py-6">
        Microsoft phi3
        <svg xmlns="http://www.w3.org/2000/svg" width="32" height="32" viewBox="0 0 32 32" fill="none">
          <path d="M21.5623 14.6665L14.4103 7.51452L16.2959 5.62891L26.6668 15.9999L16.2959 26.3707L14.4103 24.4851L21.5623 17.3332H5.3335V14.6665H21.5623Z" fill="#ECECEC"/>
        </svg>
      </a>
    </li>
    <li class="border border-gray-02 hover:border-gold transition-colors">
      <a href="https://jozu.ml/repository/jozu/fine-tuning" class="flex items-center justify-between px-14 py-6">
        Fine-tuning
        <svg xmlns="http://www.w3.org/2000/svg" width="32" height="32" viewBox="0 0 32 32" fill="none">
          <path d="M21.5623 14.6665L14.4103 7.51452L16.2959 5.62891L26.6668 15.9999L16.2959 26.3707L14.4103 24.4851L21.5623 17.3332H5.3335V14.6665H21.5623Z" fill="#ECECEC"/>
        </svg>
      </a>
    </li>
    <li class="border border-gray-02 hover:border-gold transition-colors">
      <a href="https://jozu.ml/repository/jozu/rag-pipeline" class="flex items-center justify-between px-14 py-6">
        Rag pipeline
        <svg xmlns="http://www.w3.org/2000/svg" width="32" height="32" viewBox="0 0 32 32" fill="none">
          <path d="M21.5623 14.6665L14.4103 7.51452L16.2959 5.62891L26.6668 15.9999L16.2959 26.3707L14.4103 24.4851L21.5623 17.3332H5.3335V14.6665H21.5623Z" fill="#ECECEC"/>
        </svg>
      </a>
    </li>
    <li class="border border-gray-02 hover:border-gold transition-colors">
      <a href="https://jozu.ml/repository/jozu/yolo-v10" class="flex items-center justify-between px-14 py-6">
        Object detection
        <svg xmlns="http://www.w3.org/2000/svg" width="32" height="32" viewBox="0 0 32 32" fill="none">
          <path d="M21.5623 14.6665L14.4103 7.51452L16.2959 5.62891L26.6668 15.9999L16.2959 26.3707L14.4103 24.4851L21.5623 17.3332H5.3335V14.6665H21.5623Z" fill="#ECECEC"/>
        </svg>
      </a>
    </li>
  </ul>
</div>

<div class="mt-32 md:mt-40 xl:mt-60 px-6 md:px-12 content-container">
  <div class="kit-cards md:!grid-cols-2 w-fit mx-auto">
    <div class="kit-card max-w-[370px] flex flex-col">
      <h3 class="!text-cornflower">ModelKit</h3>

      <div class="flex-1 mt-8 space-y-4">
        <p class="p2">The ModelKit is an OCI compliant package (like a container, but more fully featured) that contains everything needed to integrate with a model, or deploy it to production.</p>
        <p class="p2">The ModelKit holds the serialized model, dataset, hyperparameters, input / output structure, and validation criteria. Kitfiles define a ModelKit in a modular and easy-to-understand way.</p>
      </div>

      <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 48 48" fill="none" class="mt-10 md:mt-16 xl:mt-20 size-8 md:size-12">
        <path d="M6 16L18.0064 4H39.9956C41.1026 4 42 4.91062 42 5.9836V42.0164C42 43.112 41.1102 44 40.0132 44H7.9868C6.88952 44 6 43.1002 6 41.9864V16ZM20 8V18H10V40H38V8H20Z" class="fill-cornflower" />
      </svg>

      <div>
        <a href="/docs/modelkit/intro/">LEARN MORE</a>
      </div>
    </div>
    <div class="kit-card max-w-[370px] flex flex-col">
      <h3 class="!text-salmon">Kit cli</h3>

      <div class="flex-1 mt-8 space-y-4">
        <p class="p2">The Kit CLI is a command line interface (CLI) that performs actions on ModelKits.</p>
        <p class="p2">You can: build and version ModelKits; push or pull them from a model registry; run them locally with a RESTful API we generate for your model automatically, and deploy them to staging or production.</p>
      </div>

      <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 48 32" fill="none" class="mt-10 md:mt-16 xl:mt-20 size-8 md:size-12">
        <path d="M36.1238 25.0052H54.7501" class="stroke-salmon" stroke-width="6.7732" stroke-linecap="square" stroke-linejoin="round"/>
        <path d="M0 5.25L15.8041 16.8209L0 26.6985" class="stroke-salmon" stroke-width="6.7732" stroke-linecap="square" stroke-linejoin="round"/>
      </svg>

      <div>
        <a href="/docs/cli/installation/">LEARN MORE</a>
      </div>
    </div>
  </div>
</div>

<div class="mt-32 md:mt-40 xl:mt-60 px-6 md:px-12 text-center content-container">
  <h2>WhAt’s suppoRted<span class="font-heading font-extralight">?</span></h2>
  <p class="p1 mt-8 mb-22">Kit was designed to work with the tools your team already uses.</p>

  <div class="space-y-12 relative marquee-gradients">

    <Vue3Marquee>
      <div class="flex justify-center items-center mx-10">
        <img src="/images/logos/jupyter@2x.png" alt="jupyter logo" width="48" height="56" class="opacity-65">
      </div>
      <div class="flex justify-center items-center mx-10">
        <img src="/images/logos/docker@2x.png" alt="docker logo" width="160" height="36" class="opacity-65">
      </div>
      <div class="flex justify-center items-center mx-10">
        <img src="/images/logos/dvc@2x.png" alt="dvc logo" width="48" height="30" class="opacity-65">
      </div>
      <div class="flex justify-center items-center mx-10">
        <img src="/images/logos/huggingface@2x.png" alt="hugging face logo" width="200" height="44" class="opacity-65">
      </div>
      <div class="flex justify-center items-center mx-10">
        <img src="/images/logos/github@2x.png" alt="github logo" width="120" height="31" class="opacity-65">
      </div>
      <div class="flex justify-center items-center mx-10">
        <img src="/images/logos/gitlab@2x.png" alt="gitlab logo" width="110" height="25" class="opacity-65">
      </div>
      <div class="flex justify-center items-center mx-10">
        <img src="/images/logos/jfrog@2x.png" alt="jfrog logo" width="110" height="29" class="opacity-65">
      </div>
      <div class="flex justify-center items-center mx-10">
        <img src="/images/logos/azureml@2x.png" alt="azureml logo" width="35" height="37" class="opacity-65">
      </div>
    </Vue3Marquee>

    <Vue3Marquee
      direction="reverse">
      <div class="flex justify-center items-center mx-10">
        <img src="/images/logos/databricks@2x.png" alt="data bricks logo" width="148" height="25" class="opacity-65">
      </div>
      <div class="flex justify-center items-center mx-10">
        <img src="/images/logos/datarobot@2x.png" alt="data robot logo" width="148" height="19" class="opacity-65">
      </div>
      <div class="flex justify-center items-center mx-10">
        <img src="/images/logos/kubernetes@2x.png" alt="kubernetes logo" width="148" height="27" class="opacity-65">
      </div>
      <div class="flex justify-center items-center mx-10">
        <img src="/images/logos/mlflow@2x.png" alt="mlflow logo" width="90" height="33" class="opacity-65">
      </div>
      <div class="flex justify-center items-center mx-10">
        <img src="/images/logos/nvidia@2x.png" alt="nvidia logo" width="58" height="43" class="opacity-65">
      </div>
      <div class="flex justify-center items-center mx-10">
        <img src="/images/logos/openshift@2x.png" alt="openshift logo" width="48" height="51" class="opacity-65">
      </div>
      <div class="flex justify-center items-center mx-10">
        <img src="/images/logos/tensorflow@2x.png" alt="tensorflow logo" width="148" height="33" class="opacity-65">
      </div>
    </Vue3Marquee>

    <Vue3Marquee>
      <div class="flex justify-center items-center mx-10">
        <img src="/images/logos/amazonsagemaker@2x.png" alt="amazon sage maker logo" width="130" height="40" class="opacity-65">
      </div>
      <div class="flex justify-center items-center mx-10">
        <img src="/images/logos/circleci@2x.png" alt="circle ci logo" width="110" height="32" class="opacity-65">
      </div>
      <div class="flex justify-center items-center mx-10">
        <img src="/images/logos/prefect@2x.png" alt="prefect logo" width="120" height="15" class="opacity-65">
      </div>
      <div class="flex justify-center items-center mx-10">
        <img src="/images/logos/ray@2x.png" alt="ray logo" width="80" height="31" class="opacity-65">
      </div>
      <div class="flex justify-center items-center mx-10">
        <img src="/images/logos/runai@2x.png" alt="runai logo" width="48" height="33" class="opacity-65">
      </div>
      <div class="flex justify-center items-center mx-10">
        <img src="/images/logos/vertexai@2x.png" alt="vertex logo" width="120" height="39" class="opacity-65">
      </div>
      <div class="flex justify-center items-center mx-10">
        <img src="/images/logos/weightsbiases@2x.png" alt="weights & biases logo" width="160" height="25" class="opacity-65">
      </div>
    </Vue3Marquee>
  </div>

  <a href="/docs/modelkit/compatibility/" class="kit-button mt-22">SEE FULL LIST</a>
</div>

<div class="mt-32 md:mt-40 xl:mt-60 overflow-x-hidden">
  <h2 class="text-center">WhAt people sAy About us</h2>

  <div class="my-22 inline-flex items-center gap-0 md:gap-4 relative md:left-1/2 transition quotes-container">
    <button
      v-for="(quote, index) in quotes" :key="index"
      class="quote-bg p-10 pt-24 md:pt-10 md:pl-32 border border-solid border-gray-02 hover:border-white
        flex flex-col justify-between xs:items-center
        text-left md:self-stretch min-h-[344px] lg:basis-auto max-w-[664px] space-y-10 opacity-50 cursor-pointer text-off-white hover:text-white
        min-w-[calc(100vw-3rem)] md:min-w-[664px] mx-6 md:mx-0"
      :class="{ '!opacity-100': activeQuote === index }"
      @click="activeQuote = index">
      <p class="p2">{{ quote.text }}</p>
      <p class="text-gray-06 text-xl">
        <strong>{{ quote.authorName }}</strong>, {{ quote.authorTitle }} @ {{ quote.authorCompany }}
      </p>
    </button>
  </div>

  <nav class="flex gap-4 items-center justify-center">
    <button v-for="(_, index) in quotes" :key="index"
     @click="activeQuote = index">
      <svg xmlns="http://www.w3.org/2000/svg" width="17" height="16" viewBox="0 0 17 16" fill="none"
        class="opacity-20"
        :class="{ '!opacity-100': activeQuote === index }">
        <path d="M0.333496 16V0H17.0002L0.333496 16Z" fill="#FFF"/>
      </svg>
    </button>
  </nav>
</div>

<div class="max-w-3xl mx-auto my-32 md:my-40 lg:my-60 faq-section content-container">
  <h2 class="text-center mb-10 md:mb-14 lg:mb-22">fAq</h2>

  <Accordion content-class="space-y-[1em]">
    <template #title>Are ModelKits a versioning solution or a packaging solution?</template>

    <p class="mt-6">
      ModelKits do both. With a ModelKit, you can package all the parts of your AI project in one shareable asset, and tag them with a version.
      ModelKits were designed for the model development lifecycle, where projects are handed off from data science teams to application teams to deployment teams. Versioning and packaging makes it easy for team members to find the datasets and configurations that map to a specific model version.
      You can <a href="/docs/overview/" class="underline">read more details about KitOps in our overview</a>.
    </p>
  </Accordion>

  <Accordion content-class="space-y-[1em]">
    <template #title>How do I get started with Kit?</template>

    <p class="mt-6">The easiest way to get started is to follow our <a href="/docs/get-started/" class="underline">Quick Start</a>, where you’ll learn how to:</p>

    <ul class="space-y-2 list-disc list-inside">
      <li>Package up a model, notebook, and datasets into a single ModelKit you can use with your existing tools</li>
      <li>Share the ModelKit with others through your public or private registry</li>
      <li>Grab only the assets you need from the ModelKit for testing, integration, local running, or deployment</li>
    </ul>

  </Accordion>

  <Accordion content-class="space-y-[1em]">
    <template #title>Can I see if something changed between ModelKits?</template>

    <p class="mt-6">Yes [choir sings hallelujah], each ModelKit includes SHA digests for the ModelKit and every artifact it holds so you can quickly see if something changed between ModelKit versions. </p>
  </Accordion>

  <Accordion content-class="space-y-[1em]">
    <template #title>What are the benefits of using Kit?</template>

    <p class="mt-6">Increased speed: Teams can work faster with a centralized and versioned package for their AI project coordination. ModelKits eliminate hunting for datasets or code, and make it obvious which datasets and configurations are needed for each model. Handoffs can be automated and executed quickly and with confidence.</p>
    <p>Reduced risk: ModelKits are self-verifying. Both the ModelKit itself and all the artifacts added to it are tamper-proof. Anyone can quickly and easily verify when something may have changed.</p>
    <p>Improved efficiency: Models stored in ModelKits can be run locally for experimentation or application integration, or packaged for deployment with a single command. Any artifact in a ModelKit can be separately pulled saving time and space on local or shared machines. This makes it easy for data scientists, application developers, and DevOps engineers to find and grab the pieces they need to do their job without being overwhelmed with unnecessary files.</p>
  </Accordion>

  <Accordion content-class="space-y-[1em]">
    <template #title>What tools are compatible with Kit?</template>

    <p class="mt-6">ModelKits store their assets as OCI-compatible artifacts. This makes them compatible with nearly every development and deployment tool and registry in use today.</p>
  </Accordion>

  <Accordion content-class="space-y-[1em]">
    <template #title>Where are ModelKits stored?</template>
    <p class="mt-6">ModelKits can be stored in any OCI-compliant registry - for example in a container registry like Docker Hub or Jozu Hub, or your favorite cloud vendor’s container registry, they can even be stored in an artifact repository like Artifactory.</p>
  </Accordion>

  <Accordion content-class="space-y-[1em]">
    <template #title>Is KitOps open source and free to use?</template>

    <p class="mt-6">Yes, it is licensed with the Apache 2.0 license and welcomes all users and contributors. If you’re <a href="https://github.com/kitops-ml/kitops/blob/main/CONTRIBUTING.md" class="underline">interested in contributing</a>, let us know.</p>
  </Accordion>

  <Accordion content-class="space-y-[1em]">
    <template #title>Are ModelKits a replacement for Docker containers?</template>

    <p class="mt-6">No, ModelKits complement containers - in fact, KitOps can take a ModelKit and generate a container for the model automatically. However, not all models should be deployed inside containers - sometimes it’s more efficient and faster to deploy an init container linked to the model for deployment. Datasets may also not need to be in containers - many datasets are easier to read and manipulate for training and validation when they’re not in a container. Finally, each container is still separate so even if you do want to put everything in its own container it’s not clear to people outside the AI project which datasets go with which models and which configurations.</p>
  </Accordion>

  <Accordion content-class="space-y-[1em]">
    <template #title>Why would I use KitOps for versioning instead of Git?</template>

    <p class="mt-6">Models and datasets in AI projects are often 10s or 100s of GB in size. Git was designed to work with many small files that can be easily diff’ed between versions. Git treats models and datasets stored in LFS (large file storage) as atomic blobs and can’t differentiate between versions of them. This makes it both inefficient and dangerous since it’s easy for someone to tamper with the models and datasets in the LFS without Git knowing. Finally, once you use LFS, a clone is no longer guaranteed to be the same as the original repo, because the repo refers to an LFS server that is independent of the clone and can change independently.</p>
  </Accordion>

  <Accordion content-class="space-y-[1em]">
    <template #title>My MLOps tools do versioning, why do I need Kit?</template>

    <p class="mt-6">KitOps is the only standards-based and open source solution for packaging and versioning AI project assets. Popular MLOps tools use proprietary and often closed formats to lock you into their ecosystem. This makes handoffs between MLOps tool users and non-MLOps tool users (like your application development and DevOps teams) unnecessarily hard. The future of MLOps tools is still being written, and it’s likely that many will be acquired or shut down and the cost of moving projects from one proprietary format to another is high. By using the OCI standard that’s already supported by nearly every tool on the planet, ModelKits give you a future-proofed solution for packaging and versioning that is compatible with both your MLOps tools and development / DevOps tools so everyone can collaborate regardless of the tools they use.</p>
  </Accordion>

  <Accordion content-class="space-y-[1em]">
    <template #title>Is enterprise support available for Kit?</template>

    <p class="mt-6">Enterprise support for ModelKits and the Kit CLI is available from <a href="https://www.jozu.com/" class="underline" target="_blank">Jozu</a>.</p>
  </Accordion>
</div>

<div class="mt-32 md:mt-40 xl:mt-60 px-6 md:px-12 content-container">
  <h2 class="text-center">How to Get inVolVeD<span class="font-heading font-extralight">?</span></h2>

  <div class="space-y-6 w-fit mx-auto mt-22">
    <a href="https://discord.gg/Tapeh8agYy" v-ga-track="{ category: 'button', label: 'join the kitops discord', location: 'how to get involved' }" class="border border-gray-02 p-8 md:px-14 md:py-10 flex justify-between gap-8 items-center hover:border-gold transition-colors">
      <div class="p1">Join the KitOps Discord</div>

      <div class="size-8 md:size-12">
        <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 43 32" fill="none">
          <g clip-path="url(#clip0_8_909)">
            <path d="M36.476 2.67995C33.6728 1.40994 30.7134 0.508951 27.6736 0C27.2576 0.735868 26.8812 1.49298 26.546 2.26816C23.308 1.78531 20.0151 1.78531 16.777 2.26816C16.4416 1.49306 16.0653 0.735957 15.6495 0C12.6076 0.513249 9.64634 1.41638 6.84027 2.68659C1.26951 10.8427 -0.240636 18.7962 0.514437 26.6368C3.77681 29.022 7.42835 30.8361 11.3103 32C12.1844 30.8366 12.9579 29.6024 13.6225 28.3105C12.3601 27.8439 11.1417 27.2683 9.98138 26.5903C10.2868 26.3711 10.5854 26.1453 10.874 25.9261C14.2504 27.4974 17.9355 28.312 21.6666 28.312C25.3976 28.312 29.0827 27.4974 32.4591 25.9261C32.751 26.1619 33.0497 26.3877 33.3517 26.5903C32.1891 27.2694 30.9685 27.8461 29.7039 28.3138C30.3677 29.6052 31.1412 30.8383 32.0161 32C35.9014 30.8407 39.5557 29.0276 42.8187 26.6401C43.7046 17.5475 41.3052 9.66708 36.476 2.67995ZM14.5789 21.8149C12.4748 21.8149 10.7364 19.9253 10.7364 17.6007C10.7364 15.276 12.4144 13.3699 14.5722 13.3699C16.7301 13.3699 18.455 15.276 18.4181 17.6007C18.3811 19.9253 16.7233 21.8149 14.5789 21.8149ZM28.7542 21.8149C26.6467 21.8149 24.915 19.9253 24.915 17.6007C24.915 15.276 26.593 13.3699 28.7542 13.3699C30.9154 13.3699 32.6269 15.276 32.5899 17.6007C32.553 19.9253 30.8986 21.8149 28.7542 21.8149Z" fill="white"/>
          </g>
          <defs>
            <clipPath id="clip0_8_909">
              <rect width="42.6667" height="32" fill="white" transform="translate(0.333252)"/>
            </clipPath>
          </defs>
        </svg>
      </div>
    </a>

    <a href="https://github.com/kitops-ml/kitops" v-ga-track="{ category: 'button', label: 'contribute to kit', location: 'how to get involved' }" class="border border-gray-02 p-8 md:px-14 md:py-10 flex justify-between gap-8 items-center hover:border-gold transition-colors">
      <div class="p1">Contribute to Kit</div>

      <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 48 48" fill="none" class="size-8 md:size-12">
        <path d="M32.3432 21.9998L21.6152 11.2718L24.4436 8.44336L40 23.9998L24.4436 39.556L21.6152 36.7276L32.3432 25.9998H8V21.9998H32.3432Z" fill="#ECECEC"/>
      </svg>
    </a>

    <a href="https://github.com/kitops-ml/kitops" v-ga-track="{ category: 'button', label: 'star the repo', location: 'how to get involved' }" class="border border-gray-02 p-8 md:px-14 md:py-10 flex justify-between gap-8 items-center hover:border-gold transition-colors">
      <div class="p1">Star the repo on GitHub</div>

      <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 48 48" fill="none" class="size-8 md:size-12">
        <g clip-path="url(#clip0_79_2469)">
          <path fill-rule="evenodd" clip-rule="evenodd" d="M23.9285 0C10.6967 0 0 11 0 24.6085C0 35.4865 6.85371 44.6945 16.3616 47.9535C17.5504 48.1985 17.9858 47.424 17.9858 46.7725C17.9858 46.202 17.9466 44.2465 17.9466 42.209C11.2903 43.676 9.90417 39.2755 9.90417 39.2755C8.83445 36.4235 7.24947 35.6905 7.24947 35.6905C5.07086 34.183 7.40816 34.183 7.40816 34.183C9.82482 34.346 11.0929 36.709 11.0929 36.709C13.2318 40.457 16.6785 39.398 18.0651 38.746C18.263 37.157 18.8973 36.057 19.5708 35.446C14.2619 34.8755 8.67625 32.757 8.67625 23.3045C8.67625 20.6155 9.62645 18.4155 11.1321 16.7045C10.8945 16.0935 10.0624 13.567 11.3701 10.1855C11.3701 10.1855 13.3905 9.5335 17.9461 12.7115C19.8965 12.1728 21.908 11.8988 23.9285 11.8965C25.9489 11.8965 28.0085 12.182 29.9104 12.7115C34.4665 9.5335 36.4869 10.1855 36.4869 10.1855C37.7946 13.567 36.962 16.0935 36.7244 16.7045C38.2697 18.4155 39.1807 20.6155 39.1807 23.3045C39.1807 32.757 33.5951 34.8345 28.2465 35.446C29.1184 36.22 29.8707 37.6865 29.8707 40.009C29.8707 43.309 29.8315 45.9575 29.8315 46.772C29.8315 47.424 30.2674 48.1985 31.4557 47.954C40.9636 44.694 47.8173 35.4865 47.8173 24.6085C47.8565 11 37.1207 0 23.9285 0Z" fill="white"/>
        </g>
        <defs>
          <clipPath id="clip0_79_2469">
            <rect width="48" height="48" fill="white"/>
          </clipPath>
        </defs>
      </svg>
    </a>
  </div>
</div>
</template>

<!-- Our custom home styles -->
<style scoped src="@theme/assets/css/home.css"></style>

<style scoped>
.input {
  @apply border border-off-white text-off-white;
  @apply focus:border-gold;
  @apply placeholder:text-gray-05 placeholder:opacity-100;
  @apply block px-4 py-2 flex-1 bg-transparent w-full;
  @apply outline-none focus:!outline-none;
}

.quote-bg {
  background-image: url("data:image/svg+xml,%3Csvg width='30' height='23' viewBox='0 0 30 23' fill='none' xmlns='http://www.w3.org/2000/svg'%3E%3Cpath d='M0.0875627 15.504C0.0875627 14.4 0.375563 13.32 0.951563 12.264V12.192L7.57556 0.887997H10.5996L5.84756 9.024L6.63956 8.952C8.41556 8.952 9.92756 9.6 11.1756 10.896C12.4716 12.144 13.1196 13.68 13.1196 15.504C13.1196 17.328 12.4716 18.888 11.1756 20.184C9.92756 21.432 8.41556 22.056 6.63956 22.056C4.81556 22.056 3.25556 21.432 1.95956 20.184C0.711563 18.888 0.0875627 17.328 0.0875627 15.504ZM23.4156 8.952C25.2396 8.952 26.7756 9.6 28.0236 10.896C29.3196 12.144 29.9676 13.68 29.9676 15.504C29.9676 17.28 29.3196 18.816 28.0236 20.112C26.7756 21.408 25.2396 22.056 23.4156 22.056C21.6396 22.056 20.1036 21.408 18.8076 20.112C17.5116 18.816 16.8636 17.28 16.8636 15.504C16.8636 14.352 17.1756 13.272 17.7996 12.264V12.192L24.3516 0.887997H27.4476L22.6956 9.024L23.4156 8.952Z' fill='%23FFAF52'/%3E%3C/svg%3E%0A");
  background-position: 40px 40px;
  background-repeat: no-repeat;
}

.quotes-container {
  transform: v-bind(quotesOffsetMobile);
}

@media screen(md) {
  .quotes-container {
    transform: v-bind(quotesOffsetDesktop);
  }
}
</style>
