<script setup lang="ts">
import { useSidebar } from 'vitepress/theme'
import { useData } from 'vitepress/client'

const { theme, frontmatter, page, site } = useData()
const { hasSidebar } = useSidebar()
const year = (new Date()).getFullYear()

</script>

<template>
<footer
  v-if="theme.footer && frontmatter.footer !== false"
  class="bg-black px-12 py-24 xl:px-16 text-center"
  :class="{ 'hidden': hasSidebar }">

  <div class="flex flex-col md:flex-row gap-16 items-center justify-center">
    <img src="/logo.svg" class="w-32" alt="kitops logo" />

    <div class="flex items-center gap-4">
      <span aria-label="made with love by Jozu">Made with <span class="text-red-500">♥️</span> by</span>
      <a href="https://jozu.com" target="_blank">
        <svg xmlns="http://www.w3.org/2000/svg"  width="56" height="18" viewBox="0 0 56 18" fill="none" title="Jozu">
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

  <nav
    v-if="theme.nav"
    class="text-sm flex xs:flex-wrap justify-center items-center gap-x-8 gap-y-4 md:gap-14 my-16">
    <template v-for="item in theme.nav" :key="item.text">
      <a v-if="'link' in item" :href="item.link" class="font-bold">
        {{ item.text }}
      </a>
    </template>
  </nav>

  <div class="text-center space-y-6 text-sm font-bold">
    <ul class="flex items-center gap-10 md:gap-14 justify-center">
      <li>
        <a href="https://discord.gg/3eDb4yAN" target="_blank">
          <span class="sr-only">discord</span>
          <svg xmlns="http://www.w3.org/2000/svg" width="25" height="18" fill="none">
            <path fill="#ECECEC" d="M20.83 1.507A20.002 20.002 0 0 0 15.88 0c-.234.414-.446.84-.634 1.276a18.631 18.631 0 0 0-5.495 0C9.56.84 9.35.414 9.115 0A20.143 20.143 0 0 0 4.16 1.511C1.027 6.1.177 10.573.602 14.983A20.04 20.04 0 0 0 6.675 18a14.49 14.49 0 0 0 1.3-2.075 12.98 12.98 0 0 1-2.048-.968c.172-.123.34-.25.502-.374a14.39 14.39 0 0 0 12.142 0c.164.133.332.26.502.374-.654.382-1.34.707-2.052.97.373.726.808 1.42 1.3 2.073a19.952 19.952 0 0 0 6.077-3.015c.498-5.115-.851-9.547-3.568-13.478ZM8.513 12.271c-1.183 0-2.161-1.063-2.161-2.37 0-1.308.944-2.38 2.157-2.38 1.214 0 2.184 1.072 2.164 2.38-.021 1.307-.954 2.37-2.16 2.37Zm7.974 0c-1.186 0-2.16-1.063-2.16-2.37 0-1.308.944-2.38 2.16-2.38 1.215 0 2.178 1.072 2.157 2.38-.02 1.307-.951 2.37-2.157 2.37Z"/>
          </svg>
        </a>
      </li>
      <li>
        <a href="https://github.com/kitops-ml/kitops" target="_blank">
          <span class="sr-only">github</span>
          <svg xmlns="http://www.w3.org/2000/svg" width="25" height="24" fill="none" title="">
            <path fill="#ECECEC" fill-rule="evenodd" d="M12.464 0C5.848 0 .5 5.5.5 12.304c0 5.44 3.427 10.043 8.18 11.673.595.122.813-.265.813-.59 0-.286-.02-1.264-.02-2.282-3.328.733-4.02-1.467-4.02-1.467-.536-1.426-1.328-1.793-1.328-1.793-1.09-.753.08-.753.08-.753 1.207.081 1.841 1.263 1.841 1.263 1.07 1.873 2.793 1.344 3.487 1.018.099-.795.416-1.345.752-1.65-2.654-.285-5.447-1.345-5.447-6.07 0-1.345.475-2.445 1.228-3.3-.119-.306-.535-1.57.12-3.26 0 0 1.01-.326 3.287 1.263.975-.27 1.981-.407 2.991-.408 1.01 0 2.04.143 2.991.408 2.278-1.59 3.288-1.263 3.288-1.263.654 1.69.238 2.954.12 3.26.772.855 1.227 1.955 1.227 3.3 0 4.725-2.792 5.764-5.467 6.07.436.387.812 1.12.812 2.282 0 1.65-.02 2.974-.02 3.381 0 .326.219.713.813.591 4.754-1.63 8.18-6.234 8.18-11.673C24.429 5.5 19.06 0 12.465 0Z" clip-rule="evenodd"/>
          </svg>
        </a>
      </li>
      <li>
        <a href="https://x.com/Kit_Ops" target="_blank">
          <span class="sr-only">x</span>
          <svg width="24" height="24" fill="none" xmlns="http://www.w3.org/2000/svg">
            <g clip-path="url(#a)">
              <path d="M13.482 10.464 20.701 2H18.99l-6.268 7.35L7.714 2H1.94l7.572 11.114L1.94 21.99h1.71l6.62-7.761 5.288 7.76h5.775l-7.852-11.525Zm-2.344 2.747-.767-1.106-6.104-8.806h2.628l4.926 7.107.767 1.106 6.403 9.238h-2.628l-5.225-7.538Z" fill="#ECECEC"/>
            </g>
            <defs>
              <clipPath id="a">
                <path fill="#fff" d="M1.94 2h19.393v20H1.94z"/>
              </clipPath>
            </defs>
          </svg>
        </a>
      </li>
      <li>
        <a href="https://www.youtube.com/@Jozu_ml" target="_blank">
          <svg xmlns="http://www.w3.org/2000/svg" width="32" height="32" viewBox="0 0 256 256" xml:space="preserve">
            <path d="M88.119 23.338a11.273 11.273 0 0 0-7.957-7.957C73.144 13.5 45 13.5 45 13.5s-28.144 0-35.162 1.881a11.273 11.273 0 0 0-7.957 7.957C0 30.356 0 45 0 45s0 14.644 1.881 21.662a11.273 11.273 0 0 0 7.957 7.957C16.856 76.5 45 76.5 45 76.5s28.144 0 35.162-1.881a11.273 11.273 0 0 0 7.957-7.957C90 59.644 90 45 90 45s0-14.644-1.881-21.662zM36 58.5v-27L59.382 45 36 58.5z" fill="#ECECEC" transform="matrix(2.81 0 0 2.81 1.407 1.407)"/>
          </svg>
        </a>
      </li>
      <li>
        <a href="https://dev.to/kitops" target="_blank">
          <svg xmlns="http://www.w3.org/2000/svg" width="32" height="32" viewBox="0 0 448 512">
            <path d="M120.12 208.29c-3.88-2.9-7.77-4.35-11.65-4.35H91.03v104.47h17.45c3.88 0 7.77-1.45 11.65-4.35 3.88-2.9 5.82-7.25 5.82-13.06v-69.65c-.01-5.8-1.96-10.16-5.83-13.06zM404.1 32H43.9C19.7 32 .06 51.59 0 75.8v360.4C.06 460.41 19.7 480 43.9 480h360.2c24.21 0 43.84-19.59 43.9-43.8V75.8c-.06-24.21-19.7-43.8-43.9-43.8zM154.2 291.19c0 18.81-11.61 47.31-48.36 47.25h-46.4V172.98h47.38c35.44 0 47.36 28.46 47.37 47.28l.01 70.93zm100.68-88.66H201.6v38.42h32.57v29.57H201.6v38.41h53.29v29.57h-62.18c-11.16.29-20.44-8.53-20.72-19.69V193.7c-.27-11.15 8.56-20.41 19.71-20.69h63.19l-.01 29.52zm103.64 115.29c-13.2 30.75-36.85 24.63-47.44 0l-38.53-144.8h32.57l29.71 113.72 29.57-113.72h32.58l-38.46 144.8z" fill="#ECECEC"/>
          </svg>
        </a>
      </li>
    </ul>

    <ul class="flex flex-wrap lg:flex-nowrap items-center justify-center text-white gap-x-4 gap-y-2 lg:gap-10 list-none">
      <li><a href="https://jozu.com/privacy-policy" target="_blank">Privacy policy</a></li>
      <li>
        <div class="w-px h-2.5 bg-gray-05"></div>
      </li>
      <li><a href="https://jozu.com/terms-of-service" target="_blank">Terms of service</a></li>
      <li class="hidden lg:block">
        <div class="w-px h-2.5 bg-gray-05"></div>
      </li>
      <li><a href="https://jozu.com/acceptable-use" target="_blank">Acceptable use policy</a></li>
    </ul>

    <p class="font-normal text-off-white">This site conforms to <strong>Web Content Accessibility Guidelines 2.1</strong></p>
    <p class="font-normal">
      Copyright © KitOps a Series of LF Projects, LLC For website terms of use, trademark policy and other project policies please see <a href="https://lfprojects.org/policies/" class="text-gold hover:underline" target="_blank">lfprojects.org/policies/</a>
    </p>
  </div>
</footer>
</template>
