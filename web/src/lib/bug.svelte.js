// Глобальний перемикач модалки «Повідомити про помилку» (Svelte 5 руни в .svelte.js).
let open = $state(false);

export function openBug() {
	open = true;
}
export function closeBug() {
	open = false;
}
export function bugOpen() {
	return open;
}
