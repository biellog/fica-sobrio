//go:build js && wasm

// Programa em Go que roda no navegador (WebAssembly) e dá vida às páginas do site:
// contador do celular de exemplo, calculadora, compartilhar, o mini-game e a página de convite.
// Cada página diz quem é em <body data-pagina="...">.
package main

import (
	"strings"
	"syscall/js"
	"time"

	"ficasobrio/site/logica"
)

var (
	janela = js.Global()
	doc    = janela.Get("document")
)

func porId(id string) js.Value { return doc.Call("getElementById", id) }

func existe(v js.Value) bool { return !v.IsNull() && !v.IsUndefined() }

func texto(id, s string) {
	if el := porId(id); existe(el) {
		el.Set("textContent", s)
	}
}

// ao liga um evento a uma função Go (as funções ficam vivas enquanto a página estiver aberta).
func ao(alvo js.Value, evento string, f func(e js.Value)) {
	if !existe(alvo) {
		return
	}
	alvo.Call("addEventListener", evento, js.FuncOf(func(_ js.Value, args []js.Value) any {
		var e js.Value
		if len(args) > 0 {
			e = args[0]
		}
		f(e)
		return nil
	}))
}

func main() {
	switch doc.Get("body").Get("dataset").Get("pagina").String() {
	case "inicio":
		contadorDemo()
		calculadora()
		compartilhar()
		iniciarJogo()
	case "convite":
		convite()
	}
	doc.Get("documentElement").Get("classList").Call("add", "go-pronto")
	select {} // mantém o Go rodando para responder aos cliques
}

// ---------- Contador do celular de exemplo ----------
func contadorDemo() {
	inicio := time.Now().Add(-logica.MinutosDemo * time.Minute)
	atualizar := func() {
		dias, resto := logica.TempoDemo(int64(time.Since(inicio) / time.Minute))
		texto("dias-demo", itoa(dias))
		texto("horas-demo", resto)
	}
	atualizar()
	go func() {
		for range time.Tick(30 * time.Second) {
			atualizar()
		}
	}()
}

// ---------- Calculadora ----------
func calculadora() {
	campo := porId("gasto")
	if !existe(campo) {
		return
	}
	calc := func(js.Value) {
		e := logica.Calcular(logica.LerGasto(campo.Get("value").String()))
		texto("r-mes", e.Mes)
		texto("r-ano", e.Ano)
		texto("r-5", e.Cinco)
	}
	ao(campo, "input", calc)
	calc(js.Undefined())
}

// ---------- Compartilhar ----------
func compartilhar() {
	ao(porId("compartilhar"), "click", func(e js.Value) {
		e.Call("preventDefault")
		txt := logica.TextoCompartilhar(janela.Get("location").Get("href").String())
		nav := janela.Get("navigator")
		if existe(nav.Get("share")) {
			dados := map[string]any{"title": "Fica Sóbrio", "text": txt}
			p := nav.Call("share", dados)
			p.Call("catch", js.FuncOf(func(js.Value, []js.Value) any { return nil }))
			return
		}
		janela.Call("open", logica.LinkWhatsApp(txt), "_blank")
	})
}

// ---------- Página de convite ----------
func convite() {
	codigo, valido := logica.CodigoConvite(janela.Get("location").Get("href").String())
	el, ok := porId("codigo"), porId("ok")
	aviso := func(classe, msg string) {
		if existe(ok) {
			ok.Set("className", classe)
			ok.Set("textContent", msg)
		}
	}
	if !valido {
		el.Set("textContent", "?")
		aviso("ok erro", "Link de convite incompleto. Peça o código para quem te convidou.")
	} else {
		el.Set("textContent", codigo)
	}
	ao(porId("copiar"), "click", func(js.Value) {
		if !valido {
			return
		}
		feito := func() { aviso("ok", "✅ Código copiado! Agora é só colar no app.") }
		selecionar := func() {
			r := doc.Call("createRange")
			r.Call("selectNodeContents", el)
			s := janela.Call("getSelection")
			s.Call("removeAllRanges")
			s.Call("addRange", r)
			if doc.Call("execCommand", "copy").Truthy() {
				feito()
			} else {
				aviso("ok", "Segure o código para copiar.")
			}
		}
		cb := janela.Get("navigator").Get("clipboard")
		if existe(cb) && janela.Get("isSecureContext").Truthy() {
			cb.Call("writeText", codigo).Call("then",
				js.FuncOf(func(js.Value, []js.Value) any { feito(); return nil }),
				js.FuncOf(func(js.Value, []js.Value) any { selecionar(); return nil }))
			return
		}
		selecionar()
	})
	abrir := porId("abrir")
	android := strings.Contains(strings.ToLower(janela.Get("navigator").Get("userAgent").String()), "android")
	if existe(abrir) {
		if !valido || !android {
			abrir.Get("style").Set("display", "none")
		} else {
			abrir.Set("href", logica.LinkAbrirApp(codigo)) // abre o app se estiver instalado; senão, cai no site
		}
	}
}

func itoa(n int64) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}
