FROM node:20-alpine AS builder

WORKDIR /app

# Copy everything but defined in docker ignore file
COPY . .

# Copy dependencies and install
# COPY package.json package-lock.json* pnpm-lock.yaml* yarn.lock* ./

# # Install dependencies (adjust if you use yarn or pnpm)
# RUN npm ci
RUN npm install

# Copy all source files
# COPY . .

# Build the Next.js app
RUN npm run build

# ----------------------
# 2. Production Stage
# ----------------------
FROM node:20-alpine AS runner

WORKDIR /app

# # Install only production dependencies
# COPY package.json package-lock.json* ./
# # RUN npm ci --omit=dev
# RUN npm install

# # Copy built output and static assets
# COPY --from=builder /app/.next .next
# COPY --from=builder /app/public public
# COPY --from=builder /app/next.config.ts ./
# COPY --from=builder /app/package.json ./
# COPY --from=builder /app/tsconfig.json ./

# Copy standalone build output
COPY --from=builder /app/.next/standalone ./
COPY --from=builder /app/public ./public
COPY --from=builder /app/.next/static ./.next/static

# Start Next.js in standalone mode
EXPOSE 3000

# CMD ["npx", "next", "start"]
CMD ["node", "server.js"]
