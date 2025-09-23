#!/bin/sh

sed -i "s|%%VITE_AUTH_API_BASE_URL%%|${VITE_AUTH_API_BASE_URL}|g" ./dist/index.html
sed -i "s|%%VITE_STORAGE_API_BASE_URL%%|${VITE_STORAGE_API_BASE_URL}|g" ./dist/index.html

npm run release --no-cache