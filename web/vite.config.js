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
	}
});
