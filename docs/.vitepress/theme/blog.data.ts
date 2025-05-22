import * as cheerio from 'cheerio'
import fs from 'fs'
import path from 'path'
import crypto from 'crypto'
import { fileURLToPath } from 'url'

export type Post = {
  title: string,
  author: string,
  description: string,
  published_time: string,
  site_name: string,
  image: string,
  icon: string,
  url: string,
  tags: string[]
}
const __filename = fileURLToPath(import.meta.url)
const __dirname = path.dirname(__filename)

const POSTS_PATH = path.resolve(__dirname, './posts.json')
const CACHE_PATH = path.resolve(__dirname, './posts.cache.json')

// get a sha to use as digest based on the file contents
function getFileDigest(filePath: string) {
  const fileBuffer = fs.readFileSync(filePath)
  return crypto.createHash('sha1').update(fileBuffer).digest('hex')
}

// load the cache with the given digest
function loadCache(digest: string): Post[] | null {
  // No cache? nothing to do then
  if (!fs.existsSync(CACHE_PATH)) {
    return null
  }
  const { hash, posts } = JSON.parse(fs.readFileSync(CACHE_PATH, 'utf-8'))
  return hash === digest ? posts : null
}

// Save the given posts with the given digest to the cache
function saveCache(digest: string, posts: Post[]) {
  if (!fs.existsSync(POSTS_PATH)) {
    fs.mkdirSync(POSTS_PATH)
  }
  fs.writeFileSync(CACHE_PATH, JSON.stringify({ hash: digest, posts }, null, 2))
}

async function getPostData(post: Post): Promise<Post | undefined> {
  try {
    const html = await (await fetch(post.url)).text()
    const $ = cheerio.load(html)

    const getMetaTag = (name: keyof Post) => (
      post[name] ||
      $(`meta[name=${name}]`).attr("content") ||
      $(`meta[property="og:${name}"]`).attr("content") ||
      $(`meta[property="twitter${name}"]`).attr("content")
    )

    return {
      url: post.url,
      tags: post.tags || [],
      author: getMetaTag('author'),
      title: getMetaTag('title') || $('title').text(),
      description: getMetaTag('description'),
      published_time: post.published_time || $('meta[property="article:published_time"]').attr('content'),
      site_name: getMetaTag('site_name'),
      image: getMetaTag('image') || $('meta[property="og:image:url"]').attr('content'),
      icon: $('link[rel="icon"]').attr('href') || $('link[rel="shortcut icon"]').attr('href') || $('link[rel="alternate icon"]').attr('href')
    }
  } catch (e) {
    console.error(`failed to fetch: ${post.url}`, e)
    return undefined
  }
}

export default {
  async load() {
    const digest = getFileDigest(POSTS_PATH)
    const cached = loadCache(digest)

    if (cached) {
      console.log('Valid cache found, using cached posts')
      return cached
    }

    const postsUrls: Post[] = JSON.parse(fs.readFileSync(POSTS_PATH, 'utf-8'))
    const posts: Post[] = []

    for (const post of postsUrls) {
      console.log(`fetching blog post ${post.url}`)
      const data = await getPostData(post)

       // Skip the post that has no url, which is probably a 404 page.
      if (!data?.url) continue;

      posts.push(data)
    }

    // Update (or create) the cache with the posts
    saveCache(digest, posts)

    return posts
  }
}
