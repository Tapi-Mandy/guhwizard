#!/bin/bash
# ===============================================================
# --- GUHWIZARD -------------------------------------------------
# ===============================================================

# Colors (Bash Side - Midnight Rose Theme)
MAGENTA='\033[1;35m'
ROSE='\033[0;31m' 
NC='\033[0m' # No Color

echo -e "${MAGENTA}[*] Initializing guhwizard...${NC}"

# 1. Check for Sudo
if [ "$EUID" -eq 0 ]; then
  echo -e "${ROSE}[!] Please run this script as a standard user (not root).${NC}"
  exit 1
fi

# 2. Dependencies
echo -e "${MAGENTA}[*] Installing system dependencies...${NC}"
sudo pacman -Sy --noconfirm python python-pip git base-devel > /dev/null 2>&1

# 3. TUI Libraries
echo -e "${MAGENTA}[*] Setting up Python TUI libraries...${NC}"
pip install rich questionary --break-system-packages --no-warn-script-location

if [ $? -ne 0 ]; then
    echo -e "${ROSE}[!] Failed to install Python libraries. Exiting.${NC}"
    exit 1
fi

# 4. Python Installer Generation
echo -e "${MAGENTA}[*] Launching guhwizard...${NC}"

cat << 'EOF' > installer.py
import os
import sys
import subprocess
import shutil
import time
from rich.console import Console
from rich.panel import Panel
from rich.text import Text
from rich.table import Table
from rich.align import Align
from rich.progress import Progress, SpinnerColumn, TextColumn
import questionary

# --- Configuration ---
# NOTE: If you change this URL, the script will automatically adapt to the new folder name.
REPO_URL = "https://github.com/Tapi-Mandy/guhwm"
REPO_NAME = REPO_URL.split("/")[-1].replace(".git", "")

console = Console()

# --- Colors & Styles (Midnight Rose Theme) ---
C_PRIMARY = "bold magenta"     
C_ACCENT = "#ff5faf"           
C_DARK = "#5f005f"             
C_DIM = "dim #d75f87"          

# --- Package Definitions ---
class Pkg:
    def __init__(self, name, desc, pkg_name=None, is_aur=False, binary_name=None, service_name=None):
        self.name = name
        self.desc = desc
        self.pkg_name = pkg_name if pkg_name else name.lower()
        self.is_aur = is_aur
        self.binary_name = binary_name if binary_name else self.pkg_name
        self.service_name = service_name if service_name else self.pkg_name

# --- Utilities ---
def run_cmd(cmd, shell=False, show_output=False):
    try:
        if show_output:
            subprocess.check_call(cmd, shell=shell)
        else:
            subprocess.check_call(cmd, shell=shell, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
        return True
    except subprocess.CalledProcessError:
        return False

def clear():
    os.system('cls' if os.name == 'nt' else 'clear')

def print_header():
    clear()
    ascii_art = r"""
              _               _                  _ 
   __ _ _   _| |__  __      _(_)______ _ _ __ __| |
  / _` | | | | '_ \ \ \ /\ / / |_  / _` | '__/ _` |
 | (_| | |_| | | | | \ V  V /| |/ / (_| | | | (_| |
  \__, |\__,_|_| |_|  \_/\_/ |_/___\__,_|_|  \__,_|
  |___/                                            
    """
    logo = Text(ascii_art, style=C_PRIMARY)
    subtitle = Text("Guh Window Manager Installer", style=C_DIM)
    content = Align.center(Text.assemble(logo, "\n", subtitle))
    console.print(Panel(content, border_style=C_DARK, expand=True))

def center_print(text_obj):
    console.print(Align.center(text_obj))

def install_config_file(src_path, dest_dir, file_name):
    if not os.path.exists(src_path): return
    full_dest_dir = os.path.expanduser(dest_dir)
    full_dest_path = os.path.join(full_dest_dir, file_name)
    os.makedirs(full_dest_dir, exist_ok=True)
    if os.path.exists(full_dest_path):
        if questionary.confirm(f"Config file '{file_name}' already exists in {dest_dir}. Overwrite?").ask():
            shutil.copy(src_path, full_dest_path)
            console.print(Align.center(f"[dim]Overwrote {file_name}[/dim]"))
        else:
            console.print(Align.center(f"[dim]Skipped {file_name}[/dim]"))
    else:
        shutil.copy(src_path, full_dest_path)
        console.print(Align.center(f"[dim]Installed {file_name}[/dim]"))

# --- Core Logic ---

def install_pacman_packages(packages):
    if not packages: return
    cmd = ["sudo", "pacman", "-S", "--noconfirm", "--needed"] + packages
    run_cmd(cmd, show_output=True)

def install_aur_package(package_name, helper):
    if not helper: return False
    cmd = [helper, "-S", "--noconfirm", "--needed", package_name]
    return run_cmd(cmd, show_output=True)

def setup_aur_helper(choice):
    if choice == "None": return None
    if shutil.which(choice.lower()): return choice.lower()
    console.print(Align.center(f"[yellow]Installing {choice}...[/yellow]"))
    build_dir = os.path.expanduser(f"~/{choice.lower()}_build_temp")
    repo = f"https://aur.archlinux.org/{choice.lower()}-bin.git"
    try:
        run_cmd(["git", "clone", repo, build_dir], show_output=True)
        os.chdir(build_dir)
        os.system("makepkg -si --noconfirm")
        os.chdir(os.path.expanduser("~"))
        shutil.rmtree(build_dir)
        return choice.lower()
    except Exception as e:
        console.print(Align.center(f"[red]Failed to install {choice}: {e}[/red]"))
        return None

# --- Categories ---

base_pkgs = [
    "xorg", "xorg-xinit", "libx11", "libxinerama", "libxft", "imlib2", "freetype2",
    "kitty", "picom", "rofi", "feh", "zip", "unzip", "jq", "alsa-utils", 
    "noto-fonts", "noto-fonts-cjk", "noto-fonts-emoji", 
    "ttf-dejavu", "ttf-fira-code", "ttf-jetbrains-mono", "ttf-jetbrains-mono-nerd"
]

browsers = [
    Pkg("Brave", "Privacy-focused browser blocking trackers", "brave-bin", is_aur=True),
    Pkg("Firefox", "Fast, Private & Safe Web Browser", "firefox"),
    Pkg("Librewolf", "Fork of Firefox focused on privacy", "librewolf-bin", is_aur=True),
    Pkg("Lynx", "Text-based web browser", "lynx"),
]

comm_apps = [
    Pkg("Discord", "All-in-one voice and text chat", "discord"),
    Pkg("Telegram", "Official Telegram Desktop client", "telegram-desktop"),
    Pkg("Vesktop", "The cutest Discord client", "vesktop-bin", is_aur=True),
    Pkg("Webcord", "Discord client that uses the web version", "webcord-bin", is_aur=True),
]

dev_tools = [
    Pkg("Emacs", "Extensible, customizable text editor", "emacs"),
    Pkg("Nano", "Simple terminal text editor", "nano"),
    Pkg("Neovim", "Fork of Vim aiming to improve user experience", "neovim"),
    Pkg("Sublime", "Sophisticated text editor for code", "sublime-text-4", is_aur=True),
    Pkg("Vim", "Highly configurable text editor", "vim"),
    Pkg("VSCodium", "Free/Libre Open Source binary of VSCode", "vscodium-bin", is_aur=True),
]

wall_apps = [
    Pkg("Waypaper", "GUI for wallpaper backends", "waypaper", is_aur=True)
]

misc_apps = [
    Pkg("Htop", "Interactive process viewer", "htop"),
    Pkg("Krita", "A full-featured free digital painting studio", "krita"),
    Pkg("Mpv", "Command line video player", "mpv"),
    Pkg("Redshift", "Adjusts screen color temperature", "redshift"),
    Pkg("Uwufetch", "Cute system information fetcher", "uwufetch"),
    Pkg("Yazi", "Blazing fast terminal file manager", "yazi"),
]

shells = [
    Pkg("Bash", "The GNU Bourne Again shell", "bash"),
    Pkg("Ksh", "KornShell, a classic Unix shell", "ksh"),
    Pkg("Oh My Zsh", "Community-driven framework for Zsh", "zsh"), 
    Pkg("Zsh", "Shell designed for advanced use", "zsh"),
]

dms = [
    Pkg("LightDM", "Lightweight display manager", "lightdm"),
    Pkg("Ly", "TUI display manager", "ly"),
    Pkg("SDDM", "QML based display manager", "sddm"),
]

# --- Main Execution ---

def main():
    print_header()

    # 1. Install Base
    print_header()
    center_print(Text("Base Packages", style=C_PRIMARY))
    console.print(Align.center("[dim]Installing base packages via pacman...[/dim]"))
    print() 
    install_pacman_packages(base_pkgs)
    print()
    center_print(Text("✔ Base packages installed.", style="green"))
    time.sleep(2)

    # 2. Welcome
    print_header()
    welcome_text = Text("Welcome to the guhwm installer.\nThis will set up your environment, install applications, and configure the window manager.", justify="center")
    welcome_text.stylize(C_ACCENT)
    console.print(Panel(Align.center(welcome_text), border_style=C_DARK, title="Welcome", title_align="center"))
    if not questionary.confirm("Ready to proceed?").ask():
        sys.exit()

    # 3. AUR Helper
    print_header()
    center_print(Text("AUR Helper Selection", style=C_PRIMARY))
    center_print(Text("Required for Brave, Vesktop, Waypaper, etc.", style="dim"))
    aur_choice = questionary.select("Choose an AUR helper:", choices=["Yay", "Paru", "None"]).ask()
    aur_helper = setup_aur_helper(aur_choice)

    # 4. Categories
    categories = [
        ("Browsers", browsers),
        ("Communication", comm_apps),
        ("Developer Tools", dev_tools),
        ("Wallpaper Manager", wall_apps),
        ("Miscellaneous Tools", misc_apps),
    ]

    for cat_name, pkg_list in categories:
        print_header()
        available_pkgs = [p for p in pkg_list if not (p.is_aur and not aur_helper)]
        disabled_pkgs = [p for p in pkg_list if p.is_aur and not aur_helper]

        table = Table(title=f"{cat_name}", border_style=C_DARK, header_style=C_PRIMARY)
        table.add_column("Software", style=C_ACCENT, no_wrap=True, justify="center")
        table.add_column("Source", style="cyan", justify="center")
        table.add_column("Description", style="white")

        for p in available_pkgs:
            source = "AUR" if p.is_aur else "Pacman"
            table.add_row(p.name, source, p.desc)
        if disabled_pkgs:
             table.add_row("[dim]Others[/dim]", "[dim]AUR[/dim]", f"[dim]({len(disabled_pkgs)} hidden)[/dim]")

        console.print(Align.center(table))
        
        choices = [p.name for p in available_pkgs] + ["None"]
        selected_names = questionary.checkbox(f"Select {cat_name} to install:", choices=choices, instruction=" ").ask()

        if selected_names and "None" not in selected_names:
            for name in selected_names:
                pkg = next((p for p in available_pkgs if p.name == name), None)
                if pkg:
                    if pkg.is_aur and aur_helper:
                        console.print(Align.center(f"Installing {pkg.name} (AUR)..."))
                        install_aur_package(pkg.pkg_name, aur_helper)
                    else:
                        console.print(Align.center(f"Installing {pkg.name} (Pacman)..."))
                        install_pacman_packages([pkg.pkg_name])

    # 5. Shell
    print_header()
    s_table = Table(title="Shell Selection", border_style=C_DARK, header_style=C_PRIMARY)
    s_table.add_column("Shell", style=C_ACCENT, justify="center")
    s_table.add_column("Description", justify="center")
    for s in shells: s_table.add_row(s.name, s.desc)
    console.print(Align.center(s_table))

    shell_choice = questionary.select("Which shell should be the default?", choices=[s.name for s in shells]).ask()
    if shell_choice:
        sel_shell = next(s for s in shells if s.name == shell_choice)
        install_pacman_packages([sel_shell.pkg_name])
        if shell_choice == "Oh My Zsh":
            console.print(Align.center("Installing Oh My Zsh..."))
            install_pacman_packages(["zsh", "curl", "git"])
            os.system('sh -c "$(curl -fsSL https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh)" "" --unattended')
            try: subprocess.run(["chsh", "-s", "/bin/zsh", os.environ.get("USER", "")])
            except: console.print(Align.center("[red]Could not auto-change shell. Do it manually.[/red]"))
        else:
            bin_path = f"/bin/{sel_shell.pkg_name}"
            try: subprocess.run(["chsh", "-s", bin_path, os.environ.get("USER", "")])
            except: pass

    # 6. Display Manager
    print_header()
    dm_table = Table(title="Display Managers", border_style=C_DARK, header_style=C_PRIMARY)
    dm_table.add_column("Manager", style=C_ACCENT, justify="center")
    dm_table.add_column("Description", justify="center")
    for d in dms: dm_table.add_row(d.name, d.desc)
    console.print(Align.center(dm_table))

    dm_choice = questionary.select("Select a Login Manager:", choices=[d.name for d in dms] + ["None"]).ask()
    if dm_choice != "None":
        sel_dm = next(d for d in dms if d.name == dm_choice)
        install_pacman_packages([sel_dm.pkg_name])
        if sel_dm.name == "LightDM": install_pacman_packages(["lightdm-gtk-greeter"])
        svc = sel_dm.service_name
        console.print(Align.center(f"Enabling {svc} service..."))
        os.system(f"sudo systemctl enable {svc}")

    # 7. Install GUHWM
    print_header()
    console.print(Align.center(Text(f"Installing {REPO_NAME}...", style=C_PRIMARY)))
    
    # FIX: DYNAMIC CLEANUP
    if os.path.exists(REPO_NAME):
        console.print(Align.center(f"[yellow]Removing previous {REPO_NAME} directory...[/yellow]"))
        os.system(f"sudo rm -rf {REPO_NAME}")
    
    run_cmd(["git", "clone", REPO_URL], show_output=True)
    
    if os.path.exists(REPO_NAME):
        # Configs
        console.print(Align.center("Checking configuration files..."))
        install_config_file(f"{REPO_NAME}/picom.conf", "~/.config/picom", "picom.conf")
        install_config_file(f"{REPO_NAME}/config.rasi", "~/.config/rofi", "config.rasi")
        
        # Wallpapers
        console.print(Align.center("Installing Wallpapers..."))
        os.system("sudo mkdir -p /usr/share/backgrounds/guhwm_wallpapers")
        if os.path.exists(f"{REPO_NAME}/Wallpapers"):
            os.system(f"sudo cp -r {REPO_NAME}/Wallpapers/* /usr/share/backgrounds/guhwm_wallpapers/")

        # Mod Key
        print_header()
        center_print(Text("Modifier Key Selection", style=C_PRIMARY))
        mod_choice = questionary.select("Which key as 'Mod' key?", choices=["Alt (Default / Mod1)", "Windows/Super (Mod4)"]).ask()
        if "Windows" in mod_choice:
            console.print(Align.center("[yellow]Applying Windows/Super key...[/yellow]"))
            c_path = f"{REPO_NAME}/dwm/config.def.h"
            if os.path.exists(c_path):
                with open(c_path, 'r') as f: c = f.read()
                c = c.replace("#define MODKEY Mod1Mask", "#define MODKEY Mod4Mask")
                with open(c_path, 'w') as f: f.write(c)

        # Compilation
        console.print(Align.center(Text("Compiling guhwm...", style=C_PRIMARY)))
        console.print(Align.center("[dim]Output is enabled to debug compilation errors.[/dim]"))
        
        targets = ["dwm", "slstatus"]
        for target in targets:
            t_path = os.path.join(REPO_NAME, target)
            if os.path.exists(t_path):
                console.print(Align.center(f"Compiling {target}..."))
                
                # Auto-Patch config.mk
                config_mk = os.path.join(t_path, "config.mk")
                if os.path.exists(config_mk):
                    try:
                        with open(config_mk, "r") as f: mk_data = f.read()
                        mk_data = mk_data.replace("/usr/X11R6/include", "/usr/include")
                        mk_data = mk_data.replace("/usr/X11R6/lib", "/usr/lib")
                        if "/usr/include/freetype2" not in mk_data:
                            mk_data = mk_data.replace("FREETYPEINC = /usr/include", "FREETYPEINC = /usr/include/freetype2")
                        mk_data = mk_data.replace("PREFIX = /usr/local", "PREFIX = /usr")
                        with open(config_mk, "w") as f: f.write(mk_data)
                        console.print(Align.center("[green]Patched config.mk for Arch Linux paths.[/green]"))
                    except Exception as e:
                        console.print(Align.center(f"[red]Warning: Could not patch config.mk: {e}[/red]"))

                os.chdir(t_path)
                
                # Clean, Build, Install
                exit_code = os.system("sudo make PREFIX=/usr clean install")
                
                if exit_code != 0:
                    console.print(Align.center(f"[bold red]CRITICAL ERROR: Failed to compile {target}.[/bold red]"))
                    console.print(Align.center("[red]See the error output above. Exiting.[/red]"))
                    sys.exit(1)
                    
                os.chdir("../..")
            else:
                console.print(Align.center(f"[yellow]Warning: {target} folder not found.[/yellow]"))

        # Verify Installation
        if not os.path.exists("/usr/bin/dwm"):
             console.print(Align.center(Text("[CRITICAL] dwm binary not found in /usr/bin. Compilation failed.", style="bold red")))
             sys.exit(1)

        # Session Script
        console.print(Align.center("Creating session startup script..."))
        wall_path = "/usr/share/backgrounds/guhwm_wallpapers/guhwm_midnight-rose.jpg"
        
        session_script = f"""#!/bin/sh
# --- guhwm session ---
/usr/bin/feh --bg-fill {wall_path} &
/usr/bin/picom -b &
/usr/bin/slstatus &
exec /usr/bin/dwm
"""
        with open("guhwm-session", "w") as f: f.write(session_script)
        os.system("sudo mv guhwm-session /usr/bin/guhwm-session")
        os.system("sudo chmod +x /usr/bin/guhwm-session")

        # Desktop Entry
        desktop_entry = """[Desktop Entry]
Encoding=UTF-8
Name=guhwm
Comment=Guh Window Manager
Exec=/usr/bin/guhwm-session
Icon=guhwm
Type=Application
"""
        with open("guhwm.desktop", "w") as f: f.write(desktop_entry)
        os.system("sudo mv guhwm.desktop /usr/share/xsessions/")
        
        console.print(Align.center(Text("✔ guhwm installed successfully.", style="green")))
    else:
        console.print(Align.center(Text("Failed to clone repository.", style="red")))

    # 8. Finish
    print_header()
    final_text = Text("\nInstallation Complete!\n\nguhwm and your selected applications are installed.\n", justify="center", style="bold green")
    console.print(Align.center(Panel(final_text, border_style="green", expand=False)))
    
    if questionary.confirm("Do you want to reboot now?").ask():
        os.system("sudo reboot")
    else:
        console.print("Exiting installer...")

if __name__ == "__main__":
    try:
        main()
    except KeyboardInterrupt:
        print("\nCancelled by user.")
        sys.exit(0)
EOF

# 5. Run
if [ -f "installer.py" ]; then
    python3 installer.py
else
    echo -e "${ROSE}[!] Error: installer.py was not created.${NC}"
fi

# 6. Cleanup
rm -f installer.py
