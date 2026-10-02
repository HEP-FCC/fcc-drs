#!/bin/sh
# Copies the npm-managed frontend libraries' minified browser builds out of
# node_modules into static/vendor/. Run after `npm ci`. Shared by the
# Makefile's `assets` target and the Dockerfile's `assets` build stage so
# the two never drift apart.
set -eu

VENDOR="${1:-static/vendor}"

mkdir -p "$VENDOR/katex/fonts" "$VENDOR/fonts/files"
cp node_modules/htmx.org/dist/htmx.min.js             "$VENDOR/"
cp node_modules/marked/lib/marked.umd.js              "$VENDOR/marked.min.js"
cp node_modules/dompurify/dist/purify.min.js          "$VENDOR/"
cp node_modules/bulma/css/bulma.min.css               "$VENDOR/"
cp node_modules/katex/dist/katex.min.css              "$VENDOR/katex/"
cp node_modules/katex/dist/katex.min.js               "$VENDOR/katex/"
cp node_modules/katex/dist/contrib/auto-render.min.js "$VENDOR/katex/"
cp node_modules/katex/dist/fonts/*.woff2              "$VENDOR/katex/fonts/"
cp node_modules/@fontsource-variable/inter/wght.css                  "$VENDOR/fonts/inter.css"
cp node_modules/@fontsource-variable/inter/files/*-wght-normal.woff2 "$VENDOR/fonts/files/"
