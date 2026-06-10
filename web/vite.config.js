import { sveltekit } from '@sveltejs/kit/vite';
import tailwindcss from '@tailwindcss/vite';
import { defineConfig } from 'vite';
import { execSync } from 'child_process';
import { fileURLToPath } from 'node:url';

const commitHash = execSync('git rev-parse --short HEAD').toString().trim();
const commitDate = execSync('git log -1 --format=%cd --date=short').toString().trim();

// rombik-engine експортується умовно (rombik-source → TS-джерела, default → dist).
// Для вебу резолвимо ПРЯМО на джерела аліасом — хірургічно, без глобального override
// resolve.conditions (той викидав дефолтні module/browser і ламав резолв Svelte/
// CodeMirror → серверні ентрі потрапляли в браузерний бандл, вікно коду не рендерилось).
const engineSrc = fileURLToPath(new URL('../packages/engine/src/index.ts', import.meta.url));

export default defineConfig({
	plugins: [tailwindcss(), sveltekit()],
	define: {
		__COMMIT_HASH__: JSON.stringify(commitHash),
		__COMMIT_DATE__: JSON.stringify(commitDate)
	},
	resolve: {
		alias: { 'rombik-engine': engineSrc }
	},
	optimizeDeps: {
		// rombik-engine — workspace-джерело (TS), не пребандлити. web-tree-sitter
		// застосунок не імпортує (бере статичний tree-sitter.js), а в npm-workspace
		// він hoisted у корінь — пребандл vite падав ENOENT у web/node_modules.
		exclude: ['rombik-engine', 'web-tree-sitter']
	}
});
