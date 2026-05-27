package main

import (
	"fmt"
	"os"
)

const (
	R    = "\033[31m"
	G    = "\033[32m"
	Y    = "\033[33m"
	C    = "\033[36m"
	BOLD = "\033[1m"
	DIM  = "\033[2m"
	RST  = "\033[0m"
)

func salvar(caminho string, validos []resultado) {
	f, err := os.Create(caminho)
	if err != nil {
		fmt.Fprintf(os.Stderr, R+"[!] "+RST+"nao foi possivel salvar: %v\n", err)
		return
	}
	defer f.Close()

	for _, r := range validos {
		fmt.Fprintf(f, "%s:%s\n", r.p.usuario, r.p.senha)
	}
	fmt.Printf(DIM+"[*] "+RST+"resultados salvos em %s\n", caminho)
}
