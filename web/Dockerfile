# Development stage
FROM node:20-alpine

WORKDIR /app

# Copy package files
COPY package*.json ./

# Install dependencies
RUN npm ci

# Copy source code
COPY . .

# Expose port 3000
EXPOSE 3000

# Use a shell form to allow environment variable substitution
CMD if [ "$NODE_ENV" = "production" ]; then \
        npm run build && \
        npx serve -s dist -l 3000; \
    else \
        npm run dev -- --host 0.0.0.0; \
    fi 