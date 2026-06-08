<script>
	import { onMount } from 'svelte';
	import { EditorView, basicSetup } from 'codemirror';
	import { EditorState, Compartment } from '@codemirror/state';
	import { python } from '@codemirror/lang-python';
	import { oneDark } from '@codemirror/theme-one-dark';

	let { value = $bindable('') } = $props();
	let el;
	let view;
	const themeC = new Compartment(); // окремий слот під тему — перемикаємо на льоту

	const isDark = () => document.documentElement.classList.contains('dark');
	const themeExt = () => (isDark() ? oneDark : []);

	let isInternal = false;
	
	onMount(() => {
		view = new EditorView({
			parent: el,
			state: EditorState.create({
				doc: value,
				extensions: [
					basicSetup,
					python(),
					themeC.of(themeExt()),
					EditorView.updateListener.of((u) => {
						if (u.docChanged) {
							isInternal = true;
							value = u.state.doc.toString();
							setTimeout(() => { isInternal = false; }, 0);
						}
					}),
					EditorView.theme({
						'&': { height: '100%', fontSize: '13px' },
						'.cm-scroller': { overflow: 'auto', fontFamily: 'ui-monospace, monospace' },
						'&.cm-focused': { outline: 'none' }
					})
				]
			})
		});
		const obs = new MutationObserver(() => {
			view.dispatch({ effects: themeC.reconfigure(themeExt()) });
		});
		obs.observe(document.documentElement, { attributes: true, attributeFilter: ['class'] });
		return () => {
			obs.disconnect();
			view.destroy();
		};
	});

	$effect(() => {
		if (view && !isInternal && value !== view.state.doc.toString()) {
			view.dispatch({ changes: { from: 0, to: view.state.doc.length, insert: value } });
		}
	});
</script>

<div bind:this={el} class="h-full min-h-0 overflow-hidden"></div>
