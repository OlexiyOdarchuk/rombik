import adapter from '@sveltejs/adapter-static';
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';

/** Повністю статичний сайт (без сервера). SPA-fallback для клієнтського роутингу. */
const config = {
	preprocess: vitePreprocess(),
	kit: {
		adapter: adapter({ fallback: '200.html' })
	}
};

export default config;
