#!/bin/bash
set -e

# GuhWizard Fresh Wrapper Script

INSTALLER="./guhwizard"
SUDOERS_FILE="/etc/sudoers.d/99-no-password-until-reboot"

echo "--- GuhWizard Fresh Orchestrator ---"

# 1. Always rebuild to ensure we are running the latest code
echo "Building the latest installer binary..."
go build -o guhwizard ./cmd/guhwizard
chmod +x guhwizard

# 2. Check for sudoers persistence
if [ ! -f "$SUDOERS_FILE" ]; then
    echo
    echo "[IMPORTANT] Root Setup Required"
    echo "This installer will set up a temporary sudo rule (valid until reboot)"
    echo "so you don't have to enter your password repeatedly during installation."
    echo "This requires ONE-TIME sudo authentication now."
    echo
    
    # Run installer in root-setup mode
    # We use -v to ensure we have a valid timestamp before running the command
    sudo -v
    if sudo "$INSTALLER" --root-setup; then
        echo "Root setup complete. Passwordless mode active until reboot."
    else
        echo "Root setup failed. You may not have sudo permissions or an error occurred."
        exit 1
    fi
else
    echo "Sudo persistence already configured ($SUDOERS_FILE). Proceeding..."
fi

echo
echo "Launching TUI Installer..."
echo "---------------------------"

# 3. Run the actual installer
# Since we just authenticated with 'sudo -v' and/or 'root-setup',
# and we have 'timestamp_timeout=-1', the Go app will run without prompting.
"$INSTALLER"

