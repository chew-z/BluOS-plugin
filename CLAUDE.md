# CLAUDE.md

## GOLANG

@./GOLANG.md

### Learning about Golang packages

@./GODOC.md

## Language server (LSP) is your superpower

1) After every file edit, the language server pushes diagnostics: type errors, missing imports, undefined variables. Claude Code sees these immediately and fixes them in the same turn, before you ever see the error.

After writing or editing code, check LSP diagnostics and fix errors before proceeding.

2) Beyond automatic diagnostics, Claude Code can explicitly ask the language server questions:

* goToDefinition — "Where is processOrder defined?" → exact file and line
* findReferences — "Find all places that call validateUser" → every call site with location
* hover — "What type is the config variable?" → full type signature and docs
* documentSymbol — "List all functions in this file" → every symbol with location
* workspaceSymbol — "Find the PaymentService class" → search symbols across the entire project
* goToImplementation — "What classes implement AuthProvider?" → concrete implementations of interfaces
* incomingCalls / outgoingCalls — "What calls processPayment?" → full call hierarchy tracing

Use LSP for code navigation — it's fast, precise, and avoids reading entire files:

LSP uses gopls for Golang

## Working with Postgres database

- Your access to Postgres database IS and SHOULD BE read-only
- You are allowed to use MCP tools for read-only access to Postges database
- You are FORBIDEN from writing hack scripts that would write to Postgres database for quick fixes. THIS IS STRICTLY FORBIDDEN HACK!
- For changing database structure, adding tables, database functions, indexes etc. you MUST create SQL scripts in `migrations` folder
- Don't use psql expressions in migration scripts.
- User runs migrations scripts himself via Postico Postgres console
