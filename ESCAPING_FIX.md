# HTML Double-Escaping Fix

## Problem
Text with apostrophes and special characters was being double/multi-escaped, resulting in text like:
- `don&amp;amp;amp;amp;amp;#39;t` instead of `don't`
- `Re-runs don&amp;amp;amp;#39;t work` instead of `Re-runs don't work`

## Root Cause
1. **Backend** was HTML-escaping input before storing in database
2. **Frontend** was HTML-escaping again when rendering (including in form values)
3. **Each edit** would re-escape the already-escaped data, compounding the issue

## Solution

### Backend Changes (`api/internal/util/sanitize.go`)
- **Removed** `html.EscapeString()` from `SanitizeString()`
- Now only trims whitespace
- HTML escaping is handled on the frontend where it belongs (output sanitization, not input)

### Database Migration (`api/scripts/fix_escaped_data.sql`)
- Created migration script to unescape existing double-escaped data
- Recursively unescapes HTML entities until no more escaping is found
- Applies to: tickets, acceptance_criteria, ticket_updates, and projects tables

### Test Updates
- Updated tests to reflect that `SanitizeString()` no longer escapes HTML
- Added test case for preserving apostrophes

## Verification

Tested the following scenarios:
1. ✅ Create ticket with apostrophes: `don't` stays as `don't`
2. ✅ Create ticket with quotes: `"properly"` stays as `"properly"`
3. ✅ Create ticket with ampersands: `A & B` stays as `A & B`
4. ✅ Edit ticket multiple times: no compounding of escaping
5. ✅ Acceptance criteria with special characters work correctly

## Security Note

HTML escaping is still performed on the **frontend** when rendering content to prevent XSS attacks. The escaping just happens at the correct layer (output) rather than at input/storage.

This follows the principle: **Sanitize on output, not on input**.

## Files Changed

- `api/internal/util/sanitize.go` - Removed HTML escaping
- `api/internal/util/sanitize_test.go` - Updated tests
- `api/scripts/fix_escaped_data.sql` - Migration to fix existing data

## Running the Migration

If you have existing production data with double-escaped content, run:

```bash
docker exec -i flyhalf-db psql -U flyhalf -d flyhalf < api/scripts/fix_escaped_data.sql
```

Or connect to your production database and execute the SQL script.
