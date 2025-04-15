// https://vitepress.dev/guide/custom-theme
import { h, ref } from 'vue'
import { type Theme, inBrowser } from 'vitepress'
import DefaultTheme from 'vitepress/theme'
import PlatformSelect from './components/PlatformSelect.vue'
import PlatformSnippet from './components/PlatformSnippet.vue'
import GithubStartButton from './components/GithubStartButton.vue'
import DiscordBanner from './components/DiscordBanner.vue'
import './assets/css/fonts.css'
import './assets/css/tailwind.css'
import './style.css'

const isPlatformModalOpen = ref(false)

const isProd = import.meta.env.PROD

export default {
  extends: DefaultTheme,
  Layout: () => {
    if (inBrowser) {
      // Google tag manager
      const noscript = document.createElement('noscript')
      const ifr = document.createElement('iframe')
      ifr.src = 'https://www.googletagmanager.com/ns.html?id=GTM-TFFZXCQW';
      ifr.height = 0;
      ifr.width = 0;
      ifr.style = 'display:none;visibility:hidden;';
      noscript.appendChild(ifr)
      document.body.insertBefore(noscript, document.body.firstElementChild)
    }

    return h(DefaultTheme.Layout, null, {
      // https://vitepress.dev/guide/extending-default-theme#layout-slots
      'sidebar-nav-after': () => h(PlatformSelect),
      'nav-bar-content-after': () => h(GithubStartButton, {
        class: 'ml-4 pt-2'
      }),
      'doc-after': () => h('div', {
        class: 'mt-20 text-off-white',
        style: { 'font-size': '0.8rem' }
      }, [
        h('p', [
          'Copyright © KitOps a Series of LF Projects, LLC For website terms of use, trademark policy and other project policies please see ',
          h('a', {
            href: 'https://lfprojects.org/policies/',
            target: '_blank',
            class: 'text-gold hover:underline'
          }, 'lfprojects.org/policies/')
        ]),
      ]),
    })
  },
  enhanceApp({ app, router, siteData }) {
    app.component('PlatformSnippet', PlatformSnippet)
    app.provide('isPlatformModalOpen', isPlatformModalOpen)
    app.component('DiscordBanner', DiscordBanner);
  }
} satisfies Theme
