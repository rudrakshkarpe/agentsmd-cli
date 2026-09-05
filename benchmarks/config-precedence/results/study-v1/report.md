# config-precedence benchmark

Model: `gpt-5.6-luna (low reasoning)` · Trials per condition: 3 · Generated: 2026-09-05T11:04:10Z

| Condition | Trial | Passed | Tokens | Commands | Duration |
|---|---:|:---:|---:|---:|---:|
| baseline | 1 | true | 102024 | 5 | 30.0s |
| baseline | 2 | true | 79022 | 3 | 27.9s |
| baseline | 3 | true | 90724 | 5 | 26.0s |
| learned | 1 | true | 44132 | 3 | 20.0s |
| learned | 2 | true | 54274 | 2 | 23.1s |
| learned | 3 | true | 78169 | 5 | 24.9s |

**baseline:** 100% success, median 90724 tokens, median 5 commands, median 27.9s.

**learned:** 100% success, median 54274 tokens, median 3 commands, median 23.1s.
