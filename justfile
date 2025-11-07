# justfile

# Node.js is bundled with Zed — expose it for recipes that need npm
export PATH := env_var('HOME') + "/.local/share/zed/node/node-v24.11.0-linux-x64/bin:" + env_var('PATH')

# run the web server (requires built frontend: just build-web)
run-server:
    go run cmd/server/main.go

# build the SvelteKit frontend
build-web:
    cd web && npm run build

gci-all:
    for file in $(find . -name '*.go'); do \
        gci write "$file" -s "alias" -s "standard" -s "default" -s "prefix(plown)"; \
    done
