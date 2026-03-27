Paste an image from the local clipboard into this session.

Instructions:
1. Run `ccimg` using the Bash tool. It connects to the local ccimgd daemon via TCP on 127.0.0.1:9998 (available through SSH reverse tunnel), receives the clipboard image, saves it as a temporary PNG file, and prints the file path.
2. Use the Read tool to read the printed file path. This will display the image to Claude.

```bash
ccimg
```
