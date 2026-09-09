# Let AGENTS.md Write Itself

Animated web presentation for AGNTCon + MCPCon Japan 2026. This copy lives
with the implementation so the product demo and talk claims can be reviewed
together.

## Run locally

From the repository root:

```bash
python3 -m http.server 8080 --directory presentation
```

Open <http://localhost:8080>.

## Controls

- Next reveal, then slide: `Arrow Right`, `Space`, `Enter`, or `Page Down`
- Previous reveal, then slide: `Arrow Left`, `Backspace`, or `Page Up`
- First or last slide: `Home` or `End`
- Slide overview: `Esc`
- Speaker notes: `N`
- Sources and scope: `S`
- Keyboard help: `?`
- Dark or light theme: `T`
- Play, pause, replay, or scrub progressive beats with the visible controls
- Replay the active CLI animation or reset the active diagram: `R`
- Full screen: `F`

The URL hash identifies the active slide, so `#8` opens slide 8 directly.
Click the left third of the stage to move back and the rest to move forward.
Slide 3 is a presenter-controlled AGENTS.md introduction. Use the visible
play, replay, and scrub controls, or advance it one beat at a time.

## Session

**Let AGENTS.md Write Itself**

Presented by Rudraksh Karpe and Satyam Soni.

- September 10, 2026, 10:40–11:05 JST
- Hall C
- 17 slides
- 21 minutes of planned material, including the controlled AGENTS.md
  introduction, two recorded CLI sequences, and
  four minutes for questions

See [TALK-NOTES.md](TALK-NOTES.md) for the stage plan and
[SOURCES.md](SOURCES.md) for source provenance. [CLAIM-LEDGER.md](CLAIM-LEDGER.md)
separates confirmed behavior, dated snapshots, community opinion, and
illustrative motion.

## Source

The original material was copied from
[`satyampsoni/agentsmd-agntcon-japan`](https://github.com/satyampsoni/agentsmd-agntcon-japan)
at commit `544295b9768b65059359cf6e7a762f43a4cad82f`, then updated with the
implemented agentsmd workflow.

The refreshed layout and interaction model take structural inspiration from
[`abhishekpanditofficial/agentcon-japan`](https://github.com/abhishekpanditofficial/agentcon-japan).
This includes the fixed 16:9 stage, staggered reveals, overview, notes, help,
and hash navigation.
The replayable terminal follows the approach used by
[`rohitg00/agentmemory`](https://github.com/rohitg00/agentmemory/tree/main/website),
with scripts rewritten around real agentsmd commands and evidence.
The decision-workflow layout is informed by
[`get-vix/vix`](https://github.com/get-vix/vix), while the nodes describe only
the capture, reflection, evaluation, and promotion behavior implemented here.
The light-first system anatomy and presenter-controlled motion adapt the visual
grammar of
[`Inside an Inference Request`](https://rudrakshkarpe.com/presentations/inside-an-inference-request)
through the
[`animated-technical-illustrations`](https://github.com/shivaylamba/animated-technical-illustrations)
skill.

## Rebuild the CLI recordings

The checked-in GIFs are generated from
[`scripts/render_cli_gifs.mjs`](scripts/render_cli_gifs.mjs). The script uses
Playwright and FFmpeg so the terminal media remains reproducible rather than a
hand-edited animation.

## Validate the interactive deck

With the local server running, use the bundled workspace Node.js packages:

```bash
WORKSPACE_NODE_MODULES=/path/to/bundled/node_modules \
  node presentation/scripts/validate_deck.mjs
```

The check covers all 17 slides at 320px, 768px, and desktop widths, along with
deep links, keyboard stepping, replay, image loading, source visibility, and
the reduced-motion final state.
