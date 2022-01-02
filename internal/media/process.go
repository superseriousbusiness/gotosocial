package media

import "context"

type mediaProcessingFunction func(ctx context.Context, data []byte, contentType string, accountID string)
