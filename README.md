# Simplix

CRM de atendimento multicanal com suporte a WhatsApp (Meta Cloud API e QuePasa), pipeline de contatos, conversas em tempo real e painel de relatórios.

## Stack

| Camada | Tecnologia |
|--------|-----------|
| Backend | Go 1.25 · chi · pgx · goose · zerolog |
| Frontend | Nuxt 3 · Nuxt UI · VueUse |
| Banco | PostgreSQL 16 |
| WhatsApp | Meta Cloud API + QuePasa (self-hosted) |
| Infra | Docker Compose |

## Estrutura

```
simplix/
├── backend/          # API REST em Go
│   ├── cmd/api/      # Entrypoint do servidor
│   ├── cmd/migrate/  # Runner de migrations
│   ├── internal/
│   │   ├── domain/   # Models e tipos
│   │   ├── handlers/ # HTTP handlers
│   │   ├── repository/
│   │   └── service/  # WhatsApp, QuePasa, SSE
│   └── migrations/   # SQL (goose)
├── frontend/         # App Nuxt 3
│   └── app/
│       ├── pages/
│       ├── components/
│       └── composables/
└── docker-compose.yml
```

## Pré-requisitos

- Docker e Docker Compose
- Go 1.25+
- Node.js 20+ e pnpm

## Subindo o ambiente

**1. Infraestrutura (Postgres + QuePasa):**
```bash
docker compose up -d
```

**2. Backend:**
```bash
cp backend/.env.example backend/.env
# edite backend/.env com suas configurações

cd backend
go run ./cmd/migrate   # executa as migrations
go run ./cmd/api       # inicia o servidor na :8080
```

**3. Frontend:**
```bash
cd frontend
pnpm install
pnpm dev               # inicia em http://localhost:3000
```

## Variáveis de ambiente (backend)

| Variável | Descrição |
|----------|-----------|
| `DATABASE_URL` | Connection string do PostgreSQL |
| `JWT_SECRET` | Segredo para assinar tokens JWT |
| `PORT` | Porta do servidor (padrão: `8080`) |
| `ENV` | `development` ou `production` |
| `PUBLIC_URL` | URL pública do servidor (para registro de webhooks) |

## Funcionalidades

- Autenticação JWT com roles (admin / agent)
- Gestão de contatos com pipeline e score
- Empresas vinculadas a contatos
- Conversas multicanal (WhatsApp Meta, QuePasa, manual)
- Mensagens em tempo real via SSE
- Labels, notas, atributos customizáveis
- Relatórios com métricas e timeseries
- Webhooks de saída configuráveis
- Gerenciamento de usuários e inboxes
