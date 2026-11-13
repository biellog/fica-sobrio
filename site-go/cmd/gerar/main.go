// Gerador do site do Fica Sóbrio: monta as páginas a partir dos modelos (html/template),
// copia os arquivos estáticos e o wasm_exec.js oficial do Go para a pasta de saída.
//
// Uso (na pasta site-go):
//
//	GOOS=js GOARCH=wasm go build -ldflags="-s -w" -trimpath -o ../_site/app.wasm ./wasm
//	go run ./cmd/gerar -saida ../_site
package main

import (
	"flag"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"ficasobrio/site/logica"
)

// Dados é o que os modelos das páginas enxergam.
type Dados struct {
	APK, Site, Versao string
	DiasDemo          int64
	HorasDemo         string
	GastoPadrao       string
	Calc              logica.Economia
	Ano               int
}

// paginas: modelo -> arquivo gerado.
var paginas = map[string]string{
	"inicio.html":      "index.html",
	"convite.html":     "convite.html",
	"privacidade.html": "privacidade.html",
}

func novosDados(agora time.Time) Dados {
	dias, horas := logica.TempoDemo(logica.MinutosDemo)
	return Dados{
		APK:         logica.APKURL,
		Site:        logica.SiteURL,
		Versao:      strconv.FormatInt(agora.Unix(), 36),
		DiasDemo:    dias,
		HorasDemo:   horas,
		GastoPadrao: strconv.Itoa(int(logica.GastoPadrao)),
		Calc:        logica.Calcular(logica.GastoPadrao),
		Ano:         agora.Year(),
	}
}

// Gerar escreve o site em [saida] a partir de [modelos] e [estatico].
func Gerar(modelos, estatico, saida, wasmExec string, d Dados) error {
	if err := os.MkdirAll(saida, 0o755); err != nil {
		return err
	}
	for modelo, arquivo := range paginas {
		t, err := template.New(modelo).Option("missingkey=error").ParseFiles(
			filepath.Join(modelos, modelo), filepath.Join(modelos, "_go.html"))
		if err != nil {
			return fmt.Errorf("modelo %s: %w", modelo, err)
		}
		f, err := os.Create(filepath.Join(saida, arquivo))
		if err != nil {
			return err
		}
		err = t.Execute(f, d)
		if e := f.Close(); err == nil {
			err = e
		}
		if err != nil {
			return fmt.Errorf("página %s: %w", arquivo, err)
		}
	}
	if err := copiarPasta(estatico, saida); err != nil {
		return err
	}
	return copiar(wasmExec, filepath.Join(saida, "wasm_exec.js"))
}

func copiarPasta(de, para string) error {
	return filepath.WalkDir(de, func(p string, e fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(de, p)
		destino := filepath.Join(para, rel)
		if e.IsDir() {
			return os.MkdirAll(destino, 0o755)
		}
		return copiar(p, destino)
	})
}

func copiar(de, para string) error {
	in, err := os.Open(de)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(para)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}

func main() {
	saida := flag.String("saida", "../_site", "pasta onde o site é gerado")
	modelos := flag.String("modelos", "modelos", "pasta dos modelos das páginas")
	estatico := flag.String("estatico", "estatico", "pasta dos arquivos estáticos (imagens, robots...)")
	wasmExec := flag.String("wasm-exec", filepath.Join(runtime.GOROOT(), "lib", "wasm", "wasm_exec.js"), "wasm_exec.js do Go")
	flag.Parse()
	if err := Gerar(*modelos, *estatico, *saida, *wasmExec, novosDados(time.Now())); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Site gerado em", *saida)
}
