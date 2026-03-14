# Algorithm Decisions

This document tracks all customizations made on top of the base FSRS v4 algorithm.

## 1. Random queue order

**Date:** 2026-03-14
**Change:** Shuffle the review queue randomly at session start.
**Reason:** Cards were presented in due-date order, which for bulk imports meant alphabetical. Random order provides better learning variety.
**Location:** `internal/service/review_service.go` — `StartSession()`

## 2. Re-queue "Again" cards within session

**Date:** 2026-03-14
**Change:** When a card is rated "Again", it is re-inserted into the queue approximately 10 cards later (`againRequeueGap = 10`). The card is still scheduled by FSRS normally (persisted to DB), but the re-queue gives the learner a second attempt within the same session.
**Reason:** Without this, "Again" cards disappear until the next session. Immediate reinforcement within the same session improves short-term recall.
**Location:** `internal/service/review_service.go` — `Rate()`, `againRequeueGap` constant
**Note:** The re-queued card uses its updated SRS state (post-FSRS scheduling). When rated a second time, FSRS treats it as a subsequent review, not a repeat of the first. If fewer than 10 cards remain in the queue, the card is appended at the end.
