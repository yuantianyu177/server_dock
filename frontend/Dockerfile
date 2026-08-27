# Build stage
FROM node:26.7.0-alpine3.24 AS builder

WORKDIR /app
COPY package.json package-lock.json ./
RUN npm install --global npm@12.0.2 && npm ci

COPY . .
RUN npm run build

# Runtime stage
FROM nginx:1.31.4-alpine3.24

COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
