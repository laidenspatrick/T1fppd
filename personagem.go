// personagem.go - Funções para movimentação e ações do personagem
package main

import "fmt"

// Atualiza a posição do personagem com base na tecla pressionada (WASD)
func personagemMover(tecla rune, jogo *Jogo) {
	dx, dy := 0, 0
	switch tecla {
	case 'w': dy = -1 // Move para cima
	case 'a': dx = -1 // Move para a esquerda
	case 's': dy = 1  // Move para baixo
	case 'd': dx =1  // Move para a direita
	}

	nx, ny := jogo.PosX+dx, jogo.PosY+dy

	// --- LÓGICA DE COLISÃO E INTERAÇÃO ---

    // 1. Verifique se a nova posição está fora dos limites
    if ny < 0 || ny >= len(jogo.Mapa) || nx < 0 || nx >= len(jogo.Mapa[ny]) {
        return
    }

    // 2. Pegue o elemento na nova posição
    elementoNaPosicao := jogo.Mapa[ny][nx]

    // 3. Verifique a colisão com a armadilha
    if elementoNaPosicao.simbolo == armadilha.Elemento.simbolo {
        jogo.StatusMsg = "FIM DE JOGO! Você caiu na armadilha!"
        go func() { armadilha.ProximidadeJogador <- true }()
        // O loop principal deve quebrar. Uma forma de fazer isso é mover
        // a lógica de retorno para a função que chama personagemMover.
        // Já que você está chamando essa função de dentro de personagemExecutarAcao,
        // a lógica de retorno `false` deve estar lá.
        return 
    }

    // 4. Verifique a colisão com o portal
    if elementoNaPosicao.simbolo == portal.Elemento.simbolo {
        jogo.StatusMsg = "Você entrou no portal e foi teletransportado!"
        go func() { portal.Teletransportar <- true }()
        // Implemente a lógica de teletransporte. Exemplo:
        jogo.Mapa[jogo.PosY][jogo.PosX] = jogo.UltimoVisitado // Limpa a posição antiga
        jogo.PosX, jogo.PosY = 1, 1 // Nova posição
        jogo.UltimoVisitado = jogo.Mapa[jogo.PosY][jogo.PosX] // Salva o elemento da nova posição
        jogo.Mapa[jogo.PosY][jogo.PosX] = Personagem // Coloca o personagem na nova posição
        return
    }
    
    // 5. Se o movimento for para um elemento tangível, não faça nada
    if elementoNaPosicao.tangivel {
        return
    }

    // 6. Se nenhuma das condições acima foi satisfeita, mova o personagem
    jogoMoverElemento(jogo, jogo.PosX, jogo.PosY, dx, dy)
    jogo.PosX, jogo.PosY = nx, ny

	nx, ny = jogo.PosX+dx, jogo.PosY+dy
	// Verifica se o movimento é permitido e realiza a movimentação
	if jogoPodeMoverPara(jogo, nx, ny) {
		jogoMoverElemento(jogo, jogo.PosX, jogo.PosY, dx, dy)
		jogo.PosX, jogo.PosY = nx, ny
	}
}

// Define o que ocorre quando o jogador pressiona a tecla de interação
// Neste exemplo, apenas exibe uma mensagem de status
// Você pode expandir essa função para incluir lógica de interação com objetos
func personagemInteragir(jogo *Jogo) {
	// Atualmente apenas exibe uma mensagem de status
	jogo.StatusMsg = fmt.Sprintf("Interagindo em (%d, %d)", jogo.PosX, jogo.PosY)
}

// Processa o evento do teclado e executa a ação correspondente
func personagemExecutarAcao(ev EventoTeclado, jogo *Jogo) bool {
	switch ev.Tipo {
	case "sair":
		// Retorna false para indicar que o jogo deve terminar
		return false
	case "interagir":
		// Executa a ação de interação
		personagemInteragir(jogo)
	case "mover":
		// Move o personagem com base na tecla
		personagemMover(ev.Tecla, jogo)
	}
	return true // Continua o jogo
}
