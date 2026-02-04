const esbuild = require('esbuild')
const postcss = require('postcss')
const tailwindcss = require('@tailwindcss/postcss')
const fs = require('fs')
const path = require('path')

const args = process.argv.slice(2)
const isDev = args.includes('--dev')
const isWatch = args.includes('--watch')

// PostCSS plugin for esbuild
const postcssPlugin = {
  name: 'postcss',
  setup(build) {
    build.onLoad({ filter: /\.css$/ }, async (args) => {
      const css = await fs.promises.readFile(args.path, 'utf8')

      const result = await postcss([tailwindcss()]).process(css, {
        from: args.path,
        to: path.join(__dirname, 'static', path.basename(args.path)),
      })

      return {
        contents: result.css,
        loader: 'css',
      }
    })
  },
}

const buildOptions = {
  entryPoints: ['./assets/scripts.js', './assets/styles.css'],
  bundle: true,
  outdir: './static',
  minify: !isDev,
  sourcemap: isDev,
  plugins: [postcssPlugin],
  logLevel: 'info',
}

async function build() {
  try {
    if (isWatch) {
      const ctx = await esbuild.context(buildOptions)
      await ctx.watch()
      console.log('Watching for changes...')
    } else {
      await esbuild.build(buildOptions)
      console.log('Build complete!')
    }
  } catch (error) {
    console.error('Build failed:', error)
    process.exit(1)
  }
}

build()
