# Roteiro de Apresentação — FinançasLua

Dashboard financeiro pessoal feito com **Go** (backend) e **Lua** (motor de análise),
para a disciplina de Linguagens de Programação.

> **Regra de ouro da apresentação:** não mostramos código. Falamos sobre a aplicação
> funcionando e sobre *por que* cada linguagem foi escolhida para cada papel.
> Quem estiver na tela deve navegar na aplicação enquanto o outro fala.

---

## 0. Divisão sugerida (quem fala o quê)

| Bloco | Tempo aprox. | Conteúdo |
|-------|--------------|----------|
| Abertura + demo | 3–4 min | O que é, navegar pelas telas |
| Papel do Go | 2–3 min | Backend, servidor, renderização |
| Papel do Lua | 2–3 min | Algoritmos de análise e previsão |
| Integração Go ↔ Lua | 2 min | O ponto mais interessante do projeto |
| HTML/CSS (o terceiro recurso) | 1–2 min | Justificativa |
| Fechamento | 1 min | O que aprendemos |

---

## 1. Abertura (gancho)

> "Todo mundo já tentou controlar gastos numa planilha e desistiu. A gente quis montar
> um painel financeiro pessoal que não só mostra os números, mas **interpreta** eles:
> diz se você está poupando bem, alerta quando uma categoria estoura o orçamento e
> ainda **prevê** o gasto do mês seguinte."

Frase de posicionamento:

> "O projeto usa **duas linguagens com papéis bem separados**: Go cuida de toda a
> infraestrutura — servidor web, rotas, dados, renderização das telas — e Lua é o
> **cérebro analítico**, onde ficam as regras financeiras e o algoritmo de previsão."

Por que essa divisão é boa de mencionar: ela mostra que a escolha das linguagens foi
**intencional**, cada uma no que faz de melhor — não foi "usar duas por usar".

---

## 2. Demonstração da aplicação (passeio pelas telas)

Navegue na ordem abaixo. A cada tela, diga *o que ela faz* e *qual linguagem está por trás*.

1. **Visão Geral**
   - KPIs (renda, gastos, saldo, taxa de poupança), gráfico de rosca por categoria e
     barras de tendência mensal.
   - Apontar a **"pílula de insight"** no topo: *"essa frase não é fixa — ela é gerada
     pelo script Lua a cada carregamento, analisando a situação financeira atual."*
   - Apontar os **gráficos**: *"são desenhados pelo próprio servidor, sem JavaScript —
     já volto nesse ponto."*

2. **Gastos** — listagem por categoria, busca/filtro, e formulário para adicionar gasto.
   - Adicionar um gasto ao vivo para mostrar que a aplicação é interativa.

3. **Metas** — cards de progresso (valor atual / alvo, prazo). Fazer um depósito numa meta.

4. **Calendário** — eventos financeiros do mês, com navegação entre meses.

5. **Previsões** — gráfico de linha com a previsão de gasto.
   - *"Esse número de previsão vem de um algoritmo de média móvel ponderada escrito em Lua."*

6. **Scripts Lua** — a própria aplicação exibe os scripts que estão rodando no backend.
   - Ótimo momento de transição: *"a aplicação é transparente sobre isso, então vamos
     falar de cada linguagem."*

---

## 3. O papel do Go (backend e plataforma)

**Mensagem central:** *Go é a fundação — ele sustenta a aplicação inteira.*

O que dizer que o Go faz:

- **Servidor web nativo.** O servidor HTTP usa a **biblioteca padrão do Go** (`net/http`),
  sem framework externo. Isso é uma marca registrada da linguagem: ela já vem "com pilhas
  incluídas" para construir serviços web de produção.
- **Roteamento e formulários.** Todas as rotas (gastos, metas, orçamento, depósitos, reset)
  são tratadas pelo Go: ele recebe os formulários, valida, atualiza os dados e redireciona.
- **Renderização das páginas.** O Go monta o HTML final no servidor a partir de templates,
  já preenchido com os dados — o navegador recebe a página pronta.
- **Os gráficos.** Os gráficos de rosca, barras e linha são **desenhados pelo Go como SVG**,
  diretamente no servidor.

**Especificidades do Go que ajudaram (vale destacar):**

- **Compilado para um binário único** — a aplicação vira um executável só, sem precisar
  instalar interpretador ou dependências na máquina. Fácil de rodar e distribuir.
- **Concorrência e desempenho** — o `net/http` lida com várias requisições simultâneas
  de forma eficiente, algo natural na linguagem.
- **Tipagem estática** — os dados financeiros (gastos, metas, eventos) são modelados como
  *structs* com tipos definidos, o que reduz erros e deixa claro o formato dos dados.
- **Biblioteca padrão forte** — servidor, templates de HTML e formatação vêm de fábrica.

---

## 4. O papel do Lua (motor de análise / scripting)

**Mensagem central:** *Lua é onde mora a "inteligência" financeira — as regras de negócio.*

O Lua é usado para **dois scripts**, ambos rodando dentro do backend Go:

- **`insights.lua` — análise de saúde financeira.**
  Recebe renda, gastos e metas, e devolve frases de diagnóstico: calcula a **taxa de poupança**,
  aplica a **regra 50/30/20**, alerta quando o orçamento de uma categoria estoura e avisa
  quando uma meta está quase concluída. É o texto que aparece na pílula de insight.

- **`forecast.lua` — previsão de gastos.**
  Implementa uma **média móvel ponderada**: os últimos meses entram com pesos diferentes
  (os mais recentes pesam mais), gerando a previsão do próximo mês.

**Por que Lua se encaixa bem nesse papel (especificidades a destacar):**

- **Linguagem de scripting feita para ser embarcada.** O propósito clássico do Lua é ser
  *embutido* dentro de outro programa (foi muito usado assim em jogos, ex.: roteiros e regras).
  Aqui é exatamente isso: o Go é o programa principal e o Lua é embutido para rodar regras.
- **As regras ficam separadas do servidor.** A lógica financeira vive em arquivos `.lua`
  próprios. Dá para **mudar uma regra** (ex.: o limite da taxa de poupança, ou os pesos da
  previsão) **sem mexer no backend** — bastaria editar o script.
- **Sintaxe leve e enxuta** — ideal para expressar fórmulas e regras de negócio de forma
  legível, sem o "peso" de uma linguagem de sistema.

> Analogia útil: *"Go é o motor e o chassi do carro; Lua é o computador de bordo que decide
> o que os números significam."*

---

## 5. A integração Go ↔ Lua (o ponto mais interessante)

Esse é o trecho que mais impressiona, porque mostra **duas linguagens conversando**.

O que explicar (sem código):

- O backend Go **carrega os scripts Lua** e os executa através de um **interpretador Lua
  embarcado** (a biblioteca **gopher-lua**, um Lua implementado em Go).
- Na prática: o Go **passa os dados** (renda, lista de gastos, metas) para dentro do
  ambiente Lua, o script Lua **roda os cálculos** e **devolve o resultado** (a previsão ou
  as frases de insight) de volta para o Go, que então mostra na tela.
- É uma divisão de responsabilidades real: **o Go nunca calcula as regras financeiras —
  ele delega isso ao Lua.**

Por que isso é relevante para a disciplina:

> "É um exemplo concreto de **interoperabilidade entre linguagens**: uma linguagem
> compilada e estática (Go) hospedando uma linguagem interpretada e dinâmica (Lua),
> cada uma fazendo o que faz melhor."

---

## 6. Por que usamos HTML/CSS além de Go e Lua (justificativa)

Esta pergunta provavelmente vai vir do professor. Resposta curta e firme:

> "A proposta era uma **aplicação web**, e a web tem uma camada de apresentação obrigatória.
> HTML e CSS **não são linguagens de programação** — são linguagens de **marcação e estilo**.
> Eles não têm lógica, não tomam decisões: só descrevem *como a página aparece*. Toda a
> lógica continua 100% em Go e Lua."

Reforço importante — **o que NÃO usamos**:

> "Conscientemente **não usamos JavaScript**. Normalmente uma aplicação web precisa de JS
> para gráficos e interatividade, e isso seria uma **terceira linguagem de programação**.
> Para manter o foco em Go e Lua, fizemos tudo **renderizado no servidor**: os gráficos são
> **SVG gerado pelo Go** e a interatividade usa **formulários HTML puros**. Resultado:
> nenhuma linha de programação fora de Go e Lua."

Essa é a sacada que deixa a justificativa convincente: HTML/CSS é só a "moldura"; nós até
**evitamos de propósito** o JavaScript para não fugir das duas linguagens escolhidas.

---

## 7. Fechamento

Resumo em uma frase:

> "FinançasLua mostra duas linguagens com papéis complementares: **Go** como plataforma —
> servidor, dados e renderização — e **Lua** como motor de regras e previsões, com as duas
> integradas via gopher-lua. A camada web é HTML/CSS pura, sem JavaScript, justamente para
> manter o foco nas linguagens da disciplina."

O que aprendemos (escolha 1–2 para falar):

- Quando faz sentido **embarcar** uma linguagem de scripting dentro de outra.
- As diferenças práticas entre **compilada/estática** (Go) e **interpretada/dinâmica** (Lua).
- Como **separar lógica de negócio** (Lua) da **infraestrutura** (Go) deixa o projeto mais organizado.

---

## Apêndice — Possíveis perguntas do professor

- **"Por que não fizeram tudo em Go?"** → Daria, mas o objetivo era usar duas linguagens
  e mostrar integração. Além disso, deixar as regras em Lua permite alterá-las sem recompilar.
- **"Por que Lua e não Python/JS?"** → Lua é a linguagem de scripting *embarcável* por
  excelência, leve, e é uma das linguagens vistas na disciplina; combina com o Go como host.
- **"Os gráficos são uma biblioteca pronta?"** → Não; são SVG desenhados pelo próprio Go,
  justamente para não depender de JavaScript.
- **"Onde ficam os dados?"** → Em memória no servidor (Go), com dados de exemplo. Próximo
  passo natural seria persistir em banco (SQLite).
- **"O que é gopher-lua?"** → Um interpretador da linguagem Lua escrito em Go, que permite
  rodar scripts Lua de dentro de um programa Go.
</content>
</invoke>
