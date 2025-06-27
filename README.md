# Waveland

I’ve been annoyed with how much effort it takes to collaborate on coding real-time. Screensharing takes too much back and forth, and git workflows are slow.

Tools like VS Code Live Share or CoderPad nail real-time collab, but only if you live inside their IDE. If you prefer Vim, Neovim, JetBrains, or even Nano, you’re back to the same issues again.

So I built waveland, a tool that makes your collab sessions feel like editing on Google Docs.

## Quick Start

**Install:**
```bash
curl -sSf https://raw.githubusercontent.com/go-johnnyhe/waveland/main/install.sh | sh
```

**Start a session:**
```bash
waveland start .
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

Built with Go + WebSockets + Cloudflared. Open source.