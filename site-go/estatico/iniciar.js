// Dá a partida no site em Go (WebAssembly). Toda a lógica está em site-go/wasm (Go).
addEventListener('DOMContentLoaded', () => {
  const go = new Go();
  WebAssembly.instantiateStreaming(fetch('app.wasm?v=' + document.documentElement.dataset.versao), go.importObject).then((r) => go.run(r.instance));
});
