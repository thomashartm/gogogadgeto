# Docker Containers for GoGoGadgeto

This directory contains Docker configurations for building pre-configured containers used by GoGoGadgeto.

## Kali Linux Security Tools Container

### Overview

The Kali container (`gogogadgeto/kali-tools`) provides a pre-built environment with security and information gathering tools ready to use.

### Pre-installed Tools

- **Network Scanning**: nmap, masscan, netdiscover
- **DNS/Domain**: whois, dig, nslookup, host
- **Web Testing**: nikto, dirb, gobuster (with wordlists), whatweb
- **Network Utils**: curl, wget, netcat, traceroute, ping
- **Enumeration**: enum4linux, sublist3r, theharvester
- **File Sharing**: smbclient, showmount
- **System Tools**: net-tools, rpcbind

### Wordlists & Security Data

The container includes comprehensive wordlists for brute force and fuzzing:

- **DIRB Lists** (`/usr/share/wordlists/dirb/`):
  - `common.txt` - Common web directories (35KB) - **Default for gobuster**
  - `big.txt` - Large directory list (180KB)
  - Various language-specific lists

- **SecLists** (`/usr/share/wordlists/seclists/`):
  - `Discovery/Web-Content/` - Web content discovery
  - `Fuzzing/` - Various fuzzing wordlists  
  - `Passwords/` - Password lists for brute force
  - `Payloads/` - Security payload collections

- **Password Lists**:
  - `rockyou.txt` - Popular passwords (140MB, extracted)

#### Wordlist Usage Examples
```bash
# Gobuster with default DIRB common list
gobuster dir -u http://target.com -w /usr/share/wordlists/dirb/common.txt

# Gobuster with larger DIRB list  
gobuster dir -u http://target.com -w /usr/share/wordlists/dirb/big.txt

# Gobuster with SecLists
gobuster dir -u http://target.com -w /usr/share/wordlists/seclists/Discovery/Web-Content/common.txt

# DIRB with custom wordlist
dirb http://target.com /usr/share/wordlists/dirb/big.txt
```

### Building the Container

#### Quick Build
```bash
cd docker
./build-kali.sh
```

#### Build Options
```bash
# Build without cache (clean build)
./build-kali.sh --no-cache

# Quiet build (less output)
./build-kali.sh --quiet

# Both options
./build-kali.sh --no-cache --quiet
```

#### Background Build
```bash
# Build in background and log output
nohup ./build-kali.sh > build.log 2>&1 &

# Monitor progress
tail -f build.log
```

### Testing the Container

The build script automatically tests the container, but you can manually test:

```bash
# Test nmap
docker run --rm gogogadgeto/kali-tools nmap --version

# Test nikto
docker run --rm gogogadgeto/kali-tools nikto -Version

# Test gobuster and wordlists
docker run --rm gogogadgeto/kali-tools gobuster --help
docker run --rm gogogadgeto/kali-tools test -f /usr/share/wordlists/dirb/common.txt && echo "Wordlists OK"

# Interactive shell
docker run --rm -it gogogadgeto/kali-tools /bin/bash
```

### Usage in GoGoGadgeto

The container is automatically used when you start the server. The Go code references:

```go
Image: "gogogadgeto/kali-tools:latest"
```

### Container Details

- **Base Image**: kalilinux/kali-rolling
- **Size**: ~2.8GB (includes comprehensive wordlists)
- **Working Dir**: /workspace
- **Health Check**: Verifies tools and wordlists are available
- **Network**: Enabled for information gathering
- **Wordlists**: Pre-extracted and ready to use

### Build Time

- **First build**: ~5-10 minutes (downloads packages)
- **Subsequent builds**: ~1-2 minutes (uses cache)

### Troubleshooting

#### Build Fails
```bash
# Clean build without cache
./build-kali.sh --no-cache

# Check Docker is running
docker info
```

#### Tool Not Found
```bash
# Verify tool installation
docker run --rm gogogadgeto/kali-tools which nmap

# Check all tools
docker run --rm gogogadgeto/kali-tools ls /usr/bin | grep -E "(nmap|nikto|gobuster)"
```

#### Container Won't Start
```bash
# Check image exists
docker images gogogadgeto/kali-tools

# Rebuild if corrupted
./build-kali.sh --no-cache
```

### Adding New Tools

1. Edit `Dockerfile.kali`
2. Add tool to the `apt-get install` section
3. Update the tools list in `../agent/tools/tools.go`
4. Rebuild: `./build-kali.sh --no-cache`

### Performance Tips

- Build the container before starting development
- Use `--quiet` for automated builds
- Consider using a Docker registry for team sharing
- The container can be built once and used multiple times 