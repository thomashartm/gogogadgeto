package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"gogogajeto/util"
	"log"
	"strings"
	"time"

	"github.com/cloudwego/eino-ext/components/tool/commandline"
	"github.com/cloudwego/eino-ext/components/tool/commandline/sandbox"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

// NewKaliSandbox creates a Kali Linux container for security tools
func NewKaliSandbox(ctx context.Context) *sandbox.DockerSandbox {
	sb, err := sandbox.NewDockerSandbox(ctx, &sandbox.Config{
		Image:          "gogogadgeto/kali-tools:latest", // Use pre-built container
		HostName:       "kali-pentest",
		WorkDir:        "/workspace",
		MemoryLimit:    1024 * 1024 * 1024, // 1GB for Kali tools
		CPULimit:       2.0,                // More CPU for security tools
		NetworkEnabled: true,               // Enable network for information gathering
		Timeout:        time.Minute * 2,    // Normal timeout since tools are pre-installed
	})
	if err != nil {
		log.Fatal(err)
	}
	err = sb.Create(ctx)
	if err != nil {
		log.Fatal(err)
	}

	util.LogMessage("Kali Linux container ready with pre-installed security tools")
	return sb
}

// Predefined Kali information gathering tools (matching pre-built container)
var KaliInfoGatheringTools = map[string]string{
	"nmap":         "Network discovery and security auditing",
	"masscan":      "Fast network scanner",
	"netdiscover":  "Network discovery tool",
	"whois":        "Domain registration information lookup",
	"dig":          "DNS lookup utility (from dnsutils)",
	"nslookup":     "DNS lookup utility (from dnsutils)",
	"host":         "DNS lookup utility (from dnsutils)",
	"ping":         "Network connectivity test (built-in)",
	"traceroute":   "Network path tracing",
	"netstat":      "Network connections and statistics (from net-tools)",
	"ss":           "Socket statistics (built-in)",
	"curl":         "HTTP client for web requests",
	"wget":         "Web content downloader",
	"nc":           "Netcat for network connections (netcat-openbsd)",
	"nikto":        "Web vulnerability scanner",
	"dirb":         "Web directory brute forcer",
	"gobuster":     "Directory/file brute forcer with comprehensive wordlists",
	"whatweb":      "Web technology identification",
	"enum4linux":   "Linux/Samba enumeration tool",
	"smbclient":    "SMB client for file sharing",
	"showmount":    "NFS exports information (from nfs-common)",
	"rpcinfo":      "RPC services information (from rpcbind)",
	"sublist3r":    "Subdomain enumeration tool",
	"theharvester": "Email and subdomain harvester",
}

// Available wordlists in the container
var AvailableWordlists = map[string]string{
	"dirb/common.txt": "Common web directories (DIRB default)",
	"dirb/big.txt":    "Large directory list (DIRB)",
	"seclists/Discovery/Web-Content/common.txt": "Common web content (SecLists)",
	"rockyou.txt": "Popular passwords list",
}

// KaliInfoGatheringTool implements a Kali Linux information gathering tool
type KaliInfoGatheringTool struct {
	sandbox commandline.Operator
}

func NewKaliInfoGatheringTool(ctx context.Context, sb commandline.Operator) *KaliInfoGatheringTool {
	return &KaliInfoGatheringTool{
		sandbox: sb,
	}
}

func (k *KaliInfoGatheringTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	// Create description with available tools
	var toolsList []string
	for tool, desc := range KaliInfoGatheringTools {
		toolsList = append(toolsList, fmt.Sprintf("%s: %s", tool, desc))
	}

	// Create wordlists information
	var wordlistsList []string
	for wordlist, desc := range AvailableWordlists {
		wordlistsList = append(wordlistsList, fmt.Sprintf("%s: %s", wordlist, desc))
	}

	description := fmt.Sprintf(`Kali Linux Information Gathering Tool

This tool provides access to information gathering and reconnaissance tools from Kali Linux.

Available tools:
%s

Available wordlists:
%s

Usage:
- tool: The information gathering tool to use (e.g., "nmap", "whois", "dig", "gobuster")
- target: The target to investigate (IP address, domain, URL, etc.)
- options: Additional command line options for the tool (optional)

Examples:
- {"tool": "nmap", "target": "192.168.1.1", "options": "-sV -sC"}
- {"tool": "whois", "target": "example.com"}
- {"tool": "dig", "target": "example.com", "options": "MX"}
- {"tool": "nikto", "target": "http://example.com"}
- {"tool": "gobuster", "target": "http://example.com"} (uses default dir mode + common wordlist)
- {"tool": "gobuster", "target": "http://example.com", "options": "-w /usr/share/wordlists/dirb/big.txt"}
- {"tool": "gobuster", "target": "http://example.com", "options": "dir -w /usr/share/wordlists/seclists/Discovery/Web-Content/common.txt"}
- {"tool": "gobuster", "target": "example.com", "options": "dns -w /usr/share/wordlists/seclists/Discovery/DNS/bitquark-subdomains-top100000.txt"}
- {"tool": "dirb", "target": "http://example.com"} (uses default common wordlist)
- {"tool": "dirb", "target": "http://example.com", "options": "/usr/share/wordlists/dirb/big.txt"}

Security Notice: This tool is for authorized security testing only. Ensure you have permission before scanning any targets.`, strings.Join(toolsList, "\n"), strings.Join(wordlistsList, "\n"))

	return &schema.ToolInfo{
		Name: "kali_info_gathering",
		Desc: description,
	}, nil
}

func (k *KaliInfoGatheringTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	// Parse JSON arguments
	var params struct {
		Tool    string `json:"tool"`
		Target  string `json:"target"`
		Options string `json:"options,omitempty"`
	}

	if err := json.Unmarshal([]byte(argumentsInJSON), &params); err != nil {
		return "", fmt.Errorf("failed to parse input parameters: %v", err)
	}

	// Validate tool
	if params.Tool == "" {
		return "", fmt.Errorf("tool parameter is required")
	}

	if params.Target == "" {
		return "", fmt.Errorf("target parameter is required")
	}

	// Check if tool is in allowed list
	if _, exists := KaliInfoGatheringTools[params.Tool]; !exists {
		availableTools := make([]string, 0, len(KaliInfoGatheringTools))
		for tool := range KaliInfoGatheringTools {
			availableTools = append(availableTools, tool)
		}
		return "", fmt.Errorf("tool '%s' is not available. Available tools: %s", params.Tool, strings.Join(availableTools, ", "))
	}

	// Build command with special handling for different tools
	var command string

	switch params.Tool {
	case "gobuster":
		// Gobuster requires specific syntax: gobuster <mode> -u <url> [options]
		// Default to 'dir' mode if no mode specified in options
		if params.Options != "" {
			// Check if mode is already specified in options
			if strings.Contains(params.Options, "dir") || strings.Contains(params.Options, "dns") || strings.Contains(params.Options, "vhost") || strings.Contains(params.Options, "fuzz") {
				command = fmt.Sprintf("%s %s -u %s", params.Tool, params.Options, params.Target)
			} else {
				// Default to dir mode and append options
				command = fmt.Sprintf("%s dir -u %s %s", params.Tool, params.Target, params.Options)
			}
		} else {
			// Default gobuster dir scan with common wordlist
			command = fmt.Sprintf("%s dir -u %s -w /usr/share/wordlists/dirb/common.txt", params.Tool, params.Target)
		}
	case "dirb":
		// DIRB syntax: dirb <url> [wordlist] [options]
		if params.Options != "" {
			command = fmt.Sprintf("%s %s %s", params.Tool, params.Target, params.Options)
		} else {
			// Default dirb with common wordlist
			command = fmt.Sprintf("%s %s /usr/share/wordlists/dirb/common.txt", params.Tool, params.Target)
		}
	default:
		// Standard command format for other tools
		if params.Options != "" {
			command = fmt.Sprintf("%s %s %s", params.Tool, params.Options, params.Target)
		} else {
			command = fmt.Sprintf("%s %s", params.Tool, params.Target)
		}
	}

	// Tools that may return useful output even with non-zero exit codes
	// We append "|| true" to ensure exit code 0 while preserving all output
	toolsWithValidNonZeroOutput := map[string]bool{
		"whois":      true, // whois returns exit code 1 for "no match" but shows useful info
		"nmap":       true, // nmap may return exit code 1 for various reasons but still provide scan results
		"dig":        true, // dig may return exit code 1 for NXDOMAIN but still shows DNS info
		"nslookup":   true, // nslookup may return exit code 1 for failed lookups but shows info
		"host":       true, // host may return exit code 1 for failed lookups but shows info
		"nikto":      true, // nikto may return exit code 1 but still provide scan results
		"ping":       true, // ping may return exit code 1 for unreachable hosts but shows attempts
		"traceroute": true, // traceroute may return exit code 1 but shows partial route
	}

	// For tools that can have useful output with non-zero exit codes, append "|| true"
	if toolsWithValidNonZeroOutput[params.Tool] {
		command = command + " || true"
		util.LogMessage(fmt.Sprintf("Appending '|| true' to command for tool: %s", params.Tool))
	}

	// Execute command in Kali sandbox using RunCommand
	result, err := k.sandbox.RunCommand(ctx, command)
	if err != nil {
		return "", fmt.Errorf("failed to execute %s: %v", params.Tool, err)
	}

	return fmt.Sprintf("Kali %s Results:\n%s", params.Tool, result), nil
}

// NewKaliCommandLineTool creates Kali Linux information gathering tools
func NewKaliCommandLineTool(ctx context.Context, kaliSb commandline.Operator) []tool.BaseTool {
	kaliTool := NewKaliInfoGatheringTool(ctx, kaliSb)
	return []tool.BaseTool{kaliTool}
}
