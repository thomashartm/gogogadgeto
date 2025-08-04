# 🔍 GoGo Gadgeto Scanner

```
       .-""""""-.
      /          \
     |  .-""-.    |
     | /      \   |
     ||   ()   |  |
     |\        /  |
     | \  __  /   |
     |  '-..-'    |
     |    ||      |
     |    ||      |
     |   _||_     |
     |  [____]    |
     |    ||      |
     |    ||      |
    /|    ||      |\
   | |   _||_     | |
   | |  [____]    | |
   | |    ||      | |
   |_|____||______| |
     |    ||      |
     |    ||      |
     \____||_____/
          ||
       ___||___
      [_______]
   
   "Go Go Gadget Scanner!"
```

*An homage to the classic cartoon character Inspector Gadget*

An AI-powered security analysis and penetration testing tool with real-time chat interface and integrated tool execution capabilities.

## 🚀 Features

- **Real-time AI Chat Interface** - Interactive WebSocket-based communication with AI assistant
- **Security Analysis Tools** - Built-in penetration testing and web application analysis capabilities
- **Code Execution Environment** - Sandboxed Python script execution with Docker containers
- **Session Management** - Persistent chat sessions with auto-save functionality
- **Tool Flow Visualization** - Real-time reasoning and tool execution tracking
- **Preset Commands** - Pre-configured security analysis prompts
- **Responsive UI** - Modern React-based interface with syntax highlighting

## 🏗️ Architecture

### Backend (Go)
- **WebSocket Server** - Real-time communication using Gorilla WebSocket
- **AI Integration** - Powered by Cloudwego Eino framework
- **Sandboxed Execution** - Docker-based Python environment for safe code execution
- **Tool Management** - Dynamic tool registration and execution system

### Frontend (React)
- **Component-based Architecture** - Modular UI components
- **Real-time Updates** - WebSocket integration for live chat
- **Session Persistence** - Local storage for chat history
- **Syntax Highlighting** - Code formatting with react-syntax-highlighter

## 📋 Prerequisites

- **Go** 1.19 or higher
- **Node.js** 16 or higher
- **Docker** (for sandboxed execution)
- **Make** (for build automation)

## 🛠️ Installation

### 1. Clone the Repository
```bash
git clone <repository-url>
cd gogogadgeto
```

### 2. Backend Setup
```bash
cd server
go mod download
```

### 3. Frontend Setup
```bash
cd ui
npm install
```

### 4. Environment Configuration
Create a `.env` file in the server directory:
```env
# Add your API keys and configuration here
OPENAI_API_KEY=your_api_key_here
# Other environment variables as needed
```

## 🚀 Running the Application

### Development Mode

#### Start the Backend Server
```bash
cd server
make run
```
The server will start on `http://localhost:8080`

#### Start the Frontend Development Server
```bash
cd ui
npm start
```
The UI will be available at `http://localhost:5173`

### Production Build

#### Build Frontend
```bash
cd ui
npm run build
```

#### Build Backend
```bash
cd server
go build -o gogogajeto
```

## 📁 Project Structure

```
gogogadgeto/
├── server/                 # Go backend
│   ├── main.go            # Main server entry point
│   ├── chatmodel.go       # AI chat model integration
│   ├── tools.go           # Tool management and sandboxing
│   ├── prompts.go         # Prompt templates
│   ├── log.go             # Logging utilities
│   └── Makefile           # Build automation
├── ui/                    # React frontend
│   ├── src/
│   │   ├── app.jsx        # Main application component
│   │   ├── main.jsx       # Application entry point
│   │   ├── logo.jsx       # Inspector Gadget logo component
│   │   ├── presets.jsx    # Preset commands configuration
│   │   ├── components/    # UI components
│   │   │   ├── ChatPanel.jsx
│   │   │   ├── ResultsPanel.jsx
│   │   │   ├── ToolFlowPanel.jsx
│   │   │   ├── TopMenu.jsx
│   │   │   └── SessionControls.jsx
│   │   └── utils/         # Utility modules
│   │       ├── rendering.jsx
│   │       └── sessionManager.js
│   ├── package.json
│   └── index.html
└── README.md
```

## 🔧 Usage

### Basic Chat
1. Open the application in your browser
2. Type your security analysis questions in the chat input
3. View AI responses with syntax-highlighted code blocks
4. Monitor tool execution in the reasoning panel

### Preset Commands
- Use the preset selector below the chat input
- Choose from pre-configured security analysis commands
- Commands are automatically inserted into the chat input

### Session Management
- Sessions are automatically saved to local storage
- Chat history persists between browser sessions
- Use session controls to manage multiple analysis sessions

### Tool Execution
- Ask the AI to create and execute Python scripts
- Code runs in a secure Docker sandbox environment
- View execution results in real-time

## 🔐 Security Features

- **Sandboxed Execution** - All code runs in isolated Docker containers
- **Network Isolation** - Sandbox containers have no network access
- **Resource Limits** - CPU and memory constraints prevent resource abuse
- **Input Sanitization** - All user inputs are properly sanitized
- **No Direct File System Access** - Code execution is contained within sandbox

## 🎨 UI Components

- **TopMenu** - Application header with Inspector Gadget logo
- **ChatPanel** - Main chat interface with message history
- **ToolFlowPanel** - Real-time reasoning and tool execution tracking
- **ResultsPanel** - Analysis results and data visualization
- **SessionControls** - Session management and export functionality

## 🧪 Development

### Adding New Tools
1. Implement the tool interface in `server/tools.go`
2. Register the tool in the tool management system
3. Add tool-specific prompts in `server/prompts.go`

### Adding Preset Commands
1. Edit `ui/src/presets.jsx`
2. Add new preset objects with name and prompt properties
3. Presets will automatically appear in the UI selector

### Customizing the UI
1. Modify components in `ui/src/components/`
2. Update styling and layout as needed
3. All components use modern React hooks and functional components

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test thoroughly
5. Submit a pull request

## 📄 License

See the LICENSE file for details.

## 🎭 Acknowledgments

This application is an homage to the classic cartoon character Inspector Gadget. The "GoGo Gadgeto" name is used as a tribute to the beloved character and is not intended for commercial purposes.

---

**Warning**: This tool is designed for authorized security testing only. Always ensure you have proper authorization before conducting any security analysis or penetration testing activities.
