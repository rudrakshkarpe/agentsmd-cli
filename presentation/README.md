# Let AGENTS.md Write Itself

Animated web presentation for AGNTCon + MCPCon Japan 2026. This copy lives
with the implementation so the product demo, benchmark evidence, and talk
claims can be reviewed together.

## Run locally

From the repository root:

```bash
python3 -m http.server 8080 --directory presentation
```

Open <http://localhost:8080>.

## Controls

- Next slide: `Arrow Right`, `Space`, or `Page Down`
- Previous slide: `Arrow Left` or `Page Up`
- First or last slide: `Home` or `End`
- Full screen: `F`

## Session

**Let AGENTS.md Write Itself**

Presented by Rudraksh Karpe and Satyam Soni.

- September 10, 2026, 10:40–11:05 JST
- Hall C
- 14 slides
- 22.5 minutes of planned material, including two live terminal sequences and
  four minutes for questions

See [TALK-NOTES.md](TALK-NOTES.md) for the stage plan and
[SOURCES.md](SOURCES.md) for claim provenance.

## Source

The original material was copied from
[`satyampsoni/agentsmd-agntcon-japan`](https://github.com/satyampsoni/agentsmd-agntcon-japan)
at commit `544295b9768b65059359cf6e7a762f43a4cad82f`, then updated with the
implemented agentsmd workflow and the checked-in `study-v1` result.

The refreshed layout takes structural inspiration from
[`abhishekpanditofficial/agentcon-japan`](https://github.com/abhishekpanditofficial/agentcon-japan).
The replayable terminal follows the approach used by
[`rohitg00/agentmemory`](https://github.com/rohitg00/agentmemory/tree/main/website),
with scripts rewritten around real agentsmd commands and evidence.
