package cleaner_test

import (
	"context"
	"errors"
	"time"

	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtscontext"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/util"
)

func copyMap(in map[string]*gtsmodel.Emoji) map[string]*gtsmodel.Emoji {
	out := make(map[string]*gtsmodel.Emoji, len(in))

	for k, v1 := range in {
		v2 := new(gtsmodel.Emoji)
		*v2 = *v1
		out[k] = v2
	}

	return out
}

func (suite *CleanerTestSuite) TestEmojiUncacheRemote() {
	suite.testEmojiUncacheRemote(
		context.Background(),
		mapvals(suite.emojis),
	)
}

func (suite *CleanerTestSuite) TestEmojiUncacheRemoteDryRun() {
	suite.testEmojiUncacheRemote(
		gtscontext.SetDryRun(context.Background()),
		mapvals(suite.emojis),
	)
}

func (suite *CleanerTestSuite) TestEmojiFixBroken() {
	suite.testEmojiFixBroken(
		context.Background(),
		mapvals(suite.emojis),
	)
}

func (suite *CleanerTestSuite) TestEmojiFixBrokenDryRun() {
	suite.testEmojiFixBroken(
		gtscontext.SetDryRun(context.Background()),
		mapvals(suite.emojis),
	)
}

func (suite *CleanerTestSuite) TestEmojiPruneUnused() {
	suite.testEmojiPruneUnused(
		context.Background(),
		mapvals(suite.emojis),
	)
}

func (suite *CleanerTestSuite) TestEmojiPruneUnusedDryRun() {
	suite.testEmojiPruneUnused(
		gtscontext.SetDryRun(context.Background()),
		mapvals(suite.emojis),
	)
}

func (suite *CleanerTestSuite) TestEmojiFixCacheStates() {
	// Copy testrig emojis + mark
	// rainbow emoji as uncached
	// so there's something to fix.
	emojis := copyMap(suite.emojis)
	emojis["rainbow"].Cached = util.Ptr(false)

	suite.testEmojiFixCacheStates(
		context.Background(),
		mapvals(emojis),
	)
}

func (suite *CleanerTestSuite) TestEmojiFixCacheStatesDryRun() {
	// Copy testrig emojis + mark
	// rainbow emoji as uncached
	// so there's something to fix.
	emojis := copyMap(suite.emojis)
	emojis["rainbow"].Cached = util.Ptr(false)

	suite.testEmojiFixCacheStates(
		gtscontext.SetDryRun(context.Background()),
		mapvals(emojis),
	)
}

func (suite *CleanerTestSuite) testEmojiUncacheRemote(ctx context.Context, emojis []*gtsmodel.Emoji) {
	var uncacheIDs []string

	// Test state.
	t := suite.T()

	// Get max remote cache days to keep.
	days := config.GetMediaRemoteCacheDays()
	olderThan := time.Now().Add(-24 * time.Hour * time.Duration(days))

	for _, emoji := range emojis {
		// Check whether this emoji should be uncached.
		ok, err := suite.shouldUncacheEmoji(ctx, emoji, olderThan)
		if err != nil {
			t.Fatalf("error checking whether emoji should be uncached: %v", err)
		}

		if ok {
			// Mark this emoji ID as to be uncached.
			uncacheIDs = append(uncacheIDs, emoji.ID)
		}
	}

	// Attempt to uncache remote emojis.
	found, err := suite.cleaner.Emoji().UncacheRemote(ctx, olderThan)
	if err != nil {
		t.Errorf("error uncaching remote emojis: %v", err)
		return
	}

	// Check expected were uncached.
	if found != len(uncacheIDs) {
		t.Errorf("expected %d emojis to be uncached, %d were", len(uncacheIDs), found)
		return
	}

	if gtscontext.DryRun(ctx) {
		// nothing else to test.
		return
	}

	for _, id := range uncacheIDs {
		// Fetch the emoji by ID that should now be uncached.
		emoji, err := suite.state.DB.GetEmojiByID(ctx, id)
		if err != nil {
			t.Fatalf("error fetching emoji from database: %v", err)
		}

		// Check cache state.
		if *emoji.Cached {
			t.Errorf("emoji %s@%s should have been uncached", emoji.Shortcode, emoji.Domain)
		}

		// Check that the emoji files in storage have been deleted.
		if ok, err := suite.state.Storage.Has(ctx, emoji.ImagePath); err != nil {
			t.Fatalf("error checking storage for emoji: %v", err)
		} else if ok {
			t.Errorf("emoji %s@%s image path should not exist", emoji.Shortcode, emoji.Domain)
		} else if ok, err := suite.state.Storage.Has(ctx, emoji.ImageStaticPath); err != nil {
			t.Fatalf("error checking storage for emoji: %v", err)
		} else if ok {
			t.Errorf("emoji %s@%s image static path should not exist", emoji.Shortcode, emoji.Domain)
		}
	}
}

func (suite *CleanerTestSuite) shouldUncacheEmoji(ctx context.Context, emoji *gtsmodel.Emoji, after time.Time) (bool, error) {
	if emoji.ImageRemoteURL == "" {
		// Local emojis are never uncached.
		return false, nil
	}

	if emoji.Cached == nil || !*emoji.Cached {
		// Emoji is already uncached.
		return false, nil
	}

	// Get related accounts using this emoji (if any).
	accounts, err := suite.state.DB.GetAccountsUsingEmoji(ctx, emoji.ID)
	if err != nil {
		return false, err
	}

	// Check if accounts are recently updated.
	for _, account := range accounts {
		if account.FetchedAt.After(after) {
			return false, nil
		}
	}

	// Get related statuses using this emoji (if any).
	statuses, err := suite.state.DB.GetStatusesUsingEmoji(ctx, emoji.ID)
	if err != nil {
		return false, err
	}

	// Check if statuses are recently updated.
	for _, status := range statuses {
		if status.FetchedAt.After(after) {
			return false, nil
		}
	}

	return true, nil
}

func (suite *CleanerTestSuite) testEmojiFixBroken(ctx context.Context, emojis []*gtsmodel.Emoji) {
	var fixIDs []string

	// Test state.
	t := suite.T()

	for _, emoji := range emojis {
		// Check whether this emoji should be fixed.
		ok, err := suite.shouldFixBrokenEmoji(ctx, emoji)
		if err != nil {
			t.Fatalf("error checking whether emoji should be fixed: %v", err)
		}

		if ok {
			// Mark this emoji ID as to be fixed.
			fixIDs = append(fixIDs, emoji.ID)
		}
	}

	// Attempt to fix broken emojis.
	found, err := suite.cleaner.Emoji().FixBroken(ctx)
	if err != nil {
		t.Errorf("error fixing broken emojis: %v", err)
		return
	}

	// Check expected were fixed.
	if found != len(fixIDs) {
		t.Errorf("expected %d emojis to be fixed, %d were", len(fixIDs), found)
		return
	}

	if gtscontext.DryRun(ctx) {
		// nothing else to test.
		return
	}

	for _, id := range fixIDs {
		// Fetch the emoji by ID that should now be fixed.
		emoji, err := suite.state.DB.GetEmojiByID(ctx, id)
		if err != nil {
			t.Fatalf("error fetching emoji from database: %v", err)
		}

		// Ensure category was cleared.
		if emoji.CategoryID != "" {
			t.Errorf("emoji %s@%s should have empty category", emoji.Shortcode, emoji.Domain)
		}
	}
}

func (suite *CleanerTestSuite) shouldFixBrokenEmoji(ctx context.Context, emoji *gtsmodel.Emoji) (bool, error) {
	if emoji.CategoryID == "" {
		// no category issue.
		return false, nil
	}

	// Get the related category for this emoji.
	category, err := suite.state.DB.GetEmojiCategory(ctx, emoji.CategoryID)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		return false, nil
	}

	return (category == nil), nil
}

func (suite *CleanerTestSuite) testEmojiPruneUnused(ctx context.Context, emojis []*gtsmodel.Emoji) {
	var pruneIDs []string

	// Test state.
	t := suite.T()

	for _, emoji := range emojis {
		// Check whether this emoji should be pruned.
		ok, err := suite.shouldPruneEmoji(ctx, emoji)
		if err != nil {
			t.Fatalf("error checking whether emoji should be pruned: %v", err)
		}

		if ok {
			// Mark this emoji ID as to be pruned.
			pruneIDs = append(pruneIDs, emoji.ID)
		}
	}

	// Attempt to prune emojis.
	found, err := suite.cleaner.Emoji().PruneUnused(ctx)
	if err != nil {
		t.Errorf("error fixing broken emojis: %v", err)
		return
	}

	// Check expected were pruned.
	if found != len(pruneIDs) {
		t.Errorf("expected %d emojis to be pruned, %d were", len(pruneIDs), found)
		return
	}

	if gtscontext.DryRun(ctx) {
		// nothing else to test.
		return
	}

	for _, id := range pruneIDs {
		// Fetch the emoji by ID that should now be pruned.
		emoji, err := suite.state.DB.GetEmojiByID(ctx, id)
		if err != nil && !errors.Is(err, db.ErrNoEntries) {
			t.Fatalf("error fetching emoji from database: %v", err)
		}

		// Ensure gone.
		if emoji != nil {
			t.Errorf("emoji %s@%s should have been pruned", emoji.Shortcode, emoji.Domain)
		}
	}
}

func (suite *CleanerTestSuite) shouldPruneEmoji(ctx context.Context, emoji *gtsmodel.Emoji) (bool, error) {
	if emoji.ImageRemoteURL == "" {
		// Local emojis are never pruned.
		return false, nil
	}

	// Get related accounts using this emoji (if any).
	accounts, err := suite.state.DB.GetAccountsUsingEmoji(ctx, emoji.ID)
	if err != nil {
		return false, err
	} else if len(accounts) > 0 {
		return false, nil
	}

	// Get related statuses using this emoji (if any).
	statuses, err := suite.state.DB.GetStatusesUsingEmoji(ctx, emoji.ID)
	if err != nil {
		return false, err
	} else if len(statuses) > 0 {
		return false, nil
	}

	return true, nil
}

func (suite *CleanerTestSuite) testEmojiFixCacheStates(ctx context.Context, emojis []*gtsmodel.Emoji) {
	var fixIDs []string

	// Test state.
	t := suite.T()

	for _, emoji := range emojis {
		// Check whether this emoji should be fixed.
		ok, err := suite.shouldFixEmojiCacheState(ctx, emoji)
		if err != nil {
			t.Fatalf("error checking whether emoji should be fixed: %v", err)
		}

		if ok {
			// Mark this emoji ID as to be fixed.
			fixIDs = append(fixIDs, emoji.ID)
		}
	}

	// Attempt to fix broken emoji cache states.
	found, err := suite.cleaner.Emoji().FixCacheStates(ctx)
	if err != nil {
		t.Errorf("error fixing broken emojis: %v", err)
		return
	}

	// Check expected were fixed.
	if found != len(fixIDs) {
		t.Errorf("expected %d emojis to be fixed, %d were", len(fixIDs), found)
		return
	}

	if gtscontext.DryRun(ctx) {
		// nothing else to test.
		return
	}

	for _, id := range fixIDs {
		// Fetch the emoji by ID that should now be fixed.
		emoji, err := suite.state.DB.GetEmojiByID(ctx, id)
		if err != nil {
			t.Fatalf("error fetching emoji from database: %v", err)
		}

		// Ensure emoji cache state has been fixed.
		ok, err := suite.shouldFixEmojiCacheState(ctx, emoji)
		if err != nil {
			t.Fatalf("error checking whether emoji should be fixed: %v", err)
		} else if ok {
			t.Errorf("emoji %s@%s cache state should have been fixed", emoji.Shortcode, emoji.Domain)
		}
	}
}

func (suite *CleanerTestSuite) shouldFixEmojiCacheState(ctx context.Context, emoji *gtsmodel.Emoji) (bool, error) {
	// Check whether emoji image path exists.
	haveImage, err := suite.state.Storage.Has(ctx, emoji.ImagePath)
	if err != nil {
		return false, err
	}

	// Check whether emoji static path exists.
	haveStatic, err := suite.state.Storage.Has(ctx, emoji.ImageStaticPath)
	if err != nil {
		return false, err
	}

	switch exists := (haveImage && haveStatic); {
	case emoji.Cached != nil &&
		*emoji.Cached && !exists:
		// (cached can be nil in tests)
		// Cached but missing files.
		return true, nil

	case emoji.Cached != nil &&
		!*emoji.Cached && exists:
		// (cached can be nil in tests)
		// Uncached but unexpected files.
		return true, nil

	default:
		// No cache state issue.
		return false, nil
	}
}
