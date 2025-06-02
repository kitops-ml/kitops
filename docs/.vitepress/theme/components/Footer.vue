<script setup lang="ts">
import { useSidebar } from 'vitepress/theme'
import { useData } from 'vitepress/client'
import KernelKrunch from './KernelKrunch.vue'
import vGaTrack from '@theme/directives/ga'

const { theme, frontmatter, page, site } = useData()
const { hasSidebar } = useSidebar()
</script>

<template>
<footer
  v-if="theme.footer && frontmatter.footer !== false"
  class="bg-black"
  :class="{ 'hidden': hasSidebar }">
  <div class="container mx-auto px-6 lg:px-12 pt-10 pb-24 lg:py-24 xl:px-16 grid grid-cols-1 lg:grid-cols-2">
    <div class="flex flex-col lg:flex-row items-center gap-8">
      <KernelKrunch class="size-30 scale-50 lg:scale-100 origin-bottom" />

      <div class="flex flex-col gap-4">
        <img src="/logo.svg" class="w-32 mx-auto lg:mx-0 lg:w-48" alt="kitops logo" />

        <div class="flex items-center gap-0 lg:gap-4 p2">
          <span aria-label="made with love by Jozu">Made with <span class="text-red-500">♥️</span> by</span>
          <a href="https://jozu.com" target="_blank">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-4.5 lg:h-7" width="88" height="28" viewBox="0 0 56 18" fill="none" title="Jozu">
              <path d="M21.0996 0H23.3535V2.37709H21.0996V0Z" fill="white"/>
              <path d="M25.542 9.09594C25.542 5.99557 27.567 3.80322 30.441 3.80322C33.315 3.80322 35.3413 5.99557 35.3413 9.09594C35.3413 12.1963 33.295 14.4099 30.441 14.4099C27.587 14.4099 25.542 12.2796 25.542 9.09594ZM30.441 13.0241C32.3429 13.0241 33.501 11.4331 33.501 9.09594C33.501 6.75877 32.3025 5.20943 30.441 5.20943C28.5795 5.20943 27.3823 6.76005 27.3823 9.09594C27.3823 11.4318 28.5396 13.0241 30.441 13.0241Z" fill="white"/>
              <path d="M36.7441 12.6312L42.2844 6.07713C42.4292 5.89111 42.5736 5.6847 42.7397 5.51864H36.9505V4.13367H44.6832V5.60146L39.184 12.094L38.6871 12.7141H44.8072V14.099H36.7437V12.6312H36.7441Z" fill="white"/>
              <path d="M46.7461 10.3778V4.13416H48.5243V9.98537C48.5243 11.9284 49.1648 13.0242 50.9218 13.0242C52.6789 13.0242 53.5478 11.7012 53.5478 9.51012V4.13458H55.3252V14.0995H53.5894V12.7145C53.1142 13.7886 51.894 14.41 50.4678 14.41C47.9862 14.41 46.7461 12.921 46.7461 10.3778Z" fill="white"/>
              <path d="M17.0852 5.62179H21.4691V11.8306C21.4691 14.945 20.217 15.1229 17.0806 16.0811L17.0857 17.9999C20.9734 17.2066 23.3539 16.1355 23.3539 11.831V4.13489H17.0857V5.62264L17.0852 5.62179Z" fill="white"/>
              <path d="M6.42505 5.62354L6.42802 16.0803C9.5594 15.1221 10.8106 14.9467 10.8106 11.8336V5.62354H6.42505Z" fill="white"/>
              <path d="M6.42495 16.0795C3.29357 15.1214 2.03941 14.9443 2.03941 11.8311V5.62361H6.42495V4.13416H0.15625V11.8311C0.15625 16.1347 2.53674 17.2058 6.42495 17.9992L6.42792 16.0918V16.0804L6.42495 16.0795Z" fill="white"/>
            </svg>
          </a>
        </div>
      </div>
    </div>

    <nav class="flex flex-wrap lg:flex-nowrap justify-between lg:justify-end gap-4 text-sm mt-12 lg:mt-0">
      <div class="space-y-4 flex-1 max-w-44">
        <div class="text-gray-06">Install KitOps</div>
        <ul class="space-y-2">
          <li class="font-bold">
            <a href="/docs/cli/installation/#🪟-windows-install" class="flex gap-4 hocus:!underline">
              <svg xmlns="http://www.w3.org/2000/svg" width="20" height="21" viewBox="0 0 20 21" fill="none">
                <path d="M2.5 16.3337H17.5V18.0003H2.5V16.3337ZM10.8333 11.477L15.8925 6.41699L17.0708 7.59533L10 14.667L2.92917 7.59616L4.1075 6.41699L9.16667 11.4753V2.16699H10.8333V11.477Z" fill="white"/>
              </svg>
              Windows
            </a>
          </li>
          <li class="font-bold">
            <a href="/docs/cli/installation/#%F0%9F%8D%8E-macos-install" class="flex gap-4 hocus:!underline">
              <svg xmlns="http://www.w3.org/2000/svg" width="20" height="21" viewBox="0 0 20 21" fill="none">
                <path d="M2.5 16.3337H17.5V18.0003H2.5V16.3337ZM10.8333 11.477L15.8925 6.41699L17.0708 7.59533L10 14.667L2.92917 7.59616L4.1075 6.41699L9.16667 11.4753V2.16699H10.8333V11.477Z" fill="white"/>
              </svg>
              Mac
            </a>
          </li>
          <li class="font-bold">
            <a href="/docs/cli/installation/#%F0%9F%90%A7-linux-install" class="flex gap-4 hocus:!underline">
              <svg xmlns="http://www.w3.org/2000/svg" width="20" height="21" viewBox="0 0 20 21" fill="none">
                <path d="M2.5 16.3337H17.5V18.0003H2.5V16.3337ZM10.8333 11.477L15.8925 6.41699L17.0708 7.59533L10 14.667L2.92917 7.59616L4.1075 6.41699L9.16667 11.4753V2.16699H10.8333V11.477Z" fill="white"/>
              </svg>
              Linux
            </a>
          </li>
        </ul>
      </div>
      <div class="space-y-4 flex-1 max-w-44">
        <div class="text-gray-06">KitOps</div>
        <ul class="space-y-2">
          <li class="font-bold">
            <a href="/docs/why-kitops/#why-use-kitops" class="hocus:!underline">
              Why Kit?
            </a>
          </li>
          <li class="font-bold">
            <a href="/docs/use-cases/#how-kitops-is-used-%F0%9F%9B%A0%EF%B8%8F" class="hocus:!underline">
              How does it work?
            </a>
          </li>
        </ul>
      </div>
      <div class="space-y-4 flex-1 max-w-44 min-w-full md:min-w-min">
        <div class="text-gray-06">Resources</div>
        <ul class="space-y-2">
          <li class="font-bold">
            <a href="/docs/overview/" class="hocus:!underline">
              Docs
            </a>
          </li>
          <li class="font-bold">
            <a href="/blog/" class="hocus:!underline">
              Blog
            </a>
          </li>
        </ul>
      </div>
    </nav>

    <div class="text-center lg:text-left lg:col-span-2 mt-16">
      <div class="flex gap-14 !mt-6 justify-center lg:justify-start">
        <a href="https://discord.gg/Tapeh8agYy" target="_blank" class="hocus:!text-gold" v-ga-track="{ category: 'button', label: 'join the kitops discord', location: 'footer' }">
          <svg xmlns="http://www.w3.org/2000/svg" class="size-6" viewBox="0 0 43 30" fill="currentColor">
            <g clip-path="url(#clip0_8_909)">
              <path d="M36.476 2.67995C33.6728 1.40994 30.7134 0.508951 27.6736 0C27.2576 0.735868 26.8812 1.49298 26.546 2.26816C23.308 1.78531 20.0151 1.78531 16.777 2.26816C16.4416 1.49306 16.0653 0.735957 15.6495 0C12.6076 0.513249 9.64634 1.41638 6.84027 2.68659C1.26951 10.8427
                -0.240636 18.7962 0.514437 26.6368C3.77681 29.022 7.42835 30.8361 11.3103 32C12.1844 30.8366 12.9579 29.6024 13.6225 28.3105C12.3601 27.8439 11.1417 27.2683 9.98138 26.5903C10.2868 26.3711 10.5854 26.1453 10.874 25.9261C14.2504 27.4974 17.9355 28.312 21.6666 28.312C25.3976 28.312
                29.0827 27.4974 32.4591 25.9261C32.751 26.1619 33.0497 26.3877 33.3517 26.5903C32.1891 27.2694 30.9685 27.8461 29.7039 28.3138C30.3677 29.6052 31.1412 30.8383 32.0161 32C35.9014 30.8407 39.5557 29.0276 42.8187 26.6401C43.7046 17.5475 41.3052 9.66708 36.476 2.67995ZM14.5789
                21.8149C12.4748 21.8149 10.7364 19.9253 10.7364 17.6007C10.7364 15.276 12.4144 13.3699 14.5722 13.3699C16.7301 13.3699 18.455 15.276 18.4181 17.6007C18.3811 19.9253 16.7233 21.8149 14.5789 21.8149ZM28.7542 21.8149C26.6467 21.8149 24.915 19.9253 24.915 17.6007C24.915 15.276 26.593
                13.3699 28.7542 13.3699C30.9154 13.3699 32.6269 15.276 32.5899 17.6007C32.553 19.9253 30.8986 21.8149 28.7542 21.8149Z"/>
            </g>
          </svg>
        </a>

        <a href="https://github.com/kitops-ml/kitops" target="_blank" class="hocus:!text-gold" v-ga-track="{ category: 'button', label: 'contribute to kit', location: 'how to get involved' }">
          <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 48 48" fill="none" class="size-6">
            <g clip-path="url(#clip0_79_2469)">
              <path fill-rule="evenodd" clip-rule="evenodd" d="M23.9285 0C10.6967 0 0 11 0 24.6085C0 35.4865 6.85371 44.6945 16.3616 47.9535C17.5504 48.1985 17.9858 47.424 17.9858 46.7725C17.9858 46.202 17.9466 44.2465 17.9466 42.209C11.2903 43.676 9.90417 39.2755 9.90417 39.2755C8.83445 36.4235
                7.24947 35.6905 7.24947 35.6905C5.07086 34.183 7.40816 34.183 7.40816 34.183C9.82482 34.346 11.0929 36.709 11.0929 36.709C13.2318 40.457 16.6785 39.398 18.0651 38.746C18.263 37.157 18.8973 36.057 19.5708 35.446C14.2619 34.8755 8.67625 32.757 8.67625 23.3045C8.67625 20.6155 9.62645
                18.4155 11.1321 16.7045C10.8945 16.0935 10.0624 13.567 11.3701 10.1855C11.3701 10.1855 13.3905 9.5335 17.9461 12.7115C19.8965 12.1728 21.908 11.8988 23.9285 11.8965C25.9489 11.8965 28.0085 12.182 29.9104 12.7115C34.4665 9.5335 36.4869 10.1855 36.4869 10.1855C37.7946 13.567 36.962
                16.0935 36.7244 16.7045C38.2697 18.4155 39.1807 20.6155 39.1807 23.3045C39.1807 32.757 33.5951 34.8345 28.2465 35.446C29.1184 36.22 29.8707 37.6865 29.8707 40.009C29.8707 43.309 29.8315 45.9575 29.8315 46.772C29.8315 47.424 30.2674 48.1985 31.4557 47.954C40.9636 44.694 47.8173 35.4865
                47.8173 24.6085C47.8565 11 37.1207 0 23.9285 0Z" fill="currentColor"/>
            </g>
            <defs>
              <clipPath id="clip0_79_2469">
                <rect width="48" height="48" fill="white"/>
              </clipPath>
            </defs>
          </svg>
        </a>

        <a href="https://x.com/Kit_Ops" target="_blank" class="hocus:!text-gold" v-ga-track="{ category: 'button', label: 'contribute to kit', location: 'how to get involved' }">
          <svg xmlns="http://www.w3.org/2000/svg" class="size-5" viewBox="0 0 21 21" fill="none">
            <g clip-path="url(#clip0_41_3377)">
              <path d="M12.4815 8.6762L19.7013 0.211914H17.9904L11.7215 7.56133L6.71445 0.211914H0.939453L8.51104 11.3255L0.939453 20.2016H2.65042L9.27062 12.4404L14.5584 20.2016H20.3334L12.4811 8.6762H12.4815ZM10.1381 11.4235L9.37092 10.3168L3.2669 1.51093H5.89485L10.8209 8.61754L11.588
                9.7242L17.9913 18.9617H15.3633L10.1381 11.4239V11.4235Z" fill="currentColor"/>
            </g>
          </svg>
        </a>
      </div>

      <ul class="!mt-6 flex justify-center lg:justify-start items-center text-sm font-bold flex-wrap lg:flex-nowrap gap-x-6">
        <li><a href="https://jozu.com/privacy-policy" target="_blank" class="hocus:!underline">Privacy policy</a></li>
        <li class="lg:hidden">
          <div class="w-px h-2.5 bg-gray-05"></div>
        </li>
        <li><a href="https://jozu.com/terms-of-service" target="_blank" class="hocus:!underline">Terms of service</a></li>
        <li><a href="https://jozu.com/acceptable-use" target="_blank" class="hocus:!underline">Acceptable use policy</a></li>
      </ul>

      <div class="flex flex-col lg:flex-row gap-2">
        <p class="font-normal text-off-white !mt-6 text-xs lg:text-sm">KitOps is licensed under the Apache 2.0 open source license.</p>
        <p class="font-normal text-off-white !mt-6 text-xs lg:text-sm">This site conforms to <strong class="block lg:inline">Web Content Accessibility Guidelines 2.1</strong></p>
      </div>

      <svg xmlns="http://www.w3.org/2000/svg" aria-label="CNCF sandbox project logo" width="131" height="36" class="!mt-6 mx-auto lg:mx-0" role="img" viewBox="-3.65 -3.40 358.80 85.80">
        <path id="XMLID_207_" d="M96.9 24.5c2.7 0 4.9-1.1 6.5-3.2l3.5 3.6c-2.8 3.1-6 4.7-9.8 4.7-3.7 0-6.8-1.2-9.3-3.6-2.4-2.4-3.6-5.4-3.6-9 0-3.6 1.2-6.6 3.7-9 2.5-2.4 5.5-3.6 9.1-3.6 4 0 7.3 1.5 10 4.6l-3.4 3.8c-1.7-2.1-3.9-3.2-6.4-3.2-2 0-3.8.7-5.2 2-1.5 1.3-2.2 3.1-2.2 5.4s.7 4.1 2.1 5.4c1.4 1.5 3 2.1 5 2.1z" class="fill-white"/>
        <path id="XMLID_205_" d="M111.2 29.4V5h5.4v19.5H127v4.8h-15.8z" class="fill-white"/>
        <path id="XMLID_157_" d="M150.8 26c-2.5 2.4-5.5 3.6-9.1 3.6-3.6 0-6.6-1.2-9.1-3.6-2.5-2.4-3.7-5.4-3.7-9 0-3.6 1.2-6.6 3.7-9 2.5-2.4 5.5-3.6 9.1-3.6 3.6 0 6.6 1.2 9.1 3.6 2.5 2.4 3.7 5.4 3.7 9 0 3.6-1.3 6.6-3.7 9zm-1.9-9c0-2.2-.7-4-2.1-5.6-1.4-1.5-3.1-2.3-5.1-2.3-2 0-3.7.8-5.1 2.3-1.4 1.5-2.1 3.4-2.1 5.6 0 2.2.7 4 2.1 5.5 1.4 1.5 3.1 2.3 5.1 2.3 2 0 3.7-.8 5.1-2.3 1.4-1.4 2.1-3.3 2.1-5.5z" class="fill-white"/>
        <path id="XMLID_155_" d="M165.5 23.1c.9 1.1 2.1 1.7 3.7 1.7 1.5 0 2.7-.6 3.6-1.7.9-1.1 1.3-2.7 1.3-4.6V5h5.4v13.6c0 3.5-1 6.2-2.9 8.1-2 1.9-4.4 2.8-7.5 2.8-3 0-5.5-1-7.5-2.9s-2.9-4.6-2.9-8.1V5h5.4v13.4c0 2 .5 3.5 1.4 4.7z" class="fill-white"/>
        <path id="XMLID_151_" d="M203.8 8.2c2.3 2.1 3.5 5.1 3.5 8.9s-1.1 6.8-3.4 9-5.7 3.3-10.4 3.3h-8.3V5h8.6c4.3 0 7.7 1.1 10 3.2zm-4 14.5c1.3-1.3 2-3.1 2-5.5s-.7-4.2-2-5.5c-1.3-1.3-3.4-1.9-6.1-1.9h-3v14.8h3.4c2.5-.1 4.4-.7 5.7-1.9z" class="fill-white"/>
        <path id="XMLID_103_" d="M238.2 5h5.4v24.3h-5.4L226.6 14v15.3h-5.4V5h5.1l11.9 15.7V5z" class="fill-white"/>
        <path id="XMLID_100_" d="M267 29.4l-2.3-5.3h-10.2l-2.3 5.3h-5.8L257 5h5.3l10.5 24.3H267zm-7.4-17.1l-3.1 7.1h6.1l-3-7.1z" class="fill-white"/>
        <path id="XMLID_98_" d="M284.1 9.7v19.6h-5.4V9.7h-6.9V5H291v4.7h-6.9z" class="fill-white"/>
        <path id="XMLID_96_" d="M295 5h5.4v24.3H295V5z" class="fill-white"/>
        <path id="XMLID_94_" d="M315.4 20.5L321.6 5h5.9l-9.8 24.3h-4.6L303.3 5h5.9l6.2 15.5z" class="fill-white"/>
        <path id="XMLID_12_" d="M347.9 5v4.8h-12.1v5h10.9v4.6h-10.9v5h12.5v4.8h-17.9V5h17.5z" class="fill-white"/>
        <g>
          <path d="M84.2 68.9l2.4-2.8c3.5 3.2 6.9 4.8 11.7 4.8 4.6 0 7.6-2.4 7.6-5.8V65c0-3.2-1.7-5-8.8-6.5-7.8-1.7-11.4-4.2-11.4-9.8v-.1c0-5.3 4.7-9.3 11.2-9.3 5 0 8.5 1.4 12 4.2l-2.2 3c-3.2-2.6-6.3-3.7-9.8-3.7-4.4 0-7.2 2.4-7.2 5.5v.1c0 3.2 1.7 5 9.2 6.6 7.6 1.7 11.1 4.4 11.1 9.6v.1c0 5.8-4.9 9.6-11.6 9.6-5.7 0-10.1-1.8-14.2-5.4zm47.6-29.3h3.6l15.5 34.2h-4.1l-4-9h-18.5l-4 9h-3.9l15.4-34.2zm9.5 21.8L133.6 44l-7.8 17.3h15.5zm18.8-21.6h3.6l21.4 27.3V39.8h3.7v34h-3.1l-22-27.9v27.9H160v-34zm40.8 0h11.8c10.7 0 18.1 7.3 18.1 16.9v.1c0 9.6-7.4 17-18.1 17h-11.8v-34zm11.8 30.5c8.6 0 14.1-5.8 14.1-13.4v-.1c0-7.5-5.5-13.5-14.1-13.5h-8v26.9h8zm28.6-30.5h14.5c3.9 0 6.9 1.1 8.9 3 1.4 1.5 2.2 3.3 2.2 5.4v.1c0 4.4-2.7 6.7-5.4 7.9 4 1.2 7.3 3.5 7.3 8.2v.1c0 5.8-4.9 9.3-12.3 9.3h-15.1v-34zM255.1 55c4.6 0 7.8-2.1 7.8-6.1v-.1c0-3.4-2.7-5.5-7.5-5.5h-10.3V55h10zm1.4 15.4c5.1 0 8.4-2.3 8.4-6.1v-.1c0-3.7-3.1-5.8-9-5.8h-10.7v12h11.3zm20.8-13.5c0-9.5 7-17.7 17.4-17.7 10.3 0 17.3 8.1 17.3 17.5v.1c0 9.4-7 17.6-17.4 17.6s-17.3-8.1-17.3-17.5zm30.7 0c0-7.8-5.6-14.1-13.4-14.1-7.8 0-13.3 6.2-13.3 13.9v.1c0 7.7 5.6 14 13.4 14 7.7.1 13.3-6.2 13.3-13.9zm22.6-.4L318 39.8h4.5l10.4 13.9 10.3-13.9h4.5l-12.6 16.6 13.1 17.4h-4.6l-10.9-14.6-10.9 14.6h-4.4l13.2-17.3z" class="fill-white"/>
        </g>
        <g>
          <path d="M13.6 14.1V16v-1.9zm16.1 51.3c.1 0 0 0 0 0z" class="fill-white"/>
          <path d="M27.3 14.1l-13.7.1V16l.1 11.8-.1 9v14.8c1.8-1.5 3.5-3.3 5.4-4.8 2.8-2.4 6-3.5 9.5-3.1l4.3 1.8c1.1-1.6 2.8-2.3 4.8-2.3 1.6 0 3.3.2 4.9.2.2-.6.4-1.1.5-1.7.5-2.9 1.6-5.5 2.5-8.2 1.6-5.1 3-10.3 4.9-15.2.1-.9 1-1.1 1.5-1.7h2.5l.7.7c1 .8 1.7 1.7 1.1 3.1l-8.7 22.8c.2 1.1 1.1.7 1.7.7 2.8.1 3.4.8 3 3.5-.1.9 0 1.9 0 2.8 1.5-.6 2.9-1.2 4.5-1.6 3-.7 5.7.2 8.1 2l.1-7.9V14H36.3l-9 .1z" class="fill-white opacity-20" />
          <path d="M53 20.4zM32.2 57.2c.1 0 .1 0 0 0 .1 0 .1 0 0 0z" class="fill-white"/>
          <path d="M65 65.3v-.1.1zm-13.7.1c5.8 0 7.8 4.3 13.5 4.3.3 0-.1-4.5.1-4.5V51.8l-.1-.1v-1.1c-2.5-1.9-5.1-2.8-8.1-2-1.5.4-3 .9-4.5 1.6 0 .5-.1 1-.3 1.5-1 2.6-2.7 4.9-3.9 7.4-.3.7-.7 1.8-1.8 1.5-.1 0-.2 0-.3-.1-1.7.1-3.5.2-4.9-1.1-.5-.5-1.2-.2-1.8-.3-1.7.2-3.3.3-5 .7-1.4-.1-2.9 1-4.2 1.1-.9.5-3.1 1-4.1-.2h-.1s-.1 0-.1-.1l.3-.5c1.5-3.9 4.5-10.2 6.2-13.7.1-.3.3-.6.5-.8l-4.3-1.8c-3.5-.5-6.7.7-9.5 3.1-1.8 1.5-3.5 3.3-5.4 4.8V65.5h1.2c.1-.1-1.6 1.3-1.5 1.2 4.9 0 9.1-1.2 14-1.2" class="fill-white opacity-40" />
          <path d="M46.7 65.4H65v-.2h-.8m-38.4-4.5c.1 0 .1 0 0 0 .1 0 .1 0 0 0 0-.1 0 0 0 0zm20.4-.1z" class="fill-white"/>
          <path d="M53 20.3c0 .1 0 .1 0 0 .3-.4.5-.9.7-1.3.2-1.2-.6-1.9-1.4-2.5h-.4c-.5.6-1.4.8-1.5 1.7l.2.4c2.2-.6 2.3-.5 2.4 1.7z" class="fill-white opacity-80" />
          <path d="M53.7 19l.4.4c.9-.4-.1-1.8 1-2.1l-.7-.7h-2.1c.8.5 1.6 1.2 1.4 2.4z" class="fill-white opacity-80" />
          <path d="M43.4 41.5c2.1.4 2.3-1.2 2.8-2.5 2.7-6.1 5.1-12.2 6.8-18.6v-.1c-.1-2.2-.3-2.3-2.4-1.7 0 0-.1-.3-.2-.4-1.9 5-3.3 10.2-4.9 15.2-.9 2.7-2 5.4-2.5 8.2.2 0 .3-.1.4-.1zM26.1 60l3.8-6.7c1.1-2.3 2.1-4.6 3.3-6.7.7-1 1.6-1.9 2.9-2.1.2-.1.4-.1.6-.2 2.1-.5 4.1 0 6.2.2 2.3.1 4.7.2 7 .3 1.4.1 2 1.1 1.5 2.4l-.3.6c-.9 2.1-1.4 3.1-3.1 3.1 1.3.3 1.4 1.1.7 2.7s-1.6 3.3-2.2 5c-.2.6-.4 1.1-.4 1.7.1 0 .2.1.3.1 1-3.3 2.8-6.1 4.3-9.6.8-1.3.8-1.7 1.5-2.5v3.2c.2-.5.3-1 .3-1.5 0-.9-.2-1.9 0-2.8.4-2.7-.2-3.4-3-3.5-.6 0-1.5.4-1.7-.7-.4.4-.7 1.1-1.3.3-1.2.3-2.5.1-3.7-.1-1.6-.1-3.3-.3-4.9-.2-2 0-3.7.7-4.8 2.3-.2.3-.4.5-.5.8-2 3.7-5 10-6.5 13.9z" class="fill-white opacity-60" />
          <path d="M46.3 43.5c.9-3.2 2.4-6.2 3.6-9.3 1.1-3.1 2.4-6.1 3.2-9.2.5-1.9 1.3-3.6 1.1-5.6l-.4-.4c-.2.5-.4.9-.7 1.4-1.8 6.4-4.2 12.6-6.9 18.6-.6 1.3-.8 2.9-2.8 2.5-.1 0-.2.2-.2.2-.2.6-.4 1.1-.5 1.7 1.1.1 2.3.4 3.6.1z" class="fill-white opacity-60" />
          <path d="M32.2 57.2c-.1-.1-.1-.2-.2-.3l-2.3-3.5v-.1L26.1 60l-.3.5s.1 0 .1.1c1 1.1 3.2.7 4.1.2l2-3.2c.1-.1.2-.2.2-.4zm21.9-37.8c.3 2-.6 3.8-1.1 5.6-.8 3.2-2.1 6.2-3.2 9.2-1.1 3.1-2.6 6.1-3.6 9.3.7.8.9 0 1.3-.3l8.7-22.8c.5-1.4-.1-2.3-1.1-3.1-1.2.3-.1 1.7-1 2.1zM50.5 51c-1.5 3.4-3.3 6.2-4.3 9.6h.1c1.1.3 1.5-.7 1.8-1.5 1.2-2.5 2.9-4.8 3.9-7.4v-3.2c-.7.8-.8 1.2-1.5 2.5z" class="fill-white opacity-80" />
          <path d="M48.5 53.8c.8-1.7.6-2.5-.7-2.7-.3 0-.6 0-.9-.1-2.8-.1-5.5-.3-8.3-.4-.8 0-1.6-.1-2 .7 1.2 2.5 2.4 4.9 2.5 7.8.6.1 1.3-.1 1.8.3 1.4 1.4 3.2 1.2 4.9 1.1 0-.6.2-1.1.4-1.7.8-1.7 1.6-3.3 2.3-5z" class="fill-white opacity-60" />
          <path d="M34.6 53.5v0c.9-1 1.3-1.5 2.1-2.2.5-.8 1.3-.7 2-.7 2.8.1 5.5.3 8.3.4.3.1.6.1.9.1 1.7 0 2.2-1 3.1-3.1l.3-.6c.5-1.3-.1-2.3-1.5-2.4-2.3-.2-4.7-.2-7-.3-2.1-.2-4.1-.7-6.2-.2-.2 0-.4.1-.6.2-1.3.2-2.1 1.1-2.9 2.1 1.9 1.8 3 4 1.5 6.7z" class="fill-white opacity-80" />
          <path d="M32.1 57c0 .1.1.2.2.2.9-1.3 1.4-2.3 2.3-3.6v-.2c1.5-2.7.4-4.9-1.5-6.8-1.2 2.1-2.2 4.4-3.3 6.7v.1l2.3 3.6z" class="fill-white opacity-80" />
          <path d="M36.7 51.3c-.8.7-1.2 1.2-2 2.3l-.1.1c-.9 1.2-1.5 2.3-2.3 3.6-.1.1-.2.2-.3.4l-2 3.2c1.3-.1 2.8-1.2 4.2-1.1 1.6-.4 3.3-.5 5-.7-.1-2.9-1.4-5.4-2.5-7.8z" class="fill-white opacity-60" />
          <path d="M27.3 65.4H13.6V51.7H3.3v24h24zm37.6-13.7l.1 13.5v.1l-13.7.1v10.3h24v-24H64.9zM13.6 16v-1.9h13.7V3.7h-24v24h10.4zM51.3 3.7V14H65v13.7h10.3v-24z" class="fill-white"/>
          <path d="M36.3 14h15V3.7h-24v10.4zm28.6 37.7h10.4v-24H65v15m-51.4 8.9V36.7l.1-9H3.3v24h10.3zm22.7 24.1h15V65.4h-24v10.3z" class="fill-white opacity-40" />
        </g>
      </svg>

      <p class="font-normal text-off-white !mt-2 text-xs lg:text-sm">Copyright © KitOps a Series of LF Projects, LLC For website terms of use, trademark policy and other project policies please see <a href="https://lfprojects.org/policies/" class="!text-gold hocus:!underline" target="_blank">lfprojects.org/policies/</a></p>
    </div>
  </div>
</footer>
</template>
