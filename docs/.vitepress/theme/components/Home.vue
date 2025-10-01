<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { Vue3Marquee } from 'vue3-marquee'
import axios from 'axios'
import VueTurnstile from 'vue-turnstile'
import Accordion from './Accordion.vue'

const error = ref('')
const email = ref('')
const token = ref('')
const favoriteDevOpsTool = ref('')
const isBusy = ref(false)
const isSubscribed = ref(false)
const isSuccess = ref(false)
const isProd = import.meta.env.PROD

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
<div class="mt-32 md:mt-40 px-6 md:px-12 text-center content-container">
  <div class="relative">
    <div class="bg-radial from-[--alpha(var(--color-gold)/30%)] via-[--alpha(#996931/5%)] to-transparent from-0% to-60% absolute size-full lg:size-160 top-1/2 left-1/2 -translate-1/2"></div>
    <h1 class="z-0 relative mt-4! font-medium!">The Enterprise Model Registry <br><span class="bg-gradient-to-b from-gold to-[#996931] text-transparent bg-clip-text">for Secure AI</span></h1>
    <h2 class="z-0 relative font-bold! font-sans! text-3xl! my-6!">Secure. Portable. No Lock‑In.</h2>
    <p class="z-0 relative h4 font-normal! text-off-white!">The only model registry that works anywhere containers run — from public cloud to the most locked‑down air‑gapped environments.</p>
  </div>

  <div class="flex flex-col lg:flex-row justify-center items-center gap-10 mt-10 md:mt-14 xl:mt-22">
    <div class="relative inline-block w-full max-w-48 lg:w-auto text-center lg:text-left group">
      <a href="/docs/cli/installation/#%F0%9F%AA%9F-windows-install" class="kit-button !flex justify-center lg:justify-start items-center gap-2">
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
      <a href="/docs/cli/installation/#%F0%9F%8D%8E-macos-install" class="kit-button !flex justify-center lg:justify-start items-center gap-2">
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
      <a href="/docs/cli/installation/#%F0%9F%90%A7-linux-install" class="kit-button !flex justify-center lg:justify-start items-center gap-2">
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
  <h2>Why Kitops exists</h2>
  <div class="p1 space-y-8! mt-8 max-w-lg mx-auto">
    <p>Enterprise AI teams face three problems with traditional model registries: <strong>security gaps, vendor lock‑in, and deployment friction</strong>.<br>KitOps solves them all.</p>
    <p>Unlike proprietary registries that trap you in closed ecosystems, KitOps stores models, datasets, code, and configs as OCI artifacts in <strong>your</strong> container registries — Docker Hub, ECR, GCR, Harbor, Artifactory, and beyond.</p>
    <p>With 100,000+ downloads and 18+ months in production, KitOps is trusted by organizations that can’t compromise on <strong>security, compliance, or uptime</strong>.</p>
  </div>

  <iframe width="1050" class="border-16 border-[--alpha(#2c2c2c/20%)] rounded-xl max-w-full aspect-video mt-22 mx-auto" src="https://www.youtube.com/embed/iK9mnU0prRU?si=ommsLD32Kjj4RUMu&amp;autoplay=0" title="YouTube video player" frameborder="0" allow="accelerometer; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" referrerpolicy="strict-origin-when-cross-origin" allowfullscreen></iframe>

  <div class="mt-16">
    <a href="https://jozu.com/kitops-modelpack-support" target="_blank" class="px-4 py-2.5 rounded-lg inline-flex items-center gap-2.5 bg-[#FFB052] hocus:opacity-50 text-black!">
      Enterprise Support for KitOps
      <svg xmlns="http://www.w3.org/2000/svg" width="21" height="17" viewBox="0 0 21 17" fill="none">
        <path d="M15.7625 2.20004H16.5125V11.2H15.0125V4.75942L5.79375 13.9782L5.2625 14.5094L4.20312 13.45L4.73438 12.9188L13.9531 3.70004H7.5125V2.20004H15.7625Z" fill="black"/>
      </svg>
    </a>
  </div>
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

<div id="whykitops" class="mt-32 md:mt-40 xl:mt-60 px-6 md:px-0 content-container">
  <h2 class="text-center">Why engineers choose Kitops<span class="font-heading font-extralight">?</span></h2>

  <div class="mt-10 md:mt-14 xl:mt-22 grid grid-cols-1 md:grid-cols-3 gap-3">
    <div>
      <div class="p2 text-gold!">Security That’s Built In — Not Bolted On</div>
      <ul class="p2 list-disc! list-inside! leading-tight!">
        <li class="pl-3 text-[#B8B8B8]"><span class="text-white">Immutable by design</span> — once published, artifacts can’t be tampered with</li>
        <li class="pl-3 text-[#B8B8B8]"><span class="text-white">Cryptographically signed</span> — every model, dataset, and config is verifiable</li>
        <li class="pl-3 text-[#B8B8B8]"><span class="text-white">Provenance tracking</span> — full supply chain auditability out of the box</li>
        <li class="pl-3 text-[#B8B8B8]"><span class="text-white">Defense in depth</span> — layered SHA digest verification for all artifacts</li>
      </ul>
    </div>

    <div>
      <div class="p2 text-gold! xs:mt-12">Fits Into Your Toolchain — No Re‑Platforming</div>
      <ul class="p2 list-disc! list-inside! leading-tight!">
        <li class="pl-3 text-[#B8B8B8]"><span class="text-white">OCI‑native</span> — runs on any compliant container registry</li>
        <li class="pl-3 text-[#B8B8B8]"><span class="text-white">No proprietary formats</span> — open standards mean zero lock‑in</li>
        <li class="pl-3 text-[#B8B8B8]"><span class="text-white">CI/CD‑ready</span> — GitHub Actions, GitLab CI, Jenkins, Azure DevOps, and more </li>
        <li class="pl-3 text-[#B8B8B8]"><span class="text-white">Deployment‑friendly</span> — works with Kubernetes, Docker, and existing pipelines</li>
      </ul>
    </div>

    <div>
      <div class="p2 text-gold! xs:mt-12">Enterprise‑Grade from Day One</div>
      <ul class="p2 list-disc! list-inside! leading-tight!">
        <li class="pl-3 text-[#B8B8B8]"><span class="text-white">Air‑gapped deployments</span> — fully supported</li>
        <li class="pl-3 text-[#B8B8B8]"><span class="text-white">Multi‑cloud portability</span> — same artifact runs anywhere</li>
        <li class="pl-3 text-[#B8B8B8]"><span class="text-white">Deduplicated storage</span> — cuts registry costs by up to 80%</li>
        <li class="pl-3 text-[#B8B8B8]"><span class="text-white">Automatic containerization</span> — for any model  </li>
        <li class="pl-3 text-[#B8B8B8]"><span class="text-white">RBAC integration</span> — aligns with your security policies</li>
      </ul>
    </div>
  </div>

  <div class="text-center mt-16">
    <a href="/docs/get-started/" class="bg-white rounded-lg text-black! px-4 py-2.5 inline-flex items-center gap-2.5">
      Get started
      <svg xmlns="http://www.w3.org/2000/svg" width="21" height="17" viewBox="0 0 21 17" fill="none">
        <path d="M15.7625 2.20004H16.5125V11.2H15.0125V4.75942L5.79375 13.9782L5.2625 14.5094L4.20312 13.45L4.73438 12.9188L13.9531 3.70004H7.5125V2.20004H15.7625Z" fill="black"/>
      </svg>
    </a>
  </div>
</div>

<div class="mt-32 md:mt-40 xl:mt-60 px-6 md:px-12 text-center content-container">
  <h2>
    Works with the tools
    <div class="text-gold">you Already use</div>
  </h2>

  <div class="mt-16 space-y-12! relative marquee-gradients">
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
</div>

<div class="mt-32 md:mt-40 xl:mt-60 px-6 md:px-12 content-container">
  <div class="text-center mb-6">
    <div class="inline-block p-2.5 rounded-lg border-2 border-gold text-gold bg-[--alpha(var(--color-gold)/40%)] uppercase font-bold">MISSION</div>
  </div>
  <h2 class="text-center font-heading!">Why the Standard Matters</h2>
  <div class="max-w-2xl mx-auto p1 space-y-8! mt-8 text-center text-[#B8B8B8]!">
    <p>Enterprise AI teams face three problems with traditional model registries: security gaps, vendor lock‑in, and deployment friction. KitOps solves them all.</p>
    <p>Unlike proprietary registries that trap you in closed ecosystems, KitOps stores models, datasets, code, and configs as OCI artifacts in your container registries — Docker Hub, ECR, GCR, Harbor, Artifactory, and beyond.</p>
    <p>With 150,000+ downloads and 18+ months in production, KitOps is trusted by organizations that can’t compromise on security, compliance, or uptime.</p>
  </div>
</div>

<div class="mt-32 md:mt-40 xl:mt-60 px-6 md:px-12 content-container">
  <div class="text-center mb-6">
    <div class="inline-block p-2.5 rounded-lg border-2 border-gold text-gold bg-[--alpha(var(--color-gold)/40%)] uppercase font-bold">THE ENGINEERS CHOICE</div>
  </div>
  <h2 class="text-center font-heading!">Built for Engineers Who Demand Trust</h2>

  <div class="mt-10 md:mt-14 xl:mt-22 grid grid-cols-1 md:grid-cols-3 gap-3">
    <div>
      <div class="p2 text-gold!">Proof / Momentum Focused</div>
      <ul class="p2 list-disc! list-inside! leading-tight!">
        <li class="pl-3 text-[#B8B8B8]!">Engineered for Adoption, Trusted in Production</li>
        <li class="pl-3 text-[#B8B8B8]!">The Standard Engineers Rely On</li>
        <li class="pl-3 text-[#B8B8B8]!">Why Teams Ship with KitOps</li>
      </ul>
    </div>

    <div>
      <div class="p2 text-gold! xs:mt-12">Trust & Security Focused</div>
      <ul class="p2 list-disc! list-inside! leading-tight!">
        <li class="pl-3 text-[#B8B8B8]!">Built for Engineers Who Demand Trust</li>
        <li class="pl-3 text-[#B8B8B8]!">The Secure Choice for Modern AI</li>
        <li class="pl-3 text-[#B8B8B8]!">Security by Design, Not by Accident</li>
      </ul>
    </div>

    <div>
      <div class="p2 text-gold! xs:mt-12">Integration & Workflow Focused</div>
      <ul class="p2 list-disc! list-inside! leading-tight!">
        <li class="pl-3 text-[#B8B8B8]!">Seamless Security, Native to Your Stack</li>
        <li class="pl-3 text-[#B8B8B8]!">Fits Where You Work, Scales How You Build</li>
        <li class="pl-3 text-[#B8B8B8]!">Security That Moves at CI/CD Speed</li>
      </ul>
    </div>
  </div>
</div>

<div class="mt-32 md:mt-40 xl:mt-60 px-6 md:px-12 content-container">
  <div class="text-center mb-6">
    <div class="inline-block p-2.5 rounded-lg border-2 border-gold text-gold bg-[--alpha(var(--color-gold)/40%)] uppercase font-bold">Built for Scale</div>
  </div>
  <h2 class="text-center font-heading!">Enterprise‑Grade from Day One</h2>

  <div class="mt-10 md:mt-14 xl:mt-22 ">
    <div class="grid grid-cols-1 lg:grid-cols-3 gap-3 w-max mx-auto">
      <div>
        <div class="h-full bg-black p-6 rounded-xl border border-[#363636] p1 w-full max-w-[340px] mx-auto">
          Air-gapped deployments

          <div class="flex gap-3 mt-5 text-gold! leading-tight p2">
            <svg xmlns="http://www.w3.org/2000/svg" width="24" height="25" viewBox="0 0 24 25" fill="none" class="min-w-6 mt-1">
              <path d="M3 12.26C3 17.2306 7.02944 21.26 12 21.26C16.9706 21.26 21 17.2306 21 12.26C21 7.28945 16.9706 3.26001 12 3.26001C7.02944 3.26001 3 7.28945 3 12.26Z" fill="#EB9D42" fill-opacity="0.39"/>
              <path d="M15 10.26L11 14.26L9 12.26M12 21.26C7.02944 21.26 3 17.2306 3 12.26C3 7.28945 7.02944 3.26001 12 3.26001C16.9706 3.26001 21 7.28945 21 12.26C21 17.2306 16.9706 21.26 12 21.26Z" stroke="#FFAF52" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
            </svg>
            Fully supported
          </div>
        </div>
      </div>

      <div>
        <div class="h-full bg-black p-6 rounded-xl border border-[#363636] p1 w-full max-w-[340px] mx-auto">
          Multi-cloud portability

          <div class="flex gap-3 mt-5 text-gold! leading-tight p2">
            <svg xmlns="http://www.w3.org/2000/svg" width="24" height="25" viewBox="0 0 24 25" fill="none" class="min-w-6 mt-1">
              <path d="M3 12.26C3 17.2306 7.02944 21.26 12 21.26C16.9706 21.26 21 17.2306 21 12.26C21 7.28945 16.9706 3.26001 12 3.26001C7.02944 3.26001 3 7.28945 3 12.26Z" fill="#EB9D42" fill-opacity="0.39"/>
              <path d="M15 10.26L11 14.26L9 12.26M12 21.26C7.02944 21.26 3 17.2306 3 12.26C3 7.28945 7.02944 3.26001 12 3.26001C16.9706 3.26001 21 7.28945 21 12.26C21 17.2306 16.9706 21.26 12 21.26Z" stroke="#FFAF52" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
            </svg>
            Same artifact runs everywhere
          </div>
        </div>
      </div>

      <div>
        <div class="h-full bg-black p-6 rounded-xl border border-[#363636] p1 w-full max-w-[340px] mx-auto">
          Deduplicated storage

          <div class="flex gap-3 mt-5 text-gold! leading-tight p2">
            <svg xmlns="http://www.w3.org/2000/svg" width="24" height="25" viewBox="0 0 24 25" fill="none" class="min-w-6 mt-1">
              <path d="M3 12.26C3 17.2306 7.02944 21.26 12 21.26C16.9706 21.26 21 17.2306 21 12.26C21 7.28945 16.9706 3.26001 12 3.26001C7.02944 3.26001 3 7.28945 3 12.26Z" fill="#EB9D42" fill-opacity="0.39"/>
              <path d="M15 10.26L11 14.26L9 12.26M12 21.26C7.02944 21.26 3 17.2306 3 12.26C3 7.28945 7.02944 3.26001 12 3.26001C16.9706 3.26001 21 7.28945 21 12.26C21 17.2306 16.9706 21.26 12 21.26Z" stroke="#FFAF52" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
            </svg>
            Cuts registry costs by up to 80%
          </div>
        </div>
      </div>
    </div>

    <div class="grid grid-cols-1 lg:grid-cols-2 gap-3 mt-0 lg:mt-4 w-max mx-auto">
      <div>
        <div class="h-full bg-black p-6 rounded-xl border border-[#363636] p1 w-full max-w-[340px] mx-auto">
          Automatic containerization

          <div class="flex gap-3 mt-5 text-gold! leading-tight p2">
            <svg xmlns="http://www.w3.org/2000/svg" width="24" height="25" viewBox="0 0 24 25" fill="none" class="min-w-6 mt-1">
              <path d="M3 12.26C3 17.2306 7.02944 21.26 12 21.26C16.9706 21.26 21 17.2306 21 12.26C21 7.28945 16.9706 3.26001 12 3.26001C7.02944 3.26001 3 7.28945 3 12.26Z" fill="#EB9D42" fill-opacity="0.39"/>
              <path d="M15 10.26L11 14.26L9 12.26M12 21.26C7.02944 21.26 3 17.2306 3 12.26C3 7.28945 7.02944 3.26001 12 3.26001C16.9706 3.26001 21 7.28945 21 12.26C21 17.2306 16.9706 21.26 12 21.26Z" stroke="#FFAF52" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
            </svg>
            For any model
          </div>
        </div>
      </div>

      <div>
        <div class="h-full bg-black p-6 rounded-xl border border-[#363636] p1 w-full max-w-[340px] mx-auto">
          RBAC integration

          <div class="flex gap-3 mt-5 text-gold! leading-tight p2">
            <svg xmlns="http://www.w3.org/2000/svg" width="24" height="25" viewBox="0 0 24 25" fill="none" class="min-w-6 mt-1">
              <path d="M3 12.26C3 17.2306 7.02944 21.26 12 21.26C16.9706 21.26 21 17.2306 21 12.26C21 7.28945 16.9706 3.26001 12 3.26001C7.02944 3.26001 3 7.28945 3 12.26Z" fill="#EB9D42" fill-opacity="0.39"/>
              <path d="M15 10.26L11 14.26L9 12.26M12 21.26C7.02944 21.26 3 17.2306 3 12.26C3 7.28945 7.02944 3.26001 12 3.26001C16.9706 3.26001 21 7.28945 21 12.26C21 17.2306 16.9706 21.26 12 21.26Z" stroke="#FFAF52" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
            </svg>
            Aligns with your security policies
          </div>
        </div>
      </div>
    </div>
  </div>
</div>

<div class="mt-32 md:mt-40 xl:mt-60 px-6 md:px-12 content-container">
  <h2 class="text-center font-heading!">Get Started with KitOps</h2>

  <div class="kit-cards mt-8">
    <div class="kit-card flex flex-col">
      <div class="icon">
        <svg width="24" height="25" viewBox="0 0 24 25" fill="none" xmlns="http://www.w3.org/2000/svg">
          <path d="M17 15.9248H12M7 10.9248L10 13.4248L7 15.9248M3 16.725V9.125C3 8.0049 3 7.44443 3.21799 7.0166C3.40973 6.64028 3.71547 6.33454 4.0918 6.14279C4.51962 5.9248 5.08009 5.9248 6.2002 5.9248H17.8002C18.9203 5.9248 19.4796 5.9248 19.9074 6.14279C20.2837 6.33454 20.5905 6.64028 20.7822 7.0166C21 7.44401 21 8.0038 21 9.12171V16.7279C21 17.8458 21 18.4048 20.7822 18.8322C20.5905 19.2085 20.2837 19.5153 19.9074 19.707C19.48 19.9248 18.921 19.9248 17.8031 19.9248H6.19691C5.07899 19.9248 4.5192 19.9248 4.0918 19.707C3.71547 19.5153 3.40973 19.2085 3.21799 18.8322C3 18.4044 3 17.8451 3 16.725Z" stroke="#FFAF52" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
        </svg>
      </div>
      <div class="mt-8 flex flex-col flex-1 p2">
        <p>Install the KitOps CLI</p>
        <code>Brew install KitOps</code>
      </div>
    </div>

    <div class="kit-card flex flex-col">
      <div class="icon">
        <svg xmlns="http://www.w3.org/2000/svg" width="24" height="25" viewBox="0 0 24 25" fill="none" class="min-w-6">
          <path d="M21 12.9248L12 18.9248L3 12.9248M21 16.9248L12 22.9248L3 16.9248M21 8.9248L12 14.9248L3 8.9248L12 2.9248L21 8.9248Z" stroke="#FFAF52" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
        </svg>
      </div>
      <div class="mt-8 flex flex-col flex-1 p2">
        <p>Package Your First Model</p>
        <code>Kit pack . -t myregistry.com/my-model:latest</code>
      </div>
    </div>

    <div class="kit-card flex flex-col">
      <div class="icon">
        <svg xmlns="http://www.w3.org/2000/svg" width="24" height="25" viewBox="0 0 24 25" fill="none" class="min-w-6">
          <path d="M18 12.9248V17.9248C18 19.5817 15.3137 20.9248 12 20.9248C8.68629 20.9248 6 19.5817 6 17.9248V12.9248M18 12.9248V7.9248M18 12.9248C18 14.5817 15.3137 15.9248 12 15.9248C8.68629 15.9248 6 14.5817 6 12.9248M18 7.9248C18 6.26795 15.3137 4.9248 12 4.9248C8.68629 4.9248 6 6.26795 6 7.9248M18 7.9248C18 9.58166 15.3137 10.9248 12 10.9248C8.68629 10.9248 6 9.58166 6 7.9248M6 12.9248V7.9248" stroke="#FFAF52" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
        </svg>
      </div>
      <div class="mt-8 flex flex-col flex-1 p2">
        <p>Store in Your Registry</p>
        <code>Kit push myregistry.com/my-model:latest</code>
      </div>
    </div>

    <div class="kit-card flex flex-col">
      <div class="icon">
        <svg xmlns="http://www.w3.org/2000/svg" width="24" height="25" viewBox="0 0 24 25" fill="none" class="min-w-6">
          <path d="M3 12.9248H8M3 12.9248C3 17.8954 7.02944 21.9248 12 21.9248M3 12.9248C3 7.95424 7.02944 3.9248 12 3.9248M8 12.9248H16M8 12.9248C8 17.8954 9.79086 21.9248 12 21.9248M8 12.9248C8 7.95424 9.79086 3.9248 12 3.9248M16 12.9248H21M16 12.9248C16 7.95424 14.2091 3.9248 12 3.9248M16 12.9248C16 17.8954 14.2091 21.9248 12 21.9248M21 12.9248C21 7.95424 16.9706 3.9248 12 3.9248M21 12.9248C21 17.8954 16.9706 21.9248 12 21.9248" stroke="#FFAF52" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
        </svg>
      </div>
      <div class="mt-8 flex flex-col flex-1 p2">
        <p>Deploy Anywhere</p>
        <code>Kit unpack myregistry.com/my-model:latest --model</code>
      </div>
    </div>
  </div>
</div>

<div class="max-w-3xl mx-auto mt-32 md:mt-40 xl:mt-60 px-6 faq-section content-container">
  <h2 class="text-center !mb-10 md:!mb-14 lg:!mb-22">fAq</h2>

  <Accordion content-class="!space-y-[1em]">
    <template #title>Are ModelKits a versioning solution or a packaging solution?</template>

    <p class="!mt-6">
      ModelKits do both. With a ModelKit, you can package all the parts of your AI project in one shareable asset, and tag them with a version.
      ModelKits were designed for the model development lifecycle, where projects are handed off from data science teams to application teams to deployment teams. Versioning and packaging makes it easy for team members to find the datasets and configurations that map to a specific model version.
      You can <a href="/docs/overview/" class="underline">read more details about KitOps in our overview</a>.
    </p>
  </Accordion>

  <Accordion content-class="!space-y-[1em]">
    <template #title>How do I get started with Kit?</template>

    <p class="!mt-6">The easiest way to get started is to follow our <a href="/docs/get-started/" class="underline">Quick Start</a>, where you’ll learn how to:</p>

    <ul class="space-y-2 list-disc list-inside">
      <li>Package up a model, notebook, and datasets into a single ModelKit you can use with your existing tools</li>
      <li>Share the ModelKit with others through your public or private registry</li>
      <li>Grab only the assets you need from the ModelKit for testing, integration, local running, or deployment</li>
    </ul>

  </Accordion>

  <Accordion content-class="!space-y-[1em]">
    <template #title>Can I see if something changed between ModelKits?</template>

    <p class="!mt-6">Yes [choir sings hallelujah], each ModelKit includes SHA digests for the ModelKit and every artifact it holds so you can quickly see if something changed between ModelKit versions. </p>
  </Accordion>

  <Accordion content-class="!space-y-[1em]">
    <template #title>What are the benefits of using Kit?</template>

    <p class="!mt-6">Increased speed: Teams can work faster with a centralized and versioned package for their AI project coordination. ModelKits eliminate hunting for datasets or code, and make it obvious which datasets and configurations are needed for each model. Handoffs can be automated and executed quickly and with confidence.</p>
    <p>Reduced risk: ModelKits are self-verifying. Both the ModelKit itself and all the artifacts added to it are tamper-proof. Anyone can quickly and easily verify when something may have changed.</p>
    <p>Improved efficiency: Models stored in ModelKits can be run locally for experimentation or application integration, or packaged for deployment with a single command. Any artifact in a ModelKit can be separately pulled saving time and space on local or shared machines. This makes it easy for data scientists, application developers, and DevOps engineers to find and grab the pieces they need to do their job without being overwhelmed with unnecessary files.</p>
  </Accordion>

  <Accordion content-class="!space-y-[1em]">
    <template #title>What tools are compatible with Kit?</template>

    <p class="!mt-6">ModelKits store their assets as OCI-compatible artifacts. This makes them compatible with nearly every development and deployment tool and registry in use today.</p>
  </Accordion>

  <Accordion content-class="!space-y-[1em]">
    <template #title>Where are ModelKits stored?</template>
    <p class="!mt-6">ModelKits can be stored in any OCI-compliant registry - for example in a container registry like Docker Hub or Jozu Hub, or your favorite cloud vendor’s container registry, they can even be stored in an artifact repository like Artifactory.</p>
  </Accordion>

  <Accordion content-class="!space-y-[1em]">
    <template #title>Is KitOps open source and free to use?</template>

    <p class="!mt-6">Yes, it is licensed with the Apache 2.0 license and welcomes all users and contributors. If you’re <a href="https://github.com/kitops-ml/kitops/blob/main/CONTRIBUTING.md" class="underline">interested in contributing</a>, let us know.</p>
  </Accordion>

  <Accordion content-class="!space-y-[1em]">
    <template #title>Are ModelKits a replacement for Docker containers?</template>

    <p class="!mt-6">No, ModelKits complement containers - in fact, KitOps can take a ModelKit and generate a container for the model automatically. However, not all models should be deployed inside containers - sometimes it’s more efficient and faster to deploy an init container linked to the model for deployment. Datasets may also not need to be in containers - many datasets are easier to read and manipulate for training and validation when they’re not in a container. Finally, each container is still separate so even if you do want to put everything in its own container it’s not clear to people outside the AI project which datasets go with which models and which configurations.</p>
  </Accordion>

  <Accordion content-class="!space-y-[1em]">
    <template #title>Why would I use KitOps for versioning instead of Git?</template>

    <p class="!mt-6">Models and datasets in AI projects are often 10s or 100s of GB in size. Git was designed to work with many small files that can be easily diff’ed between versions. Git treats models and datasets stored in LFS (large file storage) as atomic blobs and can’t differentiate between versions of them. This makes it both inefficient and dangerous since it’s easy for someone to tamper with the models and datasets in the LFS without Git knowing. Finally, once you use LFS, a clone is no longer guaranteed to be the same as the original repo, because the repo refers to an LFS server that is independent of the clone and can change independently.</p>
  </Accordion>

  <Accordion content-class="!space-y-[1em]">
    <template #title>My MLOps tools do versioning, why do I need Kit?</template>

    <p class="!mt-6">KitOps is the only standards-based and open source solution for packaging and versioning AI project assets. Popular MLOps tools use proprietary and often closed formats to lock you into their ecosystem. This makes handoffs between MLOps tool users and non-MLOps tool users (like your application development and DevOps teams) unnecessarily hard. The future of MLOps tools is still being written, and it’s likely that many will be acquired or shut down and the cost of moving projects from one proprietary format to another is high. By using the OCI standard that’s already supported by nearly every tool on the planet, ModelKits give you a future-proofed solution for packaging and versioning that is compatible with both your MLOps tools and development / DevOps tools so everyone can collaborate regardless of the tools they use.</p>
  </Accordion>

  <Accordion content-class="!space-y-[1em]">
    <template #title>Is enterprise support available for Kit?</template>

    <p class="!mt-6">Enterprise support for ModelKits and the Kit CLI is available from <a href="https://www.jozu.com/" class="underline" target="_blank">Jozu</a>.</p>
  </Accordion>
</div>

<div class="bg-black/20 py-18 mt-32">
  <div class="text-center max-w-xl mx-auto">
    <svg class="w-[140px] h-[56px] mx-auto" viewBox="0 0 198 64" fill="none" xmlns="http://www.w3.org/2000/svg">
      <path d="M22.2887 19.9949L22.2993 57.1745C33.4331 53.7677 37.8818 53.1441 37.8818 42.0752V19.9949H22.2887Z" fill="#00ffc8" />
      <path d="M22.2887 57.1714C11.1549 53.7647 6.69567 53.135 6.69567 42.0661V19.9949H22.2887V14.6991H0V42.0661C0 57.3677 8.46397 61.1762 22.2887 63.997L22.2993 57.2152V57.1744L22.2887 57.1714Z"
        fill="#00ffc8" />
      <path d="M74.4648 -3.05176e-05H82.4788V8.45186H74.4648V-3.05176e-05Z" fill="currentColor" />
      <path d="M90.2602 32.3413C90.2602 21.3177 97.4602 13.5227 107.679 13.5227C117.898 13.5227 125.102 21.3177 125.102 32.3413C125.102 43.3648 117.827 51.2353 107.679 51.2353C97.5312 51.2353
        90.2602 43.6608 90.2602 32.3413ZM107.679 46.3079C114.441 46.3079 118.559 40.6512 118.559 32.3413C118.559 24.0313 114.298 18.5226 107.679 18.5226C101.06 18.5226 96.8033 24.0358 96.8033
        32.3413C96.8033 40.6467 100.918 46.3079 107.679 46.3079Z" fill="currentColor" />
      <path d="M130.09 44.9111L149.789 21.6076C150.304 20.9462 150.817 20.2123 151.408 19.6219H130.824V14.6975H158.318V19.9163L138.765 43.0009L136.998 45.2056H158.759V50.1299H130.088V44.9111H
        130.09Z" fill="currentColor" />
      <path d="M165.652 36.8987V14.699H171.975V35.5034C171.975 42.4119 174.252 46.3079 180.499 46.3079C186.746 46.3079 189.836 41.6041 189.836 33.8136V14.7006H196.156V50.1315H189.984V45.2071C
        188.294 49.0261 183.956 51.2353 178.885 51.2353C170.062 51.2353 165.652 45.941 165.652 36.8987Z" fill="currentColor" />
      <path d="M60.1915 19.9888H75.7785V42.0646C75.7785 53.138 71.3268 53.7707 60.1749 57.1775L60.193 64C74.0163 61.1791 82.4802 57.3707 82.4802 42.0661V14.7021H60.193V19.9918L60.1915 19.9888Z"
        fill="currentColor" />
    </svg>

    <p class="my-8! p1">Production-ready support from the creators and maintainers of the industry's leading open source AI/ML packaging standards</p>

    <div class="flex items-center justify-center">
      <a href="https://jozu.com/pricing" target="_blank" class="text-[#B8B8B8] px-4 py-3 inline-flex items-center gap-2.5 rounded-lg bg-[#141414] border border-[#363636]">
        See Support Plans
        <svg xmlns="http://www.w3.org/2000/svg" width="21" height="17" viewBox="0 0 21 17" fill="none">
          <path d="M18.2156 9.48618L18.9219 8.77993L18.2156 8.07368L13.2156 3.07368L12.5094 2.36743L11.0938 3.78306C11.1344 3.82368 12.4688 5.15806 15.0938 7.78306H2.50938V9.78306H15.0938C12.4688 12.4081 11.1344 13.7424 11.0938 13.7831L12.5094 15.1987L13.2156 14.4924L18.2156 9.49243V9.48618Z" fill="#FFAF52"/>
        </svg>
      </a>
    </div>
  </div>
</div>

</template>

<!-- Our custom home styles -->
<style scoped src="@theme/assets/css/home.css"></style>

<style scoped>
.input {
  display: block;
  border: 1px solid var(--color-off-white);
  color: var(--color-off-white);
  outline: none;
  padding-block: 8px;
  padding-inline: 16px;
  flex: 1;
  background-color: transparent;
  width: 100%;

  &::placeholder {
    color: var(--color-gray-05);
    opacity: 1;
  }

  &:focus {
    border-color: var(--color-gold);
    outline: none !important;
  }
}
</style>
