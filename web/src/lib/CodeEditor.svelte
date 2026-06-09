<script>
	import { onMount } from 'svelte';
	import { EditorView, basicSetup } from 'codemirror';
	import { EditorState, Compartment } from '@codemirror/state';
	import { python } from '@codemirror/lang-python';
	import { cpp } from '@codemirror/lang-cpp';
	import { oneDark } from '@codemirror/theme-one-dark';

	let { value = $bindable(''), lang = 'python' } = $props();
	let el;
	let view;
	const themeC = new Compartment(); // окремий слот під тему
	const langC = new Compartment(); // слот під мову

	const isDark = () => document.documentElement.classList.contains('dark');
	const themeExt = () => (isDark() ? oneDark : []);
	const langExt = () => (lang === 'cpp' ? cpp() : python());


	onMount(() => {
		view = new EditorView({
			parent: el,
			state: EditorState.create({
				doc: value,
				extensions: [
					basicSetup,
					langC.of(langExt()),
					themeC.of(themeExt()),
					EditorView.updateListener.of((u) => {
						if (u.docChanged) {
							value = u.state.doc.toString();
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
		if (view && value !== view.state.doc.toString()) {
			view.dispatch({ changes: { from: 0, to: view.state.doc.length, insert: value } });
		}
	});

	$effect(() => {
		if (view) {
			view.dispatch({ effects: langC.reconfigure(langExt()) });
		}
	});
</script>

<div bind:this={el} class="h-full min-h-0 overflow-hidden"></div>
