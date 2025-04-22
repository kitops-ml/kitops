<script setup lang="ts">
import { type Ref, inject, ref, onMounted } from 'vue'
import { useLocalStorage } from '@vueuse/core'
import { getUserOS } from '@theme/utils'
import vGaTrack from '@theme/directives/ga'

let documentClassList:DOMTokenList

onMounted(() => {
  documentClassList = document.documentElement.classList

  // On load, set the store platform
  documentClassList.add(`platform-${selectedPlatform.value}`)
})

const selectedPlatform = useLocalStorage('preferred-platform', getUserOS())

const isPlatformModalOpen = inject<Ref<boolean>>('isPlatformModalOpen', ref(false))

const closeModal = () => {
  isPlatformModalOpen.value = false
}

const changePlatform = () => {
  // clear all classes
  documentClassList.remove('platform-windows')
  documentClassList.remove('platform-mac')
  documentClassList.remove('platform-linux')
  documentClassList.add(`platform-${selectedPlatform.value}`)
}
</script>

<template>
<div class="container">
  <label>Preferred platform</label>

  <div class="select-wrapper">
    <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 12 12" fill="none">
      <path d="M5.99985 6.5857L8.47475 4.11084L9.18185 4.81795L5.99985 7.99995L2.81787 4.81795L3.52498 4.11084L5.99985 6.5857Z" fill="#ECECEC"/>
    </svg>

    <select
      @change="changePlatform"
      v-model="selectedPlatform"
      v-ga-track="{ category: 'select', action: 'change', label: 'platform selector switch', location: 'sidebar' }"
      aria-label="select your preferred platform">
      <option value="mac">Mac</option>
      <option value="windows">Windows</option>
      <option value="linux">Linux</option>
    </select>
  </div>
</div>

<ClientOnly>
  <Teleport to="body">
    <template v-if="isPlatformModalOpen">
      <div class="platform-select-modal">
        <div class="modal">
          <button class="close-button" @click="closeModal()">
            &times;
          </button>

          <div style="margin-bottom: 10px;">What platform are you using?</div>

          <div class="select-wrapper">
            <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 12 12" fill="none">
              <path d="M5.99985 6.5857L8.47475 4.11084L9.18185 4.81795L5.99985 7.99995L2.81787 4.81795L3.52498 4.11084L5.99985 6.5857Z" fill="#ECECEC"/>
            </svg>

            <select
              @change="() => {
                changePlatform();
                closeModal();
              }"
              v-ga-track="{ category: 'select', action: 'change', label: 'platform selector switch', location: 'platform switch modal' }"
              v-model="selectedPlatform"
              aria-label="select your preferred platform">
              <option value="mac">Mac</option>
              <option value="windows">Windows</option>
              <option value="linux">Linux</option>
            </select>
          </div>
        </div>
      </div>
    </template>
  </Teleport>
</ClientOnly>
</template>

<style>
.platform-snippet-mac,
.platform-snippet-linux,
.platform-snippet-windows {
  display: none;
}

html.platform-windows .platform-snippet-windows {
  display: initial;
}

html.platform-mac .platform-snippet-mac {
  display: initial;
}

html.platform-linux .platform-snippet-linux {
  display: initial;
}

.platform-select-modal {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.75);
  z-index: 50;
}

.platform-select-modal .modal {
  padding: 20px;
  border-radius: 6px;
  background: var(--vp-brand-3);
  position: absolute;
  left: 50%;
  bottom: 80px;
  transform: translateX(-50%);
  background-color: #1d1d1f;
  width: 90%;
}

@media (min-width: 640px) {
  .platform-select-modal .modal {
    bottom: auto;
    top: 50%;
    transform: translate(-50%, -50%);
    width: auto;
  }
}

.platform-select-modal .modal .close-button {
  display: block;
  margin-left: auto;
  margin-bottom: 20px;
  padding-left: 3px;
  padding-right: 3px;
  font-size: 24px;
}
</style>

<style scoped>
option {
  background-color: var(--color-night);
  color: var(--color-off-white);
}

.container {
  padding-top: 20px;
  margin-bottom: 10px;
  border-top: 1px solid var(--vp-c-divider);
}

.select-wrapper {
  position: relative;
}

.select-wrapper svg {
  position: absolute;
  right: 16px;
  top: 50%;
  transform: translateY(-50%);
  pointer-events: none;
}

.container label {
  display: block;
  font-size: 12px;
  font-weight: 700;
  margin-bottom: 8px;
  position: relative;
  color: #6A6A6A;
}

select {
  padding: 8px 24px;
  font-size: 14px;
  width: 100%;
  background: var(--color-elevation-05); /* elevation 5 */
}

select:focus-visible {
  outline: auto;
  outline-color: var(--vp-c-brand-1);
}
</style>
