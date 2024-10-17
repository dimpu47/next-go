# Stage 1: Base Node image for dependencies and builds
FROM node:18-alpine AS base

# Install libc6-compat for compatibility (optional but recommended for Alpine)
RUN apk add --no-cache libc6-compat
WORKDIR /app

# Stage 2: Install dependencies based on lockfiles
FROM base AS deps
COPY package.json yarn.lock* package-lock.json* pnpm-lock.yaml* ./
RUN \
  if [ -f yarn.lock ]; then yarn install --frozen-lockfile; \
  elif [ -f package-lock.json ]; then npm ci; \
  elif [ -f pnpm-lock.yaml ]; then yarn global add pnpm && pnpm install --frozen-lockfile; \
  else echo "Lockfile not found." && exit 1; \
  fi

# Stage 3: Build the Next.js app
FROM base AS builder
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY . .
RUN yarn build

# Stage 4: Final production image with Nginx
FROM nginx:alpine

# Copy the built static files and other assets from the build stage
COPY --from=builder /app/.next /usr/share/nginx/html/_next
COPY --from=builder /app/public /usr/share/nginx/html/public

# Copy custom Nginx configuration file (which you provide in the project)
COPY nginx.conf /etc/nginx/nginx.conf

# Expose port 80 for the Nginx server
EXPOSE 80

# Start Nginx
CMD ["nginx", "-g", "daemon off;"]
