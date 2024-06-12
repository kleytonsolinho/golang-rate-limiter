## Como funciona o rate limiter?

Ao disparar uma requisição ao endpoint ele irá verificar a quantidade de requisições por segundo que cada usuário está fazendo, utilizando o seu IP ou o Token no Header da requisição para identificar o limite de requisições por segundo que cada usuário poderá fazer.

## Como configurar o rate limiter?

Acesse o arquivo `.env` e adicione as configurações do rate limiter para que ele possar fazer a execução correta do programa.

<i>Abaixo um exemplo de como definir as variáveis de ambientes para configurar o rate limiter</i>

```typescript

RATE_LIMIT_WITH_IP_PER_SECOND=2          // Requisições/s com IP
RATE_LIMIT_WITH_TOKEN_PER_SECOND=5       // Requisições/s com Token
RATE_LIMIT_BLOCK_DURATION_IN_MINUTES=5   // Tempo de bloqueio em minutos ao atingir o limite
WEB_SERVER_PORT=8080                     // Porta da aplicação
````