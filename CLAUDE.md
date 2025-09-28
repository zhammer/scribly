# scribly - Project Context for Claude

## Overview
Scribly is a personal web application for collaborative story writing between friends. It's built with a traditional web app approach prioritizing accessibility and progressive enhancement.

## Technology Stack
- **Backend**: Go (Golang) using gorilla/mux for routing
- **Database**: PostgreSQL with Bun ORM and Sqitch for migrations
- **Frontend**: Server-rendered HTML templates with minimal JavaScript and CSS
- **Testing**: Cypress with BDD/Cucumber framework for end-to-end testing
- **Hosting**: Vercel (serverless functions via api/index.go)

## Key Architecture Principles
- Traditional web app (server-rendered HTML)
- Accessible/a11y-friendly website
- JavaScript not required for basic functionality
- Expressive architecture and code
- Tests that ensure core functionality without being invasive

## Project Structure
- `cmd/site/` - Main web application server
- `cmd/scribbot/` - Bot functionality 
- `api/` - Vercel serverless function entry point
- `internal/` - Core business logic and services
- `pkg/` - Shared packages (db, helpers)
- `embed/` - Static assets and templates
- `cypress/` - BDD/Cucumber test suite
- `sqitch/` - Database schema migrations

## Development Commands
- `docker-compose up` - Start local development server and database (runs on http://127.0.0.1:8000)
- `go run cmd/site/main.go` - Run local development server directly (alternative to Docker)
- `yarn cypress run` - Run Cypress tests (requires server running)
- `yarn cypress open` - Open Cypress GUI for interactive testing
- `sqitch deploy` - Apply database migrations

## Testing
Uses Cypress with cucumber-preprocessor for BDD-style testing. Tests cover core user flows like signup, login, story creation, and collaboration features.

## Database
PostgreSQL with structured migrations via Sqitch. Uses Bun ORM for query building and database interactions.