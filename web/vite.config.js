import { sveltekit } from '@sveltejs/kit/vite';
import tailwindcss from '@tailwindcss/vite';
import { defineConfig } from 'vite';
import { execSync } from 'child_process';

const commitHash = execSync('git rev-parse --short HEAD').toString().trim();
const commitDate = execSync('git log -1 --format=%cd --date=short').toString().trim();

export default defineConfig({
	plugins: [tailwindcss(), sveltekit()],
	define: {
		__COMMIT_HASH__: JSON.stringify(commitHash),
		__COMMIT_DATE__: JSON.stringify(commitDate)
	},
	// rombik-engine експортується умовно: «rombik-source» → TS-джерела (для монорепо),
	// «default» → зібраний dist (для npm-споживачів). Веб бере джерела (HMR на правки рушія).
	resolve: { conditions: ['rombik-source'] },
	ssr: { resolve: { conditions: ['rombik-source'] } },
	optimizeDeps: {
		// rombik-engine — workspace-джерело (TS), обробляти як source, не пребандлити.
		// web-tree-sitter застосунок не імпортує (бере статичний tree-sitter.js), а в
		// npm-workspace він hoisted у корінь — пребандл vite падав ENOENT у web/node_modules.
		exclude: ['rombik-engine', 'web-tree-sitter']
	}
});
