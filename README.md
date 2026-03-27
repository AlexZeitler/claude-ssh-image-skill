# claude-ssh-image-skill

A Claude Code skill + local daemon that enables pasting clipboard images into a Claude Code session on a remote server over SSH.

Claude Code doesn't natively support pasting images over SSH. This project solves that with a similar approach to [sshimg.nvim](https://github.com/AlexZeitler/sshimg.nvim), but specifically for Claude Code.

## How it works

```
Local Machine                            Remote Server (SSH)
┌──────────────────────┐                 ┌──────────────────────────┐
│  Clipboard (PNG)     │                 │  Claude Code             │
│        │             │                 │        │                 │
│        ▼             │                 │        ▼                 │
│  ccimgd (Port 9998)  │◄────────────────│  /paste-image Skill      │
│  - wl-paste          │  SSH Reverse    │  - TCP Request to ccimgd │
│  - Returns base64    │  Tunnel         │  - Receives base64 image │
│    image in response │  (Port 9998)    │  - Saves as temp file    │
│                      │                 │  - Read → Claude sees    │
└──────────────────────┘                 │    the image             │
                                         └──────────────────────────┘
```

1. The `/paste-image` skill runs the `ccimg` client on the remote server, which sends a TCP request to `127.0.0.1:9998` (forwarded through the SSH reverse tunnel to the local machine)
2. `ccimgd` reads the PNG image from the local clipboard (`wl-paste` on Wayland, `xclip` on X11)
3. The image is returned as base64-encoded JSON
4. `ccimg` saves it as a temporary PNG file and prints the path
5. The skill uses Claude's `Read` tool to display the image

## Requirements

- **Local machine**: `wl-paste` (Wayland, part of `wl-clipboard`) or `xclip` (X11), Go (only for building)
- **Remote server**: Claude Code

## Building

Build both binaries (daemon + client):

```bash
./build.sh
```

This produces `daemon/ccimgd` and `client/ccimg` as statically linked binaries with no runtime dependencies.

## Setup

### Local machine

Install the daemon:

```bash
cp daemon/ccimgd ~/.local/bin/
cp daemon/ccimgd.service ~/.config/systemd/user/
systemctl --user daemon-reload
systemctl --user enable --now ccimgd
```

### Remote server

Copy the client binary and skill to the remote server:

```bash
scp client/ccimg your-server:~/.local/bin/
scp skill/paste-image.md your-server:~/.claude/commands/paste-image.md
```

To avoid permission prompts each time, add this to `~/.claude/settings.json`:

```json
{
  "permissions": {
    "allow": [
      "Bash(ccimg)"
    ]
  }
}
```

### SSH connection

Connect with a reverse tunnel:

```bash
ssh -R 9998:localhost:9998 your-server
```

Or add to `~/.ssh/config`:

```
Host your-server
    RemoteForward 9998 localhost:9998
```

## Usage

In Claude Code on the remote server, copy an image to the clipboard on your local machine and run:

```
/paste-image
```

## Comparison with sshimg.nvim

| Aspect | sshimg.nvim | claude-ssh-image-skill |
|---|---|---|
| **Client** | Neovim plugin (Lua) | Claude Code skill |
| **Daemon** | `imgd` (Port 9999) | `ccimgd` (Port 9998) |
| **Image transfer** | Daemon → scp → Server | Daemon → base64 in TCP response → Client saves |
| **Request needs** | `host` + `path` | Nothing (empty object) |
| **Coexistence** | Port 9999 | Port 9998 — both can run simultaneously |

## License

MIT
