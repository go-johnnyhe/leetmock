# Waveland

Share code instantly. Perfect for interviews, pair programming, debugging.

## Quick Start

**Install:**
```bash
curl -sSf https://raw.githubusercontent.com/go-johnnyhe/waveland/main/install.sh | sh
```

**Share a file:**
```bash
waveland start main.py
```

**Join a session:**
```bash
waveland join <session-url>
```

## How it works

1. Run `waveland start filename.py` 
2. Share the generated URL with your partner
3. They run `waveland join <your-url>`
4. Both see live file updates with `→` and `←` indicators

No setup, no accounts, no configuration. Just works.

## Features

- **Zero setup** - automatic secure tunneling 
- **Live sync** - see changes instantly
- **Vim/Neovim** - auto-reload on file changes
- **Cross-platform** - works everywhere
- **Private** - secure peer-to-peer sessions

## Use Cases

- **Mock interviews** - share code with interviewer
- **Pair programming** - collaborate in real-time  
- **Debug sessions** - troubleshoot together
- **Code reviews** - walk through changes live

---

Built with Go + WebSockets. Open source.