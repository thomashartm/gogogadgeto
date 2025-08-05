# 🔍 GoGo Gadgeto Security Tool

An AI-powered security analysis and penetration testing tool with real-time chat interface, integrated security tools, and containerized execution environments.

## 📋 Prerequisites

- **Go** 1.19 or higher
- **Node.js** 16 or higher  
- **Docker** (for sandboxed execution and Kali tools)
- **Make** (for build automation)
- **OpenAI API Key** (for AI functionality)

## 🛠️ Quick Start

### 1. Clone and Setup
```bash
git clone <repository-url>
cd gogogadgeto

# Complete project setup (dependencies + containers)
make setup
```

### 2. Configure Environment
Create a `.env` file in the server directory:
```env
OPENAI_API_KEY=your_api_key_here
OPENAI_BASE_URL=https://api.openai.com/v1  # optional
OPENAI_MODEL=gpt-4o-mini                    # optional
```

### 3. Start Development Environment
```bash
# Start both servers (UI + Backend)
make dev
```
- **UI**: http://localhost:5173
- **API**: http://localhost:8080

### 4. Stop Everything
```bash
make stop-all
```

## 🚀 Central Build System

The project includes a comprehensive Makefile for easy management:

### 🏃 **Quick Commands**
```bash
make           # Show help with all available commands
make setup     # Complete project setup
make dev       # Start development environment
make prod      # Build and start production
make stop-all  # Stop all services
```

### 🐳 **Container Management**
```bash
make build-kali  # Build Kali Linux security tools container
make check-kali  # Verify Kali container status
```

### 🏗️ **Build Commands**
```bash
make build-all     # Build everything
make build-ui      # Build React frontend
make build-server  # Build Go backend
```

### 🖥️ **Server Management**
```bash
make start-all     # Start both servers
make start-server  # Start Go backend only
make start-ui      # Start React UI only
```

### 🧹 **Maintenance**
```bash
make clean-all     # Clean all build artifacts
make check-deps    # Verify all dependencies
make install-deps  # Install all dependencies
make status        # Check system status
```

## 🔧 Usage Guide

### 🔧 **Tool Execution**
Ask the AI to use security tools with natural language:
```
"Use nmap to scan 192.168.1.1 for open services"
"Get whois information for example.com"  
"Perform a DNS lookup for the MX records of google.com"
"Use nikto to scan http://testsite.com for vulnerabilities"
```

### 📋 **Prompt Presets**
Enhanced preset commands for common security tasks:
- Network scanning and port analysis
- Domain and DNS investigation  
- Web application vulnerability testing
- Information gathering workflows

### 💾 **Session Management**
- **Auto-save** every 30 seconds
- **Persistent storage** of chat history, tool flows, and UI layout
- **Session isolation** - each conversation starts fresh
- **Export capabilities** for analysis results

## 🧪 Development

### 🛠️ **Adding New Security Tools**
1. Edit `server/docker/Dockerfile.kali` to add the tool
2. Update `server/agent/tools/tools.go` with tool definitions
3. Rebuild container: `make build-kali`
4. Test: `make check-kali`

### 📋 **Adding Preset Commands** 
1. Edit `ui/src/presets.jsx`
2. Add new preset objects with descriptive names
3. Include example targets and proper syntax
4. UI automatically updates with new presets

## 🐳 Container Documentation

Detailed container information is available in:
- `server/docker/README.md` - Complete container documentation
- `server/docker/build-kali.sh` - Build script with options
- `server/docker/Dockerfile.kali` - Container definition

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/amazing-feature`
3. Make your changes and test thoroughly
4. Ensure containers build: `make build-kali`
5. Test the full system: `make dev`
6. Submit a pull request

## 🔧 Troubleshooting

### Common Issues

**Container Build Fails**
```bash
make build-kali          # Try building the container
docker system prune      # Clean up Docker if needed
make setup               # Reinstall everything
```

**Servers Won't Start**
```bash
make check-deps          # Verify all dependencies
make stop-all           # Stop any running processes
make dev                # Restart development environment
```

**API Key Issues**
```bash
# Ensure server/.env contains:
OPENAI_API_KEY=your_actual_key_here
OPENAI_MODEL=your_mode_goes_here e.g. gpt-4.1
OPENAI_API_BASE=https://api.openai.com/v1
OPENAI_API_VERSION=2023-05-15
```

**Port Conflicts**
- Backend runs on port 8080
- Frontend runs on port 5173  
- Stop other services using these ports

### Getting Help
- Check `make status` for system overview
- Review logs in terminal output
- Verify Docker containers: `docker ps`
- Test Kali tools: `make check-kali`

## 📄 License

See the LICENSE file for details.

## 🎭 Acknowledgments

This application is an homage to the classic cartoon character Inspector Gadget and the implementation language GoLang. 
The "GoGo Gadgeto" name is used as a tribute to the beloved character and his ability to pull out all purpose tools for any situation and any given moment just by saying go go gadgeto .... 


The tool leverages excellent open-source projects including:
- **Cloudwego Eino** - AI framework integration
- **Kali Linux** - Security tools and methodologies  
- **React and Vite** - Modern web interface
- **Go** - High-performance backend
- **Docker** - Containerization and sandboxing

---

**⚠️ Legal Notice**: This tool is designed for authorized security testing only. Always ensure you have proper authorization before conducting any security analysis or penetration testing activities. Users are responsible for complying with all applicable laws and regulations.
