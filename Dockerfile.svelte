# syntax=docker/dockerfile:1
FROM node:20-alpine AS build
WORKDIR /app
COPY web/package.json ./
RUN npm install --legacy-peer-deps
ARG UTHMANIC_HAFS_URL="https://raw.githubusercontent.com/thetruetruth/quran-data-kfgqpc/main/hafs/font/hafs.18.woff2"
COPY web/ ./
# Optionally fetch Uthmanic Hafs font into static assets if URL provided
RUN mkdir -p web/static/fonts \
 && if [ -n "$UTHMANIC_HAFS_URL" ]; then echo "Fetching Uthmanic Hafs from $UTHMANIC_HAFS_URL"; \
    ext=${UTHMANIC_HAFS_URL##*.}; out=web/static/fonts/uthmanic_hafs.$ext; \
    wget -qO "$out" "$UTHMANIC_HAFS_URL" || true; fi
RUN npm run build

FROM node:20-alpine
WORKDIR /srv/app
ENV PORT=8090
EXPOSE 8090
COPY --from=build /app/build ./build
RUN npm install @sveltejs/kit @sveltejs/adapter-node --no-save >/dev/null 2>&1 || true
CMD ["node","build/index.js"]
