import * as cheerio from 'cheerio'

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

const postsUrls: Post[] = require('./posts.json')

async function getPostData(post: Post) {
  try {
    const html = await (await fetch(post.url)).text()

    // Using cheerio to parse the html into actual dom nodes that we can interact.
    const $ = cheerio.load(html)

    // Tiny helper
    const getMetaTag = (name: keyof Post) => (
      post[name] ||
      $(`meta[name=${name}]`).attr("content") ||
      $(`meta[property="og:${name}"]`).attr("content") ||
      $(`meta[property="twitter${name}"]`).attr("content")
    )

    const title = getMetaTag('title') || $('title').text()
    const description = getMetaTag('description')
    const site_name = getMetaTag('site_name')
    const image = getMetaTag('image') || $('meta[property="og:image:url"]').attr('content')
    const icon = $('link[rel="icon"]').attr('href') || $('link[rel="shortcut icon"]').attr('href') || $('link[rel="alternate icon"]').attr('href')
    const author = getMetaTag('author')
    const published_time = post.published_time || $('meta[property="article:published_time"]').attr('content')
    const tags = post.tags || []

    return {
      url: post.url,
      tags,
      author,
      title,
      description,
      published_time,
      site_name,
      image,
      icon
    } as Post
  }
  catch (e) {
    console.error(`failed to fetch: ${post.url}`, e)
  }
}

export default {
  async load() {
    const posts = []

    // Let's slow down these fetch's...
    for (const post of postsUrls) {
      console.log(`fetching blog post ${post.url}`)
      const data = await getPostData(post)

       // Skip the post that has no url, which is probably a 404 page.
      if (!data ||!data.url) continue;

      posts.push(data)
    }

    return Promise.all(posts)
  }
}
