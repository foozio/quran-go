# syntax=docker/dockerfile:1
FROM node:20-alpine AS build
WORKDIR /app
COPY web/package.json ./
RUN npm install --legacy-peer-deps
COPY web/ ./
RUN npm run build

FROM node:20-alpine
WORKDIR /srv/app
ENV PORT=8090
EXPOSE 8090
COPY --from=build /app/build ./build
RUN npm install @sveltejs/kit @sveltejs/adapter-node --no-save >/dev/null 2>&1 || true
CMD ["node","build/index.js"]
