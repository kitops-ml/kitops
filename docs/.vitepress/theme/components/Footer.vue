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

    <div class="flex items-center p2 gap-4">
      <strong aria-label="made with love by Jozu">Made with <span class="text-red-500 text-2xl">♥️</span> by</strong>
      <a href="https://jozu.com" target="_blank">
        <svg xmlns="http://www.w3.org/2000/svg" width="48" height="48" viewBox="0 0 48 48" fill="none" title="Jozu">
          <path d="M4.5 5.29955L23.885 0L43.27 5.29955V48C43.27 48 32.924 42.7004 23.885 42.7004C14.846 42.7004 4.5 48 4.5 48V5.29955Z" fill="white"/>
          <path d="M15.1993 21.6987C13.7519 21.6987 12.6008 21.3457 11.7462 20.6397C10.8915 19.9337 10.3194 18.9923 10.03 17.8156L13.0282 16.7151C13.1109 17.1027 13.2349 17.4834 13.4004 17.8572C13.5796 18.2171 13.8139 18.5147 14.1034 18.7501C14.4067 18.9854 14.772 19.1031 15.1993 19.1031C15.8885 19.1031 16.4124 18.8885 16.7708 18.4594C17.143 18.0164 17.3291 17.2411 17.3291 16.1337V7.57852H20.5961V16.1337C20.5961 17.8641 20.1343 19.2277 19.2107 20.2244C18.3009 21.2073 16.9638 21.6987 15.1993 21.6987Z" fill="black"/>
          <path d="M30.5422 21.7153C29.1361 21.7153 27.8955 21.4177 26.8203 20.8224C25.745 20.2272 24.9042 19.3897 24.2976 18.3099C23.6911 17.2301 23.3878 15.9704 23.3878 14.5307C23.3878 13.091 23.6911 11.8312 24.2976 10.7514C24.9042 9.67166 25.745 8.83414 26.8203 8.23888C27.8955 7.64361 29.1361 7.34598 30.5422 7.34598C31.962 7.34598 33.2096 7.64361 34.2848 8.23888C35.36 8.83414 36.2009 9.67166 36.8074 10.7514C37.4277 11.8312 37.7379 13.091 37.7379 14.5307C37.7379 15.9704 37.4277 17.2301 36.8074 18.3099C36.2009 19.3897 35.36 20.2272 34.2848 20.8224C33.2096 21.4177 31.962 21.7153 30.5422 21.7153ZM30.5422 19.0159C31.2176 19.0159 31.838 18.8498 32.4031 18.5175C32.9683 18.1714 33.4163 17.6662 33.7472 17.0017C34.0918 16.3372 34.2641 15.5135 34.2641 14.5307C34.2641 13.5478 34.0918 12.7241 33.7472 12.0596C33.4163 11.3951 32.9683 10.8968 32.4031 10.5646C31.838 10.2185 31.2176 10.0454 30.5422 10.0454C29.8805 10.0454 29.2671 10.2185 28.7019 10.5646C28.1367 10.8968 27.6818 11.3951 27.3372 12.0596C27.0063 12.7241 26.8409 13.5478 26.8409 14.5307C26.8409 15.5135 27.0063 16.3372 27.3372 17.0017C27.6818 17.6662 28.1367 18.1714 28.7019 18.5175C29.2671 18.8498 29.8805 19.0159 30.5422 19.0159Z" fill="black"/>
          <path d="M10.4857 38.4132V35.9007L17.454 27.4286H10.7959V24.5422H21.6929V27.0548L14.7866 35.5269H22.003V38.4132H10.4857Z" fill="black"/>
          <path d="M30.5157 38.6344C28.3515 38.6344 26.8145 38.1153 25.9047 37.0771C25.0087 36.0388 24.5607 34.4538 24.5607 32.3219V24.5143H27.8277V32.5918C27.8277 33.2702 27.8897 33.8654 28.0138 34.3776C28.1516 34.876 28.4135 35.2636 28.7995 35.5405C29.1855 35.8173 29.7576 35.9558 30.5157 35.9558C31.2739 35.9558 31.8391 35.8173 32.2113 35.5405C32.5973 35.2636 32.8523 34.876 32.9763 34.3776C33.1142 33.8654 33.1831 33.2702 33.1831 32.5918V24.5143H36.4501V32.3219C36.4501 34.4538 35.9952 36.0388 35.0854 37.0771C34.1756 38.1153 32.6524 38.6344 30.5157 38.6344Z" fill="black"/>
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
