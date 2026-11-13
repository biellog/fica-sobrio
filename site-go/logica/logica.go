// Package logica tem as regras do site do Fica Sóbrio em Go puro
// (sem navegador), para serem testadas com `go test` e usadas tanto
// pelo gerador das páginas quanto pelo WebAssembly que roda no navegador.
package logica

import (
	"math"
	"strconv"
	"strings"
)

const (
	SiteURL     = "https://biellog.github.io/fica-sobrio/"
	APKURL      = "https://github.com/biellog/fica-sobrio/releases/latest/download/FicaSobrio.apk"
	PacoteApp   = "br.com.ficasobrio.fica_sobrio"
	GastoPadrao = 30.0
)

// Reais formata um valor como "R$ 10.950" (arredondado, com ponto nos milhares).
func Reais(v float64) string {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		v = 0
	}
	n := int64(math.Round(v))
	neg := n < 0
	if neg {
		n = -n
	}
	s := strconv.FormatInt(n, 10)
	var b strings.Builder
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			b.WriteByte('.')
		}
		b.WriteRune(c)
	}
	if neg {
		return "R$ -" + b.String()
	}
	return "R$ " + b.String()
}

// LerGasto entende o que a pessoa digitou ("30", "30,50", "1.234,5") e devolve
// um valor entre 0 e 1 milhão (texto inválido vira 0).
func LerGasto(s string) float64 {
	s = strings.TrimSpace(strings.ReplaceAll(s, "R$", ""))
	if strings.Contains(s, ",") {
		s = strings.ReplaceAll(s, ".", "")
		s = strings.ReplaceAll(s, ",", ".")
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil || math.IsNaN(v) || math.IsInf(v, 0) || v < 0 {
		return 0
	}
	return math.Min(v, 1e6)
}

// Economia diz quanto a pessoa junta parando de gastar [gastoDia] por dia.
type Economia struct{ Mes, Ano, Cinco string }

func Calcular(gastoDia float64) Economia {
	return Economia{Mes: Reais(gastoDia * 30), Ano: Reais(gastoDia * 365), Cinco: Reais(gastoDia * 365 * 5)}
}

// TempoDemo devolve os dias e o resto ("7h 12min") de um tempo em minutos (celular de exemplo).
func TempoDemo(minutos int64) (dias int64, resto string) {
	if minutos < 0 {
		minutos = 0
	}
	return minutos / 1440, strconv.FormatInt(minutos%1440/60, 10) + "h " + strconv.FormatInt(minutos%60, 10) + "min"
}

// MinutosDemo é o tempo sóbrio do celular de exemplo no começo (43 dias, 7h 12min).
const MinutosDemo = 43*24*60 + 7*60 + 12

// ValidarCodigo limpa e confere um código de convite (4 a 12 letras/números).
func ValidarCodigo(c string) (string, bool) {
	c = strings.ToUpper(strings.TrimSpace(c))
	if len(c) < 4 || len(c) > 12 {
		return c, false
	}
	for _, r := range c {
		if (r < 'A' || r > 'Z') && (r < '0' || r > '9') {
			return c, false
		}
	}
	return c, true
}

// CodigoConvite lê o código ?c= de um link de convite (dados de fora: tudo validado).
func CodigoConvite(link string) (string, bool) {
	if i := strings.IndexByte(link, '#'); i >= 0 {
		link = link[:i]
	}
	i := strings.IndexByte(link, '?')
	if i < 0 {
		return "", false
	}
	for _, par := range strings.Split(link[i+1:], "&") {
		chave, valor, _ := strings.Cut(par, "=")
		if chave == "c" {
			return ValidarCodigo(desescapar(valor))
		}
	}
	return "", false
}

// desescapar desfaz %XX e '+' de um parâmetro de link (o suficiente para códigos).
func desescapar(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		switch {
		case s[i] == '+':
			b.WriteByte(' ')
		case s[i] == '%' && i+2 < len(s):
			if n, err := strconv.ParseUint(s[i+1:i+3], 16, 8); err == nil {
				b.WriteByte(byte(n))
				i += 2
				continue
			}
			b.WriteByte(s[i])
		default:
			b.WriteByte(s[i])
		}
	}
	return b.String()
}

// Escapar codifica um texto para ir dentro de um link (?text=...).
func Escapar(s string) string {
	const hex = "0123456789ABCDEF"
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z', c >= '0' && c <= '9', c == '-', c == '_', c == '.', c == '~':
			b.WriteByte(c)
		case c == ' ':
			b.WriteByte('+')
		default:
			b.WriteByte('%')
			b.WriteByte(hex[c>>4])
			b.WriteByte(hex[c&15])
		}
	}
	return b.String()
}

// LinkAbrirApp abre o app instalado direto no convite (ou volta para o site).
func LinkAbrirApp(codigo string) string {
	return "intent://convite?c=" + Escapar(codigo) +
		"#Intent;scheme=ficasobrio;package=" + PacoteApp +
		";S.browser_fallback_url=" + Escapar(SiteURL) + ";end"
}

// TextoCompartilhar é a mensagem do botão "compartilhar no WhatsApp".
func TextoCompartilhar(pagina string) string {
	if i := strings.IndexByte(pagina, '#'); i >= 0 {
		pagina = pagina[:i]
	}
	return "Conheci o Fica Sóbrio, um app grátis que transforma ficar sóbrio num jogo. Cada dia sóbrio é um nível a mais: " + pagina
}

// LinkWhatsApp monta o link wa.me com o texto.
func LinkWhatsApp(texto string) string { return "https://wa.me/?text=" + Escapar(texto) }
