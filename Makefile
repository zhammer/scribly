.PHONY: env dev

env:
	vercel env pull

dev: env
	set -a && source .env.local && set +a && air
