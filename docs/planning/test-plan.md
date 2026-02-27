# Rosslib Test Plan

Current state: **zero tests** across both the Go API and Next.js webapp. This document catalogs every test we should write, organized by domain. Tests are tagged with priority (**P0** = must-have, **P1** = important, **P2** = nice-to-have) and layer (unit / integration / E2E).

---

## Table of Contents

1. [Test Infrastructure](#1-test-infrastructure)
2. [API Tests (Go)](#2-api-tests-go)
   - [Auth](#21-auth)
   - [Books](#22-books)
   - [User Books](#23-user-books)
   - [Users & Profiles](#24-users--profiles)
   - [Shelves & Collections](#25-shelves--collections)
   - [Tags & Labels](#26-tags--labels)
   - [Threads & Comments](#27-threads--comments)
   - [Book Links & Votes](#28-book-links--votes)
   - [Imports](#29-imports)
   - [Activity & Feed](#210-activity--feed)
   - [Notifications](#211-notifications)
   - [Genre Ratings](#212-genre-ratings)
   - [Feedback](#213-feedback)
   - [User Data Deletion](#214-user-data-deletion)
   - [Admin / Ghosts](#215-admin--ghosts)
   - [Helpers & Utilities](#216-helpers--utilities)
   - [Middleware](#217-middleware)
   - [OL Cache](#218-ol-cache)
3. [Webapp Tests (Next.js)](#3-webapp-tests-nextjs)
   - [Auth Utilities](#31-auth-utilities)
   - [Route Handlers](#32-route-handlers)
   - [Components — Forms](#33-components--forms)
   - [Components — Interactive](#34-components--interactive)
   - [Components — Display](#35-components--display)
   - [Pages](#36-pages)
4. [E2E Tests](#4-e2e-tests)
5. [CI Integration](#5-ci-integration)

---

## 1. Test Infrastructure

### API (Go)

**What's needed:**
- PocketBase test app helper — spin up an in-memory PocketBase instance with migrations applied, seed test data, tear down after each test
- HTTP test helper — create `httptest.Server` from the registered routes, provide methods for authenticated/unauthenticated requests
- Fixture helpers — create test users, books, shelves, tags, follows, etc.
- Mock for Open Library / Google Books external calls (use `httptest.Server` or interface injection)

**File structure:**
```
api/
  handlers/
    testutil_test.go     # shared test app, helpers, fixtures
    auth_test.go
    books_test.go
    userbooks_test.go
    users_test.go
    collections_test.go
    tags_test.go
    threads_test.go
    links_test.go
    imports_test.go
    activity_test.go
    notifications_test.go
    genreratings_test.go
    feedback_test.go
    userdata_test.go
    ghosts_test.go
    helpers_test.go
    middleware_test.go
    olcache_test.go
```

### Webapp (Next.js)

**What's needed:**
- Jest config (`jest.config.ts`) with TypeScript, jsdom environment, `@/` path alias
- React Testing Library (`@testing-library/react`, `@testing-library/jest-dom`, `@testing-library/user-event`)
- MSW (Mock Service Worker) for intercepting fetch calls to `/api/*`
- Add `"test"` and `"test:watch"` scripts to `package.json`
- Playwright for E2E tests (optional, separate config)

**File structure:**
```
webapp/
  jest.config.ts
  src/
    lib/__tests__/auth.test.ts
    components/__tests__/
      follow-button.test.tsx
      star-rating-input.test.tsx
      book-review-editor.test.tsx
      import-form.test.tsx
      shelf-picker.test.tsx
      ...
    app/api/__tests__/
      auth.test.ts
      books.test.ts
      ...
```

---

## 2. API Tests (Go)

### 2.1 Auth

**File: `auth_test.go`**

#### Login — P0, integration
- [ ] `TestLogin_ValidCredentials` — returns 200, token, user_id, username
- [ ] `TestLogin_InvalidPassword` — returns 400 with error message
- [ ] `TestLogin_NonexistentUser` — returns 400
- [ ] `TestLogin_EmptyEmail` — returns 400
- [ ] `TestLogin_EmptyPassword` — returns 400
- [ ] `TestLogin_CaseInsensitiveEmail` — email matching is case-insensitive

#### Register — P0, integration
- [ ] `TestRegister_ValidInput` — creates user, auto-creates Status tag key with 5 default values, returns token
- [ ] `TestRegister_DuplicateEmail` — returns 400
- [ ] `TestRegister_EmptyFields` — returns 400 for missing email/password/username
- [ ] `TestRegister_UsernameGeneration` — generates unique username from email prefix
- [ ] `TestRegister_StatusTagAutoCreation` — verify Status key + Want to Read / Currently Reading / Finished / DNF / Owned values exist

#### Google OAuth — P1, integration
- [ ] `TestGoogleAuth_ExistingGoogleId` — finds user by google_id, returns token
- [ ] `TestGoogleAuth_ExistingEmailNoGoogleId` — links google_id to existing user
- [ ] `TestGoogleAuth_NewUser` — creates new user with google_id

#### GetAccount — P0, integration
- [ ] `TestGetAccount_Authenticated` — returns account details including is_moderator
- [ ] `TestGetAccount_Unauthenticated` — returns 401

#### ChangePassword — P1, integration
- [ ] `TestChangePassword_Valid` — accepts old password, sets new password, can login with new
- [ ] `TestChangePassword_WrongOldPassword` — returns 400
- [ ] `TestChangePassword_EmptyNewPassword` — returns 400

#### Token Generation — P0, unit
- [ ] `TestIssueToken_ReturnsValidJWT` — token contains user id, is parseable
- [ ] `TestGenerateUsername_FromEmail` — strips domain, sanitizes, handles collisions

---

### 2.2 Books

**File: `books_test.go`**

#### SearchBooks — P0, integration
- [ ] `TestSearchBooks_LocalMatch` — finds books already in local DB
- [ ] `TestSearchBooks_EmptyQuery` — returns 400
- [ ] `TestSearchBooks_LimitCap` — results capped at 20
- [ ] `TestSearchBooks_ExternalFallback` — when local has no results, calls Open Library (mock external API)

#### LookupBook — P0, integration
- [ ] `TestLookupBook_ByISBN` — finds or creates book by ISBN
- [ ] `TestLookupBook_ByOLId` — finds or creates book by Open Library ID
- [ ] `TestLookupBook_NoParams` — returns 400 when neither isbn nor ol_id provided
- [ ] `TestLookupBook_NotFound` — returns 404 when external API has no match
- [ ] `TestLookupBook_GoogleBooksFallback` — falls back to Google Books when OL fails

#### ScanBook — P1, integration
- [ ] `TestScanBook_ValidISBN` — delegates to LookupBook, returns book

#### GetBookDetail — P0, integration
- [ ] `TestGetBookDetail_ExistingBook` — returns book with stats (avg rating, reads, want-to-read)
- [ ] `TestGetBookDetail_NonexistentBook` — returns 404
- [ ] `TestGetBookDetail_StatsAccuracy` — stats match actual user_books data

#### GetBookEditions — P1, integration
- [ ] `TestGetBookEditions_HasEditions` — returns edition list from OL (mock)
- [ ] `TestGetBookEditions_NoEditions` — returns empty list

#### GetBookStats — P0, integration
- [ ] `TestGetBookStats_WithRatings` — rating_sum, rating_count, review_count correct
- [ ] `TestGetBookStats_NoRatings` — all zeroes
- [ ] `TestGetBookStats_ReadsAndWantToRead` — counts users with Finished and Want to Read status

#### GetBookReviews — P1, integration
- [ ] `TestGetBookReviews_PublicBook` — returns reviews without auth
- [ ] `TestGetBookReviews_ViewerReviewFirst` — authenticated user's review appears first
- [ ] `TestGetBookReviews_IncludesFollowStatus` — shows whether viewer follows each reviewer
- [ ] `TestGetBookReviews_Pagination` — respects limit/offset

#### Author Routes — P1, integration
- [ ] `TestSearchAuthors_ValidQuery` — returns author results from OL
- [ ] `TestGetAuthorDetail_ValidKey` — returns author info from OL
- [ ] `TestFollowAuthor_Success` — creates author_follows record
- [ ] `TestFollowAuthor_AlreadyFollowing` — idempotent or returns error
- [ ] `TestUnfollowAuthor_Success` — deletes author_follows record
- [ ] `TestGetFollowedAuthors_ReturnsList` — returns up to 100

#### Book Follow — P1, integration
- [ ] `TestFollowBook_Success` — creates book_follows record
- [ ] `TestUnfollowBook_Success` — deletes book_follows record
- [ ] `TestGetFollowedBooks_ReturnsList` — returns followed books

---

### 2.3 User Books

**File: `userbooks_test.go`**

#### AddBook — P0, integration
- [ ] `TestAddBook_NewBook` — creates user_book, optionally sets status/rating/review
- [ ] `TestAddBook_AlreadyAdded` — returns error or updates existing
- [ ] `TestAddBook_WithRating` — sets rating on user_book
- [ ] `TestAddBook_WithReview` — sets review_text and spoiler flag
- [ ] `TestAddBook_WithStatus` — sets status tag (want-to-read, etc.)
- [ ] `TestAddBook_RecordsActivity` — activity record created in background

#### UpdateBook — P0, integration
- [ ] `TestUpdateBook_Rating` — updates rating field
- [ ] `TestUpdateBook_Review` — updates review_text
- [ ] `TestUpdateBook_Spoiler` — toggles spoiler flag
- [ ] `TestUpdateBook_Dates` — sets date_started and date_finished
- [ ] `TestUpdateBook_Progress` — sets progress (pages/percentage)
- [ ] `TestUpdateBook_SelectedEdition` — sets selected_edition_key and cover_url
- [ ] `TestUpdateBook_NotOwned` — returns 403/404 for other user's book
- [ ] `TestUpdateBook_RefreshesStats` — book_stats updated after rating change

#### DeleteBook — P0, integration
- [ ] `TestDeleteBook_Success` — deletes user_book
- [ ] `TestDeleteBook_CascadesTagValues` — book_tag_values cleaned up
- [ ] `TestDeleteBook_CascadesCollectionItems` — collection_items cleaned up
- [ ] `TestDeleteBook_NotOwned` — returns 403/404

#### GetBookStatus — P0, integration
- [ ] `TestGetBookStatus_HasStatus` — returns status slug, rating, review
- [ ] `TestGetBookStatus_NoStatus` — returns empty/null
- [ ] `TestGetBookStatus_NotInLibrary` — returns 404

#### SetBookStatus — P0, integration
- [ ] `TestSetBookStatus_WantToRead` — assigns want-to-read status
- [ ] `TestSetBookStatus_CurrentlyReading` — assigns currently-reading
- [ ] `TestSetBookStatus_Finished` — assigns finished
- [ ] `TestSetBookStatus_DNF` — assigns dnf
- [ ] `TestSetBookStatus_Owned` — assigns owned
- [ ] `TestSetBookStatus_MutualExclusivity` — setting new status removes old one (select_one behavior)
- [ ] `TestSetBookStatus_InvalidSlug` — returns 400

#### GetStatusMap — P1, integration
- [ ] `TestGetStatusMap_MultipleBooks` — returns {ol_id: status_slug} for all user books
- [ ] `TestGetStatusMap_Empty` — returns empty map for new user

#### GetUserBooks — P0, integration
- [ ] `TestGetUserBooks_AllStatuses` — returns books grouped by status
- [ ] `TestGetUserBooks_FilterByStatus` — returns only books with given status
- [ ] `TestGetUserBooks_Pagination` — respects limit parameter
- [ ] `TestGetUserBooks_LimitCap` — limit > 50 capped to 20
- [ ] `TestGetUserBooks_PrivateProfile` — returns 403 for non-followers viewing private user
- [ ] `TestGetUserBooks_OwnerCanAlwaysSee` — owner bypasses privacy

---

### 2.4 Users & Profiles

**File: `users_test.go`**

#### SearchUsers — P1, integration
- [ ] `TestSearchUsers_ByUsername` — finds user by partial username match
- [ ] `TestSearchUsers_ByDisplayName` — finds user by display_name
- [ ] `TestSearchUsers_Pagination` — paginated at 20 per page
- [ ] `TestSearchUsers_EmptyQuery` — returns all users paginated

#### GetProfile — P0, integration
- [ ] `TestGetProfile_PublicUser` — returns profile with stats (followers, following, books read, avg rating)
- [ ] `TestGetProfile_PrivateUser_AsStranger` — returns 403
- [ ] `TestGetProfile_PrivateUser_AsFollower` — returns profile
- [ ] `TestGetProfile_PrivateUser_AsSelf` — returns profile
- [ ] `TestGetProfile_StatsAccuracy` — follower/following/book counts match actual data

#### GetUserReviews — P1, integration
- [ ] `TestGetUserReviews_PublicUser` — returns reviews with book info
- [ ] `TestGetUserReviews_LimitCap` — max 50, default 20
- [ ] `TestGetUserReviews_PrivateUser` — respects privacy setting

#### UpdateProfile — P1, integration
- [ ] `TestUpdateProfile_DisplayName` — updates display_name
- [ ] `TestUpdateProfile_Bio` — updates bio
- [ ] `TestUpdateProfile_PrivacyToggle` — toggles is_private flag

#### UploadAvatar — P2, integration
- [ ] `TestUploadAvatar_ValidImage` — accepts image upload
- [ ] `TestUploadAvatar_NoFile` — returns 400

#### Follow/Unfollow — P0, integration
- [ ] `TestFollowUser_PublicUser` — creates follow with status=active
- [ ] `TestFollowUser_PrivateUser` — creates follow with status=pending
- [ ] `TestFollowUser_AlreadyFollowing` — idempotent or returns error
- [ ] `TestFollowUser_Self` — cannot follow yourself
- [ ] `TestUnfollowUser_Success` — deletes follow record
- [ ] `TestUnfollowUser_NotFollowing` — returns 404

#### Follow Requests — P0, integration
- [ ] `TestGetFollowRequests_HasPending` — returns pending follow records
- [ ] `TestGetFollowRequests_NoPending` — returns empty list
- [ ] `TestAcceptFollowRequest_Success` — changes status from pending to active
- [ ] `TestAcceptFollowRequest_NotPending` — returns 404
- [ ] `TestRejectFollowRequest_Success` — deletes follow record

---

### 2.5 Shelves & Collections

**File: `collections_test.go`**

#### GetMyShelves — P0, integration
- [ ] `TestGetMyShelves_HasShelves` — returns shelves with item counts
- [ ] `TestGetMyShelves_Empty` — returns empty list
- [ ] `TestGetMyShelves_IncludesComputedLists` — includes computed list metadata

#### CreateShelf — P0, integration
- [ ] `TestCreateShelf_Valid` — creates shelf with name and auto-generated slug
- [ ] `TestCreateShelf_DuplicateName` — handles duplicate slug
- [ ] `TestCreateShelf_EmptyName` — returns 400
- [ ] `TestCreateShelf_PublicFlag` — respects is_public setting

#### UpdateShelf — P1, integration
- [ ] `TestUpdateShelf_Name` — updates name and slug
- [ ] `TestUpdateShelf_PublicToggle` — toggles is_public
- [ ] `TestUpdateShelf_NotOwned` — returns 403

#### DeleteShelf — P1, integration
- [ ] `TestDeleteShelf_Success` — deletes shelf
- [ ] `TestDeleteShelf_CascadesItems` — collection_items deleted
- [ ] `TestDeleteShelf_NotOwned` — returns 403

#### GetUserShelves — P0, integration
- [ ] `TestGetUserShelves_PublicUser` — returns public shelves
- [ ] `TestGetUserShelves_PrivateUser` — respects privacy
- [ ] `TestGetUserShelves_IncludeBooks` — optionally includes shelf contents

#### GetShelfDetail — P0, integration
- [ ] `TestGetShelfDetail_PublicShelf` — returns shelf with books
- [ ] `TestGetShelfDetail_PrivateShelf_AsStranger` — returns 403
- [ ] `TestGetShelfDetail_PrivateShelf_AsOwner` — returns shelf

#### AddBookToShelf — P0, integration
- [ ] `TestAddBookToShelf_Success` — creates collection_item
- [ ] `TestAddBookToShelf_AlreadyOnShelf` — returns error
- [ ] `TestAddBookToShelf_NotOwnedShelf` — returns 403

#### UpdateShelfBook — P1, integration
- [ ] `TestUpdateShelfBook_Rating` — updates item rating
- [ ] `TestUpdateShelfBook_Review` — updates item review

#### RemoveBookFromShelf — P1, integration
- [ ] `TestRemoveBookFromShelf_Success` — deletes collection_item
- [ ] `TestRemoveBookFromShelf_NotOnShelf` — returns 404

---

### 2.6 Tags & Labels

**File: `tags_test.go`**

#### GetTagKeys — P0, integration
- [ ] `TestGetTagKeys_HasKeys` — returns tag keys with values
- [ ] `TestGetTagKeys_AutoCreatesStatus` — ensures Status key exists even if missing
- [ ] `TestGetTagKeys_NewUser` — auto-creates Status with 5 default values

#### CreateTagKey — P1, integration
- [ ] `TestCreateTagKey_SelectOne` — creates select_one mode key
- [ ] `TestCreateTagKey_SelectMany` — creates select_many mode key
- [ ] `TestCreateTagKey_DuplicateSlug` — handles collision
- [ ] `TestCreateTagKey_EmptyName` — returns 400

#### DeleteTagKey — P1, integration
- [ ] `TestDeleteTagKey_CustomKey` — deletes key and cascades values
- [ ] `TestDeleteTagKey_StatusKey` — prevented, returns 400
- [ ] `TestDeleteTagKey_NotOwned` — returns 403

#### CreateTagValue — P1, integration
- [ ] `TestCreateTagValue_Valid` — creates value under key
- [ ] `TestCreateTagValue_DuplicateSlug` — handles collision

#### DeleteTagValue — P1, integration
- [ ] `TestDeleteTagValue_Success` — deletes value
- [ ] `TestDeleteTagValue_CascadesBookTags` — removes book_tag_values referencing this value

#### GetBookTags — P0, integration
- [ ] `TestGetBookTags_HasTags` — returns book_tag_values for user's book
- [ ] `TestGetBookTags_NoTags` — returns empty

#### SetBookTag — P0, integration
- [ ] `TestSetBookTag_SelectOne` — replaces existing value (single select)
- [ ] `TestSetBookTag_SelectMany` — adds value without removing existing
- [ ] `TestSetBookTag_NotOwnedBook` — returns 403

#### UnsetBookTag — P1, integration
- [ ] `TestUnsetBookTag_Success` — removes tag from book
- [ ] `TestUnsetBookTagValue_Specific` — removes specific value, keeps others

#### GetUserTagKeys (public route) — P1, integration
- [ ] `TestGetUserTagKeys_PublicUser` — returns tag keys
- [ ] `TestGetUserTagKeys_PrivateUser` — respects privacy

#### GetUserTagBooks — P1, integration
- [ ] `TestGetUserTagBooks_ByTagPath` — returns books matching tag path
- [ ] `TestGetUserTagBooks_EmptyResult` — returns empty when no matches

---

### 2.7 Threads & Comments

**File: `threads_test.go`**

#### GetBookThreads — P1, integration
- [ ] `TestGetBookThreads_HasThreads` — returns threads with comment counts
- [ ] `TestGetBookThreads_NoThreads` — returns empty list
- [ ] `TestGetBookThreads_ExcludesDeleted` — soft-deleted threads not shown

#### CreateThread — P1, integration
- [ ] `TestCreateThread_Valid` — creates thread with title and body
- [ ] `TestCreateThread_EmptyTitle` — returns 400
- [ ] `TestCreateThread_EmptyBody` — returns 400
- [ ] `TestCreateThread_RequiresAuth` — returns 401 unauthenticated
- [ ] `TestCreateThread_RecordsActivity` — activity recorded

#### DeleteThread — P1, integration
- [ ] `TestDeleteThread_AsOwner` — soft deletes (sets deleted_at)
- [ ] `TestDeleteThread_NotOwner` — returns 403
- [ ] `TestDeleteThread_AsModerator` — moderator can delete any thread

#### AddComment — P1, integration
- [ ] `TestAddComment_Valid` — creates comment on thread
- [ ] `TestAddComment_EmptyBody` — returns 400
- [ ] `TestAddComment_OnDeletedThread` — returns 404

#### DeleteComment — P1, integration
- [ ] `TestDeleteComment_AsOwner` — soft deletes comment
- [ ] `TestDeleteComment_NotOwner` — returns 403

---

### 2.8 Book Links & Votes

**File: `links_test.go`**

#### GetBookLinks — P1, integration
- [ ] `TestGetBookLinks_HasLinks` — returns links with vote counts
- [ ] `TestGetBookLinks_ViewerVoteIncluded` — shows authenticated viewer's vote status
- [ ] `TestGetBookLinks_NoLinks` — returns empty

#### CreateBookLink — P1, integration
- [ ] `TestCreateBookLink_Valid` — creates link between two books
- [ ] `TestCreateBookLink_MissingTarget` — returns 400 without to_open_library_id
- [ ] `TestCreateBookLink_MissingType` — returns 400 without link_type
- [ ] `TestCreateBookLink_RequiresAuth` — returns 401

#### DeleteBookLink — P1, integration
- [ ] `TestDeleteBookLink_AsOwner` — deletes link
- [ ] `TestDeleteBookLink_NotOwner` — returns 403

#### VoteLink — P2, integration
- [ ] `TestVoteLink_Success` — creates vote record
- [ ] `TestVoteLink_AlreadyVoted` — prevents duplicate votes
- [ ] `TestUnvoteLink_Success` — removes vote

#### ProposeLinkEdit — P2, integration
- [ ] `TestProposeLinkEdit_Valid` — creates edit proposal
- [ ] `TestProposeLinkEdit_RequiresAuth` — returns 401

---

### 2.9 Imports

**File: `imports_test.go`**

#### PreviewGoodreadsImport — P0, integration
- [ ] `TestPreviewImport_ValidCSV` — streams NDJSON with book matching results
- [ ] `TestPreviewImport_EmptyFile` — returns error
- [ ] `TestPreviewImport_MalformedCSV` — handles gracefully
- [ ] `TestPreviewImport_TitleCleaning` — strips series info, subtitles, format tags
- [ ] `TestPreviewImport_AuthorCleaning` — strips author prefixes
- [ ] `TestPreviewImport_ISBNLookup` — matches by ISBN first
- [ ] `TestPreviewImport_TitleFallback` — falls back to title search when ISBN fails
- [ ] `TestPreviewImport_GoogleBooksFallback` — uses Google Books when OL fails
- [ ] `TestPreviewImport_ConcurrentWorkers` — 5 concurrent workers process rows

#### CommitGoodreadsImport — P0, integration
- [ ] `TestCommitImport_CreatesUserBooks` — bulk creates user_books from preview data
- [ ] `TestCommitImport_SetsRatings` — preserves Goodreads ratings
- [ ] `TestCommitImport_SetsStatus` — maps Goodreads shelves to status tags
- [ ] `TestCommitImport_SkipsDuplicates` — doesn't re-add books already in library
- [ ] `TestCommitImport_CreatesPendingImports` — unmatched books become pending_imports

#### Title/Author Cleaning — P0, unit
- [ ] `TestCleanGoodreadsTitle_StripsSeries` — removes "(Series Name, #1)"
- [ ] `TestCleanGoodreadsTitle_StripsSubtitle` — removes ": Subtitle"
- [ ] `TestCleanGoodreadsTitle_StripsFormat` — removes "[Kindle Edition]" etc.
- [ ] `TestCleanGoodreadsAuthor_StripsPrefix` — removes "by " prefix
- [ ] `TestTitleMatchesResult_ExactMatch` — exact title match returns true
- [ ] `TestTitleMatchesResult_CloseMatch` — near match returns true
- [ ] `TestTitleMatchesResult_FalsePositive` — unrelated title returns false
- [ ] `TestMapGoodreadsShelf_KnownShelves` — maps "read"→"finished", "to-read"→"want-to-read", "currently-reading"→"currently-reading"

#### Pending Imports — P1, integration
- [ ] `TestGetPendingImports_HasPending` — returns unresolved imports
- [ ] `TestResolvePendingImport_MatchToOL` — matches pending import to OL book ID
- [ ] `TestResolvePendingImport_Dismiss` — dismisses without matching
- [ ] `TestDeletePendingImport_Success` — deletes pending import record

---

### 2.10 Activity & Feed

**File: `activity_test.go`**

#### GetFeed — P1, integration
- [ ] `TestGetFeed_HasFollowedActivity` — shows activity from followed users
- [ ] `TestGetFeed_ExcludesUnfollowed` — doesn't show unfollowed users' activity
- [ ] `TestGetFeed_CursorPagination` — cursor-based pagination works
- [ ] `TestGetFeed_PageSize` — 30 items per page
- [ ] `TestGetFeed_EnrichesActivity` — includes book/user/thread details

#### GetUserActivity — P1, integration
- [ ] `TestGetUserActivity_ReturnsRecent` — returns last 30 activities
- [ ] `TestGetUserActivity_PrivateUser` — respects privacy

#### RecordActivity — P1, unit
- [ ] `TestRecordActivity_Shelved` — creates activity for shelving a book
- [ ] `TestRecordActivity_CreatedThread` — creates activity for thread creation
- [ ] `TestRecordActivity_FollowedAuthor` — creates activity for author follow

---

### 2.11 Notifications

**File: `notifications_test.go`**

#### GetNotifications — P1, integration
- [ ] `TestGetNotifications_HasNotifications` — returns newest first, up to 50
- [ ] `TestGetNotifications_Empty` — returns empty list

#### GetUnreadCount — P1, integration
- [ ] `TestGetUnreadCount_HasUnread` — returns correct count
- [ ] `TestGetUnreadCount_AllRead` — returns 0

#### MarkNotificationRead — P1, integration
- [ ] `TestMarkRead_Single` — marks one notification as read
- [ ] `TestMarkRead_NotOwned` — cannot mark others' notifications

#### MarkAllRead — P1, integration
- [ ] `TestMarkAllRead_Success` — marks all as read

---

### 2.12 Genre Ratings

**File: `genreratings_test.go`**

#### GetBookGenreRatings — P2, integration
- [ ] `TestGetBookGenreRatings_HasRatings` — returns aggregated genre ratings
- [ ] `TestGetBookGenreRatings_NoRatings` — returns empty

#### GetMyGenreRatings — P2, integration
- [ ] `TestGetMyGenreRatings_HasRatings` — returns user's genre ratings for book
- [ ] `TestGetMyGenreRatings_NoRatings` — returns empty

#### SetGenreRatings — P2, integration
- [ ] `TestSetGenreRatings_NewRatings` — creates genre rating records
- [ ] `TestSetGenreRatings_UpdateExisting` — upserts existing ratings
- [ ] `TestSetGenreRatings_RequiresAuth` — returns 401

---

### 2.13 Feedback

**File: `feedback_test.go`**

#### CreateFeedback — P2, integration
- [ ] `TestCreateFeedback_Bug` — creates bug report
- [ ] `TestCreateFeedback_Feature` — creates feature request
- [ ] `TestCreateFeedback_InvalidType` — returns 400 for non bug/feature
- [ ] `TestCreateFeedback_RequiresAuth` — returns 401

#### GetFeedback (admin) — P2, integration
- [ ] `TestGetFeedback_AsModerator` — returns feedback list
- [ ] `TestGetFeedback_FilterByStatus` — filters by open/closed
- [ ] `TestGetFeedback_AsNonModerator` — returns 403

#### UpdateFeedbackStatus — P2, integration
- [ ] `TestUpdateFeedbackStatus_OpenToClosed` — updates status
- [ ] `TestUpdateFeedbackStatus_InvalidStatus` — returns 400

---

### 2.14 User Data Deletion

**File: `userdata_test.go`**

#### DeleteAllData — P0, integration
- [ ] `TestDeleteAllData_RemovesUserBooks` — all user_books deleted
- [ ] `TestDeleteAllData_RemovesCollections` — all collections and items deleted
- [ ] `TestDeleteAllData_RemovesTags` — all tag keys/values/book_tag_values deleted
- [ ] `TestDeleteAllData_RemovesFollows` — all follow records deleted
- [ ] `TestDeleteAllData_RemovesActivity` — all activity records deleted
- [ ] `TestDeleteAllData_RemovesNotifications` — all notifications deleted
- [ ] `TestDeleteAllData_KeepsAccount` — user auth record preserved
- [ ] `TestDeleteAllData_BatchProcessing` — handles > 200 records (batch delete)
- [ ] `TestDeleteAllData_RefreshesBookStats` — affected book stats recalculated

---

### 2.15 Admin / Ghosts

**File: `ghosts_test.go`**

#### SeedGhosts — P2, integration
- [ ] `TestSeedGhosts_Creates10Users` — creates 10 ghost user personas
- [ ] `TestSeedGhosts_MarksAsGhost` — sets is_ghost flag
- [ ] `TestSeedGhosts_RequiresModerator` — returns 403 for non-moderators

#### SimulateGhosts — P2, integration
- [ ] `TestSimulateGhosts_AddsBooks` — ghost users add books from OL subjects
- [ ] `TestSimulateGhosts_Rates` — ghosts rate books
- [ ] `TestSimulateGhosts_Reviews` — ghosts leave reviews
- [ ] `TestSimulateGhosts_Follows` — ghosts follow each other

#### Admin Link Edits — P2, integration
- [ ] `TestGetPendingLinkEdits_AsModerator` — returns pending edits
- [ ] `TestReviewLinkEdit_Approve` — applies edit to link
- [ ] `TestReviewLinkEdit_Reject` — dismisses edit
- [ ] `TestReviewLinkEdit_AsNonModerator` — returns 403

---

### 2.16 Helpers & Utilities

**File: `helpers_test.go`**

#### Slugification — P0, unit
- [ ] `TestSlugify_BasicString` — "Hello World" → "hello-world"
- [ ] `TestSlugify_Unicode` — strips non-ASCII characters
- [ ] `TestSlugify_SpecialChars` — replaces &, @, etc. with dashes
- [ ] `TestSlugify_ConsecutiveDashes` — collapses multiple dashes
- [ ] `TestSlugify_LeadingTrailingDashes` — trims leading/trailing dashes
- [ ] `TestSlugify_EmptyString` — returns empty string

#### Tag Slugification — P0, unit
- [ ] `TestTagSlugify_PreservesSlash` — "fiction/fantasy" → "fiction/fantasy"
- [ ] `TestTagSlugify_HierarchicalTags` — "Science Fiction/Space Opera" → "science-fiction/space-opera"

#### Title Case — P1, unit
- [ ] `TestTitleCase_FromSlug` — "dark-fantasy" → "Dark Fantasy"
- [ ] `TestTitleCase_SingleWord` — "fiction" → "Fiction"

#### Privacy Check — P0, unit
- [ ] `TestCanViewProfile_PublicUser` — always returns true
- [ ] `TestCanViewProfile_PrivateUser_NoViewer` — returns false
- [ ] `TestCanViewProfile_PrivateUser_ActiveFollower` — returns true
- [ ] `TestCanViewProfile_PrivateUser_PendingFollower` — returns false
- [ ] `TestCanViewProfile_PrivateUser_Self` — returns true

#### UpsertBook — P0, integration
- [ ] `TestUpsertBook_NewBook` — creates book record with all fields
- [ ] `TestUpsertBook_ExistingBook` — finds by OL ID, updates empty fields only
- [ ] `TestUpsertBook_DoesNotOverwriteExisting` — populated fields not overwritten

#### RefreshBookStats — P1, integration
- [ ] `TestRefreshBookStats_CalculatesCorrectly` — rating_sum, rating_count, review_count, reads, want_to_read all accurate
- [ ] `TestRefreshBookStats_AfterRatingChange` — recalculates after user changes rating
- [ ] `TestRefreshBookStats_AfterBookDelete` — recalculates after user removes book

#### EnsureStatusTagKey — P0, integration
- [ ] `TestEnsureStatusTagKey_CreatesIfMissing` — creates Status key with 5 values
- [ ] `TestEnsureStatusTagKey_Idempotent` — does nothing if already exists

---

### 2.17 Middleware

**File: `middleware_test.go`**

#### OptionalAuth — P0, unit/integration
- [ ] `TestOptionalAuth_WithValidToken` — sets e.Auth to the user
- [ ] `TestOptionalAuth_WithInvalidToken` — proceeds without auth (no error)
- [ ] `TestOptionalAuth_WithoutToken` — proceeds without auth
- [ ] `TestOptionalAuth_WithExpiredToken` — proceeds without auth

#### RequireModerator — P0, unit/integration
- [ ] `TestRequireModerator_IsModerator` — allows request
- [ ] `TestRequireModerator_NotModerator` — returns 403
- [ ] `TestRequireModerator_Unauthenticated` — returns 401

---

### 2.18 OL Cache

**File: `olcache_test.go`**

#### Cache Operations — P1, unit
- [ ] `TestOLCache_StoresAndRetrieves` — cached items retrievable
- [ ] `TestOLCache_TTLExpiry` — items expire after 24 hours
- [ ] `TestOLCache_Eviction` — background eviction clears expired items
- [ ] `TestOLCache_Singleton` — newOLClient returns same instance
- [ ] `TestOLCache_Stats` — tracks hit/miss counts

---

## 3. Webapp Tests (Next.js)

### 3.1 Auth Utilities

**File: `lib/__tests__/auth.test.ts`**

#### getToken — P0, unit
- [ ] `test_getToken_returnsCookieValue` — reads 'token' cookie
- [ ] `test_getToken_noCookie` — returns null when no cookie

#### getUser — P0, unit
- [ ] `test_getUser_validToken` — decodes JWT, returns AuthUser with user_id, username, is_moderator
- [ ] `test_getUser_expiredToken` — returns null for expired JWT
- [ ] `test_getUser_malformedToken` — returns null for invalid base64
- [ ] `test_getUser_missingUsername` — falls back to username cookie
- [ ] `test_getUser_noCookies` — returns null
- [ ] `test_getUser_invalidJSON` — returns null when JWT payload isn't valid JSON

---

### 3.2 Route Handlers

**File: `app/api/__tests__/*.test.ts`**

These test that route handlers correctly proxy to the Go API and handle cookies.

#### Auth Route Handlers — P0, integration
- [ ] `test_loginHandler_proxiesToAPI` — forwards credentials, sets token+username cookies on success
- [ ] `test_loginHandler_forwardsError` — returns API error on failure
- [ ] `test_registerHandler_proxiesToAPI` — forwards registration, sets cookies
- [ ] `test_logoutHandler_clearsCookies` — clears token+username cookies, redirects to home
- [ ] `test_googleAuthHandler_redirectsToProvider` — initiates OAuth flow

#### Book Route Handlers — P1, integration
- [ ] `test_searchHandler_callsOpenLibrary` — directly calls OL (not Go API)
- [ ] `test_lookupHandler_proxiesWithAuth` — forwards token header
- [ ] `test_bookDetailHandler_proxies` — forwards workId parameter

#### Generic Proxy Pattern — P0, integration
- [ ] `test_proxyHandler_forwardsAuthHeader` — Bearer token from cookie sent to API
- [ ] `test_proxyHandler_noToken` — request sent without auth header when no cookie
- [ ] `test_proxyHandler_forwardsQueryParams` — query string passed through
- [ ] `test_proxyHandler_forwardsBody` — POST/PATCH body forwarded
- [ ] `test_proxyHandler_returnsStatusCode` — API status code preserved

---

### 3.3 Components — Forms

**File: `components/__tests__/*.test.tsx`**

#### BookReviewEditor — P0, component
- [ ] `test_reviewEditor_renders` — renders with initial values (rating, review text)
- [ ] `test_reviewEditor_starClick` — clicking star updates rating state
- [ ] `test_reviewEditor_textInput` — typing updates review text
- [ ] `test_reviewEditor_spoilerToggle` — toggling spoiler checkbox works
- [ ] `test_reviewEditor_saveSubmits` — save button calls API with correct payload
- [ ] `test_reviewEditor_cancelReverts` — cancel restores original values
- [ ] `test_reviewEditor_clearReview` — clear button removes rating + review
- [ ] `test_reviewEditor_autocompleteBookLink` — typing `[[` triggers book search
- [ ] `test_reviewEditor_dateInputs` — date started/finished inputs work
- [ ] `test_reviewEditor_disabledWhileSaving` — buttons disabled during API call
- [ ] `test_reviewEditor_showsError` — displays API error message

#### ImportForm — P0, component
- [ ] `test_importForm_fileUpload` — accepts CSV file
- [ ] `test_importForm_previewPhase` — shows streaming progress as NDJSON arrives
- [ ] `test_importForm_reviewPhase` — displays matched/unmatched books for review
- [ ] `test_importForm_configPhase` — shelf mapping configuration
- [ ] `test_importForm_commitPhase` — triggers import commit
- [ ] `test_importForm_errorDisplay` — shows upload/parse errors
- [ ] `test_importForm_pendingImports` — shows unmatched books needing resolution

#### SettingsForm — P1, component
- [ ] `test_settingsForm_displaysCurrentValues` — pre-fills display name, bio
- [ ] `test_settingsForm_updateProfile` — submits updated fields
- [ ] `test_settingsForm_avatarUpload` — file input triggers upload
- [ ] `test_settingsForm_privacyToggle` — toggles is_private
- [ ] `test_settingsForm_successMessage` — shows save confirmation
- [ ] `test_settingsForm_errorMessage` — shows API error

#### PasswordForm — P1, component
- [ ] `test_passwordForm_submit` — sends old + new password
- [ ] `test_passwordForm_mismatch` — shows error for non-matching passwords
- [ ] `test_passwordForm_emptyFields` — validation prevents empty submit

#### FeedbackForm — P2, component
- [ ] `test_feedbackForm_submitsBug` — sends bug type feedback
- [ ] `test_feedbackForm_submitsFeature` — sends feature type feedback
- [ ] `test_feedbackForm_successReset` — clears form after success

#### DeleteDataForm — P1, component
- [ ] `test_deleteDataForm_requiresConfirmation` — must type confirmation text
- [ ] `test_deleteDataForm_disabledWithoutConfirmation` — button disabled until text matches
- [ ] `test_deleteDataForm_submits` — calls delete endpoint

---

### 3.4 Components — Interactive

#### FollowButton — P0, component
- [ ] `test_followButton_notFollowing` — shows "Follow" text
- [ ] `test_followButton_pending` — shows "Requested" or pending state
- [ ] `test_followButton_following` — shows "Following" state
- [ ] `test_followButton_clickToFollow` — calls follow API, updates UI optimistically
- [ ] `test_followButton_clickToUnfollow` — calls unfollow API, updates UI
- [ ] `test_followButton_apiFailureReverts` — reverts optimistic update on error

#### StarRatingInput — P0, component
- [ ] `test_starRating_renders5Stars` — renders 5 star elements
- [ ] `test_starRating_clickSetsRating` — clicking 3rd star sets rating to 3
- [ ] `test_starRating_clickSameClears` — clicking current rating clears it
- [ ] `test_starRating_hoverHighlights` — hovering shows preview state
- [ ] `test_starRating_callsOnChange` — fires callback with new value

#### ShelfPicker — P1, component
- [ ] `test_shelfPicker_displaysOptions` — shows available shelves
- [ ] `test_shelfPicker_selectShelf` — selecting shelf calls API
- [ ] `test_shelfPicker_currentShelfHighlighted` — current shelf visually distinguished

#### BookTagPicker — P1, component
- [ ] `test_tagPicker_showsTagKeys` — displays available tag keys
- [ ] `test_tagPicker_selectValue` — selecting value calls set tag API
- [ ] `test_tagPicker_deselectValue` — deselecting calls unset tag API

#### EditionPicker — P2, component
- [ ] `test_editionPicker_showsEditions` — displays available editions
- [ ] `test_editionPicker_selectEdition` — selecting updates selected_edition

#### ReadingProgress — P2, component
- [ ] `test_readingProgress_inputPages` — accepts page number
- [ ] `test_readingProgress_inputPercentage` — accepts percentage
- [ ] `test_readingProgress_submitsUpdate` — calls API with progress data

#### NotificationBell — P1, component
- [ ] `test_notificationBell_showsBadge` — shows unread count badge
- [ ] `test_notificationBell_noBadge` — no badge when count is 0
- [ ] `test_notificationBell_clickOpensDropdown` — click opens notification list

#### QuickAddButton — P2, component
- [ ] `test_quickAdd_searchesBooks` — typing triggers book search
- [ ] `test_quickAdd_selectAddsBook` — selecting result adds book to library

#### BookFollowButton — P2, component
- [ ] `test_bookFollowButton_follow` — follows book on click
- [ ] `test_bookFollowButton_unfollow` — unfollows on second click

#### AuthorFollowButton — P2, component
- [ ] `test_authorFollowButton_follow` — follows author on click
- [ ] `test_authorFollowButton_unfollow` — unfollows on second click

---

### 3.5 Components — Display

#### BookList — P1, component
- [ ] `test_bookList_rendersBooks` — renders book covers, titles, authors
- [ ] `test_bookList_emptyState` — renders empty message
- [ ] `test_bookList_showsRatings` — displays rating stars
- [ ] `test_bookList_showsReadCount` — displays number of reads

#### ThreadList — P2, component
- [ ] `test_threadList_rendersThreads` — renders thread titles and comment counts
- [ ] `test_threadList_emptyState` — renders empty message

#### ThreadComments — P2, component
- [ ] `test_threadComments_rendersComments` — shows comment text and author
- [ ] `test_threadComments_deletedComment` — shows "[deleted]" for soft-deleted

#### ReviewText — P1, component
- [ ] `test_reviewText_rendersMarkdown` — renders markdown properly
- [ ] `test_reviewText_bookLinks` — renders `[[book]]` links as clickable
- [ ] `test_reviewText_spoilerHidden` — spoiler reviews hidden until clicked

#### Activity — P2, component
- [ ] `test_activity_rendersItems` — renders activity feed items
- [ ] `test_activity_differentTypes` — correctly renders shelved/reviewed/followed types

#### ReadingStats — P2, component
- [ ] `test_readingStats_showsMetrics` — displays books read, pages, avg rating

#### Nav — P1, component
- [ ] `test_nav_loggedIn` — shows user menu, notifications, search
- [ ] `test_nav_loggedOut` — shows login/register links

---

### 3.6 Pages

#### Server Component Pages — P1, integration
- [ ] `test_homePage_redirectsWhenLoggedIn` — redirects to /feed when authenticated
- [ ] `test_homePage_showsLanding` — shows landing page when not authenticated
- [ ] `test_bookDetailPage_fetchesData` — fetches book detail + stats + reviews
- [ ] `test_profilePage_fetchesUserData` — fetches user profile + stats
- [ ] `test_profilePage_handlesPrivateUser` — shows appropriate message for private profiles
- [ ] `test_feedPage_requiresAuth` — redirects to login when not authenticated
- [ ] `test_searchPage_handlesQueryParams` — processes search query from URL

---

## 4. E2E Tests

Using Playwright, test full user journeys through the browser.

**File structure:**
```
webapp/
  e2e/
    auth.spec.ts
    books.spec.ts
    import.spec.ts
    social.spec.ts
    shelves.spec.ts
```

#### Auth Flows — P0, E2E
- [ ] `test_e2e_register_and_login` — register new user → logout → login with credentials
- [ ] `test_e2e_login_invalid_credentials` — shows error message on bad login

#### Book Operations — P0, E2E
- [ ] `test_e2e_search_and_add_book` — search → view detail → add to library → appears in library
- [ ] `test_e2e_rate_and_review` — add book → rate 4 stars → write review → appears on book page
- [ ] `test_e2e_change_status` — add book → set Want to Read → change to Currently Reading → change to Finished
- [ ] `test_e2e_delete_book` — add book → remove from library → no longer in library

#### Shelf Management — P1, E2E
- [ ] `test_e2e_create_shelf_add_books` — create shelf → add books → view shelf → books present
- [ ] `test_e2e_delete_shelf` — create shelf → delete → no longer listed

#### Social — P1, E2E
- [ ] `test_e2e_follow_public_user` — visit profile → follow → appears in following list
- [ ] `test_e2e_follow_private_user` — follow request → accept → now following
- [ ] `test_e2e_feed_shows_followed_activity` — follow user → they add book → appears in feed

#### Import — P1, E2E
- [ ] `test_e2e_goodreads_import` — upload CSV → preview → configure shelves → import → books in library

#### Tags — P2, E2E
- [ ] `test_e2e_create_tag_and_assign` — create tag → assign to book → filter by tag → book appears

#### Threads — P2, E2E
- [ ] `test_e2e_create_thread_and_comment` — open book → create thread → add comment → all visible

---

## 5. CI Integration

### GitHub Actions Workflow

```yaml
# .github/workflows/test.yml
name: Tests
on: [push, pull_request]

jobs:
  api-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.23'
      - run: cd api && go test ./... -v -race -coverprofile=coverage.out
      - run: cd api && go tool cover -func=coverage.out

  webapp-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: '20'
      - run: cd webapp && npm ci
      - run: cd webapp && npm test -- --coverage

  webapp-lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: '20'
      - run: cd webapp && npm ci
      - run: cd webapp && npm run lint
      - run: cd webapp && npx tsc --noEmit

  e2e-tests:
    runs-on: ubuntu-latest
    needs: [api-tests, webapp-tests]
    steps:
      - uses: actions/checkout@v4
      - run: docker compose up -d
      - uses: actions/setup-node@v4
        with:
          node-version: '20'
      - run: cd webapp && npx playwright install --with-deps
      - run: cd webapp && npx playwright test
```

---

## Summary

| Layer | Area | P0 | P1 | P2 | Total |
|-------|------|----|----|----|----|
| API unit | helpers, slugify, cache, middleware | 16 | 9 | 0 | 25 |
| API integration | auth, books, userbooks, users, shelves, tags, threads, links, imports, activity, notifications, genre ratings, feedback, userdata, ghosts, admin | 60 | 65 | 28 | 153 |
| Webapp unit | auth utilities | 6 | 0 | 0 | 6 |
| Webapp component | forms, interactive, display | 15 | 24 | 16 | 55 |
| Webapp integration | route handlers, pages | 7 | 9 | 0 | 16 |
| E2E | full user journeys | 6 | 7 | 3 | 16 |
| **Total** | | **110** | **114** | **47** | **271** |

**Recommended implementation order:**
1. Test infrastructure (both API + webapp) — get the harness working first
2. API P0 unit tests (helpers, middleware) — quick wins, pure functions
3. API P0 integration tests (auth, userbooks, shelves, imports, privacy) — core functionality
4. Webapp P0 unit + component tests (auth utilities, review editor, follow button, star rating)
5. API P1 integration tests — remaining endpoints
6. Webapp P1 component + integration tests
7. E2E P0 tests — critical user journeys
8. CI pipeline
9. P2 tests — fill remaining gaps
