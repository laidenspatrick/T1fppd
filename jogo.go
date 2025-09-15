// jogo.go - Funções para manipular os elementos do jogo, como carregar o mapa e mover o personagem
package main

import (
	"bufio"
	"math/rand"
	"os"
	"time"
)

// ------------------ TIPOS BÁSICOS ------------------

// Elemento representa qualquer objeto do mapa (parede, personagem, vegetação, etc)
type Elemento struct {
	simbolo   rune
	cor       Cor
	corFundo  Cor
	tangivel  bool // Indica se o elemento bloqueia passagem
}

type Guarda struct {
	Elemento
	Perseguir        chan bool
	PararPerseguicao chan bool
}

type Portal struct {
	Elemento
	Teletransportar    chan bool
	PararTeletransporte chan bool
}

type Armadilha struct {
	Elemento
	ProximidadeJogador chan bool
	ProximidadeOutro   chan bool
	PararArmadilha     chan bool
}

// Jogo contém o estado atual do jogo
type Jogo struct {
	Mapa           [][]Elemento // grade 2D representando o mapa
	PosX, PosY     int          // posição atual do personagem
	UltimoVisitado Elemento     // elemento que estava na posição do personagem antes de mover
	StatusMsg      string       // mensagem para a barra de status
}

// ------------------ ELEMENTOS VISUAIS ------------------
var (
	Personagem = Elemento{'☺', CorCinzaEscuro, CorPadrao, true}
	Inimigo    = Elemento{'☠', CorVermelho, CorPadrao, true}
	Parede     = Elemento{'▤', CorParede, CorFundoParede, true}
	Vegetacao  = Elemento{'♣', CorVerde, CorPadrao, false}
	Vazio      = Elemento{' ', CorPadrao, CorPadrao, false}

	// Elementos autônomos
	guarda = &Guarda{
		Elemento:         Elemento{'G', CorAmarelo, CorPadrao, true},
		Perseguir:        make(chan bool),
		PararPerseguicao: make(chan bool),
	}

	portal = &Portal{
		Elemento:            Elemento{'P', CorCiano, CorPadrao, false},
		Teletransportar:     make(chan bool),
		PararTeletransporte: make(chan bool),
	}

	armadilha = &Armadilha{
		Elemento:           Elemento{'A', CorVermelho, CorPadrao, false},
		ProximidadeJogador: make(chan bool),
		ProximidadeOutro:   make(chan bool),
		PararArmadilha:     make(chan bool),
	}
)

// ------------------ CANAL DE LOCK ------------------
var mapaLock = make(chan struct{}, 1)

// ------------------ FUNÇÕES DO JOGO ------------------

// Cria e retorna uma nova instância do jogo
func jogoNovo() Jogo {
	return Jogo{UltimoVisitado: Vazio}
}

// Lê um arquivo texto linha por linha e constrói o mapa do jogo
func jogoCarregarMapa(nome string, jogo *Jogo) error {
	arq, err := os.Open(nome)
	if err != nil {
		return err
	}
	defer arq.Close()

	scanner := bufio.NewScanner(arq)
	y := 0
	for scanner.Scan() {
		linha := scanner.Text()
		var linhaElems []Elemento
		for x, ch := range linha {
			e := Vazio
			switch ch {
			case Parede.simbolo:
				e = Parede
			case Inimigo.simbolo:
				e = Inimigo
			case Vegetacao.simbolo:
				e = Vegetacao
			case Personagem.simbolo:
				jogo.PosX, jogo.PosY = x, y
			}
			linhaElems = append(linhaElems, e)
		}
		jogo.Mapa = append(jogo.Mapa, linhaElems)
		y++
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}

// Verifica se o personagem pode se mover para a posição (x, y)
func jogoPodeMoverPara(jogo *Jogo, x, y int) bool {
	if y < 0 || y >= len(jogo.Mapa) {
		return false
	}
	if x < 0 || x >= len(jogo.Mapa[y]) {
		return false
	}
	if jogo.Mapa[y][x].tangivel {
		return false
	}
	return true
}

// Move um elemento para a nova posição
func jogoMoverElemento(jogo *Jogo, x, y, dx, dy int) {
	nx, ny := x+dx, y+dy
	elemento := jogo.Mapa[y][x]
	jogo.Mapa[y][x] = jogo.UltimoVisitado
	jogo.UltimoVisitado = jogo.Mapa[ny][nx]
	jogo.Mapa[ny][nx] = elemento
}

// ------------------ GOROUTINES AUTÔNOMAS ------------------

func iniciarElementos(jogo *Jogo) {
	go comportamentoGuarda(guarda, jogo)
	go comportamentoPortal(portal, jogo)
	go comportamentoArmadilha(armadilha, jogo)
}

// GUARDA: patrulha e pode perseguir
func comportamentoGuarda(guarda *Guarda, jogo *Jogo) {
	mapaLock <- struct{}{}
	if len(jogo.Mapa) > 2 && len(jogo.Mapa[2]) > 2 {
		jogo.Mapa[2][2] = guarda.Elemento
	}
	<-mapaLock

	rand.Seed(time.Now().UnixNano())

	for {
		select {
		case <-guarda.Perseguir:
			mapaLock <- struct{}{}
			dx, dy := 0, 0
			if jogo.PosX > 2 {
				dx = 1
			} else if jogo.PosX < 2 {
				dx = -1
			}
			if jogo.PosY > 2 {
				dy = 1
			} else if jogo.PosY < 2 {
				dy = -1
			}
			jogoMoverElemento(jogo, 2, 2, dx, dy)
			<-mapaLock
		case <-guarda.PararPerseguicao:
			time.Sleep(time.Second)
		default:
			dx := rand.Intn(3) - 1
			dy := rand.Intn(3) - 1
			mapaLock <- struct{}{}
			jogoMoverElemento(jogo, 2, 2, dx, dy)
			<-mapaLock
			time.Sleep(time.Second)
		}
	}
}

// PORTAL: abre e fecha com timeout
func comportamentoPortal(portal *Portal, jogo *Jogo) {
	mapaLock <- struct{}{}
	if len(jogo.Mapa) > 4 && len(jogo.Mapa[4]) > 4 {
		jogo.Mapa[4][4] = portal.Elemento
	}
	<-mapaLock

	for {
		select {
		case <-portal.Teletransportar:
			jogo.StatusMsg = "Você entrou no portal!"
		case <-portal.PararTeletransporte:
			jogo.StatusMsg = "O portal fechou!"
		case <-time.After(5 * time.Second):
			jogo.StatusMsg = "O portal sumiu por inatividade."
		}
	}
}

// ARMADILHA: reage a proximidade
func comportamentoArmadilha(armadilha *Armadilha, jogo *Jogo) {
	mapaLock <- struct{}{}
	if len(jogo.Mapa) > 6 && len(jogo.Mapa[6]) > 6 {
		jogo.Mapa[6][6] = armadilha.Elemento
	}
	<-mapaLock

	for {
		select {
		case <-armadilha.ProximidadeJogador:
			jogo.StatusMsg = "Você caiu em uma armadilha!"
		case <-armadilha.ProximidadeOutro:
			jogo.StatusMsg = "Outro elemento acionou a armadilha!"
		case <-armadilha.PararArmadilha:
			jogo.StatusMsg = "A armadilha foi desativada."
			return
		default:
			time.Sleep(time.Second)
		}
	}
}