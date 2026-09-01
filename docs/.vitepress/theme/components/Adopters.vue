<script setup lang="ts">
import { caseStudyAdopters, listedAdopters } from '@theme/adopters'

const addYourOrgUrl = 'https://github.com/kitops-ml/kitops/issues/new?title=Adopter%3A+%3Cyour+organization%3E&body=Organization%3A%0AHow+you+use+KitOps%3A%0ALink+%28site%2C+repo%2C+or+case+study%29%3A%0ALogo+%28attach+SVG+or+2x+PNG%29%3A%0A%0AI+confirm+I+have+permission+to+list+this+organization+as+a+KitOps+adopter.'
</script>

<template>
<div class="pt-12 md:pt-20 pb-32">
  <div class="px-6 md:px-12 content-container">
    <div class="max-w-4xl">
      <h1 class="font-heading!">
        Who's using
        <span class="text-gold">KitOps</span>
      </h1>
      <p class="p1 mt-6 md:mt-10">
        KitOps is a CNCF open standards project, and ModelKits are packaged, versioned, and shipped
        by teams that need their AI/ML projects to move through the same pipelines as everything else
        they run. These are the organizations building on it.
      </p>
      <div class="mt-10">
        <a :href="addYourOrgUrl" target="_blank" rel="noopener" class="kit-button inline-flex! items-center gap-2.5">
          Add your organization
          <svg xmlns="http://www.w3.org/2000/svg" width="21" height="17" viewBox="0 0 21 17" fill="none">
            <path d="M15.7625 2.20004H16.5125V11.2H15.0125V4.75942L5.79375 13.9782L5.2625 14.5094L4.20312 13.45L4.73438 12.9188L13.9531 3.70004H7.5125V2.20004H15.7625Z" fill="currentColor"/>
          </svg>
        </a>
      </div>
    </div>
  </div>

  <!-- Section 1 — adopters with a published case study -->
  <section v-if="caseStudyAdopters.length" class="px-6 md:px-12 content-container mt-20 md:mt-28">
    <h2 class="font-heading!">Case studies</h2>
    <p class="p2 mt-4 max-w-3xl text-gray-06!">
      Teams who have written up how they put KitOps to work, in their own words.
    </p>

    <div class="mt-10 grid grid-cols-1 lg:grid-cols-2 gap-6">
      <div
        v-for="adopter in caseStudyAdopters"
        :key="adopter.name"
        class="adopter-card flex flex-col">
        <div class="h-14 flex items-center">
          <img
            v-if="adopter.logo"
            :src="adopter.logo"
            :alt="`${adopter.name} logo`"
            :width="adopter.logoWidth"
            :height="adopter.logoHeight"
            class="max-h-14 w-auto opacity-80"
            loading="lazy">
          <span v-else class="h4 text-off-white!">{{ adopter.name }}</span>
        </div>

        <p class="p2 mt-4 text-gray-06!">{{ adopter.usage }}</p>

        <blockquote v-if="adopter.quote" class="adopter-quote mt-6 mb-0!">
          <p class="p2 text-off-white! m-0!">&ldquo;{{ adopter.quote.text }}&rdquo;</p>
          <footer class="p2 mt-3 text-gray-06!">
            {{ adopter.quote.author }}<template v-if="adopter.quote.title">, {{ adopter.quote.title }}</template>
          </footer>
        </blockquote>

        <div class="mt-auto pt-6">
          <a
            :href="adopter.caseStudyUrl"
            target="_blank"
            rel="noopener"
            class="inline-flex items-center gap-2 text-gold! font-bold hocus:underline no-underline!">
            Read the full case study
            <svg xmlns="http://www.w3.org/2000/svg" width="21" height="17" viewBox="0 0 21 17" fill="none" aria-hidden="true">
              <path d="M15.7625 2.20004H16.5125V11.2H15.0125V4.75942L5.79375 13.9782L5.2625 14.5094L4.20312 13.45L4.73438 12.9188L13.9531 3.70004H7.5125V2.20004H15.7625Z" fill="currentColor"/>
            </svg>
          </a>
        </div>
      </div>
    </div>
  </section>

  <!-- Section 2 — adopters without a case study -->
  <section class="px-6 md:px-12 content-container mt-20 md:mt-28">
    <h2 class="font-heading!">More adopters</h2>
    <p class="p2 mt-4 max-w-3xl text-gray-06!">
      Organizations and projects using KitOps in their AI/ML workflows.
    </p>

    <ul v-if="listedAdopters.length" class="adopter-list grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-x-8 gap-y-4">
      <li v-for="adopter in listedAdopters" :key="adopter.name" class="adopter-list-item">
        <a
          v-if="adopter.url"
          :href="adopter.url"
          target="_blank"
          rel="noopener"
          class="adopter-name font-mono text-gold! hocus:text-off-white! no-underline!">{{ adopter.name }}</a>
        <span v-else class="adopter-name font-mono text-gold!">{{ adopter.name }}</span>
      </li>
    </ul>

    <div v-else class="adopter-card mt-10">
      <p class="p2 m-0! text-gray-06!">
        Using KitOps but not ready to write a case study? We'd still like to list you.
        <a :href="addYourOrgUrl" target="_blank" rel="noopener" class="text-gold! hocus:underline">Open an issue</a>
        with your organization and a link, and we'll add you here.
      </p>
    </div>
  </section>

  <div class="px-6 md:px-12 content-container mt-24 md:mt-32">
    <div class="text-center">
      <h2 class="font-heading!">
        Ready to package your first
        <div class="text-gold">ModelKit?</div>
      </h2>
      <div class="mt-10 flex flex-wrap justify-center gap-4">
        <a href="/docs/get-started/" class="kit-button">Get started</a>
        <a href="https://discord.gg/Tapeh8agYy" target="_blank" rel="noopener" class="kit-button kit-button-cornflower">Join the community</a>
      </div>
    </div>
  </div>
</div>
</template>

<style scoped>
.adopter-card {
  padding: 1.5rem;
  border-radius: 0.5rem;
  background-color: #222222;
  border: 1px solid #363636;
}

.adopter-quote {
  padding-left: 1rem;
  border-left: 2px solid var(--color-gold);
}

/*
 * These override `.vp-doc ul` / `.vp-doc li` in theme/style.css, which set their
 * own margins and would otherwise win on specificity over Tailwind utilities.
 */
.adopter-list {
  list-style: none;
  margin: 4rem 0 0;
  padding: 0;
}

.adopter-list-item {
  margin-top: 0;
  padding: 0.375rem 0;
}

.adopter-name {
  font-size: 15px;
  line-height: 24px;
  transition: color 150ms;
}
</style>
