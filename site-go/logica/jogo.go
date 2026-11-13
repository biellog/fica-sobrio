package logica

import (
	"math"
	"math/rand"
	"strings"
)

// Tipos do que cai no "Desvia da Tempestade".
const (
	Ruim    = "ruim"
	Agua    = "agua"
	Estrela = "estrela"
	Trevo   = "trevo"
	Vida    = "vida"
)

var ruins = []string{"⛈️", "⚡", "🌪️", "☄️"}

type Item struct {
	X, Y, V     float64
	Tipo, Emoji string
}

// Jogo é o mini-game do site: desvie das tempestades e pegue água, estrelas e trevos.
// Posições de 0 a 1 (proporção da tela); o boneco fica na altura AlturaBoneco.
type Jogo struct {
	Itens       []Item
	X           float64
	T           float64
	Pontos      int
	Vidas       int
	spawn, acum float64
	rnd         *rand.Rand
}

const AlturaBoneco = 0.88

func NovoJogo(semente int64) *Jogo {
	return &Jogo{X: 0.5, Vidas: 3, rnd: rand.New(rand.NewSource(semente))}
}

func (j *Jogo) Acabou() bool { return j.Vidas <= 0 }

// Mover coloca o boneco em x (0 a 1), sem sair da tela.
func (j *Jogo) Mover(x float64) { j.X = math.Min(0.94, math.Max(0.06, x)) }

// Passo avança o jogo dt segundos. Devolve true se bateu numa tempestade.
func (j *Jogo) Passo(dt float64) (bateu bool) {
	if j.Acabou() {
		return false
	}
	dt = math.Max(0, math.Min(dt, 0.05))
	j.T += dt
	dif := 1 + j.T/25
	j.acum += dt * 10
	if j.acum >= 1 {
		j.Pontos += int(j.acum)
		j.acum = math.Mod(j.acum, 1)
	}
	j.spawn -= dt
	if j.spawn <= 0 {
		j.spawn = math.Max(0.25, 0.9/dif)
		r := j.rnd.Float64()
		it := Item{X: 0.06 + j.rnd.Float64()*0.88, Y: -0.05, V: (0.28 + j.rnd.Float64()*0.2) * dif}
		switch {
		case r < 0.58:
			it.Tipo, it.Emoji = Ruim, ruins[j.rnd.Intn(len(ruins))]
		case r < 0.80:
			it.Tipo, it.Emoji = Agua, "💧"
		case r < 0.90:
			it.Tipo, it.Emoji = Estrela, "⭐"
		case r < 0.97:
			it.Tipo, it.Emoji = Trevo, "🍀"
		default:
			it.Tipo, it.Emoji = Vida, "❤️"
		}
		j.Itens = append(j.Itens, it)
	}
	vivos := j.Itens[:0]
	for _, it := range j.Itens {
		it.Y += it.V * dt
		if math.Abs(it.Y-AlturaBoneco) < 0.045 && math.Abs(it.X-j.X) < 0.085 {
			switch it.Tipo {
			case Ruim:
				j.Vidas--
				bateu = true
			case Agua:
				j.Pontos += 10
			case Estrela:
				j.Pontos += 30
			case Trevo:
				j.Pontos += 50
			case Vida:
				j.Vidas = min(3, j.Vidas+1)
			}
			continue
		}
		if it.Y < 1.1 {
			vivos = append(vivos, it)
		}
	}
	j.Itens = vivos
	return bateu
}

// Coracoes mostra as vidas: ❤️❤️🖤.
func (j *Jogo) Coracoes() string {
	v := max(0, min(3, j.Vidas))
	return strings.Repeat("❤️", v) + strings.Repeat("🖤", 3-v)
}
