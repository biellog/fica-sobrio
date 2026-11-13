//go:build js && wasm

package main

import (
	"math"
	"strconv"
	"syscall/js"
	"time"

	"ficasobrio/site/logica"
)

const fonteEmoji = `"Apple Color Emoji","Segoe UI Emoji","Noto Color Emoji",sans-serif`

// ---------- Jogo: Desvia da Tempestade (desenhado num <canvas>) ----------
func iniciarJogo() {
	palco, cv := porId("palco"), porId("tela-jogo")
	if !existe(palco) || !existe(cv) {
		return
	}
	ctx := cv.Call("getContext", "2d")
	telaInicio, telaFim := porId("inicio"), porId("fim")
	var (
		w, h    float64
		j       = logica.NovoJogo(time.Now().UnixNano())
		rodando bool
		ultimo  float64
		recorde int
		quadro  js.Func
		arrasta bool
		ultimoX float64
	)
	if s := lerRecorde(); s > 0 {
		recorde = s
	}

	desenhar := func() {
		ctx.Call("clearRect", 0, 0, w, h)
		ctx.Set("fillStyle", "rgba(255,255,255,.07)")
		for i := 0; i < 40; i++ {
			x := float64(i*97%100) / 100 * w
			y := math.Mod(float64(i*53%100)/100*h+j.T*40*float64(1+i%3), math.Max(h, 1))
			ctx.Call("beginPath")
			ctx.Call("arc", x, y, 1.5+float64(i%2), 0, 7)
			ctx.Call("fill")
		}
		tam := math.Max(26, math.Min(44, w*0.09))
		ctx.Set("textAlign", "center")
		ctx.Set("textBaseline", "middle")
		ctx.Set("fillStyle", "#fff")
		ctx.Set("font", strconv.Itoa(int(tam))+"px "+fonteEmoji)
		for _, it := range j.Itens {
			ctx.Call("fillText", it.Emoji, it.X*w, it.Y*h)
		}
		ctx.Set("fillStyle", "rgba(31,175,122,.25)")
		ctx.Call("beginPath")
		ctx.Call("arc", j.X*w, logica.AlturaBoneco*h, tam*0.85, 0, 7)
		ctx.Call("fill")
		ctx.Set("fillStyle", "#fff")
		ctx.Set("font", strconv.Itoa(int(tam*1.25))+"px "+fonteEmoji)
		ctx.Call("fillText", "😎", j.X*w, logica.AlturaBoneco*h)
		texto("pontos", strconv.Itoa(j.Pontos))
		texto("vidas", j.Coracoes())
	}

	medir := func(js.Value) {
		dpr := math.Min(2, janela.Get("devicePixelRatio").Float())
		if dpr <= 0 {
			dpr = 1
		}
		w, h = palco.Get("clientWidth").Float(), palco.Get("clientHeight").Float()
		cv.Set("width", w*dpr)
		cv.Set("height", h*dpr)
		ctx.Call("setTransform", dpr, 0, 0, dpr, 0, 0)
		if !rodando {
			desenhar()
		}
	}

	fim := func() {
		rodando = false
		novo := j.Pontos > recorde
		if novo {
			recorde = j.Pontos
			salvarRecorde(recorde)
		}
		texto("pontos-fim", strconv.Itoa(j.Pontos))
		if novo {
			texto("recorde-fim", "🏆 Novo recorde!")
		} else {
			texto("recorde-fim", "Seu recorde: "+strconv.Itoa(recorde))
		}
		telaFim.Set("hidden", false)
	}

	quadro = js.FuncOf(func(_ js.Value, args []js.Value) any {
		if !rodando {
			return nil
		}
		ts := args[0].Float()
		dt := math.Min(0.05, (ts-ultimo)/1000)
		ultimo = ts
		if j.Passo(dt) {
			palco.Call("animate", []any{
				map[string]any{"boxShadow": "inset 0 0 0 999px rgba(255,90,95,.3)"},
				map[string]any{"boxShadow": "none"},
			}, 180)
			if nav := janela.Get("navigator"); existe(nav.Get("vibrate")) {
				nav.Call("vibrate", 60)
			}
		}
		desenhar()
		if j.Acabou() {
			fim()
			return nil
		}
		janela.Call("requestAnimationFrame", quadro)
		return nil
	})

	comecar := func(js.Value) {
		j = logica.NovoJogo(time.Now().UnixNano())
		rodando = true
		ultimo = janela.Get("performance").Call("now").Float()
		telaInicio.Set("hidden", true)
		telaFim.Set("hidden", true)
		janela.Call("requestAnimationFrame", quadro)
	}

	// controles: arrastar (dedo ou mouse) e setas do teclado
	ao(cv, "pointerdown", func(e js.Value) {
		arrasta = true
		ultimoX = e.Get("clientX").Float()
		cv.Call("setPointerCapture", e.Get("pointerId"))
		r := cv.Call("getBoundingClientRect")
		j.Mover((ultimoX - r.Get("left").Float()) / r.Get("width").Float())
	})
	ao(cv, "pointermove", func(e js.Value) {
		if !arrasta || w == 0 {
			return
		}
		x := e.Get("clientX").Float()
		j.Mover(j.X + (x-ultimoX)/w)
		ultimoX = x
	})
	ao(cv, "pointerup", func(js.Value) { arrasta = false })
	ao(cv, "pointercancel", func(js.Value) { arrasta = false })
	ao(janela, "keydown", func(e js.Value) {
		if !rodando {
			return
		}
		switch e.Get("key").String() {
		case "ArrowLeft":
			j.Mover(j.X - 0.06)
			e.Call("preventDefault")
		case "ArrowRight":
			j.Mover(j.X + 0.06)
			e.Call("preventDefault")
		}
	})
	ao(porId("btn-jogar"), "click", comecar)
	ao(porId("btn-denovo"), "click", comecar)
	ao(janela, "resize", medir)
	medir(js.Undefined())
}

func lerRecorde() (n int) {
	defer func() { recover() }() // localStorage pode estar bloqueado
	v := janela.Get("localStorage").Call("getItem", "fs_recorde")
	if v.IsNull() {
		return 0
	}
	n, _ = strconv.Atoi(v.String())
	return n
}

func salvarRecorde(n int) {
	defer func() { recover() }()
	janela.Get("localStorage").Call("setItem", "fs_recorde", strconv.Itoa(n))
}
