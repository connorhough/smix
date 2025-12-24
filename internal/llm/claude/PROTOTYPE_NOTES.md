# Claude CLI Wrapper Prototype Findings

## Overview

This document captures observations from prototyping a wrapper around the `claude` CLI for non-interactive request-response scenarios.

## Test Results

All prototype tests passed successfully:
- `TestPrototypeCLIWrapper`: Basic CLI wrapper functionality
- `TestContextCancellation`: Context cancellation behavior
- `TestInvalidModelName`: Error handling for invalid models
- `TestOutputFormat`: Output structure analysis

## Key Findings

### 1. Context Cancellation

**Does CLI respect context cancellation?**

Yes, the CLI properly respects context cancellation via `exec.CommandContext`.

- When context deadline is exceeded, the process receives a kill signal
- Error returned: `signal: killed`
- Timeout test (100ms) successfully terminated a long-running prompt
- This confirms we can use context for request timeouts in production

### 2. Error Codes and Handling

**What error codes does it return?**

Invalid model names return:
- Exit code: `1`
- Error output format: `API Error: 404 {"type":"error","error":{"type":"not_found_error","message":"model: invalid-model-name-xyz"},"request_id":"req_..."}`

Key observations:
- The CLI returns structured JSON error messages in stderr/output
- Error format includes API error type, message, and request ID
- Exit code `1` for API-level errors (likely consistent across error types)
- Error messages are machine-parsable (JSON format)

### 3. Invalid Model Name Handling

**How does it handle invalid model names?**

- Returns exit code `1` with structured error message
- Error type: `not_found_error`
- Includes request ID for debugging
- Clear error message indicating the invalid model name

This means we can:
- Detect invalid model names programmatically
- Extract specific error details from JSON
- Provide user-friendly error messages in our wrapper

### 4. Output Format Observations

**Output structure:**

- Raw output: `"hello\n"` (6 bytes)
- Ends with newline character
- No leading whitespace
- No extraneous formatting or ANSI codes
- Clean, parsable text output

Implications:
- Output can be used directly in most contexts
- Trailing newline should be trimmed for string comparisons
- No need to strip ANSI escape codes or formatting
- Output is ready for further processing

## Performance Notes

- Basic prompt execution: ~4-7 seconds (varies by model and prompt)
- Context cancellation: Responds immediately to timeout signals
- Network latency affects response times (API-based)

## Recommendations for Production Implementation

1. **Error Handling:**
   - Parse JSON error responses for structured error details
   - Map exit codes to specific error types
   - Include request IDs in error logs for debugging

2. **Context Management:**
   - Always use context with reasonable timeouts
   - Default timeout: 30-60 seconds for typical requests
   - Allow timeout configuration per request

3. **Output Processing:**
   - Trim trailing newline from responses
   - Handle empty responses gracefully
   - No need for ANSI code stripping

4. **Model Validation:**
   - Consider validating model names before API calls
   - Provide clear error messages for unsupported models
   - Document supported model names

## Next Steps

This prototype validates that wrapping the Claude CLI is viable for non-interactive request-response scenarios. The findings will inform the final Claude Provider implementation in Phase 2 of the multi-provider LLM refactor.
