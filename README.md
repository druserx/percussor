<div align="center">

<pre>
 ██████╗ ███████╗██████╗  ██████╗██╗   ██╗███████╗███████╗ ██████╗ ██████╗
 ██╔══██╗██╔════╝██╔══██╗██╔════╝██║   ██║██╔════╝██╔════╝██╔═══██╗██╔══██╗
 ██████╔╝█████╗  ██████╔╝██║     ██║   ██║███████╗███████╗██║   ██║██████╔╝
 ██╔═══╝ ██╔══╝  ██╔══██╗██║     ██║   ██║╚════██║╚════██║██║   ██║██╔══██╗
 ██║     ███████╗██║  ██║╚██████╗╚██████╔╝███████║███████║╚██████╔╝██║  ██║
 ╚═╝     ╚══════╝╚═╝  ╚═╝ ╚═════╝ ╚═════╝╚══════╝╚══════╝ ╚═════╝ ╚═╝  ╚═╝
</pre>

![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat-square&logo=go&logoColor=white)
![Platform](https://img.shields.io/badge/platform-linux%20%7C%20windows-555555?style=flat-square&logo=linux&logoColor=white)
![Proto](https://img.shields.io/badge/proto-HTTP%20%7C%20OWA%20%7C%20Basic-blueviolet?style=flat-square)
![License](https://img.shields.io/badge/license-MIT-blue?style=flat-square)

**credential sprayer para pentesting de autenticacao HTTP, OWA e Basic Auth**

</div>

---

fiz isso depois de chegar num pentest de active directory com uma lista de 200 usuarios e precisar fazer password spray sem levar lockout. as ferramentas que eu conhecia ou tinham dependencia que nao existia no jump host, ou nao deixavam configurar o delay como eu queria. escrevi em go pra ter um binario estatico que roda em qualquer lugar sem instalar nada.

o nome e latim — *percussor* significa "aquele que golpeia", o assassino que age de surpresa. parece adequado pra uma coisa que testa credencial por credencial esperando o momento certo.

---

## o que faz

- spray em formularios HTTP (`-proto http`) com campos configuráveis
- spray em OWA (Outlook Web Access) com o fluxo correto do `auth.owa`
- spray em HTTP Basic Auth (`-proto basico`)
- delay configuravel entre tentativas pra nao levar lockout
- threads concorrentes com controle fino (`-T`)
- aceita arquivo de usuarios/senhas ou valor unico direto na linha de comando
- salva credenciais validas em arquivo (`-o`)
- binario estatico, sem dependencias externas

---

## build

```
go build -ldflags="-s -w" -o percussor .
```

ou simplesmente:

```
make
```

pra Windows:

```
make windows
```

precisa so do Go instalado. sem CGO, sem dependencias de sistema.

---

## uso

```
percussor -t <alvo> -u <usuarios> -p <senhas> [opcoes]

  -t        alvo (URL completa ou host:porta)
  -u        arquivo de usuarios ou usuario unico
  -p        arquivo de senhas ou senha unica
  -proto    protocolo: http (padrao), owa, basico
  -fu       nome do campo de usuario no form (padrao: username)
  -fp       nome do campo de senha no form (padrao: password)
  -T        threads simultaneas (padrao: 5)
  -d        delay entre tentativas em ms (padrao: 0)
  -o        salvar credenciais validas em arquivo
  -v        verboso — mostra todas as tentativas
```

---

## exemplos

spray em OWA com delay de 1.5s e 3 threads pra nao levar lockout:

```
./percussor -t https://mail.corp.local -u users.txt -p Winter2024 -proto owa -T 3 -d 1500
```

spray em form customizado com campos diferentes do padrao:

```
./percussor -t http://10.10.10.50/login -u admin -p rockyou.txt -fu user -fp pass -T 10
```

usuario e senha unicos (sem arquivo):

```
./percussor -t http://10.10.10.50/admin -u administrator -p P@ssw0rd -proto basico
```

---

## exemplo de saida

```
 ██████╗ ███████╗██████╗  ██████╗██╗   ██╗███████╗███████╗ ██████╗ ██████╗
 ██╔══██╗██╔════╝██╔══██╗██╔════╝██║   ██║██╔════╝██╔════╝██╔═══██╗██╔══██╗
 ██████╔╝█████╗  ██████╔╝██║     ██║   ██║███████╗███████╗██║   ██║██████╔╝
 ██╔═══╝ ██╔══╝  ██╔══██╗██║     ██║   ██║╚════██║╚════██║██║   ██║██╔══██╗
 ██║     ███████╗██║  ██║╚██████╗╚██████╔╝███████║███████║╚██████╔╝██║  ██║
 ╚═╝     ╚══════╝╚═╝  ╚═╝ ╚═════╝ ╚═════╝╚══════╝╚══════╝ ╚═════╝ ╚═╝  ╚═╝

  credential sprayer  |  v0.1.0

[*] alvo     : https://mail.corp.local
[*] proto    : owa
[*] usuarios : 200
[*] senhas   : 1
[*] total    : 200 tentativas
[*] threads  : 3

[+] jsilva:Winter2024
[+] m.rodrigues:Winter2024

════════════════════════════════════════
  tentativas : 200
  validos    : 2
════════════════════════════════════════
```

---

## estrutura

```
percussor/
├── main.go        — entrada, parse de args, orquestracao
├── spray.go       — logica de spray (http, owa, basic auth)
├── relatorio.go   — cores ansi, salvar resultados
├── go.mod
└── Makefile
```

---

## limitacoes

- deteccao de lockout nao implementada — configure o delay com cuidado (`-d 1500` ou mais pra OWA)
- sem suporte a LDAP/SMB ainda
- deteccao de sucesso por redirect (302) — alguns forms redirecionam mesmo em falha, ajuste conforme o alvo
- sem suporte a proxy ainda — se precisar rotear por burp adiciona `-fu` e usa o form manualmente

---

## referencias

- [Password Spraying — MITRE ATT&CK T1110.003](https://attack.mitre.org/techniques/T1110/003/)
- [OWA Password Spraying — byt3bl33d3r](https://byt3bl33d3r.github.io/practical-guide-to-ntlm-relaying-in-2017-aka-getting-a-foothold-in-under-5-minutes.html)
- [Avoiding Account Lockouts During Password Spraying](https://www.ired.team/offensive-security/initial-access/password-spraying-outlook-web-access-remote-shell)
