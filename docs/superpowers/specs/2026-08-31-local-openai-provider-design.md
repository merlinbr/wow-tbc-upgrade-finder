# Local OpenAI-Compatible Provider Design

## Goal

Make the local model served at `http://127.0.0.1:5000/v1/chat/completions` selectable in `omp` without changing the existing `llama.cpp` provider.

## Decision

Add a separate `local-ai` provider to the user's `~/.omp/agent/models.yml`:

- `baseUrl: http://127.0.0.1:5000/v1`
- `api: openai-completions`, matching the server's Chat Completions endpoint
- one explicit model entry using the server's exact model ID
- `auth: none` for an unauthenticated server, or an environment-backed `apiKey` with `authHeader: true` when the server requires Bearer authentication

The existing `llama.cpp` entry remains unchanged because it points at a different server and uses the Responses API.

## Data flow

`omp` resolves `local-ai/<model-id>` from `models.yml`, sends requests to the configured `/v1/chat/completions` endpoint, and supplies no credential or a Bearer credential according to the selected auth mode. The model can be selected explicitly with `--model local-ai/<model-id>` or assigned through `modelRoles.default`.

## Error handling

The server's current `/v1/models` probe returns HTTP 401. Configuration must therefore use the server's actual model ID and authentication requirements. A missing or invalid key should remain a visible provider error; no hard-coded secret belongs in the config file.

## Verification

After configuration, run `omp models find <model-id>` and launch one real request with `omp --model local-ai/<model-id>`. Confirm the request reaches the local server and produces a response. Do not alter unrelated providers or project files.
