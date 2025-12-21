#!/bin/bash
set -e

# GuhWizard Wrapper Script

INSTALLER="./guhwizard"
SUDOERS_FILE="/etc/sudoers.d/99-no-password-until-reboot"

# Check if installer binary exists
if [ ! -f "$INSTALLER" ]; then
    echo "Building installer..."
    go build -o guhwizard ./cmd/guhwizard
fi

echo "Welcome to the GuhWizard Installer Wrapper."

# 1. Check for sudoers persistence
if [ ! -f "$SUDOERS_FILE" ]; then
    echo "This installer requires temporary passwordless sudo access to run smoothly."
    echo "We will now set up a temporary sudoers rule (valid until reboot)."
    echo "You may be asked for your sudo password one last time."
    echo
    
    # Run installer in root-setup mode
    if sudo "$INSTALLER" --root-setup; then
        echo "Root setup complete."
    else
        echo "Root setup failed. Exiting."
        exit 1
    fi
else
    echo "Sudo persistence already configured. Proceeding..."
fi

echo
echo "Launching Installer..."
echo

# 2. Run the actual installer as USER (no sudo needed for launch, internal sudo calls will work)
"$INSTALLER"
