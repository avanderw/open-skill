# Open-Skill web app — build static SvelteKit output, serve with nginx,
# reverse-proxy /api to the service container.
# NOTE: node:22-slim (Debian/glibc) — rolldown ships glibc prebuilds; the
# alpine (musl) variant fails with "Cannot find native binding"
# (npm/cli#4828, rolldown musl rollout gap).
FROM node:22-slim AS build
WORKDIR /src
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM nginx:1.27-alpine
COPY --from=build /src/build /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]