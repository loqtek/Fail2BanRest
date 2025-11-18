#!/bin/bash
# Fail2Rest V2 Installation Script
# Usage: curl -fsSL https://raw.githubusercontent.com/loqtek/Fail2BanRest/main/install.sh | sudo bash
# Uninstall: curl -fsSL https://raw.githubusercontent.com/loqtek/Fail2BanRest/main/install.sh | sudo bash -s uninstall
# Uninstall (with config): curl -fsSL https://raw.githubusercontent.com/loqtek/Fail2BanRest/main/install.sh | sudo bash -s uninstall --remove-config

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
INSTALL_DIR="/opt/fail2rest"
SERVICE_NAME="fail2rest"
BINARY_NAME="fail2restV2"
CONFIG_DIR="/etc/fail2rest"
SERVICE_FILE="/etc/systemd/system/${SERVICE_NAME}.service"
REPO_URL="https://github.com/loqtek/Fail2BanRest.git"
REPO_RAW_URL="https://raw.githubusercontent.com/loqtek/Fail2BanRest"
REPO_BRANCH="main"

# Functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

check_root() {
    if [[ $EUID -ne 0 ]]; then
        log_error "This script must be run as root (use sudo)"
        echo ""
        log_info "Please run:"
        echo "  curl -fsSL https://raw.githubusercontent.com/loqtek/Fail2BanRest/main/install.sh | sudo bash"
        echo ""
        exit 1
    fi
}

check_dependencies() {
    log_info "Checking dependencies..."
    
    local missing_deps=()
    
    if ! command -v go &> /dev/null; then
        missing_deps+=("go")
    fi
    
    if ! command -v git &> /dev/null; then
        missing_deps+=("git")
    fi
    
    if ! command -v fail2ban-client &> /dev/null; then
        log_warning "fail2ban-client not found. Make sure fail2ban is installed."
    fi
    
    if [[ ${#missing_deps[@]} -gt 0 ]]; then
        log_error "Missing dependencies: ${missing_deps[*]}"
        log_info "Install them with: apt-get install -y ${missing_deps[*]}"
        exit 1
    fi
    
    log_success "All dependencies found"
}

check_fail2ban() {
    log_info "Checking fail2ban installation..."
    
    if ! command -v fail2ban-client &> /dev/null; then
        log_warning "fail2ban-client not found"
        read -p "Install fail2ban? (y/n): " -r
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            apt-get update
            apt-get install -y fail2ban
        else
            log_error "fail2ban is required. Exiting."
            exit 1
        fi
    fi
    
    # Check if fail2ban is running
    if ! systemctl is-active --quiet fail2ban 2>/dev/null; then
        log_warning "fail2ban service is not running"
        read -p "Start fail2ban service? (y/n): " -r
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            systemctl start fail2ban
            systemctl enable fail2ban
        fi
    fi
    
    log_success "fail2ban is available"
}

install_application() {
    log_info "Installing Fail2Rest V2..."
    
    # Create directories
    mkdir -p "$INSTALL_DIR"
    mkdir -p "$CONFIG_DIR"
    
    # Clone or update repository
    if [[ -d "$INSTALL_DIR/.git" ]]; then
        log_info "Updating existing installation..."
        cd "$INSTALL_DIR"
        git fetch origin
        git reset --hard "origin/$REPO_BRANCH" || {
            log_warning "Failed to update, continuing with existing code"
        }
    else
        log_info "Cloning repository..."
        rm -rf "$INSTALL_DIR" 2>/dev/null || true
        git clone -b "$REPO_BRANCH" "$REPO_URL" "$INSTALL_DIR" || {
            log_error "Failed to clone repository"
            log_error "Make sure git is installed and you have internet access"
            exit 1
        }
    fi
    
    cd "$INSTALL_DIR"
    
    # Build application
    log_info "Building application..."
    go mod download || {
        log_error "Failed to download dependencies"
        exit 1
    }
    
    go build -o "$BINARY_NAME" ./cmd/server || {
        log_error "Failed to build application"
        log_error "Make sure Go 1.21+ is installed"
        exit 1
    }
    
    # Build hash-password tool
    if [[ -d "./cmd/hash-password" ]]; then
        go build -o hash-password ./cmd/hash-password || {
            log_warning "Failed to build hash-password tool (optional)"
        }
    fi
    
    # Make binaries executable
    chmod +x "$BINARY_NAME"
    [[ -f hash-password ]] && chmod +x hash-password
    
    log_success "Application built successfully"
}

setup_config() {
    log_info "Setting up configuration..."
    
    local config_file="$CONFIG_DIR/config.yaml"
    
    if [[ -f "$config_file" ]]; then
        log_warning "Configuration file already exists at $config_file"
        read -p "Overwrite? (y/n): " -r
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            log_info "Keeping existing configuration"
            return
        fi
    fi
    
    # Copy example config
    cp "$INSTALL_DIR/config.example.yaml" "$config_file"
    
    # Generate JWT secret
    local jwt_secret=$(openssl rand -hex 32)
    sed -i "s/change-this-to-a-secure-random-string/$jwt_secret/" "$config_file"
    
    # Set fail2ban client path
    local f2b_path=$(which fail2ban-client)
    if [[ -n "$f2b_path" ]]; then
        sed -i "s|/usr/bin/fail2ban-client|$f2b_path|" "$config_file"
    fi
    
    # Configure permissions
    chmod 600 "$config_file"
    
    log_success "Configuration created at $config_file"
    log_warning "Please edit $config_file to configure:"
    log_warning "  - API keys or users for authentication"
    log_warning "  - Server host/port"
    log_warning "  - TLS certificates (if using HTTPS)"
}

setup_systemd() {
    log_info "Setting up systemd service..."
    
    # Check if service already exists
    if systemctl list-unit-files | grep -q "^${SERVICE_NAME}.service"; then
        log_warning "Service already exists"
        read -p "Reinstall service? (y/n): " -r
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            log_info "Keeping existing service"
            return
        fi
        systemctl stop "$SERVICE_NAME" 2>/dev/null || true
    fi
    
    # Create service file
    # Note: Running as root is required for fail2ban socket access
    cat > "$SERVICE_FILE" <<EOF
[Unit]
Description=Fail2Rest V2 - REST API for Fail2ban
After=network.target fail2ban.service
Requires=fail2ban.service

[Service]
Type=simple
User=root
Group=root
WorkingDirectory=$INSTALL_DIR
ExecStart=$INSTALL_DIR/$BINARY_NAME -config $CONFIG_DIR/config.yaml
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal

# Security settings
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/log

[Install]
WantedBy=multi-user.target
EOF
    
    log_info "Service configured to run as root (required for fail2ban access)"
    
    # Reload systemd
    systemctl daemon-reload
    
    # Enable service
    systemctl enable "$SERVICE_NAME"
    
    log_success "Systemd service created"
}

setup_sudo() {
    log_info "Checking fail2ban permissions..."
    
    # Test if fail2ban is accessible
    if fail2ban-client status &>/dev/null; then
        log_success "fail2ban is accessible (service runs as root)"
        return
    fi
    
    # Since the service runs as root, it should have access
    # But let's verify the socket exists
    if [[ -S /var/run/fail2ban/fail2ban.sock ]]; then
        log_success "fail2ban socket found (service will run as root)"
        return
    fi
    
    log_warning "fail2ban socket not found - make sure fail2ban is running"
    log_info "The service is configured to run as root, which allows access to fail2ban"
    log_info "If you prefer non-root, you can:"
    log_info "  1. Edit $SERVICE_FILE and change User/Group"
    log_info "  2. Configure sudo access for fail2ban-client"
    log_info "  3. Set use_sudo: true in $CONFIG_DIR/config.yaml"
}

start_service() {
    log_info "Starting service..."
    
    systemctl start "$SERVICE_NAME"
    
    # Wait a moment and check status
    sleep 2
    
    if systemctl is-active --quiet "$SERVICE_NAME"; then
        log_success "Service started successfully"
        
        # Get port from config
        local port=$(grep -E "^  port:" "$CONFIG_DIR/config.yaml" | awk '{print $2}' | tr -d '"' || echo "8080")
        log_info "API is available at: http://localhost:$port"
        log_info "Health check: http://localhost:$port/health"
    else
        log_error "Service failed to start"
        log_info "Check status with: systemctl status $SERVICE_NAME"
        log_info "Check logs with: journalctl -u $SERVICE_NAME -f"
    fi
}

uninstall() {
    log_warning "This will remove Fail2Rest V2 and all its files"
    
    # Check if running non-interactively (piped input)
    if [[ ! -t 0 ]]; then
        # Non-interactive mode - proceed with uninstall
        log_info "Running in non-interactive mode - proceeding with uninstall"
    else
        # Interactive mode - ask for confirmation
        echo ""
        log_warning "Type 'yes' to confirm uninstall, or anything else to cancel:"
        read -r
        if [[ ! $REPLY == "yes" ]]; then
            log_info "Uninstall cancelled"
            exit 0
        fi
    fi
    
    log_info "Stopping service..."
    systemctl stop "$SERVICE_NAME" 2>/dev/null || true
    systemctl disable "$SERVICE_NAME" 2>/dev/null || true
    
    log_info "Removing service file..."
    rm -f "$SERVICE_FILE"
    systemctl daemon-reload
    
    log_info "Removing sudoers entry..."
    rm -f /etc/sudoers.d/fail2rest
    
    log_info "Removing application files..."
    rm -rf "$INSTALL_DIR"
    
    # In non-interactive mode, keep config by default (safer)
    # In interactive mode, ask
    if [[ -t 0 ]]; then
        read -p "Remove configuration file? (y/n): " -r
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            log_info "Removing configuration..."
            rm -rf "$CONFIG_DIR"
        else
            log_info "Keeping configuration at $CONFIG_DIR"
        fi
    else
        log_info "Keeping configuration at $CONFIG_DIR (use --remove-config to delete)"
        # Check for --remove-config flag
        if [[ "${*}" == *"--remove-config"* ]]; then
            log_info "Removing configuration..."
            rm -rf "$CONFIG_DIR"
        fi
    fi
    
    log_success "Uninstall complete"
}

show_status() {
    log_info "Service status:"
    systemctl status "$SERVICE_NAME" --no-pager -l || true
    
    echo ""
    log_info "Useful commands:"
    echo "  Start:   systemctl start $SERVICE_NAME"
    echo "  Stop:    systemctl stop $SERVICE_NAME"
    echo "  Restart: systemctl restart $SERVICE_NAME"
    echo "  Status:  systemctl status $SERVICE_NAME"
    echo "  Logs:    journalctl -u $SERVICE_NAME -f"
}

main() {
    echo ""
    echo "=========================================="
    echo "  Fail2Rest V2 Installation Script"
    echo "=========================================="
    echo ""
    
    # Check for uninstall flag
    if [[ "${1:-}" == "uninstall" ]]; then
        check_root
        uninstall "$@"
        exit 0
    fi
    
    if [[ "${1:-}" == "status" ]]; then
        show_status
        exit 0
    fi
    
    # Installation
    check_root
    check_dependencies
    check_fail2ban
    install_application
    setup_config
    setup_systemd
    setup_sudo
    
    # Ask to start service
    read -p "Start service now? (y/n): " -r
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        start_service
    else
        log_info "Service installed but not started"
        log_info "Start it with: systemctl start $SERVICE_NAME"
    fi
    
    echo ""
    log_success "Installation complete!"
    echo ""
    log_info "Next steps:"
    echo "  1. Edit $CONFIG_DIR/config.yaml"
    echo "  2. Add API keys or users for authentication"
    echo "  3. Restart service: systemctl restart $SERVICE_NAME"
    echo ""
    log_info "Uninstall: curl -fsSL $REPO_RAW_URL/$REPO_BRANCH/install.sh | bash -s uninstall"
    echo ""
}

# Run main function
main "$@"

