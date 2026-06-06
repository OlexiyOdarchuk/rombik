<script>
	import { onMount } from 'svelte';
	import { EditorView, basicSetup } from 'codemirror';
	import { EditorState } from '@codemirror/state';
	import { python } from '@codemirror/lang-python';

	let { value = $bindable('') } = $props();
	let el;
	let view;

	onMount(() => {
		view = new EditorView({
			parent: el,
			state: EditorState.create({
				doc: value,
				extensions: [
					basicSetup, // номери рядків, дужки, авто-відступ, підсвітка
					python(), // мова Python: підсвітка + відступ після «:»
					EditorView.updateListener.of((u) => {
						if (u.docChanged) value = u.state.doc.toString();
					}),
					EditorView.theme({
						'&': { height: '100%', fontSize: '13px' },
						'.cm-scroller': { overflow: 'auto', fontFamily: 'ui-monospace, monospace' },
						'&.cm-focused': { outline: 'none' }
					})
				]
			})
		});
		return () => view.destroy();
	});

	// Віддзеркалити зовнішні зміни value (напр. скидання на зразок).
	$effect(() => {
		if (view && value !== view.state.doc.toString()) {
			view.dispatch({ changes: { from: 0, to: view.state.doc.length, insert: value } });
		}
	});
</script>

<div bind:this={el} class="h-full min-h-0 overflow-hidden"></div>
