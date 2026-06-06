import adapter from '@sveltejs/adapter-static';
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';

/** Повністю статичний сайт (без сервера). SPA-fallback для клієнтського роутингу.
 *  BASE_PATH порожній для власного домену; для github.io/<repo> постав /<repo>. */
const config = {
	preprocess: vitePreprocess(),
	kit: {
		adapter: adapter({ fallback: '200.html' }),
		paths: { base: process.env.BASE_PATH ?? '' }
	}
};

export default config;
