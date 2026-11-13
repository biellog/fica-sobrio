package main

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestGerarSite(t *testing.T) {
	saida := t.TempDir()
	wasmExec := filepath.Join(runtime.GOROOT(), "lib", "wasm", "wasm_exec.js")
	if err := Gerar("../../modelos", "../../estatico", saida, wasmExec, novosDados(time.Date(2026, 11, 13, 0, 0, 0, 0, time.UTC))); err != nil {
		t.Fatal(err)
	}
	ler := func(nome string) string {
		b, err := os.ReadFile(filepath.Join(saida, nome))
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}
	inicio := ler("index.html")
	for _, quer := range []string{`data-pagina="inicio"`, "R$ 900", "R$ 10.950", "R$ 54.750", ">43<", "7h 12min",
		"releases/latest/download/FicaSobrio.apk", "wasm_exec.js", "iniciar.js"} {
		if !strings.Contains(inicio, quer) {
			t.Errorf("index.html sem %q", quer)
		}
	}
	if !strings.Contains(ler("convite.html"), `data-pagina="convite"`) {
		t.Error("convite sem data-pagina")
	}
	ler("privacidade.html")
	for _, f := range []string{"img/icone.png", "robots.txt", "sitemap.xml", "iniciar.js", "wasm_exec.js"} {
		if _, err := os.Stat(filepath.Join(saida, f)); err != nil {
			t.Errorf("faltou %s", f)
		}
	}
	// nenhum JavaScript escrito à mão nas páginas: só o JSON-LD (dados para o Google)
	// e os dois arquivos que dão a partida no Go (wasm_exec.js oficial + iniciar.js)
	scripts := regexp.MustCompile(`<script[^>]*>`).FindAllString(inicio+ler("convite.html")+ler("privacidade.html"), -1)
	for _, s := range scripts {
		if !strings.Contains(s, "application/ld+json") && !strings.Contains(s, `src="wasm_exec.js`) && !strings.Contains(s, `src="iniciar.js`) {
			t.Errorf("script inesperado: %s", s)
		}
	}
	if strings.Contains(inicio, "addEventListener") || strings.Contains(inicio, "function(") {
		t.Error("ainda tem JavaScript escrito dentro da página")
	}
}
