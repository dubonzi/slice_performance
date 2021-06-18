# Ganhando performance com slices no Go

Os slices do Go são bem simples e fáceis de usar, mas existem algumas dicas que você pode seguir para ganhar performance e reduzir o uso de processamento e memória quando trabalha com um volume grande de dados. Vou falar superficialmente sobre o funcionamento dos slices mas você pode encontrar uma explicação detalhada no [blog oficial](https://blog.golang.org/slices-intro).

Aprendi essa dica enquanto procurava formas de otimizar uma de nossas apis aqui na B2W e desde então uso sempre que aplicável.

## Nosso exemplo

Para este exemplo, dividi o conteúdo do livro [Moby Dick](https://github.com/GITenberg/Moby-Dick--Or-The-Whale_2701) para criar um `slice` de strings (usando `strings.Split(book, " ")`) com um tamanho aproximado de 190000 palavras. Nosso objetivo é processar cada uma delas e criar um novo `slice` do tipo `Word` contendo nossa palavra modificada.

O código usado pode ser encontrado no [meu github](https://github.com/dubonzi/slice_performance).

Ps: Os testes de performance foram executados usando um `AMD Ryzen 5 5600X`.

Está é nossa struct `Word`:

```go
type Word struct {
	word  string
	index int
}
```

Vamos escrever uma função simples que irá processar as palavras:

```go
func ProcessWords(rawWords []string) []Word {
	words := make([]Word, 0)
	for i, w := range rawWords {
		words = append(words, Word{process(w), i})
	}

	return words
}
```
Parece tudo ok certo? Vamos agora fazer uma pequena modificação nesta função e comparar a performance de cada uma:

```go
func ProcessWordsFaster(rawWords []string) []Word {
	words := make([]Word, 0, len(rawWords))
	for i, w := range rawWords {
		words = append(words, Word{process(w), i})
	}

	return words
}
```

Executamos o teste (benchmark) e obtemos o resultado abaixo:

```shell
BenchmarkProcessWords-12         54  21833985 ns/op   30148452 B/op  194504 allocs/op
BenchmarkProcessWordsFaster-12  100  10423307 ns/op    6086001 B/op  194471 allocs/op
```

Percebeu a diferença no código? Ao adicionar um novo parâmetro à `make`, que é a capacidade inicial do `slice`, nossa função ganhou o dobro de performance e usou cerca de 5x menos memória por operação.

Similar ao `ArrayList` do Java, os slices são estruturas de tamanho dinâmico suportados por um `array` que *cresce automaticamente* à medida que novos elementos são adicionados. Isso não é um problema na maioria dos casos, mas pode levar a problemas de performance se você não ficar de olho. Saber como isso funciona é muito útil e pode te poupar recursos no futuro.

## Como o slice cresce?

Olhando para a primeira função que escrevemos, quando `make([]Word, 0)` é executado, um novo `slice` do tipo `Word` é criado com um tamanho e capacidade iniciais de 0, o que significa que nosso `array` de suporte não existe ainda então nenhuma memória foi alocada para ele.

Na primeira vez que adicionamos um elemento usando `append()`, o `array` é criado com capacidade 1 e com nosso elemento dentro. Se adicionarmos um outro elemento, Go verá que o `array` está cheio, o que significa que ele precisa "crescer". Arrays não podem crescer, então o que o Go faz é criar um novo `array` com uma capacidade suficiente para comportar nossos novos elementos e mais um espaço extra (geralmente dobrando a capacidade), e então copia todo conteúdo do `array` anterior para o novo, e por fim adiciona os novos elementos. O `slice` então passa a apontar para este novo `array`.

Isto pode ser custoso tanto com memória, já que memória precisa ser alocada para o novo array, como cpu, já que o coletor de lixo terá que fazer a limpeza do array anterior (se nenhum outro `slice` estiver apontando para ele).

Agora você pode imaginar como crescer um `slice` de capacidade 0 até ~190000 pode ser custoso, já que o array precisará ser copiado e limpo pelo coletor várias e várias vezes até obtermos o resultado final.

## Solução

A solução é simples no nosso caso, temos um `slice` de n `string`s e precisamos de um novo de n `Word`s. Ao informar uma capacidade inicial para o `slice` resultante usando `make([]Word, 0, len(rawWords))`, estamos dizendo para o Go alocar memória suficiente para guardar todas as nossas `Word`s, fazendo com que o `slice` não precise crescer.

Estamos trocando um pouco de performance inicial durante a criação do `slice` por um ganho muito maior ao adicionar elementos.

Para um pouco mais de contexto sobre o ganho de performance, observe abaixo a quantidade de vezes que o coletor de lixo foi executado durante cada teste:

#### Sem capacidade inicial

```
gc 5 @0.007s 3%: 0.004+1.1+0.006 ms clock, 0.051+0/0.93/0.74+0.081 ms cpu, 4->4->4 MB, 5 MB goal, 12 P

....

gc 373 @1.390s 3%: 0.008+0.96+0.008 ms clock, 0.10+0.007/1.5/0.74+0.096 ms cpu, 8->9->4 MB, 9 MB goal, 12 P

```

#### Com capacidade inicial

```
gc 5 @0.007s 2%: 0.004+0.49+0.004 ms clock, 0.058+0/0.88/0.59+0.055 ms cpu, 4->5->5 MB, 5 MB goal, 12 P

....

gc 174 @1.428s 1%: 0.019+1.0+0.006 ms clock, 0.23+0/1.6/0.33+0.076 ms cpu, 15->16->7 MB, 16 MB goal, 12 P

```

Deixei somente a primeira e última linha de cada pois é o que importa para nós.

Com a nossa mudança, a coleta de lixo aconteceu 50% menos vezes.

## Conclusão

Como pode ser visto, é possível ganhar bastante performance com um pequeno ajuste, porém, essa solução não se aplica a todos os casos. E se não fossemos adicionar todos os 190000 elementos do primeiro `slice` no novo? Se por exemplo, só quiséssemos adicionar palavras de 7 caracteres ou mais, estaríamos alocando memória desnecessária, pois somente uma fração das palavras seria usada. É possível adaptar a solução para estes casos, mas isso é assunto para outro post.

Obrigado por ler, este é meu primeiro artigo então fique a vontade para dar dicas e feedback. :)

