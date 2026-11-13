package logica

import (
	"strings"
	"testing"
)

func TestReais(t *testing.T) {
	casos := map[float64]string{0: "R$ 0", 900: "R$ 900", 10950: "R$ 10.950", 54750: "R$ 54.750", 1234567.6: "R$ 1.234.568", -5: "R$ -5"}
	for v, quer := range casos {
		if got := Reais(v); got != quer {
			t.Errorf("Reais(%v) = %q, quer %q", v, got, quer)
		}
	}
}

func TestLerGasto(t *testing.T) {
	casos := map[string]float64{"30": 30, " 30,5 ": 30.5, "1.234,5": 1234.5, "R$ 12": 12, "abc": 0, "-3": 0, "": 0, "NaN": 0, "1e99": 1e6}
	for s, quer := range casos {
		if got := LerGasto(s); got != quer {
			t.Errorf("LerGasto(%q) = %v, quer %v", s, got, quer)
		}
	}
}

func TestCalcular(t *testing.T) {
	e := Calcular(GastoPadrao)
	if e.Mes != "R$ 900" || e.Ano != "R$ 10.950" || e.Cinco != "R$ 54.750" {
		t.Fatalf("calculadora: %+v", e)
	}
}

func TestTempoDemo(t *testing.T) {
	d, r := TempoDemo(MinutosDemo)
	if d != 43 || r != "7h 12min" {
		t.Fatalf("demo: %d %q", d, r)
	}
	if d, r := TempoDemo(-10); d != 0 || r != "0h 0min" {
		t.Fatalf("negativo: %d %q", d, r)
	}
}

func TestCodigoConvite(t *testing.T) {
	if c, ok := CodigoConvite("https://x/convite.html?c=k7q2zp"); !ok || c != "K7Q2ZP" {
		t.Fatalf("válido: %q %v", c, ok)
	}
	if c, ok := CodigoConvite("https://x/convite.html?utm=1&c=%20ab12%20#topo"); !ok || c != "AB12" {
		t.Fatalf("com espaço codificado: %q %v", c, ok)
	}
	for _, l := range []string{"https://x/convite.html", "https://x/?c=<script>", "https://x/?c=%3Cb%3E1234", "https://x/?c=AB", "https://x/?c=ABCDEFGHIJKLMN", "%%%", "https://x/?c=ÁBCD"} {
		if _, ok := CodigoConvite(l); ok {
			t.Errorf("devia ser inválido: %q", l)
		}
	}
}

func TestLinks(t *testing.T) {
	l := LinkAbrirApp("K7Q2ZP")
	if !strings.HasPrefix(l, "intent://convite?c=K7Q2ZP#Intent;scheme=ficasobrio;package=br.com.ficasobrio.fica_sobrio;") || !strings.HasSuffix(l, ";end") {
		t.Fatalf("intent: %s", l)
	}
	txt := TextoCompartilhar("https://biellog.github.io/fica-sobrio/#baixar")
	if !strings.HasSuffix(txt, "https://biellog.github.io/fica-sobrio/") {
		t.Fatalf("texto: %s", txt)
	}
	if !strings.HasPrefix(LinkWhatsApp("oi tudo"), "https://wa.me/?text=oi+tudo") {
		t.Fatalf("whatsapp: %s", LinkWhatsApp("oi tudo"))
	}
}

func TestEscapar(t *testing.T) {
	if got := Escapar("Olá, mundo: https://a.b/?x=1&y"); got != "Ol%C3%A1%2C+mundo%3A+https%3A%2F%2Fa.b%2F%3Fx%3D1%26y" {
		t.Fatalf("escapar: %s", got)
	}
}

func TestJogo(t *testing.T) {
	j := NovoJogo(1)
	if j.Coracoes() != "❤️❤️❤️" {
		t.Fatal(j.Coracoes())
	}
	j.Mover(5)
	if j.X != 0.94 {
		t.Fatalf("mover limita: %v", j.X)
	}
	for k := 0; k < 20000 && !j.Acabou(); k++ {
		j.Passo(0.016)
	}
	if !j.Acabou() {
		t.Fatal("parado no lugar, as tempestades deviam acabar com as vidas")
	}
	if j.Pontos <= 0 || j.Coracoes() != "🖤🖤🖤" {
		t.Fatalf("fim: %d %s", j.Pontos, j.Coracoes())
	}
	if j.Passo(0.016) {
		t.Fatal("depois do fim não bate mais")
	}
}

func TestJogoDesviarDuraMais(t *testing.T) {
	parado, esperto := NovoJogo(7), NovoJogo(7)
	for k := 0; k < 3000; k++ {
		parado.Passo(0.016)
		// foge da tempestade mais próxima
		alvo := esperto.X
		for _, it := range esperto.Itens {
			if it.Tipo == Ruim && it.Y > 0.6 && it.Y < 0.95 && abs(it.X-esperto.X) < 0.15 {
				if it.X > esperto.X {
					alvo = esperto.X - 0.2
				} else {
					alvo = esperto.X + 0.2
				}
			}
		}
		esperto.Mover(alvo)
		esperto.Passo(0.016)
	}
	if esperto.Vidas < parado.Vidas && !parado.Acabou() {
		t.Fatalf("desviar devia ajudar: esperto %d parado %d", esperto.Vidas, parado.Vidas)
	}
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
