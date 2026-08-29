# Lesson: keep fail-fast ordering when moving validation into domain factories

Problem: refactoring validation into domain factories silently regressed
fail-fast ordering — an invalid document/type still paid a bcrypt hash.

Root cause: `NewUser` was moved to validate raw strings AFTER the handler
already called the expensive `hasher.Hash`; nothing enforced "all cheap checks
before expensive work" because the validation code simply relocated.

Solution: `NewUser` validates every field via the private `validateFields`
method (`newFullName`, `newDocument`, `newEmail`, `newPassword`, `parseType`)
BEFORE calling `hasher.Hash`. Invalid input never reaches the hasher.

How to avoid it again: when moving rules into the domain, re-check the
ordering of side-effectful/expensive operations at the call boundary — the
review agent (`.ai/context/code-review.md`, DDD-02) caught this one.
