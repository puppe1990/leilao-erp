# Design: leilao-erp — Mini ERP de leilão (módulo financeiro v1)

**Data:** 2026-07-27  
**App:** `leilao-erp`  
**Stack base:** [Cais](https://github.com/puppe1990/cais) (Go + Inertia.js + Svelte 5 + Tailwind + SQLite)  
**Status:** aprovado em brainstorming; aguardando plano de implementação

---

## 1. Contexto e objetivos

Negócio de **revenda de itens de leilão**. Primeiro lote real: **22 monitores por R$603,00**.

### Objetivos da v1

1. Registrar compras de leilão (lotes) com custos e rateio unitário.
2. Registrar vendas (direto e marketplace) com taxas e frete.
3. Controlar **caixa**, **contas a pagar** e **contas a receber**.
4. Ver **margem por venda** e resumo no dashboard.
5. Rodar em **VPS/Lightsail** com login, acessível de qualquer lugar.

### Fora de escopo (v1)

- Multi-empresa / multi-filial
- NF-e / fiscal
- Estoque multi-depósito
- Conciliação bancária automática (Open Finance)
- Contabilidade double-entry / balancete formal
- App mobile nativo
- Cadastro público de usuários
- Multi-usuário com papéis (estrutura de auth única agora; multi-user depois sem reescrever domínio)

---

## 2. Decisões de produto

| Decisão | Escolha |
|---------|---------|
| Escopo financeiro | Completo: caixa + a pagar/receber + custo por lote/item |
| Usuários | 1 usuário agora; multi-usuário depois |
| Canais de venda | Misto: marketplace + direto |
| Deploy v1 | VPS/Lightsail |
| Nome do app | `leilao-erp` |
| Abordagem | Domínio de leilão + financeiro **derivado** das operações |

### Regra de ouro

O usuário registra **fatos de negócio** (compra de lote, venda de item).  
Lançamentos de caixa e títulos a pagar/receber são **gerados** a partir desses fatos (com status/data ajustáveis na quitação).

---

## 3. Arquitetura

### Scaffold

```bash
# workspace: /Users/matheuspuppe/Desktop/negocios
cais new leilao-erp . --module github.com/puppe1990/leilao-erp
cais g auth
```

Layout padrão Cais:

- `cmd/server` — HTTP
- `internal/app` — bootstrap, rotas, middleware
- `internal/handlers` — Inertia handlers
- `internal/store` — SQLite + migrations
- `internal/models` — structs de domínio
- `web/src/pages` — Svelte 5 (Inertia)
- Auth por sessão (`cais g auth`); signup desabilitado ou inexistente em produção

### Módulos v1 (ordem de valor)

| Módulo | Responsabilidade |
|--------|------------------|
| Auth | Login único |
| Lotes & Itens | Compra no leilão, custos, rateio unitário |
| Vendas | Canal, preço, taxas, frete, status de pagamento |
| Financeiro | Caixa, a pagar, a receber, dashboard |

### Dinheiro e moeda

- Valores em **centavos** (`INTEGER`), nunca `float` para dinheiro.
- Moeda única: **BRL**.

---

## 4. Modelo de dados

```
Lot (lote de leilão)
├── Item[]            (cada peça)
├── PurchaseCost[]    (arremate, frete, comissão…)
│
Sale (venda de 1 item)
├── channel
├── fees / frete
└── gera → Receivable e/ou CashEntry

Payable
Receivable
CashAccount
CashEntry
```

### Campos

#### Lot

| Campo | Tipo | Notas |
|-------|------|--------|
| id | PK | |
| name | string | ex. "Monitores — leilão Jul/2026" |
| auction_source | string? | casa de leilão / URL |
| purchased_at | date | |
| status | enum | `open`, `partial`, `sold`, `closed` |
| notes | text? | |

#### Item

| Campo | Tipo | Notas |
|-------|------|--------|
| id | PK | |
| lot_id | FK → Lot | |
| sku | string? | |
| title | string | ex. "Monitor" |
| condition | string? | |
| unit_cost_cents | int | rateado; congelado na venda |
| status | enum | `in_stock`, `reserved`, `sold` |
| sale_price_hint_cents | int? | preço sugerido opcional |

#### PurchaseCost

| Campo | Tipo | Notas |
|-------|------|--------|
| id | PK | |
| lot_id | FK → Lot | |
| label | string | "Arremate", "Frete"… |
| amount_cents | int | > 0 |
| payable_id | FK? → Payable | se gerou título |

#### Sale

| Campo | Tipo | Notas |
|-------|------|--------|
| id | PK | |
| item_id | FK → Item | 1 item por venda na v1 |
| sold_at | datetime | |
| channel | enum | `direct`, `mercadolivre`, `shopee`, `olx`, `other` |
| gross_cents | int | |
| fee_cents | int | default 0 |
| shipping_cents | int | default 0 |
| net_cents | int | gross − fee − shipping |
| payment_status | enum | `received`, `pending`, `cancelled` |
| unit_cost_cents_at_sale | int | snapshot do custo do item |

#### Payable

| Campo | Tipo | Notas |
|-------|------|--------|
| id | PK | |
| description | string | |
| amount_cents | int | |
| due_on | date | |
| status | enum | `open`, `paid`, `cancelled` |
| lot_id | FK? | |
| paid_at | datetime? | |

#### Receivable

| Campo | Tipo | Notas |
|-------|------|--------|
| id | PK | |
| description | string | |
| amount_cents | int | |
| due_on | date | |
| status | enum | `open`, `received`, `cancelled` |
| sale_id | FK? | |
| received_at | datetime? | |

#### CashAccount

| Campo | Tipo | Notas |
|-------|------|--------|
| id | PK | |
| name | string | ex. "PIX principal" |
| kind | enum | `pix`, `bank`, `cash`, `other` |
| opening_balance_cents | int | default 0 |

#### CashEntry

| Campo | Tipo | Notas |
|-------|------|--------|
| id | PK | |
| account_id | FK → CashAccount | |
| direction | enum | `in`, `out` |
| amount_cents | int | > 0 |
| occurred_at | datetime | |
| category | string | ex. `compra_lote`, `venda`, `taxa`, `frete`, `ajuste` |
| memo | text? | |
| sale_id | FK? | |
| payable_id | FK? | |
| receivable_id | FK? | |
| lot_id | FK? | |

### Rateio de custo

```
total_costs = sum(PurchaseCost.amount_cents do lote)
qty         = count(itens do lote)
unit_base   = floor(total_costs / qty)
remainder   = total_costs % qty
→ primeiros `remainder` itens recebem unit_base + 1 centavo
→ demais recebem unit_base
```

**Exemplo seed:** 22 monitores, R$603,00 → `60300` centavos.

- `60300 / 22 = 2740` (R$27,40)
- resto `20` → 20 itens com R$27,41 e 2 com R$27,40

### Recálculo de custo

- Novo `PurchaseCost` no lote **recalcula** `unit_cost_cents` apenas dos itens com status `in_stock` (e `reserved`, se existirem).
- Itens `sold` **não** mudam: usam `unit_cost_cents_at_sale` na margem histórica.

### Como o financeiro nasce

| Ação do usuário | Efeito no sistema |
|-----------------|-------------------|
| Novo lote + custos “já paguei” | CashEntry(s) `out` + Payable(s) já `paid` (ou só CashEntry; preferência: Payable quitado para histórico) |
| Novo lote + custos “ainda devo” | Payable(s) `open` |
| Venda “recebi agora” (PIX) | Sale + CashEntry `in` + item `sold` |
| Venda marketplace “libera em N dias” | Sale + Receivable `open` + item `sold` |
| Quitar payable | status `paid` + CashEntry `out` |
| Quitar receivable | status `received` + CashEntry `in` |

### Margem por venda

```
margin_cents = net_cents - unit_cost_cents_at_sale
```

### Soft-delete / cancelamento

- Sem hard-delete de fatos financeiros na v1.
- Cancelamento lógico: Sale / Payable / Receivable → `cancelled`.
- Regras de estorno de CashEntry no cancelamento: na v1, cancelar só se ainda não houver quitação conflitante; ou gerar lançamento inverso de ajuste — **preferência:** só permitir cancelar Sale com `payment_status=pending` (receivable open) revertendo item para `in_stock` e cancelando receivable; vendas já recebidas exigem “estorno manual” (CashEntry ajuste) em fase posterior se necessário. Implementação mínima documentada no plano.

**Regra v1 explícita para cancelamento de venda:**

1. `payment_status=pending` + receivable `open` → cancela sale + receivable; item volta `in_stock`.
2. `payment_status=received` → **não** cancela pela UI na v1; usuário registra ajuste manual no caixa se precisar.

---

## 5. Telas e rotas

UI em **PT-BR**, mobile-friendly. Tudo autenticado (exceto login).

### Navegação

```
Dashboard
Lotes → detalhe
Vendas
Financeiro
  ├─ Caixa
  ├─ A pagar
  └─ A receber
Config (contas de caixa)
```

### Telas

| Tela | Função |
|------|--------|
| Dashboard | Saldos · a receber 7/30d · a pagar · lucro do mês · margem média · vencidos |
| Lotes (lista) | Nome, data, custo total, estoque/vendidos, status |
| Lote (detalhe) | Custos, rateio, itens · +custo · ações de pagamento |
| Novo lote (wizard) | Dados → custos (+ pago?) → itens (qtd/título) → grava e rateia |
| Vendas (lista) | Data, item, canal, bruto, taxas, líquido, status |
| Nova venda | Item em estoque · canal · valores · recebi agora vs a receber |
| Caixa | Extrato por conta · filtros · lançamento manual (categoria `ajuste`) |
| A pagar / A receber | Listas · botão Quitar → CashEntry |
| Login | E-mail/senha Cais |

### Fluxos principais

1. **Compra leilão:** wizard lote → 22 monitores R$603 já pago no PIX.
2. **Venda direta:** nova venda → PIX na hora → entrada no caixa + margem.
3. **Venda ML:** nova venda → receivable N dias → quitar quando liberar.
4. **Frete posterior:** +custo no lote → recalcula só estoque.

### Validações

- Não vender item que não está `in_stock`.
- Quitar payable/receivable só se `open`.
- Valores monetários > 0 onde aplicável (fee/shipping podem ser 0).
- Lote precisa de pelo menos 1 item e 1 custo (ou custo total > 0) para rateio válido.
- Lançamento manual de caixa: categoria `ajuste` (ou similares), sem apagar histórico ligado a sale.

### Empty state

Primeiro acesso: CTA “Registrar primeira compra de leilão”. Seed opcional dos monitores.

---

## 6. Erros e resiliência

| Situação | Resposta |
|----------|----------|
| Vender item já vendido | 422 + mensagem clara |
| Quitar título já quitado | 422 |
| Validação de form | Erros de campo Inertia |
| SQLite busy | Retry curto; se falhar, mensagem amigável |
| Sessão expirada | Redirect login + flash |
| Prod sem secrets | Boot falha (padrão Cais) |

---

## 7. Testes

TDD obrigatório (convenção Cais / AGENTS.md).

| Camada | Cobertura mínima |
|--------|------------------|
| Store | Rateio 60300/22; freeze de custo; quitação gera CashEntry; venda bloqueia item sold |
| Handlers | Fluxos compra/venda direta/venda pending + 422s |
| Frontend (Vitest) | Dashboard com saldos; form venda inválido não submete |

**Fixture de ouro:** lote monitores R$603 / 22 itens em testes e seed opcional (`cais db seed`).

---

## 8. Deploy Lightsail

```
VPS Ubuntu
  → binário server (cais build --os linux --arch amd64)
  → SQLite em volume persistente (/var/lib/leilao-erp/app.db)
  → systemd (template Cais)
  → Caddy/nginx + HTTPS
  → ENV=production, SESSION_SECRET, APP_URL, DB_PATH
```

- Signup público desligado.
- Backup diário do `.db` (checklist README; cron mínimo).
- Seed de usuário admin no primeiro boot ou via console/seed documentado.

---

## 9. Ordem de implementação (alto nível)

1. Scaffold `cais new leilao-erp` + auth + CashAccount
2. Lotes / PurchaseCost / Itens + rateio
3. Vendas + canais
4. Payable / Receivable + quitação
5. Caixa + Dashboard
6. Seed monitores + README deploy

Detalhamento task-by-task fica no **plano de implementação** (`docs/superpowers/plans/…`) após aprovação deste spec.

---

## 10. Critérios de sucesso da v1

- [ ] Usuário loga e registra o lote de 22 monitores por R$603 com custo unitário correto (centavos).
- [ ] Vende 1 item no PIX e vê entrada no caixa + margem.
- [ ] Vende 1 item no marketplace com receivable e quita depois.
- [ ] Dashboard mostra saldo, a pagar/receber e lucro do período.
- [ ] App sobe em Lightsail com HTTPS e SQLite persistente.

---

## 11. Riscos e mitigações

| Risco | Mitigação |
|-------|-----------|
| Escopo “ERP completo” estourar | Ordem de módulos; cancelamento de venda received adiado |
| Float em dinheiro | Só centavos INTEGER |
| Perda de DB na VPS | Path persistente + backup documentado |
| Recálculo de custo errado | Testes de rateio + freeze em sold |
| Generators Cais HTMX vs Svelte | Preferir páginas Svelte/Inertia alinhadas ao scaffold atual do `cais new` |

---

## Aprovação

- Abordagem B (domínio + financeiro derivado): **sim**
- Arquitetura: **sim**
- Modelo de dados: **sim**
- Telas e fluxos: **sim**
- Erros, testes, deploy: **sim**
- Nome: **leilao-erp**
