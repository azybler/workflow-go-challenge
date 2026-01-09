# ⚡ Workflow Automation Web Application

A modern web application for creating and managing workflow automations using a visual drag-and-drop editor powered by @xyflow/react. Design, edit, and manage workflows with a modern, accessible UI built with Radix UI and Tailwind CSS.

## 🛠️ Features

- Visual workflow editor with drag-and-drop interface (@xyflow/react)
- Support for multiple node types:
  - Trigger nodes (HTTP endpoints)
  - Action nodes (HTTP requests, email notifications)
  - Condition nodes (workflow branching)
- Real-time node property editing
- Responsive, accessible interface using Radix UI and Tailwind CSS
- Type-safe development with TypeScript
- State management with Zustand

## 🚀 Quick Start

### Prerequisites

- Node.js 18.x or later
- npm 9.x or later

### 1. Install dependencies

```bash
npm install
```

### 2. Start the development server

```bash
npm run dev
```

### 3. Build for production

```bash
npm run build
```

## 🏗️ Project Structure

```text
web/
├── src/
│   ├── components/     # React components
│   │   └── nodes/     # Workflow node components
│   ├── theme/         # Theme configuration and design tokens
│   ├── App.tsx        # Main application component
│   ├── main.tsx       # Application entry point
│   └── types.ts       # TypeScript type definitions
├── public/            # Static public assets
├── index.html         # HTML entry point
├── package.json       # Project dependencies and scripts
├── tsconfig.json      # TypeScript configuration
├── vite.config.ts     # Vite configuration
├── nginx.conf         # Nginx configuration for production
└── Dockerfile         # Docker configuration for production
```

## 📦 Available Scripts

- `npm run dev` - Start development server with hot reload
- `npm run build` - Build for production
- `npm run preview` - Preview production build locally
- `npm run lint` - Run ESLint
- `npm run format` - Format code with Prettier
- `npm run type-check` - Run TypeScript type checking

## 🔗 Core Dependencies

- React 18+
- @xyflow/react (React Flow)
- Zustand (State Management)
- Axios (HTTP Client)
- UUID
- Radix UI
- Tailwind CSS

## 🐳 Docker Support

The app can be containerized using the provided Dockerfile. The container uses Nginx to serve static files in production.

```bash
docker build -t workflow-web .
docker run -p 80:80 workflow-web
```
