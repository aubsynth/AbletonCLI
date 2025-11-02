#!/usr/bin/env bash
# Installation script for AbletonCLI
# Compatible with bash and zsh

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
RESET='\033[0m'

# Default installation directory
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin/abletoncli}"

echo -e "${BLUE}AbletonCLI Installer${RESET}"
echo "===================="
echo ""
echo "     _    _     _      _               ____ _     ___ "
echo "    / \  | |__ | | ___| |_ ___  _ __  / ___| |   |_ _|"
echo "   / _ \ | '_ \| |/ _ \ __/ _ \| '_ \| |   | |    | | "
echo "  / ___ \| |_) | |  __/ || (_) | | | | |___| |___ | | "
echo " /_/   \_\_.__/|_|\___|\__\___/|_| |_|\____|_____|___|"
echo ""
echo "===================="
echo ""


# Check if binary exists
if [ ! -f "abletoncli" ]; then
    echo -e "${RED}Error: abletoncli binary not found${RESET}"
    echo "Please build it first with: go build"
    exit 1
fi

echo -e "${GREEN}✓${RESET} Found abletoncli binary"

# Create installation directory if it doesn't exist
if [ ! -d "$INSTALL_DIR" ]; then
    echo ""
    echo "Creating installation directory: $INSTALL_DIR"
    sudo mkdir -p "$INSTALL_DIR"
fi

# Install the binary
echo ""
echo "Installing binary to $INSTALL_DIR..."
if sudo cp abletoncli "$INSTALL_DIR/abletoncli"; then
    sudo chmod +x "$INSTALL_DIR/abletoncli"
    echo -e "${GREEN}✓${RESET} Binary installed to $INSTALL_DIR/abletoncli"
else
    echo -e "${RED}✗${RESET} Failed to install binary"
    exit 1
fi

# Install shell completion
echo ""
echo "Installing shell completion to $INSTALL_DIR..."
if sudo cp completion.sh "$INSTALL_DIR/completion.sh"; then
    sudo chmod +x "$INSTALL_DIR/completion.sh"
    echo -e "${GREEN}✓${RESET} Completion installed to $INSTALL_DIR/completion.sh"
else
    echo -e "${RED}✗${RESET} Failed to install completion"
    exit 1
fi

# Detect shell and set appropriate RC file
SHELL_RC=""
SHELL_NAME=""

# Detect the user's default shell from SHELL environment variable
USER_SHELL="${SHELL##*/}"  # Extract just the shell name (zsh, bash, etc.)

case "$USER_SHELL" in
    zsh)
        SHELL_NAME="zsh"
        if [ -f "$HOME/.zshrc" ]; then
            SHELL_RC="$HOME/.zshrc"
        else
            echo -e "${RED}Error: Your default shell is zsh but $HOME/.zshrc does not exist${RESET}"
            echo "Please create it first with: touch $HOME/.zshrc"
            exit 1
        fi
        ;;
    bash)
        SHELL_NAME="bash"
        if [ -f "$HOME/.bashrc" ]; then
            SHELL_RC="$HOME/.bashrc"
        elif [ -f "$HOME/.bash_profile" ]; then
            SHELL_RC="$HOME/.bash_profile"
        else
            echo -e "${RED}Error: Your default shell is bash but could not find .bashrc or .bash_profile${RESET}"
            echo "Please create one with: touch $HOME/.bashrc"
            exit 1
        fi
        ;;
    *)
        echo -e "${RED}Error: Unsupported shell: $USER_SHELL${RESET}"
        echo ""
        echo "Supported shells: zsh, bash"
        echo ""
        echo "Your SHELL environment variable is: $SHELL"
        echo "Please manually add to your shell config:"
        echo "  export PATH=\"/usr/local/bin/abletoncli:\$PATH\""
        echo "  source /usr/local/bin/abletoncli/completion.sh"
        exit 1
        ;;
esac

echo ""
echo -e "${GREEN}Detected shell: $SHELL_NAME${RESET}"
echo -e "${GREEN}Config file: $SHELL_RC${RESET}"

# Offer to update shell RC file
echo ""
echo -e "${YELLOW}Would you like to add abletoncli to your PATH and enable completions?${RESET}"
read -p "This will modify your $SHELL_RC (y/n): " update_rc

if [[ "$update_rc" == "y" || "$update_rc" == "Y" ]]; then
    # Check if already in RC file
    if grep -q "abletoncli" "$SHELL_RC" 2>/dev/null; then
        echo -e "${YELLOW}Note: abletoncli entries already exist in $SHELL_RC${RESET}"
        read -p "Update anyway? (y/n): " force_update
        if [[ "$force_update" != "y" && "$force_update" != "Y" ]]; then
            echo "Skipping $SHELL_RC update"
        else
            # Remove old entries and add new ones
            grep -v "abletoncli" "$SHELL_RC" > "${SHELL_RC}.tmp" 2>/dev/null || true
            mv "${SHELL_RC}.tmp" "$SHELL_RC"
        fi
    fi
    
    if [[ "$update_rc" == "y" || "$update_rc" == "Y" ]] && [[ "${force_update:-y}" != "n" && "${force_update:-y}" != "N" ]]; then
        echo "" >> "$SHELL_RC"
        echo "# Added by abletoncli installer" >> "$SHELL_RC"
        echo "export PATH=\"\$PATH:$INSTALL_DIR\"" >> "$SHELL_RC"
        echo "source $INSTALL_DIR/completion.sh" >> "$SHELL_RC"
        echo -e "${GREEN}✓${RESET} Updated $SHELL_RC"
        
        # Source the updated RC file in current shell
        echo ""
        echo "Applying changes to current shell..."
        export PATH="$PATH:$INSTALL_DIR"
        source "$INSTALL_DIR/completion.sh"
        echo -e "${GREEN}✓${RESET} Changes applied to current shell"
        echo ""
        echo -e "${GREEN}You can now use 'abletoncli' command!${RESET}"
    fi
else
    echo ""
    echo -e "${YELLOW}Manual setup required:${RESET}"
    echo "Add to your $SHELL_RC:"
    echo -e "${BLUE}  export PATH=\"\$PATH:$INSTALL_DIR\"${RESET}"
    echo -e "${BLUE}  source $INSTALL_DIR/completion.sh${RESET}"
    echo ""
    echo "Then run 'source $SHELL_RC' or restart your terminal"
fi

echo ""
echo -e "${GREEN}Installation complete!${RESET}"
echo ""
echo "Usage:"
echo "  abletoncli migrate --replace '/old/path' --with '/new/path' --directory ."
echo "  abletoncli migrate --replace '/old/path' --with '/new/path' --dry-run"
echo "  abletoncli backup --directory . --destination ../backup/"
echo "  abletoncli --help"
