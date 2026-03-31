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
2. `ccimgd` reads the PNG image from the local clipboard (`wl-paste` on Wayland, `xclip` on X11, `pngpaste` on macOS)
3. The image is returned as base64-encoded JSON
4. `ccimg` saves it as a temporary PNG file and prints the path
5. The skill uses Claude's `Read` tool to display the image

## Requirements

- **Local machine (Linux)**: `wl-paste` (Wayland, part of `wl-clipboard`) or `xclip` (X11)
- **Local machine (macOS)**: `pngpaste` (`brew install pngpaste`)
- **Remote server**: Claude Code
- **Building from source**: Go

## Building

Build both binaries (daemon + client):

```bash
./build.sh
```

This builds statically linked binaries for all supported platforms:

- `daemon/ccimgd-linux-amd64`, `daemon/ccimgd-darwin-amd64`, `daemon/ccimgd-darwin-arm64`
- `client/ccimg-linux-amd64`, `client/ccimg-darwin-amd64`, `client/ccimg-darwin-arm64`

## Setup

### Local machine (Linux)

```bash
cp daemon/ccimgd-linux-amd64 ~/.local/bin/ccimgd
cp daemon/ccimgd.service ~/.config/systemd/user/
systemctl --user daemon-reload
systemctl --user enable --now ccimgd
```

### Local machine (macOS)

```bash
cp daemon/ccimgd-darwin-arm64 /usr/local/bin/ccimgd   # or ccimgd-darwin-amd64 for Intel
cp daemon/com.ccimgd.plist ~/Library/LaunchAgents/
launchctl load ~/Library/LaunchAgents/com.ccimgd.plist
```

### Remote server

Copy the client binary and skill to the remote server:

```bash
scp client/ccimg-linux-amd64 your-server:~/.local/bin/ccimg
scp skill/paste-image.md your-server:~/.claude/commands/paste-image.md
scp ~/.ccimgd-token your-server:~/.ccimgd-token
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

## Security

### Authentication

The daemon and client authenticate using a shared secret token stored in `~/.ccimgd-token` (generated automatically on first run). The token file is created with `0600` permissions (owner-only).

To copy the token to a remote server:

```bash
scp ~/.ccimgd-token your-server:~/.ccimgd-token
```

To override the token file location, set the `CCIMGD_TOKEN_FILE` environment variable on both daemon and client.

To rotate the token:

```bash
rm ~/.ccimgd-token        # delete on local machine
scp ~/.ccimgd-token your-server:~/.ccimgd-token  # re-copy after daemon regenerates
```

### Connection security

- The daemon only listens on `127.0.0.1:9998` (localhost)
- The SSH reverse tunnel provides transport encryption
- Connections have a 10-second read timeout and 4096-byte buffer limit
- All clipboard access is logged to the daemon's stderr with connection source and token hash

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
